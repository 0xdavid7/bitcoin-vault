package chain

import "fmt"

const ChainInfoBytesSize = 8

const SEPARATOR = "|"

type ChainType uint8

const (
	ChainTypeBitcoin ChainType = iota // 0x00
	ChainTypeEVM                      // 0x01
	ChainTypeSolana                   // 0x02
	ChainTypeCosmos                   // 0x03
)

const (
	ChainTypeBitcoinStr = "bitcoin"
	ChainTypeEVMStr     = "evm"
	ChainTypeSolanaStr  = "solana"
	ChainTypeCosmosStr  = "cosmos"
)

var ChainTypeString = map[ChainType]string{
	ChainTypeBitcoin: ChainTypeBitcoinStr,
	ChainTypeEVM:     ChainTypeEVMStr,
	ChainTypeSolana:  ChainTypeSolanaStr,
	ChainTypeCosmos:  ChainTypeCosmosStr,
}

var ChainTypeFromString = map[string]ChainType{
	ChainTypeBitcoinStr: ChainTypeBitcoin,
	ChainTypeEVMStr:     ChainTypeEVM,
	ChainTypeSolanaStr:  ChainTypeSolana,
	ChainTypeCosmosStr:  ChainTypeCosmos,
}

func ValidateChainType(chainType ChainType) bool {
	return chainType <= ChainTypeCosmos
}

func (c ChainType) String() string {
	return ChainTypeString[c]
}

func (c *ChainType) FromString(s string) error {
	chainType, ok := ChainTypeFromString[s]
	if !ok {
		return fmt.Errorf("invalid chain type")
	}
	*c = chainType
	return nil
}
