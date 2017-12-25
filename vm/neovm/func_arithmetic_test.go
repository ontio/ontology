package neovm

import (
	"testing"
	"math/big"
	"github.com/Ontology/vm/neovm/types"
)

var (
	engine = NewExecutionEngine(nil, nil, nil, nil, 0)
)

func TestOpBigInt(t *testing.T) {
	state, err := opBigInt(engine)
	t.Log("state:", state, "err:", err)

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(1))))
	engine.opCode = INC

	state, err = opBigInt(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("1 inc result 2, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())

	engine.opCode = DEC

	state, err = opBigInt(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("2 dec result 1, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())

	engine.opCode = NEGATE
	state, err = opBigInt(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("1 negate result -1, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())

	engine.opCode = ABS
	state, err = opBigInt(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("-1 negate result 1, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())
}

func TestOpNot(t *testing.T) {
	engine := NewExecutionEngine(nil, nil, nil, nil, 0)
	state, err := opNot(engine)
	t.Log("state:", state, "err:", err)

	engine.evaluationStack.Push(NewStackItem(types.NewBoolean(false)))
	engine.opCode = NOT
	state, err = opNot(engine)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("false not result true, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBoolean())

}

func TestOpNz(t *testing.T) {
	state, err := opNz(engine)
	t.Log("state:", state, "err:", err)

	if err != nil {
		t.Fatal(err)
	}

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(1))))
	engine.opCode = NZ
	state, err = opNz(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("1 nz result true, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBoolean())

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(0))))
	state, err = opNz(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("0 nz result false, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBoolean())

}

func TestBigIntZip(t *testing.T) {
	state, err := opBigIntZip(engine)
	t.Log("state:", state, "err:", err)

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(1))))
	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(2))))

	engine.opCode = ADD

	state, err = opBigIntZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("1 add 2 result 3, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(1))))

	engine.opCode = SUB

	state, err = opBigIntZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("3 sub 1 result 2, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(2))))

	engine.opCode = MUL

	state, err = opBigIntZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("2 mul 2 result 4, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(2))))

	engine.opCode = DIV

	state, err = opBigIntZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("4 mul 2 result 2, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(3))))

	engine.opCode = MOD

	state, err = opBigIntZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("2 mod 3 result 2, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(2))))

	engine.opCode = SHL

	state, err = opBigIntZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("2 shl 2 result 8, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(1))))

	engine.opCode = SHR

	state, err = opBigIntZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("8 shr 1 result 4, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(6))))

	engine.opCode = MIN

	state, err = opBigIntZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("4 min 6 result 4, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(6))))

	engine.opCode = MAX

	state, err = opBigIntZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("4 max 6 result 6, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(4))))

	engine.opCode = AND

	state, err = opBigIntZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("4 and 6 result 4, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(6))))

	engine.opCode = OR

	state, err = opBigIntZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("4 or 6 result 6, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(4))))

	engine.opCode = XOR

	state, err = opBigIntZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("4 xor 6 result 2, execute result:", engine.evaluationStack.Peek(0).GetStackItem().GetBigInteger())
}

func TestOpBoolZip(t *testing.T) {

	engine.evaluationStack.Push(NewStackItem(types.NewBoolean(false)))
	engine.evaluationStack.Push(NewStackItem(types.NewBoolean(true)))

	engine.opCode = BOOLAND

	_, err := opBoolZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("false booland true result false, execute result", engine.evaluationStack.Peek(0).GetStackItem().GetBoolean())

	engine.evaluationStack.Push(NewStackItem(types.NewBoolean(true)))
	engine.evaluationStack.Push(NewStackItem(types.NewBoolean(true)))

	engine.opCode = BOOLAND

	_, err = opBoolZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("true booland true result true, execute result", engine.evaluationStack.Peek(0).GetStackItem().GetBoolean())

	engine.evaluationStack.Push(NewStackItem(types.NewBoolean(false)))
	engine.evaluationStack.Push(NewStackItem(types.NewBoolean(false)))

	engine.opCode = BOOLOR

	_, err = opBoolZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("false boolor false result false, execute result", engine.evaluationStack.Peek(0).GetStackItem().GetBoolean())

	engine.evaluationStack.Push(NewStackItem(types.NewBoolean(true)))
	engine.evaluationStack.Push(NewStackItem(types.NewBoolean(false)))

	engine.opCode = BOOLOR

	_, err = opBoolZip(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("true boolor false result true, execute result", engine.evaluationStack.Peek(0).GetStackItem().GetBoolean())
}

func TestOpOpWithIn(t *testing.T) {
	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(4))))
	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(2))))
	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(5))))

	engine.opCode = WITHIN

	_, err := opWithIn(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("4 >= 2 && 4 < 5 result true, execute result", engine.evaluationStack.Peek(0).GetStackItem().GetBoolean())

	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(1))))
	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(2))))
	engine.evaluationStack.Push(NewStackItem(types.NewInteger(big.NewInt(3))))

	engine.opCode = WITHIN

	_, err = opWithIn(engine)

	if err != nil {
		t.Fatal(err)
	}

	t.Log("1 >= 2 && 1 < 3 result false, execute result", engine.evaluationStack.Peek(0).GetStackItem().GetBoolean())
}

