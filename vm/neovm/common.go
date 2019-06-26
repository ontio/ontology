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
	"math/big"
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
