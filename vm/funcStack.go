package vm

func opToAltStack(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	e.altStack.Push(e.Stack.Pop())

	return NONE,nil
}

func opFromAltStack(e *ExecutionEngine) (VMState,error) {

	if e.altStack.Count() < 1 {return FAULT,nil}
	e.Stack.Push(e.altStack.Pop())

	return NONE,nil
}

func op2Drop(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	e.Stack.Pop()
	e.Stack.Pop()

	return NONE,nil
}

func op2Dup(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	x1 := e.Stack.Peek()
	e.Stack.Push(x2)
	e.Stack.Push(x1)
	e.Stack.Push(x2)

	return NONE,nil
}

func op3Dup(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 3 {return FAULT,nil}
	x3 := e.Stack.Pop()
	x2 := e.Stack.Pop()
	x1 := e.Stack.Peek()
	e.Stack.Push(x2)
	e.Stack.Push(x3)
	e.Stack.Push(x1)
	e.Stack.Push(x2)
	e.Stack.Push(x3)

	return NONE,nil
}

func op2Over(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 4 {return FAULT,nil}
	x4 := e.Stack.Pop()
	x3 := e.Stack.Pop()
	x2 := e.Stack.Pop()
	x1 := e.Stack.Peek()
	e.Stack.Push(x2)
	e.Stack.Push(x3)
	e.Stack.Push(x4)
	e.Stack.Push(x1)
	e.Stack.Push(x2)

	return NONE,nil
}

func op2Rot(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 6 {return FAULT,nil}
	x6 := e.Stack.Pop();
	x5 := e.Stack.Pop()
	x4 := e.Stack.Pop()
	x3 := e.Stack.Pop()
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	e.Stack.Push(x3)
	e.Stack.Push(x4)
	e.Stack.Push(x5)
	e.Stack.Push(x6)
	e.Stack.Push(x1)
	e.Stack.Push(x2)

	return NONE,nil
}

func op2Swap(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 6 {return FAULT,nil}
	x4 := e.Stack.Pop()
	x3 := e.Stack.Pop()
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	e.Stack.Push(x3)
	e.Stack.Push(x4)
	e.Stack.Push(x1)
	e.Stack.Push(x2)

	return NONE,nil
}

func opIfDup(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	if e.Stack.Peek() != nil {
		e.Stack.Push(e.Stack.Peek())
	}

	return NONE,nil
}

func opDepth(e *ExecutionEngine) (VMState,error) {

	e.Stack.Push(NewStackItem(e.Stack.Count()))

	return NONE,nil
}

func opDrop(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	e.Stack.Pop();

	return NONE,nil
}

func opDup(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	e.Stack.Push(e.Stack.Peek())

	return NONE,nil
}

func opNip(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 := e.Stack.Pop()
	e.Stack.Pop()
	e.Stack.Push(x2)

	return NONE,nil
}

func opOver(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	x2 :=  e.Stack.Pop()
	x1 :=  e.Stack.Peek()
	e.Stack.Push(x2)
	e.Stack.Push(x1)

	return NONE,nil
}

func opPick(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	n := int(e.Stack.Pop().ToBigInt().Int64())
	if n < 0  {return FAULT,nil}
	if e.Stack.Count() < n+1 {return FAULT,nil}
	buffer := []StackItem{}
	for i := 0; i < n; i++ {
		buffer = append(buffer,*e.Stack.Pop())
	}
	xn := e.Stack.Peek()
	for i := n-1; i >= 0; i-- {
		e.Stack.Push(&buffer[i])
	}
	e.Stack.Push(xn)

	return NONE,nil
}

func opRoll(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	n := int(e.Stack.Pop().ToBigInt().Int64())
	if n < 0  {return FAULT,nil}
	if n == 0  {return NONE,nil}
	if e.Stack.Count() < n+1 {return FAULT,nil}
	buffer := []StackItem{}
	for i := 0; i < n; i++ {
		buffer = append(buffer,*e.Stack.Pop())
	}
	xn := e.Stack.Pop()
	for i := n-1; i >= 0; i-- {
		e.Stack.Push(&buffer[i])
	}
	e.Stack.Push(xn)

	return NONE,nil
}

func opRot(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 3 {return FAULT,nil}
	x3 := e.Stack.Pop()
	x2 := e.Stack.Pop()
	x1 := e.Stack.Pop()
	e.Stack.Push(x2)
	e.Stack.Push(x3)
	e.Stack.Push(x1)

	return NONE,nil
}

func opSwap(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 { return FAULT,nil }
	x2 := e.Stack.Pop();
	x1 := e.Stack.Pop();
	e.Stack.Push(x2);
	e.Stack.Push(x1);

	return NONE,nil
}

func opTuck(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 { return FAULT,nil }
	x2 := e.Stack.Pop();
	x1 := e.Stack.Pop();
	e.Stack.Push(x2);
	e.Stack.Push(x1);
	e.Stack.Push(x2);

	return NONE,nil
}

func opCat(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 { return FAULT,nil }
	x2 := e.Stack.Pop();
	x1 := e.Stack.Pop();
	b1 := x1.GetBytesArray()
	b2 := x2.GetBytesArray()
	if (len(b1) != len(b2)) {return FAULT,nil}

	r := ByteArrZip(b1,b2,OP_CONCAT)
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}

func opSubStr(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 3 {return FAULT,nil}
	count := int(e.Stack.Pop().ToBigInt().Int64())
	if count < 0  {return FAULT,nil}
	index := int(e.Stack.Pop().ToBigInt().Int64())
	if index < 0  {return FAULT,nil}
	x := e.Stack.Pop()
	s := x.GetBytesArray()

	for _,b := range s{
		//p.Skip(index).Take(count) : need test
		b = b[index + count :]
	}
	e.Stack.Push(NewStackItem(s))

	return NONE,nil
}

func opLeft(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	count := int(e.Stack.Pop().ToBigInt().Int64())
	if count < 0  {return FAULT,nil}
	x := e.Stack.Pop()
	s := x.GetBytesArray()
	for _,b := range s{
		b = b[count:]
	}
	e.Stack.Push(NewStackItem(s))

	return NONE,nil
}

func opRight(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 2 {return FAULT,nil}
	count := int(e.Stack.Pop().ToBigInt().Int64())
	if count < 0  {return FAULT,nil}
	x := e.Stack.Pop()
	s := x.GetBytesArray()
	for _,b := range s{
		len := len(b)
		if len < count {return FAULT,nil}
		b = b[0:len - count]

	}
	e.Stack.Push(NewStackItem(s))

	return NONE,nil
}

func opSize(e *ExecutionEngine) (VMState,error) {

	if e.Stack.Count() < 1 {return FAULT,nil}
	x := e.Stack.Peek()
	s := x.GetBytesArray()
	r := []int{}
	for _,b := range s{
		r = append(r,len(b))
	}
	e.Stack.Push(NewStackItem(r))

	return NONE,nil
}