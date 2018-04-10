// Copyright 2017 The Ontology Authors
// This file is part of the Ontology library.
//
// The Ontology library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ontology library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ontology library. If not, see <http://www.gnu.org/licenses/>.

package neovm

import (
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/vm/neovm/errors"
	"github.com/ontio/ontology/vm/neovm/interfaces"
)

func NewExecutionEngine(container interfaces.CodeContainer, crypto interfaces.Crypto, table interfaces.CodeTable, service InteropServices) *ExecutionEngine {
	var engine ExecutionEngine

	engine.crypto = crypto
	engine.table = table

	engine.codeContainer = container
	engine.invocationStack = NewRandAccessStack()
	engine.opCount = 0

	engine.evaluationStack = NewRandAccessStack()
	engine.altStack = NewRandAccessStack()
	engine.state = BREAK

	engine.context = nil
	engine.opCode = 0

	engine.service = NewInteropService()

	if service != nil {
		engine.service.MergeMap(service.GetServiceMap())
	}
	return &engine
}

type ExecutionEngine struct {
	crypto  interfaces.Crypto
	table   interfaces.CodeTable
	service *InteropService

	codeContainer   interfaces.CodeContainer
	invocationStack *RandomAccessStack
	opCount         int

	evaluationStack *RandomAccessStack
	altStack        *RandomAccessStack
	state           VMState

	context *ExecutionContext

	//current opcode
	opCode OpCode
	gas    int64
}

func (e *ExecutionEngine) Create(caller common.Address, code []byte) ([]byte, error) {
	return code, nil
}

func (e *ExecutionEngine) Call(caller common.Address, code, input []byte) ([]byte, error) {
	e.LoadCode(code, false)
	e.LoadCode(input, false)
	err := e.Execute()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (e *ExecutionEngine) GetCodeContainer() interfaces.CodeContainer {
	return e.codeContainer
}

func (e *ExecutionEngine) GetState() VMState {
	return e.state
}

func (e *ExecutionEngine) GetEvaluationStack() *RandomAccessStack {
	return e.evaluationStack
}

func (e *ExecutionEngine) GetEvaluationStackCount() int {
	return e.evaluationStack.Count()
}

func (e *ExecutionEngine) GetExecuteResult() bool {
	if e.evaluationStack.Count() < 1 {
		return false
	}
	return e.evaluationStack.Pop().GetStackItem().GetBoolean()
}

func (e *ExecutionEngine) CurrentContext() (*ExecutionContext, error) {
	if e.invocationStack.Count() < 1 {
		log.Error("[CurrentContext], Get current context fail!")
		return nil, errors.ERR_OVER_STACK_LEN
	}
	context := e.invocationStack.Peek(0).GetExecutionContext()
	if context == nil {
		return nil, errors.ERR_CURRENT_CONTEXT_NIL
	}
	return context, nil
}

func (e *ExecutionEngine) CallingContext() (*ExecutionContext, error) {
	if e.invocationStack.Count() < 2 {
		log.Error("[CallingContext], Get calling context fail!")
		return nil, errors.ERR_OVER_STACK_LEN
	}
	context := e.invocationStack.Peek(1).GetExecutionContext()
	if context == nil {
		return nil, errors.ERR_CALLING_CONTEXT_NIL
	}
	return context, nil
}

func (e *ExecutionEngine) EntryContext() (*ExecutionContext, error) {
	if e.invocationStack.Count() < 1 {
		log.Error("[EntryContext], Get entry context fail!")
		return nil, errors.ERR_OVER_STACK_LEN
	}
	context := e.invocationStack.Peek(e.invocationStack.Count() - 1).GetExecutionContext()
	if context == nil {
		return nil, errors.ERR_ENTRY_CONTEXT_NIL
	}
	return context, nil
}

func (e *ExecutionEngine) LoadCode(script []byte, pushOnly bool) {
	e.invocationStack.Push(NewExecutionContext(e, script, pushOnly, nil))
}

func (e *ExecutionEngine) Execute() error {
	e.state = e.state & (^BREAK)
	for {
		if e.state == FAULT || e.state == HALT || e.state == BREAK {
			break
		}
		err := e.StepInto()
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *ExecutionEngine) StepInto() error {
	if e.invocationStack.Count() == 0 {
		e.state = HALT
		return nil
	}
	context, err := e.CurrentContext()
	if err != nil {
		return err
	}
	var opCode OpCode

	if context.GetInstructionPointer() >= len(context.Code) {
		opCode = RET
	} else {
		o, err := context.OpReader.ReadByte()
		if err == io.EOF {
			e.state = FAULT
			return err
		}
		opCode = OpCode(o)
	}
	e.opCode = opCode
	e.context = context
	if !e.checkStackSize() {
		return errors.ERR_OVER_LIMIT_STACK
	}
	state, err := e.ExecuteOp()

	if state == HALT || state == FAULT {
		e.state = state
		return err
	}
	for _, v := range context.BreakPoints {
		if v == uint(context.InstructionPointer) {
			e.state = HALT
			return nil
		}
	}
	return nil
}

func (e *ExecutionEngine) ExecuteOp() (VMState, error) {
	if e.opCode > PUSH16 && e.opCode != RET && e.context.PushOnly {
		return FAULT, errors.ERR_BAD_VALUE
	}

	if e.opCode >= PUSHBYTES1 && e.opCode <= PUSHBYTES75 {
		PushData(e, e.context.OpReader.ReadBytes(int(e.opCode)))
		return NONE, nil
	}

	opExec := OpExecList[e.opCode]
	if opExec.Exec == nil {
		return FAULT, errors.ERR_NOT_SUPPORT_OPCODE
	}

	if opExec.Validator != nil {
		if err := opExec.Validator(e); err != nil {
			return FAULT, err
		}
	}
	return opExec.Exec(e)
}

func (e *ExecutionEngine) StepOut() {
	e.state = e.state & (^BREAK)
	c := e.invocationStack.Count()
	for {
		if e.state == FAULT || e.state == HALT || e.state == BREAK || e.invocationStack.Count() >= c {
			break
		}
		e.StepInto()
	}
}

func (e *ExecutionEngine) StepOver() {
	if e.state == FAULT || e.state == HALT {
		return
	}
	e.state = e.state & (^BREAK)
	c := e.invocationStack.Count()
	for {
		if e.state == FAULT || e.state == HALT || e.state == BREAK || e.invocationStack.Count() > c {
			break
		}
		e.StepInto()
	}
}

func (e *ExecutionEngine) AddBreakPoint(position uint) {
	e.context.BreakPoints = append(e.context.BreakPoints, position)
}

func (e *ExecutionEngine) RemoveBreakPoint(position uint) bool {
	if e.invocationStack.Count() == 0 {
		return false
	}
	bs := make([]uint, 0)
	breakPoints := e.context.BreakPoints
	for _, v := range breakPoints {
		if v != position {
			bs = append(bs, v)
		}
	}
	e.context.BreakPoints = bs
	return true
}

func (e *ExecutionEngine) checkStackSize() bool {
	size := 0
	if e.opCode < PUSH16 {
		size = 1
	} else {
		switch e.opCode {
		case DEPTH, DUP, OVER, TUCK:
			size = 1
		case UNPACK:
			item := Peek(e)
			if item == nil {
				return false
			}
			size = len(item.GetStackItem().GetArray())
		}
	}
	size += e.evaluationStack.Count() + e.altStack.Count()
	if uint32(size) > Stack_LIMIT {
		return false
	}
	return true
}
