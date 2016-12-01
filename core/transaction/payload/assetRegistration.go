package payload

import (
	"GoOnchain/core/asset"
	"GoOnchain/common"
	"GoOnchain/core/contract"
)

type AssetRegistration struct {
	Asset *asset.Asset
	Amount common.Fixed8
	Precision byte
	Issuer common.ECPoint
	Controller contract.Address
}


func (a *AssetRegistration) Data() []byte {
	//TODO: implement AssetRegistration.Data()
	return  []byte{0}
}