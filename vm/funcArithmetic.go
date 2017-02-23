package vm

func op1Add(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	x := e.Stack.Pop();
	ints := BigIntOp(x.GetIntArray(),OP_1ADD)
	e.Stack.Push(NewStackItem(ints))

	return NONE,nil
}

func op1Sub(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	x := e.Stack.Pop();
	ints := BigIntOp(x.GetIntArray(),OP_1SUB)
	e.Stack.Push(NewStackItem(ints))

	return NONE,nil
}

func op2Mul(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	x := e.Stack.Pop();
	ints := BigIntOp(x.GetIntArray(),OP_2MUL)
	e.Stack.Push(NewStackItem(ints))

	return NONE,nil
}

func op2Div(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	x := e.Stack.Pop();
	ints := BigIntOp(x.GetIntArray(),OP_2DIV)
	e.Stack.Push(NewStackItem(ints))

	return NONE,nil
}

func opNegate(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	x := e.Stack.Pop();
	ints := BigIntOp(x.GetIntArray(),OP_NEGATE)
	e.Stack.Push(NewStackItem(ints))

	return NONE,nil
}

func opAbs(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	x := e.Stack.Pop();
	ints := BigIntOp(x.GetIntArray(),OP_ABS)
	e.Stack.Push(NewStackItem(ints))

	return NONE,nil
}

func opNot(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	x := e.Stack.Pop();
	bools := BoolArrayOp(x.GetBoolArray(),OP_NOT)
	e.Stack.Push(NewStackItem(bools))

	return NONE,nil
}

func op0NotEqual(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	x := e.Stack.Pop()
	bools := BigIntsComp(x.GetIntArray(),OP_0NOTEQUAL)
	e.Stack.Push(NewStackItem(bools))

	return NONE,nil
}

func opAdd(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntZip(b2, b1,OP_ADD)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opSub(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntZip(b2, b1,OP_SUB)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opMul(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntZip(b2, b1,OP_MUL)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opDiv(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntZip(b2, b1,OP_DIV)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opMod(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntZip(b2, b1,OP_MOD)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opLShift(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntZip(b2, b1,OP_LSHIFT)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opRShift(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntZip(b2, b1,OP_RSHIFT)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opBoolAnd(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetBoolArray()
	b2 := x2.GetBoolArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BoolsZip(b2, b1,OP_BOOLAND)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opBoolOr(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetBoolArray()
	b2 := x2.GetBoolArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BoolsZip(b2, b1,OP_BOOLOR)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opNumEqual(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntsMultiComp(b2, b1,OP_NUMEQUAL)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opNumNotEqual(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntsMultiComp(b2, b1,OP_NUMNOTEQUAL)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opLessThan(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntsMultiComp(b2, b1,OP_LESSTHAN)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opGreaterThan(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntsMultiComp(b2, b1,OP_GREATERTHAN)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opLessThanOrEqual(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntsMultiComp(b2, b1,OP_LESSTHANOREQUAL)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opGreaterThanOrEqual(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetIntArray()
	b2 := x2.GetIntArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BigIntsMultiComp(b2, b1,OP_GREATERTHANOREQUAL)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opMin(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetBoolArray()
	b2 := x2.GetBoolArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BoolsZip(b2, b1,OP_MIN)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opMax(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	b1 := x1.GetBoolArray()
	b2 := x2.GetBoolArray()

	if (len(b1) != len(b2)) {return FAULT,nil}
	r := BoolsZip(b2, b1,OP_MAX)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opWithIn(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 3 {return FAULT,nil}
	b := e.Stack.Pop().ToBigInt()
	a := e.Stack.Pop().ToBigInt()
	x := e.Stack.Pop().ToBigInt()

	comp := (a.Cmp(x) <= 0) && (x.Cmp(b) < 0)
	e.Stack.Push(NewStackItem(comp))

	return NONE,nil
}


