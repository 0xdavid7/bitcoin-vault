package tests

import (
	"encoding/hex"
	"fmt"
	"log"
	"testing"

	"github.com/scalarorg/bitcoin-vault/ffi/go-psbt"
)

const PSBT_HEX = "70736274ff0100520200000001e0a68346c9118f584c22c9afa89b641e06127d1b1fa661788ea922261dee37600000000000fdffffff012823000000000000160014acd07b22adf2299c56909c9ca537fd2c58127ecc000000000001012b102700000000000022512054bfa5690019d09073d75d1094d6eb9a551a5d61b0fcfc1fd474da6bfea88627010304000000004215c150929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac007e94635a4727997d13497f6529f00a9ca291c2e6e10253eb995eecd130a9eeb4520f02e0d96250daf3ed999f12a2a7c3c198e7d26f6bef5add3ef764831004d256fad20992b50ef84354a4c0b5831bc90b36b5da98f7fc8969df5f4c88f5ec270b0dfbbacc02116992b50ef84354a4c0b5831bc90b36b5da98f7fc8969df5f4c88f5ec270b0dfbb25019e450b1a6179e18dd5ab6aeff0e5172728cb84fc236261768579eb5252cd574a000000002116f02e0d96250daf3ed999f12a2a7c3c198e7d26f6bef5add3ef764831004d256f25019e450b1a6179e18dd5ab6aeff0e5172728cb84fc236261768579eb5252cd574a0000000001172050929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac0011820867e83e93516ecde27680f5af69af0bd633f9918874b975c7e65c0b2419047ee0000"

const EXPECTED_HEX = "70736274ff0100520200000001e0a68346c9118f584c22c9afa89b641e06127d1b1fa661788ea922261dee37600000000000fdffffff012823000000000000160014acd07b22adf2299c56909c9ca537fd2c58127ecc000000000001012b102700000000000022512054bfa5690019d09073d75d1094d6eb9a551a5d61b0fcfc1fd474da6bfea88627010304000000004114f02e0d96250daf3ed999f12a2a7c3c198e7d26f6bef5add3ef764831004d256f9e450b1a6179e18dd5ab6aeff0e5172728cb84fc236261768579eb5252cd574a40b21c79a3f1196e8d8d309eff56b4ca2f39cb2957c0a540f66aed88d1ca33bdcaea2434cc02c71c30bb2ceaa629dcdf2fd2b6a5efef019cd07bde292edeb2230d4215c150929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac007e94635a4727997d13497f6529f00a9ca291c2e6e10253eb995eecd130a9eeb4520f02e0d96250daf3ed999f12a2a7c3c198e7d26f6bef5add3ef764831004d256fad20992b50ef84354a4c0b5831bc90b36b5da98f7fc8969df5f4c88f5ec270b0dfbbacc02116992b50ef84354a4c0b5831bc90b36b5da98f7fc8969df5f4c88f5ec270b0dfbb25019e450b1a6179e18dd5ab6aeff0e5172728cb84fc236261768579eb5252cd574a000000002116f02e0d96250daf3ed999f12a2a7c3c198e7d26f6bef5add3ef764831004d256f25019e450b1a6179e18dd5ab6aeff0e5172728cb84fc236261768579eb5252cd574a0000000001172050929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac0011820867e83e93516ecde27680f5af69af0bd633f9918874b975c7e65c0b2419047ee0000"

const PRIVKEY_HEX = "7ddd6c59e93689262760f9258bb205e92d353d1d6a97c7d9d986c247fcffce1e"

func TestSignPsbtBySingleKey(t *testing.T) {
	psbtBytes, err := hex.DecodeString(PSBT_HEX)
	if err != nil {
		log.Fatal(err)
	}

	privkeyBytes, err := hex.DecodeString(PRIVKEY_HEX)
	if err != nil {
		log.Fatal(err)
	}

	signedPsbt, err := psbt.SignPsbtBySingleKey(
		psbtBytes,               // []byte containing PSBT
		privkeyBytes,            // []byte containing private key
		psbt.NetworkKindTestnet, // TestNet
		false,                   // finalize
	)
	if err != nil {
		log.Fatal(err)
	}

	if hex.EncodeToString(signedPsbt) != EXPECTED_HEX {
		t.Fatal("Signed PSBT does not match expected value")
	}

	fmt.Println("Signed PSBT: ", hex.EncodeToString(signedPsbt))
}
