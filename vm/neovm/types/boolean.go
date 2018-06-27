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

	"github.com/ontio/ontology/vm/neovm/interfaces"
)

type Boolean struct {
	value bool
}

func NewBoolean(value bool) *Boolean {
	var this Boolean
	this.value = value
	return &this
}

func (this *Boolean) Equals(other StackItems) bool {
	if this == other {
		return true
	}
	b, err := other.GetByteArray()
	if err != nil {
		return false
	}

	tb, err := this.GetByteArray()
	if err != nil {
		return false
	}

	return bytes.Equal(tb, b)
}

func (this *Boolean) GetBigInteger() (*big.Int, error) {
	if this.value {
		return big.NewInt(1), nil
	}
	return big.NewInt(0), nil
}

func (this *Boolean) GetBoolean() (bool, error) {
	return this.value, nil
}

func (this *Boolean) GetByteArray() ([]byte, error) {
	if this.value {
		return []byte{1}, nil
	}
	return []byte{0}, nil
}

func (this *Boolean) GetInterface() (interfaces.Interop, error) {
	return nil, fmt.Errorf("%s", "Not support boolean to interface")
}

func (this *Boolean) GetArray() ([]StackItems, error) {
	return nil, fmt.Errorf("%s", "Not support boolean to array")
}

func (this *Boolean) GetStruct() ([]StackItems, error) {
	return nil, fmt.Errorf("%s", "Not support boolean to struct")
}

func (this *Boolean) GetMap() (map[StackItems]StackItems, error) {
	return nil, fmt.Errorf("%s", "Not support boolean to map")
}
