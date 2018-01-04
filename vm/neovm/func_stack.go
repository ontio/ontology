package neovm

func opToDupFromAltStack(e *ExecutionEngine) (VMState, error) {
	Push(e, e.altStack.Peek(0))
	return NONE, nil
}

func opToAltStack(e *ExecutionEngine) (VMState, error) {
	e.altStack.Push(Pop(e))
	return NONE, nil
}

func opFromAltStack(e *ExecutionEngine) (VMState, error) {
	Push(e, e.altStack.Pop())
	return NONE, nil
}

func opXDrop(e *ExecutionEngine) (VMState, error) {
	n := PopInt(e)
	e.evaluationStack.Remove(n)
	return NONE, nil
}

func opXSwap(e *ExecutionEngine) (VMState, error) {
	n := PopInt(e)
	if n == 0 {
		return NONE, nil
	}
	e.evaluationStack.Swap(0, n)
	return NONE, nil
}

func opXTuck(e *ExecutionEngine) (VMState, error) {
	n := PopInt(e)
	e.evaluationStack.Insert(n, Peek(e))
	return NONE, nil
}

func opDepth(e *ExecutionEngine) (VMState, error) {
	PushData(e, Count(e))
	return NONE, nil
}

func opDrop(e *ExecutionEngine) (VMState, error) {
	Pop(e)
	return NONE, nil
}

func opDup(e *ExecutionEngine) (VMState, error) {
	Push(e, Peek(e))
	return NONE, nil
}

func opNip(e *ExecutionEngine) (VMState, error) {
	x2 := Pop(e)
	Pop(e)
	Push(e, x2)
	return NONE, nil
}

func opOver(e *ExecutionEngine) (VMState, error) {
	x2 := Pop(e)
	x1 := Peek(e)

	Push(e, x2)
	Push(e, x1)
	return NONE, nil
}

func opPick(e *ExecutionEngine) (VMState, error) {
	n := PopInt(e)
	Push(e, e.evaluationStack.Peek(n))
	return NONE, nil
}

func opRoll(e *ExecutionEngine) (VMState, error) {
	n := PopInt(e)
	if n == 0 {
		return NONE, nil
	}
	Push(e, e.evaluationStack.Remove(n))
	return NONE, nil
}

func opRot(e *ExecutionEngine) (VMState, error) {
	x3 := Pop(e)
	x2 := Pop(e)
	x1 := Pop(e)
	Push(e, x2)
	Push(e, x3)
	Push(e, x1)
	return NONE, nil
}

func opSwap(e *ExecutionEngine) (VMState, error) {
	x2 := Pop(e)
	x1 := Pop(e)
	Push(e, x2)
	Push(e, x1)
	return NONE, nil
}

func opTuck(e *ExecutionEngine) (VMState, error) {
	x2 := Pop(e)
	x1 := Pop(e)
	Push(e, x2)
	Push(e, x1)
	Push(e, x2)
	return NONE, nil
}

