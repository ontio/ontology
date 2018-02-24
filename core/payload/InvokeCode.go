package payload

import (
	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/types"
	"io"
)

//type InvokeCode struct {
//	CodeHash common.Uint160
//	Code     []byte
//}

type InvokeCode struct {
	GasLimit  common.Fixed64
	Code      types.VmCode
	Params []byte
}

func (self *InvokeCode) Data() []byte {
	return []byte{0}
}

func (self *InvokeCode) Serialize(w io.Writer) error {
	self.GasLimit.Serialize(w)
	err := self.Code.Serialize(w)
	if err != nil {
		return err
	}
	err = serialization.WriteVarBytes(w, self.Params)

	return err
}

func (self *InvokeCode) Deserialize(r io.Reader) error {
	self.GasLimit.Deserialize(r)
	self.Code.Deserialize(r)

	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	self.Params = buf

	return nil
}
