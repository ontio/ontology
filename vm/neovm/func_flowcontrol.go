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
	"github.com/ontio/ontology/vm/neovm/errors"
)

func opNop(e *ExecutionEngine) (VMState, error) {
	return NONE, nil
}

func opJmp(e *ExecutionEngine) (VMState, error) {
	offset := int(e.Context.OpReader.ReadInt16())

	offset = e.Context.GetInstructionPointer() + offset - 3

	if offset < 0 || offset > len(e.Context.Code) {
		return FAULT, errors.ERR_FAULT
	}
	var fValue = true

	if e.OpCode > JMP {
		if EvaluationStackCount(e) < 1 {
			return FAULT, errors.ERR_UNDER_STACK_LEN
		}
		var err error
		fValue, err = PopBoolean(e)
		if err != nil {
			return FAULT, err
		}
		if e.OpCode == JMPIFNOT {
			fValue = !fValue
		}
	}

	if fValue {
		e.Context.SetInstructionPointer(int64(offset))
	}
	return NONE, nil
}

func opCall(e *ExecutionEngine) (VMState, error) {
	context := e.Context.Clone()
	e.Context.SetInstructionPointer(int64(e.Context.GetInstructionPointer() + 2))
	e.OpCode = JMP
	e.PushContext(context)
	return opJmp(e)
}

func opRet(e *ExecutionEngine) (VMState, error) {
	e.PopContext()
	return NONE, nil
}
