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
	"crypto/sha1"
	"crypto/sha256"
	"hash"
	"math/big"

	"github.com/ontio/ontology/vm/neovm/interfaces"
	"github.com/ontio/ontology/vm/neovm/types"
)

type BigIntSorter []big.Int

func (c BigIntSorter) Len() int {
	return len(c)
}
func (c BigIntSorter) Swap(i, j int) {
	if i >= 0 && i < len(c) && j >= 0 && j < len(c) {
		// Unit Test modify
		c[i], c[j] = c[j], c[i]
	}
}
func (c BigIntSorter) Less(i, j int) bool {
	if i >= 0 && i < len(c) && j >= 0 && j < len(c) {
		// Unit Test modify
		return c[i].Cmp(&c[j]) < 0
	}

	return false
}

func ToBigInt(data interface{}) *big.Int {
	var bi big.Int
	switch t := data.(type) {
	case int64:
		bi.SetInt64(int64(t))
	case int32:
		bi.SetInt64(int64(t))
	case int16:
		bi.SetInt64(int64(t))
	case int8:
		bi.SetInt64(int64(t))
	case int:
		bi.SetInt64(int64(t))
	case uint64:
		bi.SetUint64(uint64(t))
	case uint32:
		bi.SetUint64(uint64(t))
	case uint16:
		bi.SetUint64(uint64(t))
	case uint8:
		bi.SetUint64(uint64(t))
	case uint:
		bi.SetUint64(uint64(t))
	case big.Int:
		bi = t
	case *big.Int:
		bi = *t
	}
	return &bi
}

func Concat(array1 []byte, array2 []byte) []byte {
	var r []byte
	r = append(r, array1...)
	return append(r, array2...)
}

func BigIntOp(bi *big.Int, op OpCode) *big.Int {
	nb := new(big.Int)
	switch op {
	case INC:
		nb.Add(bi, big.NewInt(int64(1)))
	case DEC:
		nb.Sub(bi, big.NewInt(int64(1)))
	case NEGATE:
		nb.Neg(bi)
	case ABS:
		nb.Abs(bi)
	default:
		nb.Set(bi)
	}
	return nb
}

func BigIntZip(ints1 *big.Int, ints2 *big.Int, op OpCode) *big.Int {
	nb := new(big.Int)
	switch op {
	case AND:
		nb.And(ints1, ints2)
	case OR:
		nb.Or(ints1, ints2)
	case XOR:
		nb.Xor(ints1, ints2)
	case ADD:
		nb.Add(ints1, ints2)
	case SUB:
		nb.Sub(ints1, ints2)
	case MUL:
		nb.Mul(ints1, ints2)
	case DIV:
		nb.Quo(ints1, ints2)
	case MOD:
		nb.Rem(ints1, ints2)
	case SHL:
		nb.Lsh(ints1, uint(ints2.Int64()))
	case SHR:
		nb.Rsh(ints1, uint(ints2.Int64()))
	case MIN:
		c := ints1.Cmp(ints2)
		if c <= 0 {
			nb.Set(ints1)
		} else {
			nb.Set(ints2)
		}
	case MAX:
		c := ints1.Cmp(ints2)
		if c <= 0 {
			nb.Set(ints2)
		} else {
			nb.Set(ints1)
		}
	}
	return nb
}

func BigIntComp(bigint *big.Int, op OpCode) bool {
	var nb bool
	switch op {
	case NZ:
		nb = bigint.Cmp(big.NewInt(int64(0))) != 0
	}
	return nb
}

func BigIntMultiComp(ints1 *big.Int, ints2 *big.Int, op OpCode) bool {
	var nb bool
	switch op {
	case NUMEQUAL:
		nb = ints1.Cmp(ints2) == 0
	case NUMNOTEQUAL:
		nb = ints1.Cmp(ints2) != 0
	case LT:
		nb = ints1.Cmp(ints2) < 0
	case GT:
		nb = ints1.Cmp(ints2) > 0
	case LTE:
		nb = ints1.Cmp(ints2) <= 0
	case GTE:
		nb = ints1.Cmp(ints2) >= 0
	}
	return nb
}

func BoolZip(bi1 bool, bi2 bool, op OpCode) bool {
	var nb bool
	switch op {
	case BOOLAND:
		nb = bi1 && bi2
	case BOOLOR:
		nb = bi1 || bi2
	}
	return nb
}

func WithInOp(int1 *big.Int, int2 *big.Int, int3 *big.Int) bool {
	b1 := BigIntMultiComp(int1, int2, GTE)
	b2 := BigIntMultiComp(int1, int3, LT)
	return BoolZip(b1, b2, BOOLAND)
}

func NewStackItem(data interface{}) types.StackItems {
	var stackItem types.StackItems
	switch data.(type) {
	case int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, *big.Int, big.Int:
		stackItem = types.NewInteger(ToBigInt(data))
	case *types.Integer:
		stackItem = data.(*types.Integer)
	case *types.Array:
		stackItem = data.(*types.Array)
	case *types.Map:
		stackItem = data.(*types.Map)
	case *types.Boolean:
		stackItem = data.(*types.Boolean)
	case *types.ByteArray:
		stackItem = data.(*types.ByteArray)
	case *types.Struct:
		stackItem = data.(*types.Struct)
	case bool:
		stackItem = types.NewBoolean(data.(bool))
	case []byte:
		stackItem = types.NewByteArray(data.([]byte))
	case []types.StackItems:
		stackItem = types.NewArray(data.([]types.StackItems))
	case types.StackItems:
		stackItem = data.(types.StackItems)
	case interfaces.Interop:
		stackItem = types.NewInteropInterface(data.(interfaces.Interop))
	default:
		panic("NewStackItemInterface Invalid Type!")
	}
	return stackItem
}

func Hash(b []byte, e *ExecutionEngine) []byte {
	var sh hash.Hash
	var bt []byte
	switch e.OpCode {
	case SHA1:
		sh = sha1.New()
		sh.Write(b)
		bt = sh.Sum(nil)
	case SHA256:
		sh = sha256.New()
		sh.Write(b)
		bt = sh.Sum(nil)
	}
	return bt
}

func PopBigInt(e *ExecutionEngine) *big.Int {
	x := PopStackItem(e)
	return x.GetBigInteger()
}

func PopInt(e *ExecutionEngine) int {
	x := PopBigInt(e)
	n := int(x.Int64())
	return n
}

func PopBoolean(e *ExecutionEngine) bool {
	x := PopStackItem(e)
	return x.GetBoolean()
}

func PopArray(e *ExecutionEngine) []types.StackItems {
	x := PopStackItem(e)
	return x.GetArray()
}

func PopMap(e *ExecutionEngine) map[types.StackItems]types.StackItems {
	x := PopStackItem(e)
	return x.GetMap()
}

func Pop(e *ExecutionEngine) []types.StackItems {
	x := PopStackItem(e)
	return x.GetArray()
}

func PopInteropInterface(e *ExecutionEngine) interfaces.Interop {
	x := PopStackItem(e)
	return x.GetInterface()
}

func PopByteArray(e *ExecutionEngine) []byte {
	x := PopStackItem(e)
	return x.GetByteArray()
}

func PopStackItem(e *ExecutionEngine) types.StackItems {
	return e.EvaluationStack.Pop()
}

func PeekArray(e *ExecutionEngine) []types.StackItems {
	x := PeekStackItem(e)
	return x.GetArray()
}

func PeekInteropInterface(e *ExecutionEngine) interfaces.Interop {
	x := PeekStackItem(e)
	return x.GetInterface()
}

func PeekInt(e *ExecutionEngine) int {
	x := PeekBigInteger(e)
	n := int(x.Int64())
	return n
}

func PeekBigInteger(e *ExecutionEngine) *big.Int {
	x := PeekStackItem(e)
	return x.GetBigInteger()
}

func PeekStackItem(e *ExecutionEngine) types.StackItems {
	return e.EvaluationStack.Peek(0)
}

func PeekNInt(i int, e *ExecutionEngine) int {
	x := PeekNBigInt(i, e)
	n := int(x.Int64())
	return n
}

func PeekNBigInt(i int, e *ExecutionEngine) *big.Int {
	x := PeekNStackItem(i, e)
	return x.GetBigInteger()
}

func PeekNByteArray(i int, e *ExecutionEngine) []byte {
	x := PeekNStackItem(i, e)
	return x.GetByteArray()
}

func PeekNStackItem(i int, e *ExecutionEngine) types.StackItems {
	return e.EvaluationStack.Peek(i)
}

func EvaluationStackCount(e *ExecutionEngine) int {
	return e.EvaluationStack.Count()
}

func Push(e *ExecutionEngine, element types.StackItems) {
	e.EvaluationStack.Push(element)
}

func Count(e *ExecutionEngine) int {
	return e.EvaluationStack.Count()
}

func PushData(e *ExecutionEngine, data interface{}) {
	e.EvaluationStack.Push(NewStackItem(data))
}
