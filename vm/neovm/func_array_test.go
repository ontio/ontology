package neovm

import (
	"testing"
	"math/big"
	"github.com/Ontology/vm/neovm/types"
)

func TestOpArraySize(t *testing.T) {
	engine.opCode = ARRAYSIZE

	bs := []byte{0x51, 0x52}
	i := big.NewInt(1)

	is := []types.StackItemInterface{types.NewByteArray(bs), types.NewInteger(i)}
	PushData(engine, is);

	_, err := opArraySize(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("op array size result 2, execute result:", engine.GetEvaluationStack().Peek(0).GetStackItem().GetBigInteger())
}

func TestOpPack(t *testing.T) {
	engine.opCode = PACK

	bs := []byte{0x51, 0x52}
	i := big.NewInt(1)
	n := 2

	PushData(engine, bs)

	PushData(engine, i)

	PushData(engine, n)

	if _, err := opPack(engine); err != nil {
		t.Fatal(err)
	}
	array := engine.GetEvaluationStack().Peek(0).GetStackItem().GetArray()

	for _, v := range array {
		t.Log("value:", v.GetByteArray())
	}
}

func TestOpUnPack(t *testing.T) {
	engine.opCode = UNPACK

	if _, err := opUnpack(engine); err != nil {
		t.Fatal(err)
	}
	t.Log(engine.GetEvaluationStack().Pop().GetStackItem().GetBigInteger())
	t.Log(engine.GetEvaluationStack().Pop().GetStackItem().GetBigInteger())
	t.Log(engine.GetEvaluationStack().Pop().GetStackItem().GetByteArray())

}

func TestOpPickItem(t *testing.T) {
	engine.opCode = PICKITEM

	bs := []byte{0x51, 0x52}
	i := big.NewInt(1)

	is := []types.StackItemInterface{types.NewByteArray(bs), types.NewInteger(i)}
	PushData(engine, is)

	PushData(engine, 0)

	if _, err := opPickItem(engine); err != nil {
		t.Fatal(err)
	}
	t.Log(engine.GetEvaluationStack().Pop().GetStackItem().GetByteArray())

}


