package neovm

import (
	"github.com/Ontology/vm/neovm/interfaces"
	"io"
	_ "math/big"
	_ "sort"
	. "github.com/Ontology/vm/neovm/errors"
	"github.com/Ontology/common"
	"fmt"
	"reflect"
	"github.com/Ontology/vm/neovm/types"
	"github.com/Ontology/common/log"
)

func NewExecutionEngine(container interfaces.ICodeContainer, crypto interfaces.ICrypto, table interfaces.ICodeTable, service IInteropService) *ExecutionEngine {
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
	crypto          interfaces.ICrypto
	table           interfaces.ICodeTable
	service         *InteropService

	codeContainer   interfaces.ICodeContainer
	invocationStack *RandomAccessStack
	opCount         int

	evaluationStack *RandomAccessStack
	altStack        *RandomAccessStack
	state           VMState

	context         *ExecutionContext

	//current opcode
	opCode          OpCode
	gas             int64
}

func (e *ExecutionEngine) Create(caller common.Uint160, code []byte) ([]byte, error) {
	return code, nil
}

func (e *ExecutionEngine) Call(caller common.Uint160, code, input []byte) ([]byte, error) {
	e.LoadCode(code, false)
	e.LoadCode(input, false)
	err := e.Execute()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (e *ExecutionEngine) GetCodeContainer() interfaces.ICodeContainer {
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

func (e *ExecutionEngine) ExecutingCode() ([]byte, error) {
	if e.invocationStack.Count() < 1 {
		log.Error("[ExecutingCode], Get execution context fail!")
		return nil, ErrOverStackLen
	}
	context := e.invocationStack.Peek(0).GetExecutionContext()
	if context == nil {
		return nil, ErrExecutionContextNil
	}
	return context.Code, nil
}

func (e *ExecutionEngine) CurrentContext() (*ExecutionContext, error) {
	if e.invocationStack.Count() < 1 {
		log.Error("[CurrentContext], Get current context fail!")
		return nil, ErrOverStackLen
	}
	context := e.invocationStack.Peek(0).GetExecutionContext()
	if context == nil {
		return nil, ErrCurrentContextNil
	}
	return context, nil
}

func (e *ExecutionEngine) CallingContext() (*ExecutionContext, error) {
	if e.invocationStack.Count() < 2 {
		log.Error("[CallingContext], Get calling context fail!")
		return nil, ErrOverStackLen
	}
	context := e.invocationStack.Peek(1).GetExecutionContext()
	if context == nil {
		return nil, ErrCallingContextNil
	}
	return context, nil
}

func (e *ExecutionEngine) EntryContext() (*ExecutionContext, error) {
	if e.invocationStack.Count() < 3 {
		log.Error("[EntryContext], Get entry context fail!")
		return nil, ErrOverStackLen
	}
	context := e.invocationStack.Peek(e.invocationStack.Count() - 1).GetExecutionContext()
	if context == nil {
		return nil, ErrEntryContextNil
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
		return ErrOverLimitStack
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
		return FAULT, ErrBadValue
	}

	if e.opCode >= PUSHBYTES1 && e.opCode <= PUSHBYTES75 {
		PushData(e, e.context.OpReader.ReadBytes(int(e.opCode)))
		return NONE, nil
	}

	opExec := OpExecList[e.opCode]
	if opExec.Exec == nil {
		return FAULT, ErrNotSupportOpCode
	}
	fmt.Println("op:", opExec.Name)
	s := e.evaluationStack.Count()
	for i := 0; i<s;i++ {
		item := e.evaluationStack.Peek(i).GetStackItem()
		fmt.Print("type:", reflect.TypeOf(item))
		fmt.Print(" ")
		switch v := item.(type) {
		case *types.Integer:
			fmt.Print("value:", v.GetBigInteger())
		case  *types.Boolean:
			fmt.Print("value:", v.GetBoolean())
		case *types.ByteArray:
			fmt.Print("value:", v.GetByteArray())
		case *types.InteropInterface:
			fmt.Print("value:", v.GetInterface())
		case *types.Array:
			fmt.Print("value:", v.GetArray())
		}
		fmt.Print(" ")
	}
	fmt.Println()
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
	if uint32(size) > StackLimit {
		return false
	}
	return true
}
