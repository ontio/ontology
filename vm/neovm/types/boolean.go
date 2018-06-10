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

type Boolean struct {
	value bool
}

func NewBoolean(value bool) *Boolean {
	var this Boolean
	this.value = value
	return &this
}

func (this *Boolean) Equals(other StackItems) bool {
	if _, ok := other.(*Boolean); !ok {
		return false
	}
	if this.value != other.GetBoolean() {
		return false
	}
	return true
}

func (this *Boolean) GetBigInteger() *big.Int {
	if this.value {
		return big.NewInt(1)
	}
	return big.NewInt(0)
}

func (this *Boolean) GetBoolean() bool {
	return this.value
}

func (this *Boolean) GetByteArray() []byte {
	if this.value {
		return []byte{1}
	}
	return []byte{0}
}

func (this *Boolean) GetInterface() interfaces.Interop {
	return nil
}

func (this *Boolean) GetArray() []StackItems {
	return nil
}

func (this *Boolean) GetStruct() []StackItems {
	return nil
}

func (this *Boolean) GetMap() map[StackItems]StackItems {
	return nil
}
