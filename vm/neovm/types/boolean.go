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
	"github.com/Ontology/vm/neovm/interfaces"
)

type Boolean struct {
	value bool
}

func NewBoolean(value bool) *Boolean {
	var b Boolean
	b.value = value
	return &b
}

func (b *Boolean) Equals(other StackItemInterface) bool {
	if _, ok := other.(*Boolean); !ok {
		return false
	}
	if b.value != other.GetBoolean() {
		return false
	}
	return true
}

func (b *Boolean) GetBigInteger() *big.Int {
	if b.value {
		return big.NewInt(1)
	}
	return big.NewInt(0)
}

func (b *Boolean) GetBoolean() bool {
	return b.value
}

func (b *Boolean) GetByteArray() []byte {
	if b.value {
		return []byte{1}
	}
	return []byte{0}
}

func (b *Boolean) GetInterface() interfaces.IInteropInterface {
	return nil
}

func (b *Boolean) GetArray() []StackItemInterface {
	return []StackItemInterface{b}
}

func (b *Boolean) GetStruct() []StackItemInterface {
	return []StackItemInterface{b}
}

