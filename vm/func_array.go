package vm

func opArraySize(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 1 {
		return FAULT, nil
	}
	arr := AssertStackItem(e.evaluationStack.Pop()).GetArray()
	err := pushData(e, len(arr))
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func opPack(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 1 {
		return FAULT, nil
	}
	size := int(AssertStackItem(e.evaluationStack.Pop()).GetBigInteger().Int64())
	if size < 0 || size > e.evaluationStack.Count() {
		return FAULT, nil
	}
	items := NewStackItems()
	for {
		if size == 0 {
			break
		}
		items = append(items, AssertStackItem(e.evaluationStack.Pop()))
		size--
	}
	err := pushData(e, items)
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func opUnpack(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 1 {
		return FAULT, nil
	}
	arr := AssertStackItem(e.evaluationStack.Pop()).GetArray()
	l := len(arr)
	for i := l - 1; i >= 0; i-- {
		e.evaluationStack.Push(arr[i])
	}
	err := pushData(e, l)
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func opPickItem(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 1 {
		return FAULT, nil
	}
	index := int(AssertStackItem(e.evaluationStack.Pop()).GetBigInteger().Int64())
	if index < 0 {
		return FAULT, nil
	}
	items := AssertStackItem(e.evaluationStack.Pop()).GetArray()
	if index >= len(items) {
		return FAULT, nil
	}
	err := pushData(e, items[index])
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}
