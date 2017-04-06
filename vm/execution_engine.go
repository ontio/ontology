package vm

import (
	"DNA/vm/interfaces"
	"DNA/vm/utils"
	"io"
	_ "math/big"
	_ "sort"
)

const MAXSTEPS int = 1200

func NewExecutionEngine(container interfaces.IScriptContainer, crypto interfaces.ICrypto, maxSteps int, table interfaces.IScriptTable, service *InteropService) *ExecutionEngine {
	var engine ExecutionEngine

	engine.crypto = crypto
	engine.table = table

	engine.scriptContainer = container
	engine.invocationStack = utils.NewRandAccessStack()
	engine.opCount = 0

	engine.evaluationStack = utils.NewRandAccessStack()
	engine.altStack = utils.NewRandAccessStack()
	engine.state = BREAK

	engine.context = nil
	engine.opCode = 0

	engine.maxSteps = maxSteps

	if service != nil {
		engine.service = service
	}

	engine.service = NewInteropService()

	return &engine
}

type ExecutionEngine struct {
	crypto  interfaces.ICrypto
	table   interfaces.IScriptTable
	service *InteropService

	scriptContainer interfaces.IScriptContainer
	invocationStack *utils.RandomAccessStack
	opCount         int

	maxSteps int

	evaluationStack *utils.RandomAccessStack
	altStack        *utils.RandomAccessStack
	state           VMState

	context *ExecutionContext

	//current opcode
	opCode OpCode
}

func (e *ExecutionEngine) GetState() VMState {
	return e.state
}

func (e *ExecutionEngine) GetEvaluationStack() *utils.RandomAccessStack {
	return e.evaluationStack
}

func (e *ExecutionEngine) GetExecuteResult() bool {
	return AssertStackItem(e.evaluationStack.Pop()).GetBoolean()
}

func (e *ExecutionEngine) ExecutingScript() []byte {
	context := AssertExecutionContext(e.invocationStack.Peek(0))
	if context != nil {
		return context.Script
	}
	return nil
}

func (e *ExecutionEngine) CallingScript() []byte {
	if e.invocationStack.Count() > 1 {
		context := AssertExecutionContext(e.invocationStack.Peek(1))
		if context != nil {
			return context.Script
		}
		return nil
	}
	return nil
}

func (e *ExecutionEngine) EntryScript() []byte {
	context := AssertExecutionContext(e.invocationStack.Peek(e.invocationStack.Count() - 1))
	if context != nil {
		return context.Script
	}
	return nil
}

func (e *ExecutionEngine) LoadScript(script []byte, pushOnly bool) {
	e.invocationStack.Push(NewExecutionContext(script, pushOnly, nil))
}

func (e *ExecutionEngine) Execute() {
	e.state = e.state & (^BREAK)
	for {
		if e.state == FAULT || e.state == HALT || e.state == BREAK {
			break
		}
		e.StepInto()
	}
}

func (e *ExecutionEngine) StepInto() {
	if e.invocationStack.Count() == 0 {
		e.state = VMState(e.state | HALT)
	}
	if e.state&HALT == HALT || e.state&FAULT == FAULT {
		return
	}
	context := AssertExecutionContext(e.invocationStack.Pop())
	if context.InstructionPointer >= len(context.Script) {
		e.opCode = RET
	}
	for {
		opCode, err := context.OpReader.ReadByte()
		if err == io.EOF && opCode == 0 {
			return
		}
		e.opCount++
		state, err := e.ExecuteOp(OpCode(opCode), context)
		if state == VMState(HALT) {
			e.state = VMState(e.state | HALT)
			return
		}
	}
}

func (e *ExecutionEngine) ExecuteOp(opCode OpCode, context *ExecutionContext) (VMState, error) {
	if opCode > PUSH16 && opCode != RET && context.PushOnly {
		return FAULT, nil
	}
	if opCode > PUSH16 && e.opCount > e.maxSteps {
		return FAULT, nil
	}
	if opCode >= PUSHBYTES1 && opCode <= PUSHBYTES75 {
		err := pushData(e, context.OpReader.ReadBytes(int(opCode)))
		if err != nil {
			return FAULT, err
		}
		return NONE, nil
	}
	e.opCode = opCode
	e.context = context
	opExec := OpExecList[opCode]
	if opExec.Exec == nil {
		return FAULT, nil
	}
	state, err := opExec.Exec(e)
	if err != nil {
		return state, err
	}
	return NONE, nil
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
	//b := e.context.BreakPoints
	//b = append(b, position)
}

func (e *ExecutionEngine) RemoveBreakPoint(position uint) bool {
	//if e.invocationStack.Count() == 0 { return false }
	//b := e.context.BreakPoints
	return true
}
