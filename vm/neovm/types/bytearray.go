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
	"bytes"
	"fmt"
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
	if this == other {
		return true
	}

	a1 := this.value
	a2, err := other.GetByteArray()
	if err != nil {
		return false
	}

	return bytes.Equal(a1, a2)
}

func (this *ByteArray) GetBigInteger() (*big.Int, error) {
	return common.BigIntFromNeoBytes(this.value), nil
}

func (this *ByteArray) GetBoolean() (bool, error) {
	for _, b := range this.value {
		if b != 0 {
			return true, nil
		}
	}
	return false, nil
}

func (this *ByteArray) GetByteArray() ([]byte, error) {
	return this.value, nil
}

func (this *ByteArray) GetInterface() (interfaces.Interop, error) {
	return nil, fmt.Errorf("%s", "Not support byte array to interface")
}

func (this *ByteArray) GetArray() ([]StackItems, error) {
	return nil, fmt.Errorf("%s", "Not support byte array to array")
}

func (this *ByteArray) GetStruct() ([]StackItems, error) {
	return nil, fmt.Errorf("%s", "Not support byte array to struct")
}

func (this *ByteArray) GetMap() (map[StackItems]StackItems, error) {
	return nil, fmt.Errorf("%s", "Not support byte array to map")
}
