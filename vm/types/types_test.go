package types

import (
	"testing"
	"math/big"
)

func TestTypes(t *testing.T) {
	i := NewInteger(big.NewInt(1))
	ba := NewByteArray([]byte{1})
	b := NewBoolean(false)
	a1 := NewArray([]StackItem{i})
	//a2 := NewArray([]StackItem{ba})
	t.Log(i.GetByteArray())
	t.Log(ba.GetBoolean())
	t.Log(b.Equals(NewBoolean(false)))
	t.Log(a1.Equals(NewArray([]StackItem{NewInteger(big.NewInt(1))})))
}
