package exec

import (
	"bytes"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/vm/wasmvm/wasm"
	"testing"
)

func TestEnvCall_GetParams(t *testing.T) {
	envCall := &EnvCall{}
	if envCall.GetParams() != nil {
		t.Error("TestEnvCall_GetParams should return nil")
	}
	envCall.envParams = []uint64{uint64(1), uint64(2)}
	param := envCall.GetParams()
	if len(param) != 2 {
		t.Error("TestEnvCall_GetParams should have 2 params")
	}
}

func TestEnvCall_GetReturns(t *testing.T) {
	envCall := &EnvCall{}
	if envCall.GetReturns() != false {
		t.Error("TestEnvCall_GetReturns should return false")
	}

	envCall.envReturns = true
	if envCall.GetReturns() != true {
		t.Error("TestEnvCall_GetReturns should return true")

	}
}

func TestVM_GetMemory(t *testing.T) {

	code := "0061736d01000000010c0260027f7f017f60017f017f03030200010710020361646400000673717561726500010a11020700200120006a0b0700200020006c0b"
	bs, err := common.HexToBytes(code)
	if err != nil {
		t.Error("TestVM_GetMemory read code failed")
	}
	bf := bytes.NewBuffer(bs)
	m, err := wasm.ReadModule(bf, importer)
	if err != nil {
		t.Error("TestVM_GetMemory read code failed")
	}
	vm, err := NewVM(m)
	if err != nil {
		t.Error("TestVM_GetMemory read code failed")
	}
	memory := vm.GetMemory()
	if memory == nil {
		t.Fatal("TestVM_GetMemory memory should not be nil")
	}
	if len(memory.Memory) != wasmPageSize {
		t.Error("TestVM_GetMemory memory size is not correct")
	}

}
