package neovm

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/vm/neovm/types"
	"github.com/stretchr/testify/assert"
)

func randInt64() *big.Int {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	r := binary.LittleEndian.Uint64(buf)
	right := big.NewInt(int64(r))
	return right
}

func genBBInt() (*big.Int, *big.Int) {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	left := common.BigIntFromNeoBytes(buf)
	_, _ = rand.Read(buf)
	right := common.BigIntFromNeoBytes(buf)
	return left, right
}

func genBLInt() (*big.Int, *big.Int) {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	left := common.BigIntFromNeoBytes(buf)
	right := randInt64()
	return left, right
}

func genLBInt() (*big.Int, *big.Int) {
	right, left := genBLInt()
	return left, right
}

func genLLInt() (*big.Int, *big.Int) {
	left := randInt64()
	right := randInt64()
	return left, right
}

type IntOp func(left, right *big.Int) ([]byte, error)

func compareIntOpInner(t *testing.T, left, right *big.Int, func1, func2 IntOp) {

	val1, err := func1(left, right)
	val2, err2 := func2(left, right)
	if err != nil || err2 != nil {
		return
	}

	assert.Equal(t, val1, val2)
}

func compareIntOp(t *testing.T, func1, func2 IntOp) {
	const N = 10000
	for i := 0; i < N; i++ {
		left, right := genBBInt()
		compareIntOpInner(t, left, right, func1, func2)
		left, right = genLLInt()
		compareIntOpInner(t, left, right, func1, func2)
		left, right = genBLInt()
		compareIntOpInner(t, left, right, func1, func2)
		left, right = genLBInt()
		compareIntOpInner(t, left, right, func1, func2)
	}
}

func TestIntValue_Abs(t *testing.T) {
	compareIntOp(t, func(left, right *big.Int) ([]byte, error) {
		abs := big.NewInt(0).Abs(left)
		return common.BigIntToNeoBytes(abs), nil
	}, func(left, right *big.Int) ([]byte, error) {
		val, err := types.IntValFromBigInt(left)
		assert.Nil(t, err)
		val = val.Abs()

		return val.ToNeoBytes(), nil
	})
}

func TestIntValue_Other(t *testing.T) {
	opcodes := []OpCode{MOD, AND, OR, XOR, ADD, SUB, MUL, DIV, SHL, SHR, MAX, MIN}
	for _, opcode := range opcodes {
		compareIntOp(t, func(left, right *big.Int) ([]byte, error) {
			return compareFuncBigInt(left, right, opcode)
		}, func(left, right *big.Int) ([]byte, error) {
			return compareFuncIntValue(left, right, opcode)
		})
	}
}

func compareFuncIntValue(left, right *big.Int, opcode OpCode) ([]byte, error) {
	lhs, err := types.IntValFromBigInt(left)
	if err != nil {
		return nil, err
	}
	rhs, err := types.IntValFromBigInt(right)
	if err != nil {
		return nil, err
	}
	var val types.IntValue
	switch opcode {
	case AND:
		val, err = lhs.And(rhs)
	case OR:
		val, err = lhs.Or(rhs)
	case XOR:
		val, err = lhs.Xor(rhs)
	case ADD:
		val, err = lhs.Add(rhs)
	case SUB:
		val, err = lhs.Sub(rhs)
	case MUL:
		val, err = lhs.Mul(rhs)
	case DIV:
		val, err = lhs.Div(rhs)
	case MOD:
		val, err = lhs.Mod(rhs)
	case SHL:
		val, err = lhs.Lsh(rhs)
	case SHR:
		val, err = lhs.Rsh(rhs)
	case MIN:
		val, err = lhs.Min(rhs)
	case MAX:
		val, err = lhs.Max(rhs)
	}
	return val.ToNeoBytes(), err
}

func compareFuncBigInt(left, right *big.Int, opcode OpCode) ([]byte, error) {
	if opcode == SHL {

		if right.Sign() < 0 {
			return nil, fmt.Errorf("neg num")
		}

		if left.Sign() != 0 && right.Cmp(big.NewInt(MAX_SIZE_FOR_BIGINTEGER*8)) > 0 {
			return nil, fmt.Errorf("the biginteger over max size 32bit")
		}

		if CheckBigInteger(new(big.Int).Lsh(left, uint(right.Int64()))) == false {
			return nil, fmt.Errorf("the biginteger over max size 32bit")
		}
	}
	nb := BigIntZip(left, right, opcode)
	return common.BigIntToNeoBytes(nb), nil
}
