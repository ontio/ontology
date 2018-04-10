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

// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wasm

import (
	"fmt"
	"io"

	"github.com/ontio/ontology/vm/wasmvm/wasm/leb128"
)

// ValueType represents the type of a valid value in Wasm
type ValueType int8

const (
	ValueTypeI32 ValueType = -0x01
	ValueTypeI64 ValueType = -0x02
	ValueTypeF32 ValueType = -0x03
	ValueTypeF64 ValueType = -0x04
)

var valueTypeStrMap = map[ValueType]string{
	ValueTypeI32: "i32",
	ValueTypeI64: "i64",
	ValueTypeF32: "f32",
	ValueTypeF64: "f64",
}

func (t ValueType) String() string {
	str, ok := valueTypeStrMap[t]
	if !ok {
		str = fmt.Sprintf("<unknown value_type %d>", int8(t))
	}
	return str
}

// TypeFunc represents the value type of a function
const TypeFunc int = -0x20

func readValueType(r io.Reader) (ValueType, error) {
	v, err := leb128.ReadVarint32(r)
	return ValueType(v), err
}

// BlockType represents the signature of a structured block
type BlockType ValueType // varint7
const BlockTypeEmpty BlockType = -0x40

func readBlockType(r io.Reader) (BlockType, error) {
	b, err := leb128.ReadVarint32(r)
	return BlockType(b), err
}

func (b BlockType) String() string {
	if b == BlockTypeEmpty {
		return "<empty block>"
	}
	return ValueType(b).String()
}

// ElemType describes the type of a table's elements
type ElemType int // varint7
// ElemTypeAnyFunc descibres an any_func value
const ElemTypeAnyFunc ElemType = -0x10

func readElemType(r io.Reader) (ElemType, error) {
	b, err := leb128.ReadVarint32(r)
	return ElemType(b), err
}

func (t ElemType) String() string {
	if t == ElemTypeAnyFunc {
		return "anyfunc"
	}

	return "<unknown elem_type>"
}

// FunctionSig describes the signature of a declared function in a WASM module
type FunctionSig struct {
	// value for the 'func` type constructor
	Form int8
	// The parameter types of the function
	ParamTypes  []ValueType
	ReturnTypes []ValueType
}

func (f FunctionSig) String() string {
	return fmt.Sprintf("<func %v -> %v>", f.ParamTypes, f.ReturnTypes)
}

type InvalidTypeConstructorError struct {
	Wanted int
	Got    int
}

func (e InvalidTypeConstructorError) Error() string {
	return fmt.Sprintf("wasm: invalid type constructor: wanted %d, got %d", e.Wanted, e.Got)
}

func readFunction(r io.Reader) (FunctionSig, error) {
	f := FunctionSig{}

	form, err := leb128.ReadVarint32(r)
	if err != nil {
		return f, err
	}

	f.Form = int8(form)

	paramCount, err := leb128.ReadVarUint32(r)
	if err != nil {
		return f, err
	}
	f.ParamTypes = make([]ValueType, paramCount)

	for i := range f.ParamTypes {
		f.ParamTypes[i], err = readValueType(r)
		if err != nil {
			return f, err
		}
	}

	returnCount, err := leb128.ReadVarUint32(r)
	if err != nil {
		return f, err
	}

	f.ReturnTypes = make([]ValueType, returnCount)
	for i := range f.ReturnTypes {
		vt, err := readValueType(r)
		if err != nil {
			return f, err
		}
		f.ReturnTypes[i] = vt
	}

	return f, nil
}

// GlobalVar describes the type and mutability of a declared global variable
type GlobalVar struct {
	Type    ValueType // Type of the value stored by the variable
	Mutable bool      // Whether the value of the variable can be changed by the set_global operator
}

func readGlobalVar(r io.Reader) (*GlobalVar, error) {
	g := &GlobalVar{}
	var err error

	g.Type, err = readValueType(r)
	if err != nil {
		return nil, err
	}

	m, err := leb128.ReadVarUint32(r)
	if err != nil {
		return nil, err
	}

	g.Mutable = m == 1

	return g, nil
}

// Table describes a table in a Wasm module.
type Table struct {
	// The type of elements
	ElementType ElemType
	Limits      ResizableLimits
}

func readTable(r io.Reader) (*Table, error) {
	t := Table{}
	var err error

	t.ElementType, err = readElemType(r)
	if err != nil {
		return nil, err
	}

	lims, err := readResizableLimits(r)
	if err != nil {
		return nil, err
	}

	t.Limits = *lims
	return &t, err
}

type Memory struct {
	Limits ResizableLimits
}

func readMemory(r io.Reader) (*Memory, error) {
	lim, err := readResizableLimits(r)
	if err != nil {
		return nil, err
	}

	return &Memory{*lim}, nil
}

// External describes the kind of the entry being imported or exported.
type External uint8

const (
	ExternalFunction External = 0
	ExternalTable    External = 1
	ExternalMemory   External = 2
	ExternalGlobal   External = 3
)

func (e External) String() string {
	switch e {
	case ExternalFunction:
		return "function"
	case ExternalTable:
		return "table"
	case ExternalMemory:
		return "memory"
	case ExternalGlobal:
		return "global"
	default:
		return "<unknown external_kind>"
	}
}
func readExternal(r io.Reader) (External, error) {
	bytes, err := readBytes(r, 1)

	return External(bytes[0]), err
}

// ResizableLimits describe the limit of a table or linear memory.
type ResizableLimits struct {
	Flags   uint32 // 1 if the Maximum field is valid
	Initial uint32 // initial length (in units of table elements or wasm pages)
	Maximum uint32 // If flags is 1, it describes the maximum size of the table or memory
}

func readResizableLimits(r io.Reader) (*ResizableLimits, error) {
	lim := &ResizableLimits{
		Maximum: 0,
	}
	f, err := leb128.ReadVarUint32(r)
	if err != nil {
		return nil, err
	}

	lim.Flags = f
	lim.Initial, err = leb128.ReadVarUint32(r)
	if err != nil {
		return nil, err
	}

	if lim.Flags&0x1 != 0 {
		m, err := leb128.ReadVarUint32(r)
		if err != nil {
			return nil, err
		}
		lim.Maximum = m

	}
	return lim, nil
}
