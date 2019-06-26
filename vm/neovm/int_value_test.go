/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package neovm

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
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

func CheckBigInteger(value *big.Int) bool {
	if value == nil {
		return false
	}
	if len(common.BigIntToNeoBytes(value)) > MAX_SIZE_FOR_BIGINTEGER {
		return false
	}
	return true
}

func TestRsh(t *testing.T) {
	val := types.IntValFromInt(math.MaxInt64)
	b := new(big.Int).SetUint64(math.MaxUint64)
	val2, err := types.IntValFromBigInt(b)
	assert.Nil(t, err)
	res, err := val.Rsh(val2)
	assert.Nil(t, err)

	left := new(big.Int).SetInt64(math.MaxInt64)
	right := new(big.Int).SetUint64(math.MaxUint64)
	res2 := BigIntZip(left, right, SHR)
	res22, err := types.IntValFromBigInt(res2)
	assert.Nil(t, err)
	assert.Equal(t, res, res22)
}

func TestCmp(t *testing.T) {
	a, _ := new(big.Int).SetString("73786976294838206464", 10)
	b, _ := new(big.Int).SetString("83786976294838206464", 10)
	val_a, err := types.IntValFromBigInt(a)
	assert.Nil(t, err)
	val_b, err := types.IntValFromBigInt(b)
	assert.Nil(t, err)
	assert.True(t, val_a.Cmp(val_b) < 0)

	res, err := types.IntValFromBigInt(big.NewInt(0).Not(a))
	assert.Nil(t, err)
	assert.Equal(t, val_a.Not(), res)
}

func TestIntValFromNeoBytes(t *testing.T) {
	bs := common.BigIntToNeoBytes(new(big.Int).SetUint64(math.MaxUint64))
	val, err := types.IntValFromNeoBytes(bs)
	assert.Nil(t, err)
	val2, err := types.IntValFromBigInt(new(big.Int).SetUint64(math.MaxUint64))
	assert.Nil(t, err)
	assert.Equal(t, val, val2)
}
