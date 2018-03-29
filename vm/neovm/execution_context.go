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

	"github.com/Ontology/common"
	"github.com/Ontology/vm/neovm/types"
	"github.com/Ontology/vm/neovm/utils"
	vmtypes "github.com/Ontology/vm/types"
)

type ExecutionContext struct {
	Code               []byte
	OpReader           *utils.VmReader
	PushOnly           bool
	BreakPoints        []uint
	InstructionPointer int
	CodeHash           common.Address
	engine             *ExecutionEngine
}

func NewExecutionContext(engine *ExecutionEngine, code []byte, pushOnly bool, breakPoints []uint) *ExecutionContext {
	var executionContext ExecutionContext
	executionContext.Code = code
	executionContext.OpReader = utils.NewVmReader(code)
	executionContext.PushOnly = pushOnly
	executionContext.BreakPoints = breakPoints
	executionContext.InstructionPointer = 0
	executionContext.engine = engine
	return &executionContext
}

func (ec *ExecutionContext) GetInstructionPointer() int {
	return ec.OpReader.Position()
}

func (ec *ExecutionContext) SetInstructionPointer(offset int64) {
	ec.OpReader.Seek(offset, io.SeekStart)
}

func (ec *ExecutionContext) GetCodeHash() (common.Address, error) {
	empty := common.Address{}
	if ec.CodeHash == empty {
		code := &vmtypes.VmCode{
			Code:   ec.Code,
			VmType: vmtypes.NEOVM,
		}
		ec.CodeHash = code.AddressFromVmCode()
	}
	return ec.CodeHash, nil
}

func (ec *ExecutionContext) NextInstruction() OpCode {
	return OpCode(ec.Code[ec.OpReader.Position()])
}

func (ec *ExecutionContext) Clone() *ExecutionContext {
	executionContext := NewExecutionContext(ec.engine, ec.Code, ec.PushOnly, ec.BreakPoints)
	executionContext.InstructionPointer = ec.InstructionPointer
	executionContext.SetInstructionPointer(int64(ec.GetInstructionPointer()))
	return executionContext
}

func (ec *ExecutionContext) GetStackItem() types.StackItems {
	return nil
}

func (ec *ExecutionContext) GetExecutionContext() *ExecutionContext {
	return ec
}
