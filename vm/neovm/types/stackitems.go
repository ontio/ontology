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

type StackItems interface {
	Equals(other StackItems) bool
	GetBigInteger() *big.Int
	GetBoolean() bool
	GetByteArray() []byte
	GetInterface() interfaces.Interop
	GetArray() []StackItems
	GetStruct() []StackItems
	GetMap() map[StackItems]StackItems
}

const (
	ByteArrayType byte = 0x00
	BooleanType   byte = 0x01
	IntegerType   byte = 0x02
	InterfaceType byte = 0x40
	ArrayType     byte = 0x80
	StructType    byte = 0x81
	MapType       byte = 0x82
)
