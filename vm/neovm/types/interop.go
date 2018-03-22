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
	"github.com/Ontology/common"
)

type InteropInterface struct {
	_object interfaces.IInteropInterface
}

func NewInteropInterface(value interfaces.IInteropInterface) *InteropInterface {
	var ii InteropInterface
	ii._object = value
	return &ii
}

func (ii *InteropInterface) Equals(other StackItemInterface) bool {
	if _, ok := other.(*InteropInterface); !ok {
		return false
	}
	if !common.IsEqualBytes(ii._object.ToArray(), other.GetInterface().ToArray()) {
		return false
	}
	return true
}

func (ii *InteropInterface) GetBigInteger() *big.Int {
	return big.NewInt(0)
}

func (ii *InteropInterface) GetBoolean() bool {
	if ii._object == nil {
		return false
	}
	return true
}

func (ii *InteropInterface) GetByteArray() []byte {
	return ii._object.ToArray()
}

func (ii *InteropInterface) GetInterface() interfaces.IInteropInterface {
	return ii._object
}

func (ii *InteropInterface) GetArray() []StackItemInterface {
	return []StackItemInterface{ii}
}

func (ii *InteropInterface) GetStruct() []StackItemInterface {
	return []StackItemInterface{ii}
}

