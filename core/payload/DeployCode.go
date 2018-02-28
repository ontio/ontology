package payload

import (
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/core/code"
	. "github.com/Ontology/errors"
	"github.com/Ontology/vm/types"
	"io"
)

const DeployCodePayloadVersion byte = 0x00

type DeployCode struct {
	VmType      types.VmType
	Code        *FunctionCode
	NeedStorage bool
	Name        string
	CodeVersion string
	Author      string
	Email       string
	Description string
}

func (dc *DeployCode) Data() []byte {
	// TODO: Data()

	return []byte{0}
}

func (dc *DeployCode) Serialize(w io.Writer) error {
	if dc.Code == nil {
		dc.Code = new(FunctionCode)
	}
	err := dc.Code.Serialize(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction DeployCode Code Serialize failed.")
	}

	err = serialization.WriteByte(w, byte(dc.VmType))
	if err != nil {
		return err
	}

	err = serialization.WriteBool(w, dc.NeedStorage)
	if err != nil {
		return err
	}

	err = serialization.WriteVarString(w, dc.Name)
	if err != nil {
		return err
	}

	err = serialization.WriteVarString(w, dc.CodeVersion)
	if err != nil {
		return err
	}

	err = serialization.WriteVarString(w, dc.Author)
	if err != nil {
		return err
	}

	err = serialization.WriteVarString(w, dc.Email)
	if err != nil {
		return err
	}

	err = serialization.WriteVarString(w, dc.Description)
	if err != nil {
		return err
	}

	return nil
}

func (dc *DeployCode) Deserialize(r io.Reader) error {
	if dc.Code == nil {
		dc.Code = new(FunctionCode)
	}

	err := dc.Code.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction DeployCode Code Deserialize failed.")
	}

	vmType, err := serialization.ReadByte(r)
	if err != nil {
		return err
	}
	dc.VmType = types.VmType(vmType)

	dc.NeedStorage, err = serialization.ReadBool(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction DeployCode NeedStorage Deserialize failed.")
	}

	dc.Name, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction DeployCode Name Deserialize failed.")
	}

	dc.CodeVersion, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction DeployCode CodeVersion Deserialize failed.")
	}

	dc.Author, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction DeployCode Author Deserialize failed.")
	}

	dc.Email, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction DeployCode Email Deserialize failed.")
	}

	dc.Description, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Transaction DeployCode Description Deserialize failed.")
	}

	return nil
}
