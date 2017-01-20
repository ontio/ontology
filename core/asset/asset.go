package asset

import (
	"GoOnchain/common"
	"GoOnchain/common/serialization"
	. "GoOnchain/errors"
	"errors"
	"io"
)

type AssetType byte

const (
	Currency AssetType = 0x00
	Share    AssetType = 0x01
	Invoice  AssetType = 0x10
	Token    AssetType = 0x11
)

type AssetRecordType byte

const (
	UTXO    AssetRecordType = 0x00
	Balance AssetRecordType = 0x01
)

//define the asset stucture in onchain DNA
//registered asset will be assigned to contract address
type Asset struct {
	ID         common.Uint256
	Name       string
	Precision  byte
	AssetType  AssetType
	RecordType AssetRecordType
}

func (a *Asset) Serialize(w io.Writer) error {

	a.ID.Serialize(w)
	err := serialization.WriteVarString(w, a.Name)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Asset item Name serialize failed.")
	}
	_, err = w.Write([]byte{byte(a.Precision)})
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Asset item Precision serialize failed.")
	}
	_, err = w.Write([]byte{byte(a.AssetType)})
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Asset item AssetType serialize failed.")
	}
	_, err = w.Write([]byte{byte(a.RecordType)})
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Asset item RecordType serialize failed.")
	}
	return nil
}

func (a *Asset) Deserialize(r io.Reader) error {

	a.ID.Deserialize(r)
	vars, err := serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Asset item Name deserialize failed.")
	}
	a.Name = vars
	p := make([]byte, 1)
	n, err := r.Read(p)
	if n > 0 {
		a.Precision = p[0]
	} else {
		return NewDetailErr(errors.New("Asset item Precision deserialize failed."), ErrNoCode, "")
	}
	n, err = r.Read(p)
	if n > 0 {
		a.AssetType = AssetType(p[0])
	} else {
		return NewDetailErr(errors.New("Asset item AssetType deserialize failed."), ErrNoCode, "")
	}
	n, err = r.Read(p)
	if n > 0 {
		a.RecordType = AssetRecordType(p[0])
	} else {
		return NewDetailErr(errors.New("Asset item RecordType deserialize failed."), ErrNoCode, "")
	}
	return nil
}

func GetAsset(assetId common.Uint256) *Asset {
	//TODO: GetAsset get from database
	return nil
}
