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

type Integer struct {
	value *big.Int
}

func NewInteger(value *big.Int) *Integer {
	var this Integer
	this.value = value
	return &this
}

func (this *Integer) Equals(other StackItems) bool {
	if this == other {
		return true
	}
	if other == nil {
		return false
	}

	v, err := other.GetBigInteger()
	if err == nil {
		if this.value.Cmp(v) == 0 {
			return true
		} else {
			return false
		}
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

func (this *Integer) GetBigInteger() (*big.Int, error) {
	return this.value, nil
}

func (this *Integer) GetBoolean() (bool, error) {
	if this.value.Cmp(big.NewInt(0)) == 0 {
		return false, nil
	}
	return true, nil
}

func (this *Integer) GetByteArray() ([]byte, error) {
	return common.BigIntToNeoBytes(this.value), nil
}

func (this *Integer) GetInterface() (interfaces.Interop, error) {
	return nil, fmt.Errorf("%s", "Not support integer to interface")
}

func (this *Integer) GetArray() ([]StackItems, error) {
	return nil, fmt.Errorf("%s", "Not support integer to array")
}

func (this *Integer) GetStruct() ([]StackItems, error) {
	return nil, fmt.Errorf("%s", "Not support integer to struct")
}

func (this *Integer) GetMap() (map[StackItems]StackItems, error) {
	return nil, fmt.Errorf("%s", "Not support integer to map")
}
