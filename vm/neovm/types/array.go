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

package types

import (
	"github.com/Ontology/vm/neovm/interfaces"
	"math/big"
)

type Array struct {
	_array []StackItemInterface
}

func NewArray(value []StackItemInterface) *Array {
	var a Array
	a._array = value
	return &a
}

func (a *Array) Equals(other StackItemInterface) bool {
	if _, ok := other.(*Array); !ok {
		return false
	}
	a1 := a._array
	a2 := other.GetArray()
	l1 := len(a1)
	l2 := len(a2)
	if l1 != l2 {
		return false
	}
	for i := 0; i < l1; i++ {
		if !a1[i].Equals(a2[i]) {
			return false
		}
	}
	return true
}

func (a *Array) GetBigInteger() *big.Int {
	if len(a._array) == 0 {
		return big.NewInt(0)
	}
	return a._array[0].GetBigInteger()
}

func (a *Array) GetBoolean() bool {
	if len(a._array) == 0 {
		return false
	}
	return a._array[0].GetBoolean()
}

func (a *Array) GetByteArray() []byte {
	if len(a._array) == 0 {
		return []byte{}
	}
	return a._array[0].GetByteArray()
}

func (a *Array) GetInterface() interfaces.IInteropInterface {
	if len(a._array) == 0 {
		return nil
	}
	return a._array[0].GetInterface()
}

func (a *Array) GetArray() []StackItemInterface {
	return a._array
}

func (a *Array) GetStruct() []StackItemInterface {
	return a._array
}



