package vm

func opCat(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 2 {
		return FAULT, nil
	}
	x2 := e.evaluationStack.Pop()
	x1 := e.evaluationStack.Pop()
	b1 := AssertStackItem(x1).GetByteArray()
	b2 := AssertStackItem(x2).GetByteArray()
	if len(b1) != len(b2) {
		return FAULT, nil
	}
	r := ByteArrZip(b1, b2, CAT)
	pushData(e, r)
	return NONE, nil
}

func opSubStr(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 3 {
		return FAULT, nil
	}
	count := int(AssertStackItem(e.evaluationStack.Pop()).GetBigInteger().Int64())
	if count < 0 {
		return FAULT, nil
	}
	index := int(AssertStackItem(e.evaluationStack.Pop()).GetBigInteger().Int64())
	if index < 0 {
		return FAULT, nil
	}
	x := e.evaluationStack.Pop()
	s := AssertStackItem(x).GetByteArray()
	l1 := index + count
	l2 := len(s)
	if l1 > l2 {
		return FAULT, nil
	}
	b := s[index : l2-l1+1]
	err := pushData(e, b)
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func opLeft(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 2 {
		return FAULT, nil
	}
	count := int(AssertStackItem(e.evaluationStack.Pop()).GetBigInteger().Int64())
	if count < 0 {
		return FAULT, nil
	}
	x := e.evaluationStack.Pop()
	s := AssertStackItem(x).GetByteArray()
	if count > len(s) {
		return FAULT, nil
	}
	b := s[:count]
	err := pushData(e, b)
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func opRight(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 2 {
		return FAULT, nil
	}
	count := int(AssertStackItem(e.evaluationStack.Pop()).GetBigInteger().Int64())
	if count < 0 {
		return FAULT, nil
	}
	x := e.evaluationStack.Pop()
	s := AssertStackItem(x).GetByteArray()
	l := len(s)
	if count > l {
		return FAULT, nil
	}
	b := s[l-count:]
	err := pushData(e, b)
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func opSize(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 1 {
		return FAULT, nil
	}
	x := e.evaluationStack.Peek(0)
	s := AssertStackItem(x).GetByteArray()
	err := pushData(e, len(s))
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}
