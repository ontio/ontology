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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/vm/neovm/interfaces"
)

type ByteArray struct {
	value []byte
}

func NewByteArray(value []byte) *ByteArray {
	var this ByteArray
	this.value = value
	return &this
}

func (this *ByteArray) Equals(other StackItems) bool {
	if _, ok := other.(*ByteArray); !ok {
		return false
	}
	a1 := this.value
	a2 := other.GetByteArray()
	l1 := len(a1)
	l2 := len(a2)
	if l1 != l2 {
		return false
	}
	for i := 0; i < l1; i++ {
		if a1[i] != a2[i] {
			return false
		}
	}
	return true
}

func (this *ByteArray) GetBigInteger() *big.Int {
	return common.BigIntFromNeoBytes(this.value)
}

func (this *ByteArray) GetBoolean() bool {
	for _, b := range this.value {
		if b != 0 {
			return true
		}
	}
	return false
}

func (this *ByteArray) GetByteArray() []byte {
	return this.value
}

func (this *ByteArray) GetInterface() interfaces.Interop {
	return nil
}

func (this *ByteArray) GetArray() []StackItems {
	return nil
}

func (this *ByteArray) GetStruct() []StackItems {
	return nil
}

func (this *ByteArray) GetMap() map[StackItems]StackItems {
	return nil
}
