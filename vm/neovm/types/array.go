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
	"math/big"

	"github.com/ontio/ontology/vm/neovm/interfaces"
)

type Array struct {
	_array []StackItems
}

func NewArray(value []StackItems) *Array {
	var this Array
	this._array = value
	return &this
}

func (this *Array) Equals(other StackItems) bool {
	if _, ok := other.(*Array); !ok {
		return false
	}
	a1 := this._array
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

func (this *Array) GetBigInteger() *big.Int {
	return big.NewInt(0)
}

func (this *Array) GetBoolean() bool {
	return true
}

func (this *Array) GetByteArray() []byte {
	return nil
}

func (this *Array) GetInterface() interfaces.Interop {
	return nil
}

func (this *Array) GetArray() []StackItems {
	return this._array
}

func (this *Array) GetStruct() []StackItems {
	return this._array
}

func (this *Array) GetMap() map[StackItems]StackItems {
	return nil
}

func (this *Array) Add(item StackItems) {
	this._array = append(this._array, item)
}
