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
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/ontio/ontology/vm/wasmvm/wasm/internal/readpos"
	"github.com/ontio/ontology/vm/wasmvm/wasm/leb128"
)

// SectionID is a 1-byte code that encodes the section code of both known and custom sections.
type SectionID uint8

const (
	SectionIDCustom   SectionID = 0
	SectionIDType     SectionID = 1
	SectionIDImport   SectionID = 2
	SectionIDFunction SectionID = 3
	SectionIDTable    SectionID = 4
	SectionIDMemory   SectionID = 5
	SectionIDGlobal   SectionID = 6
	SectionIDExport   SectionID = 7
	SectionIDStart    SectionID = 8
	SectionIDElement  SectionID = 9
	SectionIDCode     SectionID = 10
	SectionIDData     SectionID = 11
)

func (s SectionID) String() string {
	n, ok := map[SectionID]string{
		SectionIDCustom:   "custom",
		SectionIDType:     "type",
		SectionIDImport:   "import",
		SectionIDFunction: "function",
		SectionIDTable:    "table",
		SectionIDMemory:   "memory",
		SectionIDGlobal:   "global",
		SectionIDExport:   "export",
		SectionIDStart:    "start",
		SectionIDElement:  "element",
		SectionIDCode:     "code",
		SectionIDData:     "data",
	}[s]
	if !ok {
		return "unknown"
	}
	return n
}

// Section is a declared section in a WASM module.
type Section struct {
	Start int64
	End   int64

	ID SectionID
	// Size of this section in bytes
	PayloadLen uint32
	// Section name, empty if id != 0
	Name  string
	Bytes []byte
}

type InvalidSectionIDError SectionID

func (e InvalidSectionIDError) Error() string {
	return fmt.Sprintf("wasm: invalid section ID %d", e)
}

type InvalidCodeIndexError int

func (e InvalidCodeIndexError) Error() string {
	return fmt.Sprintf("wasm: invalid index to code section: %d", int(e))
}

var ErrUnsupportedSection = errors.New("wasm: unsupported section")

type MissingSectionError SectionID

func (e MissingSectionError) Error() string {
	return fmt.Sprintf("wasm: missing section %s", SectionID(e).String())
}

// reads a valid section from r. The first return value is true if and only if
// the module has been completely read.
func (m *Module) readSection(r *readpos.ReadPos) (bool, error) {
	var err error
	var id uint32

	logger.Println("Reading section ID")
	if id, err = leb128.ReadVarUint32(r); err != nil {
		if err == io.EOF { // no bytes were read, the reader is empty
			return true, nil
		}
		return false, err
	}
	s := Section{ID: SectionID(id)}

	logger.Println("Reading payload length")
	if s.PayloadLen, err = leb128.ReadVarUint32(r); err != nil {
		return false, nil
	}

	payloadDataLen := s.PayloadLen

	if s.ID == SectionIDCustom {
		nameLen, nameLenSize, err := leb128.ReadVarUint32Size(r)
		if err != nil {
			return false, err
		}
		payloadDataLen -= uint32(nameLenSize)
		if s.Name, err = readString(r, int(nameLen)); err != nil {
			return false, err
		}

		payloadDataLen -= uint32(len(s.Name))
	}

	logger.Printf("Section payload length: %d", payloadDataLen)

	s.Start = r.CurPos

	sectionBytes := new(bytes.Buffer)
	sectionBytes.Grow(int(payloadDataLen))
	sectionReader := io.LimitReader(io.TeeReader(r, sectionBytes), int64(payloadDataLen))

	switch s.ID {
	case SectionIDCustom:
		logger.Println("section custom")
		if err = m.readSectionCustom(sectionReader); err == nil {
			s.End = r.CurPos
			s.Bytes = sectionBytes.Bytes()
			m.Other = append(m.Other, s)
		}
	case SectionIDType:
		logger.Println("section type")
		if err = m.readSectionTypes(sectionReader); err == nil {
			s.End = r.CurPos
			s.Bytes = sectionBytes.Bytes()
			m.Types.Section = s
		}
	case SectionIDImport:
		logger.Println("section import")
		if err = m.readSectionImports(sectionReader); err == nil {
			s.End = r.CurPos
			s.Bytes = sectionBytes.Bytes()
			m.Import.Section = s
		}
	case SectionIDFunction:
		logger.Println("section function")
		if err = m.readSectionFunctions(sectionReader); err == nil {
			s.End = r.CurPos
			s.Bytes = sectionBytes.Bytes()
			m.Function.Section = s
		}
	case SectionIDTable:
		logger.Println("section table")
		if err = m.readSectionTables(sectionReader); err == nil {
			s.End = r.CurPos
			s.Bytes = sectionBytes.Bytes()
			m.Table.Section = s
		}
	case SectionIDMemory:
		logger.Println("section memory")
		if err = m.readSectionMemories(sectionReader); err == nil {
			s.End = r.CurPos
			s.Bytes = sectionBytes.Bytes()
			m.Memory.Section = s
		}
	case SectionIDGlobal:
		logger.Println("section global")
		if err = m.readSectionGlobals(sectionReader); err == nil {
			s.End = r.CurPos
			s.Bytes = sectionBytes.Bytes()
			m.Global.Section = s
		}
	case SectionIDExport:
		logger.Println("section export")
		if err = m.readSectionExports(sectionReader); err == nil {
			s.End = r.CurPos
			s.Bytes = sectionBytes.Bytes()
			m.Export.Section = s
		}
	case SectionIDStart:
		logger.Println("section start")
		if err = m.readSectionStart(sectionReader); err == nil {
			s.End = r.CurPos
			s.Bytes = sectionBytes.Bytes()
			m.Start.Section = s
		}
	case SectionIDElement:
		logger.Println("section element")
		if err = m.readSectionElements(sectionReader); err == nil {
			s.End = r.CurPos
			s.Bytes = sectionBytes.Bytes()
			m.Elements.Section = s
		}
	case SectionIDCode:
		logger.Println("section code")
		if err = m.readSectionCode(sectionReader); err == nil {
			s.End = r.CurPos
			s.Bytes = sectionBytes.Bytes()
			m.Code.Section = s
		}
	case SectionIDData:
		logger.Println("section data")
		if err = m.readSectionData(sectionReader); err == nil {
			s.End = r.CurPos
			s.Bytes = sectionBytes.Bytes()
			m.Data.Section = s
		}
	default:
		return false, InvalidSectionIDError(s.ID)
	}

	logger.Println(err)
	return false, err
}

func (m *Module) readSectionCustom(r io.Reader) error {
	_, err := io.Copy(ioutil.Discard, r)
	return err
}

// SectionTypes declares all function signatures that will be used in a module.
type SectionTypes struct {
	Section
	Entries []FunctionSig
}

func (m *Module) readSectionTypes(r io.Reader) error {
	s := &SectionTypes{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	s.Entries = make([]FunctionSig, int(count))

	for i := range s.Entries {
		if s.Entries[i], err = readFunction(r); err != nil {
			return err
		}

	}

	m.Types = s

	return nil
}

// SectionImports declares all imports that will be used in the module.
type SectionImports struct {
	Section
	Entries []ImportEntry
}

func (m *Module) readSectionImports(r io.Reader) error {
	s := &SectionImports{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}
	s.Entries = make([]ImportEntry, count)

	for i := range s.Entries {
		s.Entries[i], err = readImportEntry(r)
		if err != nil {
			return err
		}
	}

	m.Import = s
	return nil
}

func readImportEntry(r io.Reader) (ImportEntry, error) {
	i := ImportEntry{}

	modLen, err := leb128.ReadVarUint32(r)
	if err != nil {
		return i, err
	}

	if i.ModuleName, err = readString(r, int(modLen)); err != nil {
		return i, err
	}

	fieldLen, err := leb128.ReadVarUint32(r)
	if err != nil {
		return i, err
	}

	if i.FieldName, err = readString(r, int(fieldLen)); err != nil {
		return i, err
	}

	if i.Kind, err = readExternal(r); err != nil {
		return i, err
	}

	switch i.Kind {
	case ExternalFunction:
		logger.Println("importing function")
		var t uint32
		t, err = leb128.ReadVarUint32(r)
		i.Type = FuncImport{t}
	case ExternalTable:
		logger.Println("importing table")
		var table *Table

		table, err = readTable(r)
		if table != nil {
			i.Type = TableImport{*table}
		}
	case ExternalMemory:
		logger.Println("importing memory")
		var mem *Memory

		mem, err = readMemory(r)
		if mem != nil {
			i.Type = MemoryImport{*mem}
		}
	case ExternalGlobal:
		logger.Println("importing global var")
		var gl *GlobalVar
		gl, err = readGlobalVar(r)
		if gl != nil {
			i.Type = GlobalVarImport{*gl}
		}

	default:
		return i, InvalidExternalError(i.Kind)
	}

	return i, err
}

// SectionFunction declares the signature of all functions defined in the module (in the code section)
type SectionFunctions struct {
	Section
	// Sequences of indices into (FunctionSignatues).Entries
	Types []uint32
}

func (m *Module) readSectionFunctions(r io.Reader) error {
	s := &SectionFunctions{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	s.Types = make([]uint32, count)

	for i := range s.Types {
		t, err := leb128.ReadVarUint32(r)
		if err != nil {
			return err
		}
		s.Types[i] = t
	}

	m.Function = s
	return nil
}

// SectionTables describes all tables declared by a module.
type SectionTables struct {
	Section
	Entries []Table
}

func (m *Module) readSectionTables(r io.Reader) error {
	s := &SectionTables{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}
	s.Entries = make([]Table, count)

	for i := range s.Entries {
		t, err := readTable(r)
		if err != nil {
			return err
		}
		s.Entries[i] = *t
	}

	m.Table = s
	return err
}

// SectionMemories describes all linaer memories used by a module.
type SectionMemories struct {
	Section
	Entries []Memory
}

func (m *Module) readSectionMemories(r io.Reader) error {
	s := &SectionMemories{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	s.Entries = make([]Memory, count)

	for i := range s.Entries {
		m, err := readMemory(r)
		if err != nil {
			return err
		}
		s.Entries[i] = *m
	}

	m.Memory = s
	return err
}

// SectionGlobals defines the value of all global variables declared in a module.
type SectionGlobals struct {
	Section
	Globals []GlobalEntry
}

func (m *Module) readSectionGlobals(r io.Reader) error {
	s := &SectionGlobals{}

	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}
	s.Globals = make([]GlobalEntry, count)

	logger.Printf("%d global entries\n", count)
	for i := range s.Globals {
		s.Globals[i], err = readGlobalEntry(r)
		if err != nil {
			return err
		}
	}

	m.Global = s
	return nil
}

// GlobalEntry declares a global variable.
type GlobalEntry struct {
	Type    *GlobalVar // Type holds information about the value type and mutability of the variable
	Init    []byte     // Init is an initializer expression that computes the initial value of the variable
	IsEnv   bool       // true if the Global is env import
	InitVal uint64     // todo init value from env outside
}

func readGlobalEntry(r io.Reader) (e GlobalEntry, err error) {
	logger.Println("reading global_type")
	e.Type, err = readGlobalVar(r)
	if err != nil {
		logger.Println("Error!")
		return
	}
	logger.Println("reading init expr")

	// init_expr is delimited by opcode "end" (0x0b)
	e.Init, err = readInitExpr(r)
	logger.Println("Value:", e.Init)
	return e, err
}

// SectionExports declares the export section of a module
type SectionExports struct {
	Section
	Entries map[string]ExportEntry
}

type DuplicateExportError string

func (e DuplicateExportError) Error() string {
	return fmt.Sprintf("Duplicate export entry: %s", e)
}

func (m *Module) readSectionExports(r io.Reader) error {
	s := &SectionExports{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}
	s.Entries = make(map[string]ExportEntry, count)

	for i := uint32(0); i < count; i++ {
		entry, err := readExportEntry(r)
		if err != nil {
			return err
		}

		if _, exists := s.Entries[entry.FieldStr]; exists {
			return DuplicateExportError(entry.FieldStr)
		}
		s.Entries[entry.FieldStr] = entry
	}

	m.Export = s
	return nil
}

// ExportEntry represents an exported entry by the module
type ExportEntry struct {
	FieldStr string
	Kind     External
	Index    uint32
}

func readExportEntry(r io.Reader) (ExportEntry, error) {
	e := ExportEntry{}
	fieldLen, err := leb128.ReadVarUint32(r)

	if e.FieldStr, err = readString(r, int(fieldLen)); err != nil {
		return e, err
	}

	if e.Kind, err = readExternal(r); err != nil {
		return e, err
	}

	e.Index, err = leb128.ReadVarUint32(r)

	return e, err
}

// SectionStartFunction represents the start function section.
type SectionStartFunction struct {
	Section
	Index uint32 // The index of the start function into the global index space.
}

func (m *Module) readSectionStart(r io.Reader) error {
	s := &SectionStartFunction{}
	var err error

	s.Index, err = leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	m.Start = s
	return nil
}

// SectionElements describes the initial contents of a table's elements.
type SectionElements struct {
	Section
	Entries []ElementSegment
}

func (m *Module) readSectionElements(r io.Reader) error {
	s := &SectionElements{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}
	s.Entries = make([]ElementSegment, count)

	for i := range s.Entries {
		s.Entries[i], err = readElementSegment(r)
		if err != nil {
			return err
		}
	}

	m.Elements = s
	return nil
}

// ElementSegment describes a group of repeated elements that begin at a specified offset
type ElementSegment struct {
	Index  uint32 // The index into the global table space, should always be 0 in the MVP.
	Offset []byte // initializer expression for computing the offset for placing elements, should return an i32 value
	Elems  []uint32
}

func readElementSegment(r io.Reader) (ElementSegment, error) {
	s := ElementSegment{}
	var err error

	if s.Index, err = leb128.ReadVarUint32(r); err != nil {
		return s, err
	}
	if s.Offset, err = readInitExpr(r); err != nil {
		return s, err
	}

	numElems, err := leb128.ReadVarUint32(r)
	if err != nil {
		return s, err
	}
	s.Elems = make([]uint32, numElems)

	for i := range s.Elems {
		e, err := leb128.ReadVarUint32(r)
		if err != nil {
			return s, err
		}
		s.Elems[i] = e
	}

	return s, nil
}

// SectionCode describes the body for every function declared inside a module.
type SectionCode struct {
	Section
	Bodies []FunctionBody
}

func (m *Module) readSectionCode(r io.Reader) error {
	s := &SectionCode{}

	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}
	s.Bodies = make([]FunctionBody, count)
	logger.Printf("%d function bodies\n", count)

	for i := range s.Bodies {
		logger.Printf("Reading function %d\n", i)
		if s.Bodies[i], err = readFunctionBody(r); err != nil {
			return err
		}
		s.Bodies[i].Module = m
	}

	m.Code = s
	if m.Function == nil || len(m.Function.Types) == 0 {
		return MissingSectionError(SectionIDFunction)
	}
	if len(m.Function.Types) != len(s.Bodies) {
		return errors.New("The number of entries in the function and code section are unequal")
	}

	if m.Types == nil {
		return MissingSectionError(SectionIDType)
	}

	return nil
}

var ErrFunctionNoEnd = errors.New("Function body does not end with 0x0b (end)")

type FunctionBody struct {
	Module *Module // The parent module containing this function body, for execution purposes
	Locals []LocalEntry
	Code   []byte
}

func readFunctionBody(r io.Reader) (FunctionBody, error) {
	f := FunctionBody{}

	bodySize, err := leb128.ReadVarUint32(r)
	if err != nil {
		return f, err
	}

	body := make([]byte, bodySize)

	if _, err = io.ReadFull(r, body); err != nil {
		return f, err
	}

	bytesReader := bytes.NewBuffer(body)

	localCount, err := leb128.ReadVarUint32(bytesReader)
	if err != nil {
		return f, err
	}
	f.Locals = make([]LocalEntry, localCount)

	for i := range f.Locals {
		if f.Locals[i], err = readLocalEntry(bytesReader); err != nil {
			return f, err
		}
	}

	logger.Printf("bodySize: %d, localCount: %d\n", bodySize, localCount)

	code := bytesReader.Bytes()
	logger.Printf("Read %d bytes for function body", len(code))

	if code[len(code)-1] != end {
		return f, ErrFunctionNoEnd
	}

	f.Code = code[:len(code)-1]

	return f, nil
}

type LocalEntry struct {
	Count uint32    // The total number of local variables of the given Type used in the function body
	Type  ValueType // The type of value stored by the variable
}

func readLocalEntry(r io.Reader) (LocalEntry, error) {
	l := LocalEntry{}
	var err error

	l.Count, err = leb128.ReadVarUint32(r)
	if err != nil {
		return l, err
	}

	l.Type, err = readValueType(r)
	if err != nil {
		return l, err
	}

	return l, nil
}

// SectionData describes the initial values of a module's linear memory
type SectionData struct {
	Section
	Entries []DataSegment
}

func (m *Module) readSectionData(r io.Reader) error {
	s := &SectionData{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	s.Entries = make([]DataSegment, count)

	for i := range s.Entries {
		if s.Entries[i], err = readDataSegment(r); err != nil {
			return err
		}
	}

	m.Data = s
	return err
}

// DataSegment describes a group of repeated elements that begin at a specified offset in the linear memory
type DataSegment struct {
	Index  uint32 // The index into the global linear memory space, should always be 0 in the MVP.
	Offset []byte // initializer expression for computing the offset for placing elements, should return an i32 value
	Data   []byte
}

func readDataSegment(r io.Reader) (DataSegment, error) {
	s := DataSegment{}
	var err error

	if s.Index, err = leb128.ReadVarUint32(r); err != nil {
		return s, err
	}
	if s.Offset, err = readInitExpr(r); err != nil {
		return s, err
	}

	size, err := leb128.ReadVarUint32(r)
	if err != nil {
		return s, err
	}
	s.Data, err = readBytes(r, int(size))

	return s, err
}
