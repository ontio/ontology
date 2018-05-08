package wasmvm

import "testing"

func TestNewWasmStateMachine(t *testing.T) {
	sm := NewWasmStateMachine()
	if sm == nil {
		t.Fatal("NewWasmStateMachine should return a non nil state machine")
	}

	if sm.WasmStateReader == nil {
		t.Fatal("NewWasmStateMachine should return a non nil state reader")
	}

	if !sm.Exists("ContractLogDebug") {
		t.Error("NewWasmStateMachine should has ContractLogDebug service")
	}

	if !sm.Exists("ContractLogInfo") {
		t.Error("NewWasmStateMachine should has ContractLogInfo service")
	}

	if !sm.Exists("ContractLogError") {
		t.Error("NewWasmStateMachine should has ContractLogError service")
	}
}
