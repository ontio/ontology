package states

import (
	"bytes"
	"testing"

	"github.com/ontio/ontology/smartcontract/types"
)

func TestContract_Serialize_Deserialize(t *testing.T) {
	vmcode := types.VmCode{
		VmType: types.Native,
		Code:   []byte{1},
	}

	addr := vmcode.AddressFromVmCode()

	c := &Contract{
		Version: 0,
		Code:    []byte{1},
		Address: addr,
		Method:  "init",
		Args:    []byte{2},
	}
	bf := new(bytes.Buffer)
	if err := c.Serialize(bf); err != nil {
		t.Fatalf("Contract serialize error: %v", err)
	}

	v := new(Contract)
	if err := v.Deserialize(bf); err != nil {
		t.Fatalf("Contract deserialize error: %v", err)
	}
}
