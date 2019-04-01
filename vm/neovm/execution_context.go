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
	"io"

	"github.com/ontio/ontology/vm/neovm/utils"
)

type ExecutionContext struct {
	Code               []byte
	OpReader           *utils.VmReader
	InstructionPointer int
}

func NewExecutionContext(code []byte) *ExecutionContext {
	var executionContext ExecutionContext
	executionContext.Code = code
	executionContext.OpReader = utils.NewVmReader(code)
	executionContext.InstructionPointer = 0
	return &executionContext
}

func (ec *ExecutionContext) GetInstructionPointer() int {
	return ec.OpReader.Position()
}

func (ec *ExecutionContext) SetInstructionPointer(offset int64) error {
	_, err := ec.OpReader.Seek(offset, io.SeekStart)
	return err
}

func (ec *ExecutionContext) NextInstruction() OpCode {
	return OpCode(ec.Code[ec.OpReader.Position()])
}

func (self *ExecutionContext) ReadOpCode() (val OpCode, eof bool) {
	code, err := self.OpReader.ReadByte()
	if err != nil {
		eof = true
		return
	}
	val = OpCode(code)
	return val, false
}

func (ec *ExecutionContext) Clone() *ExecutionContext {
	executionContext := NewExecutionContext(ec.Code)
	executionContext.InstructionPointer = ec.InstructionPointer
	_ = executionContext.SetInstructionPointer(int64(ec.GetInstructionPointer()))
	return executionContext
}
