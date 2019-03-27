package neovm

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ripemd160"
	"math/big"
	"testing"

	"github.com/ontio/ontology/smartcontract/common"
	"github.com/ontio/ontology/vm/neovm/interfaces"
	"github.com/ontio/ontology/vm/neovm/types"
	"github.com/stretchr/testify/assert"
)

type Value interface{}

func value2json(t *testing.T, expect types.VmValue) string {
	//e, err := expect.ConvertNeoVmValueHexString()
	e, err := expect.Stringify()
	assert.Nil(t, err)
	exp, err := json.Marshal(e)
	assert.Nil(t, err)

	return string(exp)
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
	case map[interface{}]interface{}:
		mp := types.NewMapValue()
		for key, value := range v {
			mp.Set(newVmValue(t, key), newVmValue(t, value))
		}
		return types.VmValueFromMapValue(mp)
	case interfaces.Interop:
		return types.VmValueFromInteropValue(types.NewInteropValue(v))
	default:
		panic(fmt.Sprintf("newVmValue Invalid Type:%t", v))
	}
}

func newVmValueOld(t *testing.T, data Value) types.StackItems {
	switch v := data.(type) {
	case int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, *big.Int, big.Int:
		val := types.NewInteger(ToBigInt(v))
		return val
	case bool:
		return types.NewBoolean(v)
	case []byte:
		val := types.NewByteArray(v)
		return val
	case string:
		val := types.NewByteArray([]byte(v))
		return val
	case []Value:
		var arr []types.StackItems
		for _, item := range v {
			arr = append(arr, newVmValueOld(t, item))
		}

		return types.NewArray(arr)
	case map[interface{}]interface{}:
		mp := types.NewMap()
		for k, value := range v {
			mp.Add(newVmValueOld(t, k), newVmValueOld(t, value))
		}
		return mp
	case interfaces.Interop:
		return types.NewInteropInterface(v)
	default:
		panic(fmt.Sprintf("newVmValue Invalid Type:%t", v))
	}
}

func checkStackOpCode(t *testing.T, code OpCode, origin, expected []Value) {
	checkAltStackOpCode(t, code, [2][]Value{origin, {}}, [2][]Value{expected, {}})
}

func checkAltStackOpCode(t *testing.T, code OpCode, origin [2][]Value, expected [2][]Value) {
	checkAltStackOpCodeOld(t, []byte{byte(code)}, origin, expected)
	checkAltStackOpCodeNew(t, []byte{byte(code)}, origin, expected)
}

func checkMultiStackOpCode(t *testing.T, code []OpCode, origin, expected []Value) {
	var raw []byte
	for _, c := range code {
		raw = append(raw, byte(c))
	}
	checkMultiAltStackOpCode(t, raw, [2][]Value{origin, {}}, [2][]Value{expected, {}})
}
func checkMultiAltStackOpCode(t *testing.T, code []byte, origin [2][]Value, expected [2][]Value) {
	var raw []byte
	for _, c := range code {
		raw = append(raw, byte(c))
	}
	checkAltStackOpCodeOld(t, raw, origin, expected)
	checkAltStackOpCodeNew(t, raw, origin, expected)
}
func checkMultiOpCode(t *testing.T, code []byte, origin []Value, expected []Value) {
	var raw []byte
	for _, c := range code {
		raw = append(raw, byte(c))
	}
	checkAltStackOpCodeOld(t, raw, [2][]Value{origin}, [2][]Value{expected})
	checkAltStackOpCodeNew(t, raw, [2][]Value{origin}, [2][]Value{expected})
}

func checkAltStackOpCodeNew(t *testing.T, code []byte, origin [2][]Value, expected [2][]Value) {
	executor := NewExecutor(code)
	for _, val := range origin[0] {
		err := executor.EvalStack.Push(newVmValue(t, val))
		assert.Nil(t, err)
	}
	for _, val := range origin[1] {
		err := executor.AltStack.Push(newVmValue(t, val))
		assert.Nil(t, err)
	}
	err := executor.Execute()
	assert.Nil(t, err)
	assert.Equal(t, len(expected[0]), executor.EvalStack.Count())
	assert.Equal(t, len(expected[1]), executor.AltStack.Count())

	stacks := [2]*ValueStack{executor.EvalStack, executor.AltStack}
	for s, stack := range stacks {
		expect := expected[s]
		for i := 0; i < len(expect); i++ {
			val := expect[len(expect)-i-1]
			res, _ := stack.Pop()
			exp := newVmValue(t, val)
			assertEqual(t, res, exp)
		}
	}
}

func TestAltStackOpCode(t *testing.T) {
	checkAltStackOpCode(t, DUPFROMALTSTACK, [2][]Value{
		{8888},
		{9999},
	}, [2][]Value{
		{8888, 9999},
		{9999},
	})

	checkAltStackOpCode(t, TOALTSTACK, [2][]Value{
		{8888},
		{9999},
	}, [2][]Value{
		{},
		{9999, 8888},
	})

	checkAltStackOpCode(t, FROMALTSTACK, [2][]Value{
		{8888},
		{9999},
	}, [2][]Value{
		{8888, 9999},
		{},
	})
}

func TestStackOpCode(t *testing.T) {
	checkStackOpCode(t, SWAP, []Value{1, 2}, []Value{2, 1})
	checkStackOpCode(t, XDROP, []Value{3, 2, 1}, []Value{2})
	checkStackOpCode(t, XDROP, []Value{3, 2, 0}, []Value{3})
	checkStackOpCode(t, XSWAP, []Value{3, 2, 1}, []Value{2, 3})
	checkStackOpCode(t, XTUCK, []Value{3, 2, 1}, []Value{3, 2, 2})
	checkStackOpCode(t, DEPTH, []Value{1, 2}, []Value{1, 2, 2})
	checkStackOpCode(t, DROP, []Value{1, 2}, []Value{1})
	checkStackOpCode(t, DUP, []Value{1, 2}, []Value{1, 2, 2})
	checkStackOpCode(t, NIP, []Value{1, 2}, []Value{2})
	checkStackOpCode(t, OVER, []Value{1, 2}, []Value{1, 2, 1})
	checkStackOpCode(t, PICK, []Value{3, 2, 1}, []Value{3, 2, 3})
	checkStackOpCode(t, ROLL, []Value{3, 2, 1}, []Value{2, 3})
	checkStackOpCode(t, ROT, []Value{4, 3, 2, 1}, []Value{4, 2, 1, 3})
	checkStackOpCode(t, ROT, []Value{1, 2, 3}, []Value{2, 3, 1})
	checkStackOpCode(t, TUCK, []Value{1, 2}, []Value{2, 1, 2})

	checkStackOpCode(t, INVERT, []Value{2}, []Value{-3})
	checkStackOpCode(t, AND, []Value{1, 2}, []Value{0})
	checkStackOpCode(t, OR, []Value{1, 2}, []Value{3})
	checkStackOpCode(t, XOR, []Value{1, 2}, []Value{3})
	checkStackOpCode(t, EQUAL, []Value{1, 2}, []Value{false})

	checkStackOpCode(t, INC, []Value{1}, []Value{2})
	checkStackOpCode(t, DEC, []Value{2}, []Value{1})
	checkStackOpCode(t, SIGN, []Value{1}, []Value{1})
	checkStackOpCode(t, NEGATE, []Value{1}, []Value{-1})
	checkStackOpCode(t, ABS, []Value{-9999}, []Value{9999})
	checkStackOpCode(t, NOT, []Value{true}, []Value{false})

	checkStackOpCode(t, SHL, []Value{1, 2}, []Value{4})
	checkStackOpCode(t, SHR, []Value{4, 1}, []Value{2})
	checkStackOpCode(t, BOOLAND, []Value{1, 2}, []Value{1})
	checkStackOpCode(t, BOOLOR, []Value{1, 2}, []Value{1})
	checkStackOpCode(t, NUMEQUAL, []Value{1, 2}, []Value{false})
	checkStackOpCode(t, NUMNOTEQUAL, []Value{1, 2}, []Value{1})
	checkStackOpCode(t, LT, []Value{1, 2}, []Value{1})
	checkStackOpCode(t, GT, []Value{1, 2}, []Value{false})
	checkStackOpCode(t, LTE, []Value{1, 2}, []Value{1})
	checkStackOpCode(t, GTE, []Value{1, 2}, []Value{false})
	checkStackOpCode(t, MIN, []Value{1, 2}, []Value{1})
	checkStackOpCode(t, MAX, []Value{1, 2}, []Value{2})

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

	checkStackOpCode(t, NOT, []Value{1}, []Value{false})
	checkStackOpCode(t, NOT, []Value{0}, []Value{1})

	checkStackOpCode(t, NZ, []Value{0}, []Value{false})
	checkStackOpCode(t, NZ, []Value{-10}, []Value{true})
	checkStackOpCode(t, NZ, []Value{10}, []Value{true})
}

func TestArrayOpCode(t *testing.T) {
	checkStackOpCode(t, ARRAYSIZE, []Value{"12345"}, []Value{5})
	checkStackOpCode(t, ARRAYSIZE, []Value{[]Value{1, 2, 3}}, []Value{3})
	checkStackOpCode(t, ARRAYSIZE, []Value{[]Value{}}, []Value{0})

	checkStackOpCode(t, PACK, []Value{"aaa", "bbb", "ccc", 3}, []Value{[]Value{"ccc", "bbb", "aaa"}})

	checkStackOpCode(t, UNPACK, []Value{[]Value{"ccc", "bbb", "aaa"}}, []Value{"aaa", "bbb", "ccc", 3})

	checkStackOpCode(t, PICKITEM, []Value{[]Value{"ccc", "bbb", "aaa"}, 0}, []Value{"ccc"})
	checkStackOpCode(t, PICKITEM, []Value{[]Value{"ccc", "bbb", "aaa"}, 1}, []Value{"bbb"})

	// reverse will pop the value from stack
	checkStackOpCode(t, REVERSE, []Value{[]Value{"ccc", "bbb", "aaa"}}, []Value{})
	checkMultiStackOpCode(t, []OpCode{TOALTSTACK, DUPFROMALTSTACK, REVERSE, FROMALTSTACK},
		[]Value{[]Value{"ccc", "bbb", "aaa"}},
		[]Value{[]Value{"aaa", "bbb", "ccc"}},
	)

	checkMultiStackOpCode(t, []OpCode{SWAP, TOALTSTACK, DUPFROMALTSTACK, SWAP, APPEND, FROMALTSTACK},
		[]Value{[]Value{"aaa", "bbb", "ccc"}, "eee"},
		[]Value{[]Value{"aaa", "bbb", "ccc", "eee"}},
	)

	checkStackOpCode(t, WITHIN, []Value{1, 2, 3}, []Value{false})
}

func TestMapValue(t *testing.T) {
	mp := make(map[interface{}]interface{}, 0)
	mp["key"] = "value"
	mp["key2"] = "value2"

	mp2 := make(map[interface{}]interface{}, 0)
	mp2["key2"] = "value2"
	checkMultiStackOpCode(t, []OpCode{SWAP, TOALTSTACK, DUPFROMALTSTACK, SWAP, REMOVE, FROMALTSTACK},
		[]Value{mp, "key"},
		[]Value{mp2},
	)

	checkMultiStackOpCode(t, []OpCode{HASKEY}, []Value{mp, "key"}, []Value{true})
	checkMultiStackOpCode(t, []OpCode{KEYS}, []Value{mp}, []Value{[]Value{"key", "key2"}})
	checkMultiStackOpCode(t, []OpCode{VALUES}, []Value{mp}, []Value{[]Value{"value", "value2"}})
}

func TestStringOpcode(t *testing.T) {
	checkStackOpCode(t, SIZE, []Value{"12345"}, []Value{5})
	checkStackOpCode(t, CAT, []Value{"aaa", "bbb"}, []Value{"aaabbb"})
	checkStackOpCode(t, SUBSTR, []Value{"aaabbb", 1, 3}, []Value{"aab"})
	checkStackOpCode(t, LEFT, []Value{"aaabbb", 3}, []Value{"aaa"})
	checkStackOpCode(t, RIGHT, []Value{"aaabbb", 3}, []Value{"bbb"})
}

func TestPUSHDATA(t *testing.T) {
	checkStackOpCode(t, PUSH0, []Value{9999}, []Value{9999, 0})
	checkStackOpCode(t, PUSH1, []Value{9999}, []Value{9999, 1})
	checkStackOpCode(t, PUSH2, []Value{9999}, []Value{9999, 2})
	checkStackOpCode(t, PUSH4, []Value{9999}, []Value{9999, 4})
	checkStackOpCode(t, PUSHM1, []Value{1}, []Value{1, -1})
	checkStackOpCode(t, PUSH1, []Value{9999}, []Value{9999, 1})
	checkStackOpCode(t, PUSH2, []Value{9999}, []Value{9999, 2})
	checkStackOpCode(t, PUSH3, []Value{9999}, []Value{9999, 3})
	checkStackOpCode(t, PUSH4, []Value{9999}, []Value{9999, 4})
	checkStackOpCode(t, PUSH5, []Value{9999}, []Value{9999, 5})
	checkStackOpCode(t, PUSH6, []Value{9999}, []Value{9999, 6})
	checkStackOpCode(t, PUSH7, []Value{9999}, []Value{9999, 7})
	checkStackOpCode(t, PUSH8, []Value{9999}, []Value{9999, 8})
	checkStackOpCode(t, PUSH9, []Value{9999}, []Value{9999, 9})
	checkStackOpCode(t, PUSH10, []Value{9999}, []Value{9999, 10})
	checkStackOpCode(t, PUSH11, []Value{9999}, []Value{9999, 11})
	checkStackOpCode(t, PUSH12, []Value{9999}, []Value{9999, 12})
	checkStackOpCode(t, PUSH13, []Value{9999}, []Value{9999, 13})
	checkStackOpCode(t, PUSH14, []Value{9999}, []Value{9999, 14})
	checkStackOpCode(t, PUSH15, []Value{9999}, []Value{9999, 15})
	checkStackOpCode(t, PUSH16, []Value{9999}, []Value{9999, 16})
}

func TestFlowControl(t *testing.T) {
	checkMultiStackOpCode(t, []OpCode{PUSH3, DCALL, PUSH0, PUSH1, RET}, nil, []Value{1, 0, 1})
	checkMultiOpCode(t, []byte{byte(CALL), byte(0x03), byte(0x00), byte(PUSH2), byte(RET)}, nil, []Value{2, 2})
	checkMultiOpCode(t, []byte{byte(JMP), byte(0x03), byte(0x00), byte(PUSH2), byte(RET)}, nil, []Value{2})
	checkMultiOpCode(t, []byte{byte(JMPIF), byte(0x03), byte(0x00), byte(PUSH2), byte(RET)}, []Value{true}, []Value{2})
	checkMultiOpCode(t, []byte{byte(JMPIF), byte(0x04), byte(0x00), byte(PUSH2), byte(PUSH14), byte(RET)}, []Value{true}, []Value{14})
	checkMultiOpCode(t, []byte{byte(JMPIFNOT), byte(0x03), byte(0x00), byte(PUSH2), byte(RET)}, []Value{true}, []Value{2})
	checkMultiOpCode(t, []byte{byte(JMPIFNOT), byte(0x04), byte(0x00), byte(PUSH2), byte(PUSH1), byte(RET)}, []Value{true}, []Value{2, 1})
}

func TestPushData(t *testing.T) {
	checkMultiOpCode(t, []byte{byte(PUSHDATA1), byte(1), byte(2)}, nil, []Value{2})
	checkMultiOpCode(t, []byte{byte(PUSHDATA2), byte(0x01), byte(0x00), byte(2)}, nil, []Value{2})
	checkMultiOpCode(t, []byte{byte(PUSHDATA4), byte(0x01), byte(0x00), byte(0x00), byte(0x00), byte(2)}, nil, []Value{2})
}

func TestPushBytes(t *testing.T) {
	checkMultiOpCode(t, []byte{byte(PUSHBYTES1), byte(1)}, nil, []Value{1})
	code := make([]byte, 0)
	code = append(code, byte(PUSHBYTES75))
	for i := 0; i < int(PUSHBYTES75); i++ {
		code = append(code, byte(1))
	}
	code2 := make([]byte, len(code)-1, cap(code))
	copy(code2, code[1:])
	checkMultiOpCode(t, code, nil, []Value{code2})
}

func TestHashOpCode(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	temp := sha256.Sum256(data)
	md := ripemd160.New()
	md.Write(temp[:])
	checkStackOpCode(t, HASH160, []Value{data}, []Value{md.Sum(nil)})
	hash256 := sha256.Sum256(temp[:])
	checkStackOpCode(t, HASH256, []Value{data}, []Value{hash256[:]})

	sh := sha1.New()
	sh.Write(data)
	hash := sh.Sum(nil)
	checkStackOpCode(t, SHA1, []Value{data}, []Value{hash[:]})

	sh = sha256.New()
	sh.Write(data)
	hash = sh.Sum(nil)
	checkStackOpCode(t, SHA256, []Value{data}, []Value{hash[:]})
}

func TestAssertEqual(t *testing.T) {
	val1 := newVmValue(t, -12345678910)
	buf, _ := val1.AsBytes()
	val2 := newVmValue(t, buf)

	assertEqual(t, val1, val2)
}

func checkAltStackOpCodeOld(t *testing.T, code []byte, origin [2][]Value, expected [2][]Value) {
	executor := NewExecutionEngine()
	context := NewExecutionContext(code)
	executor.PushContext(context)
	for _, val := range origin[0] {
		executor.EvaluationStack.Push(newVmValueOld(t, val))
	}
	for _, val := range origin[1] {
		executor.AltStack.Push(newVmValueOld(t, val))
	}
	err := executor.Execute()
	assert.Nil(t, err)
	assert.Equal(t, len(expected[0]), executor.EvaluationStack.Count())
	assert.Equal(t, len(expected[1]), executor.AltStack.Count())

	stacks := [2]*RandomAccessStack{executor.EvaluationStack, executor.AltStack}
	for s, stack := range stacks {
		expect := expected[s]
		for i := 0; i < len(expect); i++ {
			val := expect[len(expect)-i-1]
			res := stack.Pop()
			exp := newVmValueOld(t, val)
			assertEqualOld(t, res, exp)
		}
	}
}

func oldValue2json(t *testing.T, expect types.StackItems) string {
	//e, err := common.ConvertNeoVmTypeHexString(expect)
	e, err := common.Stringify(expect)
	assert.Nil(t, err)
	exp, err := json.Marshal(e)
	assert.Nil(t, err)

	return string(exp)
}

func assertEqualOld(t *testing.T, expect, actual types.StackItems) {
	ex := oldValue2json(t, expect)
	act := oldValue2json(t, actual)

	assert.Equal(t, ex, act)
}
