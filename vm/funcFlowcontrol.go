package vm

import (
	"io"
	"time"
	"fmt"
)

func opNop(e *ExecutionEngine) (VMState,error) {
	// TODO: Sleep case

	time.Sleep( 1 * time.Millisecond )
	return NONE,nil
}

func opJmp(e *ExecutionEngine) (VMState,error) {

	offset := int(e.opReader.ReadInt16()) - 3
	offset_new := e.opReader.Position() + offset
	if offset_new < 0 || offset_new > e.opReader.Length() {
		return FAULT,ErrFault
	}
	fValue := true
	if e.opcode > OP_JMP{
		if e.Stack.Count() > 1 {
			return FAULT,ErrFault
		}
		fValue = e.Stack.Pop().ToBool()
		if e.opcode == OP_JMPIFNOT {
			fValue = !fValue
		}
		if fValue{
			e.opReader.Seek(int64(offset_new),io.SeekStart)
		}
	}

	return NONE,nil
}

func opCall(e *ExecutionEngine) (VMState,error) {

	e.Stack.Push(NewStackItem(e.opReader.Position() + 2))
	//return e.Execute(OP_JMP, e.opReader);
	return NONE,nil
}

func opRet(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	stackItem := e.Stack.Pop()
	fmt.Println( "stackItem:", stackItem )
	position := e.Stack.Pop().ToBigInt().Int64()
	fmt.Println( "position:", position )
	if position < 0 || position > int64(e.opReader.Length()){
		return  FAULT,nil
	}
	e.Stack.Push(stackItem)
	e.opReader.Seek(position,io.SeekStart)

	return NONE,nil
}

func opAppCall(e *ExecutionEngine) (VMState,error) {
/*
	if e.table == nil {return FAULT,nil}
	script_hash := e.opReader.ReadBytes(20)
	script := e.table.GetScript(script_hash)
	if script == nil {return FAULT,nil}
	if e.ExecuteProgram(script,false) {return NONE,nil}
*/
	return FAULT,nil
}

func opSysCall(e *ExecutionEngine) (VMState,error) {

	if e.service == nil {return FAULT,nil}
	success,_ := e.service.Invoke(e.opReader.ReadVarString(),e)
	if success{
		return NONE,nil
	}else{
		return FAULT,nil
	}
}

func opHaltIfNot(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	bs :=  e.Stack.Peek().GetBoolArray()
	all := true
	for _,v := range bs {
		if !v {
			all = false
			break
		}
	}

	if all {
		e.Stack.Pop()
	}else{
		return FAULT,nil
	}

	return NONE,nil
}

func opHalts(e *ExecutionEngine) (VMState,error) {

	return HALT,nil
}