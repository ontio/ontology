package vm

import "DNA/vm/utils"

type ExecutionContext struct {
	Script             []byte
	OpReader           *utils.VmReader
	PushOnly           bool
	BreakPoints        []uint
	InstructionPointer int
}

func NewExecutionContext(script []byte, pushOnly bool, breakPoints []uint) *ExecutionContext {
	var executionContext ExecutionContext
	executionContext.Script = script
	executionContext.OpReader = utils.NewVmReader(script)
	executionContext.PushOnly = pushOnly
	executionContext.BreakPoints = breakPoints
	executionContext.InstructionPointer = executionContext.OpReader.Position()
	return &executionContext
}

func (ec *ExecutionContext) NextInstruction() OpCode {
	return OpCode(ec.Script[ec.OpReader.Position()])
}

func (ec *ExecutionContext) Clone() *ExecutionContext {
	return NewExecutionContext(ec.Script, ec.PushOnly, ec.BreakPoints)
}
