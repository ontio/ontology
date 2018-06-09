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

type Integer struct {
	value *big.Int
}

func NewInteger(value *big.Int) *Integer {
	var this Integer
	this.value = value
	return &this
}

func (this *Integer) Equals(other StackItems) bool {
	if _, ok := other.(*Integer); !ok {
		return false
	}
	if this.value.Cmp(other.GetBigInteger()) != 0 {
		return false
	}
	return true
}

func (this *Integer) GetBigInteger() *big.Int {
	return this.value
}

func (this *Integer) GetBoolean() bool {
	if this.value.Cmp(big.NewInt(0)) == 0 {
		return false
	}
	return true
}

func (this *Integer) GetByteArray() []byte {
	return common.BigIntToNeoBytes(this.value)
}

func (this *Integer) GetInterface() interfaces.Interop {
	return nil
}

func (this *Integer) GetArray() []StackItems {
	return nil
}

func (this *Integer) GetStruct() []StackItems {
	return nil
}

func (this *Integer) GetMap() map[StackItems]StackItems {
	return nil
}
