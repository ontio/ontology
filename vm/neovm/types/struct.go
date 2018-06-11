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

type Struct struct {
	_array []StackItems
}

func NewStruct(value []StackItems) *Struct {
	var this Struct
	this._array = value
	return &this
}

func (this *Struct) Equals(other StackItems) bool {
	if _, ok := other.(*Struct); !ok {
		return false
	}
	a1 := this._array
	a2 := other.GetStruct()
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

func (this *Struct) GetBigInteger() *big.Int {
	return big.NewInt(0)
}

func (this *Struct) GetBoolean() bool {
	return true
}

func (this *Struct) GetByteArray() []byte {
	return nil
}

func (this *Struct) GetInterface() interfaces.Interop {
	return nil
}

func (s *Struct) GetArray() []StackItems {
	return s._array
}

func (s *Struct) GetStruct() []StackItems {
	return s._array
}

func (s *Struct) Clone() StackItems {
	var arr []StackItems
	for _, v := range s._array {
		if value, ok := v.(*Struct); ok {
			arr = append(arr, value.Clone())
		} else {
			arr = append(arr, v)
		}
	}
	return &Struct{arr}
}

func (this *Struct) GetMap() map[StackItems]StackItems {
	return nil
}

func (this *Struct) Add(item StackItems) {
	this._array = append(this._array, item)
}
