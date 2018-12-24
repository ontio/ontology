package neovm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/ontio/ontology/vm/neovm/interfaces"
	"github.com/ontio/ontology/vm/neovm/types"
	"github.com/stretchr/testify/assert"
)

type Value interface{}

func value2json(t *testing.T, expect types.VmValue) []byte {
	e, err := expect.ConvertNeoVmValueHexString()
	assert.Nil(t, err)
	exp, err := json.Marshal(e)
	assert.Nil(t, err)

	return exp
}

func assertEqual(t *testing.T, expect, actual types.VmValue) {
	assert.Equal(t, value2json(t, expect), value2json(t, actual))
}

func newVmValue(t *testing.T, data Value) types.VmValue {
	switch v := data.(type) {
	case int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, *big.Int, big.Int:
		val, err := types.VmValueFromBigInt(ToBigInt(v))
		assert.Nil(t, err)
		return val
	case bool:
		return types.VmValueFromBool(v)
	case []byte:
		val, err := types.VmValueFromBytes(v)
		assert.Nil(t, err)
		return val
	case string:
		val, err := types.VmValueFromBytes([]byte(v))
		assert.Nil(t, err)
		return val
	case []Value:
		arr := types.NewArrayValue()
		for _, item := range v {
			arr.Append(newVmValue(t, item))
		}

		return types.VmValueFromArrayVal(arr)
	case interfaces.Interop:
		return types.VmValueFromInteropValue(types.NewInteropValue(v))
	default:
		panic(fmt.Sprintf("newVmValue Invalid Type:%t", v))
	}
}

func checkStackOpCode(t *testing.T, code OpCode, origin, expected []Value) {
	executor := NewExecutor([]byte{byte(code)})
	for _, val := range origin {
		executor.EvalStack.Push(newVmValue(t, val))
	}
	err := executor.Execute()
	assert.Nil(t, err)
	stack := executor.EvalStack
	assert.Equal(t, len(expected), stack.Count())

	for i := 0; i < len(expected); i++ {
		val := expected[len(expected)-i-1]
		res, _ := stack.Pop()
		exp := newVmValue(t, val)
		assertEqual(t, res, exp)
	}
}

func TestStackOpCode(t *testing.T) {
	checkStackOpCode(t, SWAP, []Value{1, 2}, []Value{2, 1})
}

func TestArithmetic(t *testing.T) {
	checkStackOpCode(t, ADD, []Value{1, 2}, []Value{3})
	checkStackOpCode(t, SUB, []Value{1, 2}, []Value{-1})

	checkStackOpCode(t, MUL, []Value{3, 2}, []Value{6})

	checkStackOpCode(t, DIV, []Value{3, 2}, []Value{1})
	checkStackOpCode(t, DIV, []Value{103, 2}, []Value{51})

	checkStackOpCode(t, MAX, []Value{3, 2}, []Value{3})
	checkStackOpCode(t, MAX, []Value{-3, 2}, []Value{2})

	checkStackOpCode(t, MIN, []Value{3, 2}, []Value{2})
	checkStackOpCode(t, MIN, []Value{-3, 2}, []Value{-3})

	checkStackOpCode(t, SIGN, []Value{3}, []Value{1})
	checkStackOpCode(t, SIGN, []Value{-3}, []Value{-1})
	checkStackOpCode(t, SIGN, []Value{0}, []Value{0})

	checkStackOpCode(t, INC, []Value{-10}, []Value{-9})
	checkStackOpCode(t, DEC, []Value{-10}, []Value{-11})
	checkStackOpCode(t, NEGATE, []Value{-10}, []Value{10})
	checkStackOpCode(t, ABS, []Value{-10}, []Value{10})

	checkStackOpCode(t, NOT, []Value{1}, []Value{0})
	checkStackOpCode(t, NOT, []Value{0}, []Value{1})

	checkStackOpCode(t, NZ, []Value{0}, []Value{0})
	checkStackOpCode(t, NZ, []Value{-10}, []Value{1})
	checkStackOpCode(t, NZ, []Value{10}, []Value{1})
}

func TestArrayOpCode(t *testing.T) {
	checkStackOpCode(t, ARRAYSIZE, []Value{"12345"}, []Value{5})
	checkStackOpCode(t, ARRAYSIZE, []Value{[]Value{1, 2, 3}}, []Value{3})
	checkStackOpCode(t, ARRAYSIZE, []Value{[]Value{}}, []Value{0})

	checkStackOpCode(t, PACK, []Value{"aaa", "bbb", "ccc", 3}, []Value{[]Value{"ccc", "bbb", "aaa"}})

	checkStackOpCode(t, UNPACK, []Value{[]Value{"ccc", "bbb", "aaa"}}, []Value{"aaa", "bbb", "ccc"})

	checkStackOpCode(t, PICKITEM, []Value{[]Value{"ccc", "bbb", "aaa"}, 0}, []Value{"ccc"})
}

func TestAssertEqual(t *testing.T) {
	val1 := newVmValue(t, -12345678910)
	buf, _ := val1.AsBytes()
	val2 := newVmValue(t, buf)

	assertEqual(t, val1, val2)
}
