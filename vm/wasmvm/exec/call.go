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

package exec

import (
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/errors"
)

func (vm *VM) doCall(compiled compiledFunction, index int64) {
	newStack := make([]uint64, compiled.maxDepth)
	locals := make([]uint64, compiled.totalLocalVars)

	for i := compiled.args - 1; i >= 0; i-- {
		locals[i] = vm.popUint64()
	}

	// save execution context
	prevCtxt := vm.ctx

	vm.ctx = context{
		stack:   newStack,
		locals:  locals,
		code:    compiled.code,
		pc:      0,
		curFunc: index,
	}

	if compiled.isEnv {
		//set the parameters and return in vm ,these will be used by inter service
		if vm.envCall == nil {
			vm.envCall = &EnvCall{}
		}

		vm.envCall.envParams = locals
		if compiled.returns {
			vm.envCall.envReturns = true
		} else {
			vm.envCall.envReturns = false
		}
		vm.envCall.envPreCtx = prevCtxt

		v, ok := vm.Services[compiled.name]
		if ok {
			rtn, err := v(vm.Engine)
			if err != nil || !rtn {
				errmsg := ""
				if err != nil{
					errmsg = err.Error()
				}
				log.Errorf("[doCall] call method :%s failed,error:%s\n", compiled.name,errmsg)
			}
		} else {
			vm.ctx = prevCtxt
			if compiled.returns {
				vm.pushUint64(0)
			}
		}

	} else {
		rtrn := vm.execCode(false, compiled)

		// restore execution context
		vm.ctx = prevCtxt

		if compiled.returns {
			vm.pushUint64(rtrn)
		}
	}

}

var (
	// ErrSignatureMismatch is the error value used while trapping the VM when
	// a signature mismatch between the table entry and the type entry is found
	// in a call_indirect operation.
	ErrSignatureMismatch =  errors.NewErr("exec: signature mismatch in call_indirect")
	// ErrUndefinedElementIndex is the error value used while trapping the VM when
	// an invalid index to the module's table space is used as an operand to
	// call_indirect
	ErrUndefinedElementIndex =  errors.NewErr("exec: undefined element index")
)

func (vm *VM) call() {
	index := vm.fetchUint32()
	vm.doCall(vm.compiledFuncs[index], int64(index))
}

func (vm *VM) callIndirect() {
	index := vm.fetchUint32()
	fnExpect := vm.module.Types.Entries[index]
	_ = vm.fetchUint32() // reserved (https://github.com/WebAssembly/design/blob/27ac254c854994103c24834a994be16f74f54186/BinaryEncoding.md#call-operators-described-here)
	tableIndex := vm.popUint32()
	if int(tableIndex) >= len(vm.module.TableIndexSpace[0]) {
		panic(ErrUndefinedElementIndex)
	}
	elemIndex := vm.module.TableIndexSpace[0][tableIndex]
	fnActual := vm.module.FunctionIndexSpace[elemIndex]

	if len(fnExpect.ParamTypes) != len(fnActual.Sig.ParamTypes) {
		panic(ErrSignatureMismatch)
	}
	if len(fnExpect.ReturnTypes) != len(fnActual.Sig.ReturnTypes) {
		panic(ErrSignatureMismatch)
	}

	for i := range fnExpect.ParamTypes {
		if fnExpect.ParamTypes[i] != fnActual.Sig.ParamTypes[i] {
			panic(ErrSignatureMismatch)
		}
	}

	for i := range fnExpect.ReturnTypes {
		if fnExpect.ReturnTypes[i] != fnActual.Sig.ReturnTypes[i] {
			panic(ErrSignatureMismatch)
		}
	}

	vm.doCall(vm.compiledFuncs[elemIndex], int64(index))
}
