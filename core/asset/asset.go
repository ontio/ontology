package asset

import (
	"GoOnchain/common"
	"io"
	"GoOnchain/common/serialization"
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
	Precision byte
	AssetType AssetType
	RecordType AssetRecordType
}

func (a *Asset) Serialize(w io.Writer) {

	//a.ID.Serialize(w)
	serialization.WriteVarString(w,a.Name)
	w.Write([]byte{byte(a.AssetType)})
	w.Write([]byte{byte(a.RecordType)})
	w.Write([]byte{byte(a.Precision)})

}

func GetAsset(assetId common.Uint256)  *Asset{
	//TODO: GetAsset
	return nil
}
