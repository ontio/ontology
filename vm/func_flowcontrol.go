package vm

import (
	"DNA/vm/errors"
	"io"
	"time"
)

func opNop(e *ExecutionEngine) (VMState, error) {
	time.Sleep(1 * time.Millisecond)
	return NONE, nil
}

func opJmp(e *ExecutionEngine) (VMState, error) {
	offset := int(e.context.OpReader.ReadInt16())
	offset = e.context.InstructionPointer + offset - 3

	if offset < 0 || offset > len(e.context.Script) {
		return FAULT, errors.ErrFault
	}
	fValue := true
	if e.opCode > JMP {
		s := AssertStackItem(e.evaluationStack.Pop())
		fValue = s.GetBoolean()
		if e.opCode == JMPIFNOT {
			fValue = !fValue
		}
	}
	if fValue {
		e.context.InstructionPointer = offset
	}

	return NONE, nil
}

func opCall(e *ExecutionEngine) (VMState, error) {
	e.invocationStack.Push(e.context.Clone())
	e.context.InstructionPointer += 2
	opJmp(e)
	return NONE, nil
}

func opRet(e *ExecutionEngine) (VMState, error) {
	if e.invocationStack.Count() < 2 {
		return FAULT, nil
	}
	x := AssertStackItem(e.invocationStack.Pop())
	position := x.GetBigInteger().Int64()
	if position < 0 || position > int64(e.context.OpReader.Length()) {
		return FAULT, nil
	}
	e.invocationStack.Push(x)
	e.context.OpReader.Seek(position, io.SeekStart)
	return NONE, nil
}

func opAppCall(e *ExecutionEngine) (VMState, error) {
	if e.table == nil {
		return FAULT, nil
	}
	script_hash := e.context.OpReader.ReadBytes(20)
	script := e.table.GetScript(script_hash)
	if script == nil {
		return FAULT, nil
	}
	e.LoadScript(script, false)
	return NONE, nil
}

func opSysCall(e *ExecutionEngine) (VMState, error) {
	if e.service == nil {
		return FAULT, nil
	}
	success := e.service.Invoke(e.context.OpReader.ReadVarString(), e)
	if success {
		return NONE, nil
	} else {
		return FAULT, nil
	}
}
