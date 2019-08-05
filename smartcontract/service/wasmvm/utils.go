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

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/validate"
	"github.com/go-interpreter/wagon/wasm"
	"github.com/ontio/ontology/core/payload"
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

func ReadWasmModuleToVerify(dep *payload.DeployCode, verify bool) (*exec.CompiledModule, error) {
	if dep.VmType == payload.NEOVM_TYPE {
		return nil, errors.New("only wasm contract need verify")
	}

	m, err := wasm.ReadModule(bytes.NewReader(dep.Code), func(name string) (*wasm.Module, error) {
		switch name {
		case "env":
			return NewHostModule(), nil
		}
		return nil, fmt.Errorf("module %q unknown", name)
	})
	if err != nil {
		return nil, err
	}

	if m.Export == nil {
		return nil, errors.New("[Call]No export in wasm!")
	}

	if verify {
		err = validate.VerifyModule(m)
		if err != nil {
			return nil, err
		}
	}

	compiled, err := exec.CompileModule(m)
	if err != nil {
		return nil, err
	}

	return compiled, nil
}
