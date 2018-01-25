package states

import (
	"io"
	"bytes"
	. "github.com/Ontology/common/serialization"
	"github.com/Ontology/core/code"
	"github.com/Ontology/smartcontract/types"
	. "github.com/Ontology/errors"
)

type ContractState struct {
	StateBase
	Code        *code.FunctionCode
	VmType      types.VmType
	NeedStorage bool
	Name        string
	Version     string
	Author      string
	Email       string
	Description string
}

func (this *ContractState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	err := this.Code.Serialize(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState Code Serialize failed.")
	}
	err = WriteByte(w, byte(this.VmType))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState VmType Serialize failed.")
	}

	err = WriteBool(w, this.NeedStorage)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState NeedStorage Serialize failed.")
	}

	err = WriteVarString(w, this.Name)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState Name Serialize failed.")
	}
	err = WriteVarString(w, this.Version)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState Version Serialize failed.")
	}
	err = WriteVarString(w, this.Author)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState Author Serialize failed.")
	}
	err = WriteVarString(w, this.Email)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState Email Serialize failed.")
	}
	err = WriteVarString(w, this.Description)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState Description Serialize failed.")
	}
	return nil
}

func (this *ContractState) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(ContractState)
	}
	f := new(code.FunctionCode)

	err := this.StateBase.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState StateBase Deserialize failed.")
	}
	err = f.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState Code Deserialize failed.")
	}
	this.Code = f

	vmType, err := ReadByte(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState VmType Deserialize failed.")
	}
	this.VmType = types.VmType(vmType)

	this.NeedStorage, err = ReadBool(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState NeedStorage Deserialize failed.")
	}

	this.Name, err = ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState Name Deserialize failed.")
	}
	this.Version, err = ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState Version Deserialize failed.")
	}
	this.Author, err = ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState Author Deserialize failed.")
	}
	this.Email, err = ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState Email Deserialize failed.")
	}
	this.Description, err = ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "ContractState Description Deserialize failed.")
	}
	return nil
}

func (contractState *ContractState) ToArray() []byte {
	b := new(bytes.Buffer)
	contractState.Serialize(b)
	return b.Bytes()
}


