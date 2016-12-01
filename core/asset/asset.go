package asset

import (
	"GoOnchain/common"

)

type AssetType byte

const (

	Currency AssetType = 0x00
	Share AssetType = 0x01
	Invoice AssetType = 0x10
	Token AssetType =  0x11
)


type AssetRecordType byte

const (
	UTXO AssetRecordType = 0x00
	Balance AssetRecordType = 0x01
)


//define the asset stucture in onchain DNA
//registered asset will be assigned to contract address
type Asset struct {
	ID        common.Uint256
	Name      *string
	AssetType AssetType
	RecordType AssetRecordType
}
