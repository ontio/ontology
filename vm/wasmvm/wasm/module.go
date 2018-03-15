// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wasm

import (
	"errors"
	"io"
	"github.com/Ontology/vm/wasmvm/wasm/internal/readpos"
)

var ErrInvalidMagic = errors.New("wasm: Invalid magic number")

const (
	Magic   uint32 = 0x6d736100
	Version uint32 = 0x1
)

// Function represents an entry in the function index space of a module.
type Function struct {
	IsEnvFunc bool   //is the function is a env func
	Name      string //function name
	Sig       *FunctionSig
	Body      *FunctionBody
}

// Module represents a parsed WebAssembly module:
// http://webassembly.org/docs/modules/
type Module struct {
	Version uint32

	Types    *SectionTypes
	Import   *SectionImports
	Function *SectionFunctions
	Table    *SectionTables
	Memory   *SectionMemories
	Global   *SectionGlobals
	Export   *SectionExports
	Start    *SectionStartFunction
	Elements *SectionElements
	Code     *SectionCode
	Data     *SectionData

	// The function index space of the module
	FunctionIndexSpace []Function
	GlobalIndexSpace   []GlobalEntry
	// function indices into the global function space
	// the limit of each table is its capacity (cap)
	TableIndexSpace        [][]uint32
	LinearMemoryIndexSpace [][]byte

	Other []Section // Other holds the custom sections if any

	imports struct {
		Funcs    []uint32
		Globals  int
		Tables   int
		Memories int
	}
}

// ResolveFunc is a function that takes a module name and
// returns a valid resolved module.
type ResolveFunc func(name string) (*Module, error)

// ReadModule reads a module from the reader r. resolvePath must take a string
// and a return a reader to the module pointed to by the string.
func ReadModule(r io.Reader, resolvePath ResolveFunc) (*Module, error) {
	reader := &readpos.ReadPos{
		R:      r,
		CurPos: 0,
	}
	m := &Module{}
	magic, err := readU32(reader)
	if err != nil {
		return nil, err
	}
	if magic != Magic {
		return nil, ErrInvalidMagic
	}
	if m.Version, err = readU32(reader); err != nil {
		return nil, err
	}

	for {
		done, err := m.readSection(reader)
		if err != nil {
			return nil, err
		} else if done {
			break
		}
	}

	m.LinearMemoryIndexSpace = make([][]byte, 1)
	if m.Table != nil {
		m.TableIndexSpace = make([][]uint32, int(len(m.Table.Entries)))
	}

	if m.Import != nil && resolvePath != nil {
		err := m.resolveImports(resolvePath)
		if err != nil {
			return nil, err
		}
	}

	for _, fn := range []func() error{
		m.populateGlobals,
		m.populateFunctions,
		m.populateTables,
		m.populateLinearMemory,
	} {
		if err := fn(); err != nil {
			return nil, err
		}

	}

	logger.Printf("There are %d entries in the function index space.", len(m.FunctionIndexSpace))
	return m, nil
}
