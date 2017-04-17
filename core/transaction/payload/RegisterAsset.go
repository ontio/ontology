package payload

import (
	"DNA/common"
	"DNA/core/asset"
	"DNA/crypto"
	"io"
	. "DNA/errors"

)

type RegisterAsset struct {
	Asset      *asset.Asset
	Amount     common.Fixed64
	//Precision  byte
	Issuer     *crypto.PubKey
	Controller common.Uint160
}

func (a *RegisterAsset) Data() []byte {
	//TODO: implement RegisterAsset.Data()
	return []byte{0}

}

func (a *RegisterAsset) Serialize(w io.Writer)  error {
	a.Asset.Serialize(w)
	a.Amount.Serialize(w)
	//w.Write([]byte{a.Precision})
	a.Issuer.Serialize(w)
	a.Controller.Serialize(w)
	return nil
}

func (a *RegisterAsset) Deserialize(r io.Reader) error {

	//asset
	a.Asset = new(asset.Asset)
	err := a.Asset.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[RegisterAsset], Asset Deserialize failed.")
	}

	//Amount
	a.Amount = *new(common.Fixed64)
	err = a.Amount.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[RegisterAsset], Ammount Deserialize failed.")
	}

	//Precision  byte 02/10 comment out by wjj
	//p := make([]byte, 1)
	//n, err := r.Read(p)
	//if n > 0 {
	//	a.Precision = p[0]
	//} else {
	//	return NewDetailErr(err, ErrNoCode, "[RegisterAsset], Precision Deserialize failed.")
	//}

	//Issuer     *crypto.PubKey
	a.Issuer = new(crypto.PubKey)
	err = a.Issuer.DeSerialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[RegisterAsset], Ammount Deserialize failed.")
	}

	//Controller *common.Uint160
	a.Controller = *new(common.Uint160)
	err = a.Controller.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[RegisterAsset], Ammount Deserialize failed.")
	}
	return nil
}

