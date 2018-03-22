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

type Integer struct {
	value *big.Int
}

func NewInteger(value *big.Int) *Integer {
	var i Integer
	i.value = value
	return &i
}

func (i *Integer) Equals(other StackItemInterface) bool {
	if _, ok := other.(*Integer); !ok {
		return false
	}
	if i.value.Cmp(other.GetBigInteger()) != 0 {
		return false
	}
	return true
}

func (i *Integer) GetBigInteger() *big.Int {
	return i.value
}

func (i *Integer) GetBoolean() bool {
	if i.value.Cmp(big.NewInt(0)) == 0 {
		return false
	}
	return true
}

func (i *Integer) GetByteArray() []byte {
	return ConvertBigIntegerToBytes(i.value)
}

func (i *Integer) GetInterface() interfaces.IInteropInterface {
	return nil
}

func (i *Integer) GetArray() []StackItemInterface {
	return []StackItemInterface{i}
}

func (i *Integer) GetStruct() []StackItemInterface {
	return []StackItemInterface{i}
}

