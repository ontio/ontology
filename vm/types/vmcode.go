package types

import (
	"io"
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/common"
)

type VmType byte

const (
	NativeVM = VmType(0xFF)
	NEOVM    = VmType(0x80)
	WASMVM     = VmType(0x90)
	// EVM = VmType(0x90)
)

type VmCode struct {
	VmType VmType
	Code     []byte
}

func (self *VmCode) Serialize(w io.Writer) error {
	w.Write([]byte{byte(self.VmType)})
	return serialization.WriteVarBytes(w, self.Code)

}

func (self *VmCode) Deserialize(r io.Reader) error {
	var b [1]byte
	r.Read(b[:])
	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	self.VmType = VmType(b[0])
	self.Code = buf
	return nil
}

func (self *VmCode) AddressFromVmCode() Uint160 {
	u160 := ToCodeHash(self.Code)
	u160[0] = byte(self.VmType)
	return u160
}
