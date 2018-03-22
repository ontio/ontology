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

type ByteArray struct {
	value []byte
}

func NewByteArray(value []byte) *ByteArray {
	var ba ByteArray
	ba.value = value
	return &ba
}

func (ba *ByteArray) Equals(other StackItemInterface) bool {
	if _, ok := other.(*ByteArray); !ok {
		return false
	}
	a1 := ba.value
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

func (ba *ByteArray) GetBigInteger() *big.Int {
	return ConvertBytesToBigInteger(ba.value)
}

func (ba *ByteArray) GetBoolean() bool {
	for _, b := range ba.value {
		if b != 0 {
			return true
		}
	}
	return false
}

func (ba *ByteArray) GetByteArray() []byte {
	return ba.value
}

func (ba *ByteArray) GetInterface() interfaces.IInteropInterface {
	return nil
}

func (ba *ByteArray) GetArray() []StackItemInterface {
	return []StackItemInterface{ba}
}

func (ba *ByteArray) GetStruct() []StackItemInterface {
	return []StackItemInterface{ba}
}

