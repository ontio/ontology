package payload

import (
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/errors"
	"io"
	"bytes"
	vmtypes "github.com/Ontology/vm/types"
)

type DeployCode struct {
	VmType      vmtypes.VmType
	Code        []byte
	NeedStorage bool
	Name        string
	Version     string
	Author      string
	Email       string
	Description string
}

func (dc *DeployCode) Data() []byte {
	// TODO: Data()

	return []byte{0}
}

func (dc *DeployCode) Serialize(w io.Writer) error {
	var err error
	err = serialization.WriteByte(w, byte(dc.VmType))
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode VmType Serialize failed.")
	}

	err = serialization.WriteVarBytes(w, dc.Code)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Code Serialize failed.")
	}

	err = serialization.WriteBool(w, dc.NeedStorage)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode NeedStorage Serialize failed.")
	}

	err = serialization.WriteVarString(w, dc.Name)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Name Serialize failed.")
	}

	err = serialization.WriteVarString(w, dc.Version)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Version Serialize failed.")
	}

	err = serialization.WriteVarString(w, dc.Author)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Author Serialize failed.")
	}

	err = serialization.WriteVarString(w, dc.Email)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Email Serialize failed.")
	}

	err = serialization.WriteVarString(w, dc.Description)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Description Serialize failed.")
	}

	return nil
}

func (dc *DeployCode) Deserialize(r io.Reader) error {
	vmType, err := serialization.ReadByte(r)
	if err != nil {
		return err
	}
	dc.VmType = vmtypes.VmType(vmType)

	dc.Code, err = serialization.ReadVarBytes(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Code Deserialize failed.")
	}

	dc.NeedStorage, err = serialization.ReadBool(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode NeedStorage Deserialize failed.")
	}

	dc.Name, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Name Deserialize failed.")
	}

	dc.Version, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode CodeVersion Deserialize failed.")
	}

	dc.Author, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Author Deserialize failed.")
	}

	dc.Email, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Email Deserialize failed.")
	}

	dc.Description, err = serialization.ReadVarString(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "DeployCode Description Deserialize failed.")
	}

	return nil
}

func (dc *DeployCode) ToArray() []byte {
	b := new(bytes.Buffer)
	dc.Serialize(b)
	return b.Bytes()
}
