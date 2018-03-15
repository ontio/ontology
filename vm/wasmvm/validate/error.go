// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validate

import (
	"errors"
	"fmt"

	"github.com/Ontology/vm/wasmvm/wasm"
	ops "github.com/Ontology/vm/wasmvm/wasm/operators"
)

type Error struct {
	Offset   int // Byte offset in the bytecode vector where the error occurs.
	Function int // Index into the function index space for the offending function.
	Err      error
}

func (e Error) Error() string {
	return fmt.Sprintf("error while validating function %d at offset %d: %v", e.Function, e.Offset, e.Err)
}

var ErrStackUnderflow = errors.New("validate: stack underflow")

type InvalidImmediateError struct {
	ImmType string
	OpName  string
}

func (e InvalidImmediateError) Error() string {
	return fmt.Sprintf("invalid immediate for op %s at (should be %s)", e.OpName, e.ImmType)
}

type UnmatchedOpError byte

func (e UnmatchedOpError) Error() string {
	n1, _ := ops.New(byte(e))
	return fmt.Sprintf("encountered unmatched %s", n1.Name)
}

type InvalidLabelError uint32

func (e InvalidLabelError) Error() string {
	return fmt.Sprintf("invalid nesting depth %d", uint32(e))
}

type InvalidLocalIndexError uint32

func (e InvalidLocalIndexError) Error() string {
	return fmt.Sprintf("invalid index for local variable %d", uint32(e))
}

type InvalidTypeError struct {
	Wanted wasm.ValueType
	Got    wasm.ValueType
}

func (e InvalidTypeError) Error() string {
	return fmt.Sprintf("invalid type, got: %v, wanted: %v", e.Got, e.Wanted)
}

type InvalidElementIndexError uint32

func (e InvalidElementIndexError) Error() string {
	return fmt.Sprintf("invalid element index %d", uint32(e))
}

type NoSectionError wasm.SectionID

func (e NoSectionError) Error() string {
	return fmt.Sprintf("reference to non existant section (id %d) in module", wasm.SectionID(e))
}
