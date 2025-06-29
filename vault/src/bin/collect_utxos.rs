use bitcoin::consensus::serialize;
use bitcoin::hex::DisplayHex;
use bitcoin::{Address, Amount, OutPoint, TxOut};
use bitcoin::{Network, Psbt};
use electrum_client::{Client, ElectrumApi};
use rust_mempool::MempoolClient;
use std::process;
use std::sync::Arc;
use vault::core::*;
use vault::utils::*;

// export TEST_ENV=testnet4 && cargo run --package vault --bin collect_utxos

#[tokio::main(flavor = "multi_thread", worker_threads = 8)]
async fn main() {
    const VOUT: u32 = 1;
    const LIMIT: usize = 500;
    const ELECTRUM_PORT: u16 = 60001;
    const ELECTRUM_HOST: &str = "127.0.0.1";

    // Create the client once and wrap in Arc
    let client = Arc::new(
        Client::new(format!("tcp://{}:{}", ELECTRUM_HOST, ELECTRUM_PORT).as_str()).unwrap(),
    );

    let mempool_client = Arc::new(MempoolClient::new(Network::Testnet4));

    // Load environment
    let test_suite = TestSuite::new_with_loaded_env("PEPE");

    let test_account = SuiteAccount::new(AccountEnv::new(test_suite.env_path()).unwrap());

    let custodian_quorum = test_suite.env().custodian_quorum;

    // Build script and address
    let script = match <VaultManager as CustodianOnly>::locking_script(
        &test_suite.custodian_pubkeys(),
        custodian_quorum,
    ) {
        Ok(s) => s,
        Err(e) => {
            eprintln!("Failed to build locking script: {}", e);
            process::exit(1);
        }
    };
    let network = get_network_from_str(&test_suite.env().network);
    let address = match Address::from_script(&script.clone().into_script(), network) {
        Ok(addr) => addr,
        Err(e) => {
            eprintln!("Failed to create address from script: {}", e);
            process::exit(1);
        }
    };
    println!("address: {}", address);
    println!("collect address: {}", test_account.address());

    // Connect to Electrum

    println!("Connected to Electrum");

    // Fetch UTXOs
    let mut utxos = match client.script_list_unspent(&script.into_script()) {
        Ok(u) => u,
        Err(e) => {
            eprintln!("Failed to fetch UTXOs: {}", e);
            process::exit(1);
        }
    };

    // let mut utxos = utxos
    //     .iter()
    //     .filter(|utxo| utxo.height > 85000)
    //     .collect::<Vec<_>>();
    utxos.reverse();

    let utxos = utxos
        .into_iter()
        .map(|utxo| NeededUtxo {
            txid: utxo.tx_hash,
            vout: utxo.tx_pos as u32,
            amount: Amount::from_sat(utxo.value),
        })
        .collect::<Vec<_>>();

    let len = utxos.len();

    // not get last 500 utxos
    let utxos = utxos.into_iter().take(len - 500).collect::<Vec<_>>();

    println!("number of utxos: {}", utxos.len());

    // Batch and process
    let batch_futures = utxos.chunks(LIMIT).enumerate().map(|(i, utxos_chunk)| {
        let address = address.clone();
        let account_address = test_account.address().clone();
        let manager = test_suite.manager().clone();
        let custodian_pubkeys = test_suite.custodian_pubkeys().clone();
        let custodian_quorum = test_suite.env().custodian_quorum;
        let network_id = test_suite.network_id();
        let signing_privkeys = test_suite.custodian_privkeys().clone();
        let utxos_chunk: Vec<_> = utxos_chunk.to_vec();
        let mempool_client = mempool_client.clone();
        tokio::spawn(async move {
            let total: u64 = utxos_chunk.iter().map(|utxo| utxo.amount.to_sat()).sum();
            println!("Batch {}: Processing {} utxos", i + 1, utxos_chunk.len());
            let mut unstaked_psbt = match <VaultManager as CustodianOnly>::build_unlocking_psbt(
                &manager,
                &CustodianOnlyUnlockingParams {
                    inputs: utxos_chunk
                        .iter()
                        .map(|u| PreviousOutpoint {
                            outpoint: OutPoint::new(u.txid, VOUT),
                            amount_in_sats: u.amount,
                            script_pubkey: address.script_pubkey(),
                        })
                        .collect(),
                    outputs: vec![TxOut {
                        value: Amount::from_sat(total),
                        script_pubkey: account_address.script_pubkey(),
                    }],
                    custodian_pubkeys: custodian_pubkeys.clone(),
                    custodian_quorum,
                    fee_rate: 2,
                    rbf: false,
                    session_sequence: 0,
                    custodian_group_uid: [0u8; HASH_SIZE],
                },
            ) {
                Ok(psbt) => psbt,
                Err(e) => {
                    eprintln!("Failed to build PSBT for batch {}: {}", i + 1, e);
                    return;
                }
            };
            for privkey in signing_privkeys {
                let _ = <VaultManager as Signing>::sign_psbt_by_single_key(
                    &mut unstaked_psbt,
                    privkey.as_slice(),
                    network_id,
                    false,
                )
                .unwrap();
            }
            <Psbt as SignByKeyMap<bitcoin::secp256k1::All>>::finalize(&mut unstaked_psbt);
            let finalized_tx = match unstaked_psbt.extract_tx() {
                Ok(tx) => tx,
                Err(e) => {
                    eprintln!("Failed to extract tx for batch {}: {}", i + 1, e);
                    return;
                }
            };

            let tx_hex = serialize(&finalized_tx);

            let result = mempool_client
                .broadcast_transaction(tx_hex.to_lower_hex_string().as_str())
                .await;

            match result {
                Ok(tx_id) => {
                    println!("Batch {} tx_id: {:?}", i + 1, tx_id);
                }
                Err(e) => {
                    eprintln!("Broadcast error for batch {}: {:?}", i + 1, e);
                }
            }
        })
    });
    futures::future::join_all(batch_futures).await;
}
