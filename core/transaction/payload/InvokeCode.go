package payload

import (
	"io"
	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
)

type InvokeCode struct {
	CodeHash common.Uint160
	Code     []byte
}

func (ic *InvokeCode) Data(version byte) []byte {
	return []byte{0}
}

func (ic *InvokeCode) Serialize(w io.Writer, version byte) error {
	ic.CodeHash.Serialize(w)
	err := serialization.WriteVarBytes(w, ic.Code)
	if err != nil {
		return err
	}
	return nil
}

func (ic *InvokeCode) Deserialize(r io.Reader, version byte) error {
	u := new(common.Uint160)
	if err := u.Deserialize(r); err != nil {
		return err
	}
	ic.CodeHash = *u
	code, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	ic.Code = code
	return nil
}