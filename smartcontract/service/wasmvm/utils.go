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
package wasmvm

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/wagon/exec"
	"github.com/ontio/wagon/validate"
	"github.com/ontio/wagon/wasm"
)

func ReadWasmMemory(proc *exec.Process, ptr uint32, len uint32) ([]byte, error) {
	if uint64(proc.MemSize()) < uint64(ptr)+uint64(len) {
		return nil, errors.New("contract create len is greater than memory size")
	}
	keybytes := make([]byte, len)
	_, err := proc.ReadAt(keybytes, int64(ptr))
	if err != nil {
		return nil, err
	}

	return keybytes, nil
}

func checkOntoWasm(m *wasm.Module) error {
	if m.Start != nil {
		return errors.New("[Validate] start section is not allowed.")
	}

	if m.Export == nil {
		return errors.New("[Validate] No export in wasm!")
	}

	if len(m.Export.Entries) != 1 {
		return errors.New("[Validate] Can only export one entry.")
	}

	entry, ok := m.Export.Entries["invoke"]
	if ok == false {
		return errors.New("[Validate] invoke entry function does not export.")
	}

	if entry.Kind != wasm.ExternalFunction {
		return errors.New("[Validate] Can only export invoke function entry.")
	}

	//get entry index
	index := int64(entry.Index)
	//get function index
	fidx := m.Function.Types[int(index)]
	//get  function type
	ftype := m.Types.Entries[int(fidx)]

	if len(ftype.ReturnTypes) > 0 {
		return errors.New("[Validate] ExecCode error! Invoke function return sig error")
	}
	if len(ftype.ParamTypes) > 0 {
		return errors.New("[Validate] ExecCode error! Invoke function param sig error")
	}

	return nil
}

func ReadWasmModule(code []byte, verify config.VerifyMethod) (*exec.CompiledModule, error) {
	m, err := wasm.ReadModule(bytes.NewReader(code), func(name string) (*wasm.Module, error) {
		switch name {
		case "env":
			return NewHostModule(), nil
		}
		return nil, fmt.Errorf("module %q unknown", name)
	})
	if err != nil {
		return nil, err
	}

	if verify != config.NoneVerifyMethod {
		err = checkOntoWasm(m)
		if err != nil {
			return nil, err
		}

		err = validate.VerifyModule(m)
		if err != nil {
			return nil, err
		}

		switch verify {
		case config.InterpVerifyMethod:
			err = validate.VerifyWasmCodeFromRust(code)
			if err != nil {
				return nil, err
			}
		case config.JitVerifyMethod:
			err := WasmjitValidate(code)
			if err != nil {
				return nil, err
			}
		}
	}

	compiled, err := exec.CompileModule(m)
	if err != nil {
		return nil, err
	}

	return compiled, nil
}
