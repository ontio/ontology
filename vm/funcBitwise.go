package vm

import(
	"math/big"
)

func opInvert(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	x := e.Stack.Pop()
	ints := x.GetIntArray()
	var nints []big.Int
	for _,v := range ints{
		nv := v.Not(&v)
		nints = append(nints,*nv)
	}
	e.Stack.Push(NewStackItem(nints))

	return NONE,nil
}

func opAnd(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntZip(b2, b1,OP_AND)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opOr(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntZip(b2, b1,OP_OR)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opXor(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntZip(b2, b1,OP_XOR)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opEqual(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 { return FAULT,nil }
	x2 := e.Stack.Pop();
	x1 := e.Stack.Pop();
	b1 := x1.GetBytesArray()
	b2 := x2.GetBytesArray()
	if len(b1) != len(b2) {return FAULT,nil}

	var bs []bool
	len := len(b1)
	for i:=1; i<len; i++ {
		bs = append(bs,IsEqualBytes(b1[i],b2[i]))
	}
	e.Stack.Push(NewStackItem(bs))

	return NONE,nil
}