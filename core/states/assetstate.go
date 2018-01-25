package states

import (
	"io"
	"bytes"
	"github.com/Ontology/core/asset"
	. "github.com/Ontology/common"
	. "github.com/Ontology/common/serialization"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
	"fmt"
)

type AssetState struct {
	StateBase
	AssetId    Uint256
	AssetType  asset.AssetType
	Name       string
	Amount     Fixed64
	Available  Fixed64
	Precision  byte
	Owner      *crypto.PubKey
	Admin      Uint160
	Issuer     Uint160
	Expiration uint32
	IsFrozen   bool
}

func (this *AssetState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	this.AssetId.Serialize(w)
	WriteByte(w, byte(this.AssetType))
	WriteVarString(w, this.Name)
	this.Amount.Serialize(w)
	fmt.Println("[AssetState]", this.Available)
	this.Available.Serialize(w)
	WriteByte(w, this.Precision)
	this.Owner.Serialize(w)
	this.Admin.Serialize(w)
	this.Issuer.Serialize(w)
	WriteUint32(w, this.Expiration)
	WriteBool(w, this.IsFrozen)
	return nil
}

func (this *AssetState) Deserialize(r io.Reader) error {
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "AssetState StateBase Deserialize failed.")
	}
	assId := new(Uint256)
	err = assId.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "AssetState AssetId Deserialize failed.")
	}
	this.AssetId = *assId
	assetType, err := ReadByte(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "AssetState AssetType Deserialize failed.")
	}
	this.AssetType = asset.AssetType(assetType)
	name, err := ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "AssetState Name Deserialize failed.")
	}
	this.Name = name

	amount := new(Fixed64)
	err = amount.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "AssetState Amount Deserialize failed.")
	}
	this.Amount = *amount

	available := new(Fixed64)
	err = available.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "AssetState Available Deserialize failed.")
	}
	this.Available = *available

	precision, err := ReadByte(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "AssetState Precision Deserialize failed.")
	}
	this.Precision = precision

	owner := new(crypto.PubKey)
	err = owner.DeSerialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "AssetState Owner Deserialize failed.")
	}
	this.Owner = owner

	admin := new(Uint160)
	err = admin.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "AssetState Admin Deserialize failed.")
	}
	this.Admin = *admin

	issuer := new(Uint160)
	err = issuer.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "AssetState Admin Deserialize failed.")
	}
	this.Issuer = *issuer

	ex, err := ReadUint32(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "AssetState Expiration Deserialize failed.")
	}
	this.Expiration = ex
	fr, err := ReadBool(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "AssetState IsFrozen Deserialize failed.")
	}
	this.IsFrozen = fr
	return nil
}

func (assetState *AssetState) ToArray() []byte {
	b := new(bytes.Buffer)
	assetState.Serialize(b)
	return b.Bytes()
}

