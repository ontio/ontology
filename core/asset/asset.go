package asset

import (
	"DNA/common/serialization"
	. "DNA/errors"
	"errors"
	"io"
)

//AssetType
type AssetType byte

const (
	Currency AssetType = 0x00
	Share    AssetType = 0x01
	Invoice  AssetType = 0x10
	Token    AssetType = 0x11
)

const (
	MaxPrecision = 8
	MinPrecision = 0
)

type AssetRecordType byte

//onchain DNA is planed to support UTXO and Balance
const (
	UTXO    AssetRecordType = 0x00
	Balance AssetRecordType = 0x01
)

//define the asset stucture in onchain DNA
//registered asset will be assigned to contract address
type Asset struct {
	Name        string
	Description string
	Precision   byte
	AssetType   AssetType
	RecordType  AssetRecordType
}

// Serialize is the implement of SignableData interface.
func (a *Asset) Serialize(w io.Writer) error {
	err := serialization.WriteVarString(w, a.Name)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Asset], Name serialize failed.")
	}
	err = serialization.WriteVarString(w, a.Description)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Asset], Description serialize failed.")
	}
	_, err = w.Write([]byte{byte(a.Precision)})
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Asset], Precision serialize failed.")
	}
	_, err = w.Write([]byte{byte(a.AssetType)})
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Asset], AssetType serialize failed.")
	}
	_, err = w.Write([]byte{byte(a.RecordType)})
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Asset], RecordType serialize failed.")
	}
	return nil
}

// Deserialize is the implement of SignableData interface.
func (a *Asset) Deserialize(r io.Reader) error {
	name, err := serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(errors.New("[Asset], Name deserialize failed."), ErrNoCode, "")
	}
	a.Name = name
	description, err := serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(errors.New("[Asset], Description deserialize failed."), ErrNoCode, "")
	}
	a.Description = description
	p := make([]byte, 1)
	n, err := r.Read(p)
	if n > 0 {
		a.Precision = p[0]
	} else {
		return NewDetailErr(errors.New("[Asset], Precision deserialize failed."), ErrNoCode, "")
	}
	n, err = r.Read(p)
	if n > 0 {
		a.AssetType = AssetType(p[0])
	} else {
		return NewDetailErr(errors.New("[Asset], AssetType deserialize failed."), ErrNoCode, "")
	}
	n, err = r.Read(p)
	if n > 0 {
		a.RecordType = AssetRecordType(p[0])
	} else {
		return NewDetailErr(errors.New("[Asset], RecordType deserialize failed."), ErrNoCode, "")
	}
	return nil
}
