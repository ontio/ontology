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
)

const (
	ratio = 100000
	gasFree = 10 * 100000000;
)

func NewExecutionEngine(container interfaces.ICodeContainer, crypto interfaces.ICrypto, table interfaces.ICodeTable, service IInteropService, gas common.Fixed64) *ExecutionEngine {
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
	engine.gas = gasFree + gas.GetData()
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
	fmt.Println("ExecutionEngine Call:", code, input)
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
	return e.evaluationStack.Pop().GetStackItem().GetBoolean()
}

func (e *ExecutionEngine) ExecutingCode() []byte {
	context := e.invocationStack.Peek(0).GetExecutionContext()
	if context != nil {
		return context.Code
	}
	return nil
}

func (e *ExecutionEngine) CurrentContext() *ExecutionContext {
	context := e.invocationStack.Peek(0).GetExecutionContext()
	if context != nil {
		return context
	}
	return nil
}

func (e *ExecutionEngine) CallingContext() *ExecutionContext {
	context := e.invocationStack.Peek(1).GetExecutionContext()
	if context != nil {
		return context
	}
	return nil
}

func (e *ExecutionEngine) EntryContext() *ExecutionContext {
	context := e.invocationStack.Peek(e.invocationStack.Count() - 1).GetExecutionContext()
	if context != nil {
		return context
	}
	return nil
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
	context := e.CurrentContext()

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
	fmt.Println("e.opCode", e.opCode)
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
		fmt.Print(" ", e.evaluationStack.Peek(i).GetStackItem(), reflect.TypeOf(e.evaluationStack.Peek(i).GetStackItem()))
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

func (e *ExecutionEngine) getPrice() int64 {
	switch e.opCode {
	case NOP:
		return 0
	case APPCALL, TAILCALL:
		return 10
	case SYSCALL:
		return e.getPriceForSysCall()
	case SHA1, SHA256:
		return 10
	case HASH160, HASH256:
		return 20
	case CHECKSIG:
		return 100
	case CHECKMULTISIG:
		if e.evaluationStack.Count() == 0 {
			return 1
		}
		n := Peek(e).GetStackItem().GetBigInteger().Int64()
		if n < 1 {
			return 1
		}
		return int64(100 * n)
	default:
		return 1
	}
}

func (e *ExecutionEngine) getPriceForSysCall() int64 {
	context := e.context
	i := context.GetInstructionPointer() - 1
	c := len(context.Code)
	if i >= c - 3 {
		return 1
	}
	l := int(context.Code[i + 1])
	if i >= c - l - 2 {
		return 1
	}
	name := string(context.Code[i + 2:l])
	switch name {
	case "Neo.Blockchain.GetHeader":
		return 100
	case "Neo.Blockchain.GetBlock":
		return 200
	case "Neo.Blockchain.GetTransaction":
		return 100
	case "Neo.Blockchain.GetAccount":
		return 100
	case "Neo.Blockchain.RegisterValidator":
		return 1000 * 100000000 / ratio;
	case "Neo.Blockchain.GetValidators":
		return 200
	case "Neo.Blockchain.CreateAsset":
		return 5000 * 100000000 / ratio
	case "Neo.Blockchain.GetAsset":
		return 100
	case "Neo.Blockchain.CreateContract":
		return 500 * 100000000 / ratio
	case "Neo.Blockchain.GetContract":
		return 100
	case "Neo.Transaction.GetReferences":
		return 200
	case "Neo.Asset.Renew":
		return Peek(e).GetStackItem().GetBigInteger().Int64() * 5000 * 100000000 / ratio
	case "Neo.Storage.Get":
		return 100
	case "Neo.Storage.Put":
		return 1000
	case "Neo.Storage.Delete":
		return 100
	default:
		return 1
	}
}
