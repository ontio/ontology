package wasmvm

import (
	"github.com/ontio/ontology/vm/wasmvm/exec"
	"testing"
)

func TestNewWasmStateReader(t *testing.T) {
	sr := NewWasmStateReader()
	if sr == nil {
		t.Fatal("NewWasmStateReader should return a non nil state reader")
	}

	if sr.serviceMap == nil {
		t.Fatal("NewWasmStateReader should have a service map")
	}
}

func TestWasmStateReader_Register(t *testing.T) {
	sr := NewWasmStateReader()
	name := "TEST_SERVICE"
	res := sr.Register(name, func(engine *exec.ExecutionEngine) (bool, error) {
		return true, nil
	})

	if !res {
		t.Error("TestWasmStateReader_Register failed")
	}

	if !sr.Exists(name) {
		t.Error("TestWasmStateReader_Register but not stored successfully")
	}

	res, err := sr.Invoke(name, &exec.ExecutionEngine{})
	if err != nil {
		t.Error("TestWasmStateReader_Register invoke error")
	}
	if !res {
		t.Error("TestWasmStateReader_Register invoke error")
	}

	res = sr.Register(name, func(engine *exec.ExecutionEngine) (bool, error) {
		return false, nil
	})
	if res {
		t.Error("TestWasmStateReader_Register should return false while register existed function")
	}

}

func TestWasmStateReader_MergeMap(t *testing.T) {
	sr := NewWasmStateReader()

	name1 := "TEST1"
	name2 := "TEST2"
	tmpMap := make(map[string]func(engine *exec.ExecutionEngine) (bool, error))
	tmpMap[name1] = func(engine *exec.ExecutionEngine) (bool, error) {
		return true, nil
	}

	sr.Register(name2, func(engine *exec.ExecutionEngine) (bool, error) {
		return true, nil
	})
	res := sr.MergeMap(tmpMap)
	if !res {
		t.Error("TestWasmStateReader_MergeMap merge failed")
	}

	if !sr.Exists(name1) {
		t.Error("TestWasmStateReader_MergeMap should has function:" + name1)

	}

	if !sr.Exists(name2) {
		t.Error("TestWasmStateReader_MergeMap should has function:" + name2)
	}
}

func TestGetContractAddress(t *testing.T) {

}