package neovm

import (
	"testing"
)

func TestOpDcall(t *testing.T) {
	var e ExecutionEngine
	stack := NewRandAccessStack()
	e.EvaluationStack = stack
	context := NewExecutionContext(&e, []byte{0x58, 0x52, 0x00, 0x6e})
	e.PushContext(context)

	PushData(&e, 1)
	opDCALL(&e)
	e.ExecuteCode()

	if e.OpCode != PUSH2 {
		t.Fatalf("NeoVM opDCALL test failed, expect PUSH2 , get %x.", e.OpCode)
	}
}
