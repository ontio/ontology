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

package neovm

import (
	"fmt"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/vm/neovm/errors"
)

func opNop(e *ExecutionEngine) (VMState, error) {
	return NONE, nil
}

func opJmp(e *ExecutionEngine) (VMState, error) {
	offset := int(e.context.OpReader.ReadInt16())

	offset = e.context.GetInstructionPointer() + offset - 3

	if offset < 0 || offset > len(e.context.Code) {
		log.Error(fmt.Sprintf("[opJmp] offset:%v > e.contex.Code len:%v error", offset, len(e.context.Code)))
		return FAULT, errors.ERR_FAULT
	}
	var fValue = true

	if e.opCode > JMP {
		if EvaluationStackCount(e) < 1 {
			log.Error(fmt.Sprintf("[opJmp] stack count:%v > 1 error", EvaluationStackCount(e)))
			return FAULT, errors.ERR_UNDER_STACK_LEN
		}
		fValue = PopBoolean(e)
		if e.opCode == JMPIFNOT {
			fValue = !fValue
		}
	}

	if fValue {
		e.context.SetInstructionPointer(int64(offset))
	}
	return NONE, nil
}

func opCall(e *ExecutionEngine) (VMState, error) {

	e.invocationStack.Push(e.context.Clone())
	e.context.SetInstructionPointer(int64(e.context.GetInstructionPointer() + 2))
	e.opCode = JMP
	context, err := e.CurrentContext()
	if err != nil {
		return FAULT, err
	}
	e.context = context
	return opJmp(e)
}

func opRet(e *ExecutionEngine) (VMState, error) {
	e.invocationStack.Pop()
	return NONE, nil
}

func opAppCall(e *ExecutionEngine) (VMState, error) {
	codeHash := e.context.OpReader.ReadBytes(20)
	if len(codeHash) == 0 {
		codeHash = PopByteArray(e)
	}

	code, err := e.table.GetCode(codeHash)
	if code == nil {
		return FAULT, err
	}

	if e.opCode == TAILCALL {
		e.invocationStack.Pop()
	}
	e.LoadCode(code, false)
	return NONE, nil
}

func opSysCall(e *ExecutionEngine) (VMState, error) {
	s := e.context.OpReader.ReadVarString()

	success, err := e.service.Invoke(s, e)
	if success {
		return NONE, nil
	} else {
		return FAULT, err
	}
}
