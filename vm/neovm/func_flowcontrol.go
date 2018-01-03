package neovm

import (
	. "github.com/Ontology/vm/neovm/errors"
	"github.com/Ontology/common/log"
	"fmt"
)

func opNop(e *ExecutionEngine) (VMState, error) {
	return NONE, nil
}

func opJmp(e *ExecutionEngine) (VMState, error) {
	offset := int(e.context.OpReader.ReadInt16())

	offset = e.context.GetInstructionPointer() + offset - 3

	if offset < 0 || offset > len(e.context.Code) {
		log.Error(fmt.Sprintf("[opJmp] offset:%v > e.contex.Code len:%v error", offset, len(e.context.Code)))
		return FAULT, ErrFault
	}
	var fValue = true

	if e.opCode > JMP {
		if EvaluationStackCount(e) < 1 {
			log.Error(fmt.Sprintf("[opJmp] stack count:%v > 1 error", EvaluationStackCount(e)))
			return FAULT, ErrUnderStackLen
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
