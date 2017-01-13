package vm

import (
	"io"
	"sort"
	"math/big"
	"fmt"
)

const MAXSTEPS  int = 1200


func NewExecutionEngine(service IApiService,crypto ICrypto,table IScriptTable,signable ISignableObject) *ExecutionEngine {
	var engine ExecutionEngine
	engine.service = service
	engine.crypto = crypto
	engine.table = table
	engine.signable = signable
	engine.Stack =  new(OpStack)
	engine.altStack = new(OpStack)
	engine.state = BREAK
	return &engine
}

type ExecutionEngine struct  {
	crypto ICrypto
	table IScriptTable
	service IApiService
	signable ISignableObject
	nOpCount int
	Stack *OpStack
	altStack *OpStack
	executingScript []byte
	state VMState
}

func (m *ExecutionEngine) ExecuteProgram(script []byte,push_only bool) (bool){

	opReader := new(VmReader)
	//opReader := bufio.NewReader(bytes.NewReader(script))
	//var opcode ScriptOp
	for {
		//Opcode, err := opReader.Read);
		//if err == io.EOF {
		//	break
		//}

		Opcode := OpCode(opReader.ReadByte())
		if push_only && Opcode > OP_16 {
			return false
		}

		state, err := m.ExecuteOp(Opcode,opReader)
		if err != nil {
			return false
		}

		if(state == FAULT) { return false}
		if(state == HALT) {return true}
	}

	return true
}

func (engine *ExecutionEngine) Dispose(){
	//TODO
}

func (engine *ExecutionEngine) Execute(){
	//TODO
}

func ExecuteStep(){
	//TODO
}


func (engine *ExecutionEngine) ExecuteOp (opcode OpCode,opReader *VmReader) (VMState,error){

	engine.nOpCount++
	if opcode > OP_16 && engine.nOpCount > MAXSTEPS {
		return HALT,nil
	}

	if opcode > OP_PUSHBYTES1 && opcode <= OP_PUSHBYTES75 {

		engine.Stack.Push(NewStackItem(opReader.ReadBytes(int(opcode))))
		return NONE,nil
	}

	switch opcode {

	//Push value
	case OP_0:
		engine.Stack.Push(NewStackItem(new([]byte)))
		break
	case OP_PUSHDATA1:
		engine.Stack.Push(NewStackItem(opReader.ReadBytes(int(opReader.ReadByte()))))
		break
	case OP_PUSHDATA2:
		engine.Stack.Push(NewStackItem(opReader.ReadBytes(int(opReader.ReadUint16()))))
		break
	case OP_PUSHDATA4:
		engine.Stack.Push(NewStackItem(opReader.ReadBytes(int(opReader.ReadInt32()))))
		break
	case OP_1NEGATE:
	case OP_1:
	case OP_2:
	case OP_3:
	case OP_4:
	case OP_5:
	case OP_6:
	case OP_7:
	case OP_8:
	case OP_9:
	case OP_10:
	case OP_11:
	case OP_12:
	case OP_13:
	case OP_14:
	case OP_15:
	case OP_16:
		engine.Stack.Push(NewStackItem(opcode - OP_1 + 1))
		break


	//Control
	case OP_NOP:
		break
	case OP_JMP:
	case OP_JMPIF:
	case OP_JMPIFNOT:
		offset := int(opReader.ReadInt16()) - 3
		offset_new := opReader.Position() + offset
		if offset_new < 0 || offset_new > opReader.Length() {
			return FAULT,nil
		}
		fValue := true
		if opcode > OP_JMP{
			if engine.Stack.Count() > 1 {
				return FAULT,nil
			}
			fValue = engine.Stack.Pop().ToBool()
			if opcode == OP_JMPIFNOT {
				fValue = !fValue
			}
			if fValue{
				opReader.Seek(int64(offset_new),io.SeekStart)
			}
		}
		break
	case OP_CALL:
		engine.Stack.Push(NewStackItem(opReader.Position() + 2))
		return engine.ExecuteOp(OP_JMP, opReader);
	case OP_RET:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		stackItem := engine.Stack.Pop()
		fmt.Println( "stackItem:", stackItem )
		position := engine.Stack.Pop().ToBigInt().Int64()
		fmt.Println( "position:", position )
		if position < 0 || position > int64(opReader.Length()){
			return  FAULT,nil
		}
		engine.Stack.Push(stackItem)
		opReader.Seek(position,io.SeekStart)
		break
	case OP_APPCALL:
		if engine.table == nil {return FAULT,nil}
		script_hash := opReader.ReadBytes(20)
		script := engine.table.GetScript(script_hash)
		if script == nil {return FAULT,nil}
		if engine.ExecuteProgram(script,false) {return NONE,nil}
		return FAULT,nil
	case OP_SYSCALL:
		if engine.service == nil {return FAULT,nil}
		success,_ := engine.service.Invoke(opReader.ReadVarString(),engine)
		if success{
			return NONE,nil
		}else{return FAULT,nil}
	case OP_HALTIFNOT:

		if engine.Stack.Count() < 1 {return FAULT,nil}
		bs :=  engine.Stack.Peek().GetBoolArray()
		all := true
		for _,v := range bs {
			if !v {
				all = false
				break
			}
		}

		if all {
			engine.Stack.Pop()
		}else{
			{return FAULT,nil}
		}
	case OP_HALT:
		return HALT,nil



	//Stack ops
	case OP_TOALTSTACK:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		engine.altStack.Push(engine.Stack.Pop())
		break
	case OP_FROMALTSTACK:
		if engine.altStack.Count() < 1 {return FAULT,nil}
		engine.Stack.Push(engine.altStack.Pop())
		break
	case OP_2DROP:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		engine.Stack.Pop()
		engine.Stack.Pop()
		break
	case OP_2DUP:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Peek()
		engine.Stack.Push(x2)
		engine.Stack.Push(x1)
		engine.Stack.Push(x2)
		break
	case OP_3DUP:
		if engine.Stack.Count() < 3 {return FAULT,nil}
		x3 := engine.Stack.Pop()
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Peek()
		engine.Stack.Push(x2)
		engine.Stack.Push(x3)
		engine.Stack.Push(x1)
		engine.Stack.Push(x2)
		engine.Stack.Push(x3)
	case OP_2OVER:
		if engine.Stack.Count() < 4 {return FAULT,nil}
		x4 := engine.Stack.Pop()
		x3 := engine.Stack.Pop()
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Peek()
		engine.Stack.Push(x2)
		engine.Stack.Push(x3)
		engine.Stack.Push(x4)
		engine.Stack.Push(x1)
		engine.Stack.Push(x2)
	case OP_2ROT:
		if engine.Stack.Count() < 6 {return FAULT,nil}
		x6 := engine.Stack.Pop();
		x5 := engine.Stack.Pop()
		x4 := engine.Stack.Pop()
		x3 := engine.Stack.Pop()
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		engine.Stack.Push(x3)
		engine.Stack.Push(x4)
		engine.Stack.Push(x5)
		engine.Stack.Push(x6)
		engine.Stack.Push(x1)
		engine.Stack.Push(x2)
	case OP_2SWAP:
		if engine.Stack.Count() < 6 {return FAULT,nil}
		x4 := engine.Stack.Pop()
		x3 := engine.Stack.Pop()
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		engine.Stack.Push(x3)
		engine.Stack.Push(x4)
		engine.Stack.Push(x1)
		engine.Stack.Push(x2)
	case OP_IFDUP:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		if engine.Stack.Peek() != nil {
			engine.Stack.Push(engine.Stack.Peek())
		}
	case OP_DEPTH:
		engine.Stack.Push(NewStackItem(engine.Stack.Count()))
	case OP_DROP:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		engine.Stack.Pop();
	case OP_DUP:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		engine.Stack.Push(engine.Stack.Peek())
	case OP_NIP:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		engine.Stack.Pop()
		engine.Stack.Push(x2)
	case OP_OVER:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 :=  engine.Stack.Pop()
		x1 :=  engine.Stack.Peek()
		engine.Stack.Push(x2)
		engine.Stack.Push(x1)
	case OP_PICK:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		n := int(engine.Stack.Pop().ToBigInt().Int64())
		if n < 0  {return FAULT,nil}
		if engine.Stack.Count() < n+1 {return FAULT,nil}
		buffer := []StackItem{}
		for i := 0; i < n; i++ {
			buffer = append(buffer,*engine.Stack.Pop())
		}
		xn := engine.Stack.Peek()
		for i := n-1; i >= 0; i-- {
			engine.Stack.Push(&buffer[i])
		}
		engine.Stack.Push(xn)
	case OP_ROLL:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		n := int(engine.Stack.Pop().ToBigInt().Int64())
		if n < 0  {return FAULT,nil}
		if n == 0  {return NONE,nil}
		if engine.Stack.Count() < n+1 {return FAULT,nil}
		buffer := []StackItem{}
		for i := 0; i < n; i++ {
			buffer = append(buffer,*engine.Stack.Pop())
		}
		xn := engine.Stack.Pop()
		for i := n-1; i >= 0; i-- {
			engine.Stack.Push(&buffer[i])
		}
		engine.Stack.Push(xn)
	case OP_ROT:
		if engine.Stack.Count() < 3 {return FAULT,nil}
		x3 := engine.Stack.Pop()
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		engine.Stack.Push(x2)
		engine.Stack.Push(x3)
		engine.Stack.Push(x1)
	case OP_SWAP:
		if engine.Stack.Count() < 2 { return FAULT,nil }
		x2 := engine.Stack.Pop();
		x1 := engine.Stack.Pop();
		engine.Stack.Push(x2);
		engine.Stack.Push(x1);
	case OP_TUCK:
		if engine.Stack.Count() < 2 { return FAULT,nil }
		x2 := engine.Stack.Pop();
		x1 := engine.Stack.Pop();
		engine.Stack.Push(x2);
		engine.Stack.Push(x1);
		engine.Stack.Push(x2);
	case OP_CAT:
		if engine.Stack.Count() < 2 { return FAULT,nil }
		x2 := engine.Stack.Pop();
		x1 := engine.Stack.Pop();
		b1 := x1.GetBytesArray()
		b2 := x2.GetBytesArray()
		if (len(b1) != len(b2)) {return FAULT,nil}

		r := ByteArrZip(b1,b2,OP_CONCAT)
		engine.Stack.Push(NewStackItem(r))
	case OP_SUBSTR:
		if engine.Stack.Count() < 3 {return FAULT,nil}
		count := int(engine.Stack.Pop().ToBigInt().Int64())
		if count < 0  {return FAULT,nil}
		index := int(engine.Stack.Pop().ToBigInt().Int64())
		if index < 0  {return FAULT,nil}
		x := engine.Stack.Pop()
		s := x.GetBytesArray()

		for _,b := range s{
			//p.Skip(index).Take(count) : need test
			b = b[index + count :]
		}
		engine.Stack.Push(NewStackItem(s))
	case OP_LEFT:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		count := int(engine.Stack.Pop().ToBigInt().Int64())
		if count < 0  {return FAULT,nil}
		x := engine.Stack.Pop()
		s := x.GetBytesArray()
		for _,b := range s{
			b = b[count:]
		}
		engine.Stack.Push(NewStackItem(s))
	case OP_RIGHT:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		count := int(engine.Stack.Pop().ToBigInt().Int64())
		if count < 0  {return FAULT,nil}
		x := engine.Stack.Pop()
		s := x.GetBytesArray()
		for _,b := range s{
			len := len(b)
			if len < count {return FAULT,nil}
			b = b[0:len - count]

		}
		engine.Stack.Push(NewStackItem(s))
	case OP_SIZE:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		x := engine.Stack.Peek()
		s := x.GetBytesArray()
		r := []int{}
		for _,b := range s{
			r = append(r,len(b))
		}
		engine.Stack.Push(NewStackItem(r))


	//Bitwiase logic
	case OP_INVERT:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		x := engine.Stack.Pop()
		ints := x.GetIntArray()
		var nints []big.Int
		for _,v := range ints{
			nv := v.Not(&v)
			nints = append(nints,*nv)
		}
		engine.Stack.Push(NewStackItem(nints))
	case OP_AND:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntZip(b2, b1,OP_AND)
		engine.Stack.Push(NewStackItem(r))
	case OP_OR:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntZip(b2, b1,OP_OR)
		engine.Stack.Push(NewStackItem(r))
	case OP_XOR:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntZip(b2, b1,OP_XOR)
		engine.Stack.Push(NewStackItem(r))
	case OP_EQUAL:
		if engine.Stack.Count() < 2 { return FAULT,nil }
		x2 := engine.Stack.Pop();
		x1 := engine.Stack.Pop();
		b1 := x1.GetBytesArray()
		b2 := x2.GetBytesArray()
		if len(b1) != len(b2) {return FAULT,nil}

		var bs []bool
		len := len(b1)
		for i:=1; i<len; i++ {
			bs = append(bs,IsEqualBytes(b1[i],b2[i]))
		}
		engine.Stack.Push(NewStackItem(bs))


	//Numeric
	case OP_1ADD :
		if engine.Stack.Count() < 1 {return FAULT,nil}
		x := engine.Stack.Pop();
		ints := BigIntOp(x.GetIntArray(),OP_1ADD)
		engine.Stack.Push(NewStackItem(ints))
	case OP_1SUB:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		x := engine.Stack.Pop();
		ints := BigIntOp(x.GetIntArray(),OP_1SUB)
		engine.Stack.Push(NewStackItem(ints))
	case OP_2MUL:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		x := engine.Stack.Pop();
		ints := BigIntOp(x.GetIntArray(),OP_2MUL)
		engine.Stack.Push(NewStackItem(ints))
	case OP_2DIV:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		x := engine.Stack.Pop();
		ints := BigIntOp(x.GetIntArray(),OP_2DIV)
		engine.Stack.Push(NewStackItem(ints))
	case OP_NEGATE:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		x := engine.Stack.Pop();
		ints := BigIntOp(x.GetIntArray(),OP_NEGATE)
		engine.Stack.Push(NewStackItem(ints))
	case OP_ABS:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		x := engine.Stack.Pop();
		ints := BigIntOp(x.GetIntArray(),OP_ABS)
		engine.Stack.Push(NewStackItem(ints))
	case OP_NOT:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		x := engine.Stack.Pop();
		bools := BoolArrayOp(x.GetBoolArray(),OP_NOT)
		engine.Stack.Push(NewStackItem(bools))
	case OP_0NOTEQUAL:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		x := engine.Stack.Pop()
		bools := BigIntsComp(x.GetIntArray(),OP_0NOTEQUAL)
		engine.Stack.Push(NewStackItem(bools))
	case OP_ADD:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntZip(b2, b1,OP_ADD)
		engine.Stack.Push(NewStackItem(r))
	case OP_SUB:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntZip(b2, b1,OP_SUB)
		engine.Stack.Push(NewStackItem(r))
	case OP_MUL:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntZip(b2, b1,OP_MUL)
		engine.Stack.Push(NewStackItem(r))
	case OP_DIV:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntZip(b2, b1,OP_DIV)
		engine.Stack.Push(NewStackItem(r))
	case OP_MOD:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntZip(b2, b1,OP_MOD)
		engine.Stack.Push(NewStackItem(r))
	case OP_LSHIFT:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntZip(b2, b1,OP_LSHIFT)
		engine.Stack.Push(NewStackItem(r))
	case OP_RSHIFT:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntZip(b2, b1,OP_RSHIFT)
		engine.Stack.Push(NewStackItem(r))
	case OP_BOOLAND:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetBoolArray()
		b2 := x2.GetBoolArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BoolsZip(b2, b1,OP_BOOLAND)
		engine.Stack.Push(NewStackItem(r))
	case OP_BOOLOR:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetBoolArray()
		b2 := x2.GetBoolArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BoolsZip(b2, b1,OP_BOOLOR)
		engine.Stack.Push(NewStackItem(r))
	case OP_NUMEQUAL:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntsMultiComp(b2, b1,OP_NUMEQUAL)
		engine.Stack.Push(NewStackItem(r))
	case OP_NUMNOTEQUAL:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntsMultiComp(b2, b1,OP_NUMNOTEQUAL)
		engine.Stack.Push(NewStackItem(r))
	case OP_LESSTHAN:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntsMultiComp(b2, b1,OP_LESSTHAN)
		engine.Stack.Push(NewStackItem(r))
	case OP_GREATERTHAN:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntsMultiComp(b2, b1,OP_GREATERTHAN)
		engine.Stack.Push(NewStackItem(r))
	case OP_LESSTHANOREQUAL:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntsMultiComp(b2, b1,OP_LESSTHANOREQUAL)
		engine.Stack.Push(NewStackItem(r))
	case OP_GREATERTHANOREQUAL:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetIntArray()
		b2 := x2.GetIntArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BigIntsMultiComp(b2, b1,OP_GREATERTHANOREQUAL)
		engine.Stack.Push(NewStackItem(r))
	case OP_MIN:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetBoolArray()
		b2 := x2.GetBoolArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BoolsZip(b2, b1,OP_MIN)
		engine.Stack.Push(NewStackItem(r))
	case OP_MAX:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		b1 := x1.GetBoolArray()
		b2 := x2.GetBoolArray()

		if (len(b1) != len(b2)) {return FAULT,nil}
		r := BoolsZip(b2, b1,OP_MAX)
		engine.Stack.Push(NewStackItem(r))
	case OP_WITHIN:
		if engine.Stack.Count() < 3 {return FAULT,nil}
		b := engine.Stack.Pop().ToBigInt()
		a := engine.Stack.Pop().ToBigInt()
		x := engine.Stack.Pop().ToBigInt()

		comp := (a.Cmp(x) <= 0) && (x.Cmp(b) < 0)
		engine.Stack.Push(NewStackItem(comp))



	//Crypto
	case OP_SHA1:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		x := engine.Stack.Pop()
		engine.Stack.Push(NewStackItem(x.Hash(OP_SHA1,engine)))
	case OP_SHA256:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		x := engine.Stack.Pop()
		engine.Stack.Push(NewStackItem(x.Hash(OP_SHA256,engine)))
	case OP_HASH160:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		x := engine.Stack.Pop()
		engine.Stack.Push(NewStackItem(x.Hash(OP_HASH160,engine)))
	case OP_HASH256:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		x := engine.Stack.Pop()
		engine.Stack.Push(NewStackItem(x.Hash(OP_HASH256,engine)))
	case OP_CHECKSIG:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		pubkey := engine.Stack.Pop().GetBytes()
		signature := engine.Stack.Pop().GetBytes()
		ver := engine.crypto.VerifySignature(engine.signable.GetMessage(),signature,pubkey)
		engine.Stack.Push(NewStackItem(ver))
	case OP_CHECKMULTISIG:
		if engine.Stack.Count() < 4 {return FAULT,nil}
		n :=  int(engine.Stack.Pop().ToBigInt().Int64())
		if n < 1 {return FAULT,nil}
		if engine.Stack.Count() < n+2 {return FAULT,nil}
		engine.nOpCount += n
		if engine.nOpCount > MAXSTEPS {return FAULT,nil}

		pubkeys := make([][] byte,n)
		for i := 0; i < n; i++ {pubkeys[i] = engine.Stack.Pop().GetBytes()}

		m := int(engine.Stack.Pop().ToBigInt().Int64())
		if m < 1 || m > n {return FAULT,nil}
		if engine.Stack.Count() < m {return FAULT,nil}

		signatures := make([][] byte,m)
		for i := 0; i < n; i++ {signatures[i] = engine.Stack.Pop().GetBytes()}

		message := engine.signable.GetMessage()
		fSuccess := true

		for i,j := 0,0; fSuccess && i<m && j<n; {
			if(engine.crypto.VerifySignature(message,signatures[i],pubkeys[j])) {i++}
			j++
			if m - i > n - j {fSuccess = false}
		}
		engine.Stack.Push(NewStackItem(fSuccess))


	//Array
	case OP_ARRAYSIZE:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		arr :=  engine.Stack.Pop()
		engine.Stack.Push(NewStackItem(arr.Count()))
	case OP_PACK:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		c :=  int(engine.Stack.Pop().ToBigInt().Int64())
		if engine.Stack.Count() < c {return FAULT,nil}
		arr := []StackItem{}

		for{
			if(c > 0) {arr = append(arr, *engine.Stack.Pop())}
			c--
		}
		engine.Stack.Push(NewStackItem(arr))
	case OP_UNPACK:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		arr :=  engine.Stack.Pop().GetArray()
		for _,si := range arr {
			engine.Stack.Push(NewStackItem((si)))
		}
		engine.Stack.Push(NewStackItem(len(arr)))
	case OP_DISTINCT:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		engine.Stack.Push(engine.Stack.Pop().Distinct())
	case OP_SORT:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		var biSorter BigIntSorter
		biSorter = engine.Stack.Pop().GetIntArray()

		sort.Sort(biSorter)
		engine.Stack.Push(NewStackItem((biSorter)))
	case OP_REVERSE:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		arr := engine.Stack.Pop().Reverse()
		engine.Stack.Push(&arr)
	case OP_CONCAT:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		c :=  int(engine.Stack.Pop().ToBigInt().Int64())
		if c == 1 {return FAULT,nil}
		if engine.Stack.Count() < c {return FAULT,nil}
		item := engine.Stack.Pop()
		c--
		for {
			c--
			if(c>0){
				item =  engine.Stack.Pop().Concat(item)
			}
		}

		engine.Stack.Push(item)
	case OP_UNION:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		c :=  int(engine.Stack.Pop().ToBigInt().Int64())
		if c == 1 {return FAULT,nil}
		if engine.Stack.Count() < c {return FAULT,nil}
		item := engine.Stack.Pop()
		c--
		for {
			c--
			if(c>0){
				item =  engine.Stack.Pop().Concat(item)
			}
		}

		engine.Stack.Push(item.Distinct())
	case OP_INTERSECT:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		c :=  int(engine.Stack.Pop().ToBigInt().Int64())
		if c == 1 {return FAULT,nil}
		if engine.Stack.Count() < c {return FAULT,nil}
		item := engine.Stack.Pop()
		c--
		for {
			c--
			if(c>0){
				item =  engine.Stack.Pop().Intersect(item)
			}
		}

		engine.Stack.Push(item)
	case OP_EXCEPT:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		x2 := engine.Stack.Pop()
		x1 := engine.Stack.Pop()
		engine.Stack.Push(x1.Except(x2))
	case OP_TAKE:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		count :=  int(engine.Stack.Pop().ToBigInt().Int64())
		engine.Stack.Push(engine.Stack.Pop().Take(count))
	case OP_SKIP:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		count :=  int(engine.Stack.Pop().ToBigInt().Int64())
		engine.Stack.Push(engine.Stack.Pop().Take(count))
	case OP_PICKITEM:
		if engine.Stack.Count() < 2 {return FAULT,nil}
		count :=  int(engine.Stack.Pop().ToBigInt().Int64())
		engine.Stack.Push(engine.Stack.Pop().ElementAt(count))
	case OP_ALL:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		bs := engine.Stack.Pop().GetBoolArray()
		all := true
		for _,b := range bs {
			if !b {all = false; break}
		}
		engine.Stack.Push(NewStackItem(all))
	case OP_ANY:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		bs := engine.Stack.Pop().GetBoolArray()
		any := false
		for _,b := range bs {
			if b {any = true; break}
		}
		engine.Stack.Push(NewStackItem(any))
	case OP_SUM:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		is := engine.Stack.Pop().GetIntArray()
		sum := SumBigInt(is)
		engine.Stack.Push(NewStackItem(sum))
	case OP_AVERAGE:
		if engine.Stack.Count() < 1 {return FAULT,nil}
		arr := engine.Stack.Pop()
		arrCount := arr.Count()
		if arrCount < 1 {return FAULT,nil}
		is := engine.Stack.Pop().GetIntArray()
		sum := SumBigInt(is)
		avg := sum.Div(&sum,big.NewInt(int64(arrCount)))
		engine.Stack.Push(NewStackItem(*avg))
	case OP_MAXITEM:
		if engine.Stack.Count() < 1 {return  FAULT,nil}
		engine.Stack.Push(NewStackItem(MinBigInt(engine.Stack.Pop().GetIntArray())))
	case OP_MINITEM:
		if engine.Stack.Count() < 1 {return  FAULT,nil}
		engine.Stack.Push(NewStackItem(MinBigInt(engine.Stack.Pop().GetIntArray())))
	default:
		return FAULT,nil
	}

	return NONE,nil
}

func LoadScript(script []byte)(){

}

func RemoveBreakPoint(posiiton uint) bool{
	//TODO
	return false
}