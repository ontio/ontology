package neovm

func opCat(e *ExecutionEngine) (VMState, error) {
	b2 := PopByteArray(e)
	b1 := PopByteArray(e)
	r := Concat(b1, b2)
	PushData(e, r)
	return NONE, nil
}

func opSubStr(e *ExecutionEngine) (VMState, error) {
	count := PopInt(e)
	index := PopInt(e)
	arr := PopByteArray(e)
	b := arr[index : index + count]
	PushData(e, b)
	return NONE, nil
}

func opLeft(e *ExecutionEngine) (VMState, error) {
	count := PopInt(e)
	s := PopByteArray(e)
	b := s[:count]
	PushData(e, b)
	return NONE, nil
}

func opRight(e *ExecutionEngine) (VMState, error) {
	count := PopInt(e)
	arr := PopByteArray(e)
	b := arr[len(arr) - count:]
	PushData(e, b)
	return NONE, nil
}

func opSize(e *ExecutionEngine) (VMState, error) {
	x := Peek(e).GetStackItem()
	PushData(e, len(x.GetByteArray()))
	return NONE, nil
}
