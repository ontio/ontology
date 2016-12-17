package payload

import (
	"GoOnchain/core/asset"
	"GoOnchain/common"
	"GoOnchain/crypto"
	"io"
)

type AssetRegistration struct {
	Asset asset.Asset
	Amount common.Fixed64
	Precision byte
	Issuer crypto.PubKey
	Controller common.Uint160
}




func (a *AssetRegistration) Data() []byte {
	//TODO: implement AssetRegistration.Data()
	return  []byte{0}

}

func (a *AssetRegistration) Serialize(w io.Writer) {
	a.Asset.Serialize(w)
	a.Amount.Serialize(w)
	w.Write([]byte{a.Precision})
	a.Issuer.Serialize(w)
	a.Controller.Serialize(w)
}