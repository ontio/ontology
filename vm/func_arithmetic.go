package vm

func opBigInt(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 1 {
		return FAULT, nil
	}
	x := AssertStackItem(e.evaluationStack.Pop()).GetBigInteger()
	err := pushData(e, BigIntOp(x, e.opCode))
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func opNot(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 1 {
		return FAULT, nil
	}
	x := AssertStackItem(e.evaluationStack.Pop()).GetBoolean()
	err := pushData(e, !x)
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func opNz(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 1 {
		return FAULT, nil
	}
	x := AssertStackItem(e.evaluationStack.Pop()).GetBigInteger()
	err := pushData(e, BigIntComp(x, e.opCode))
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func opBigIntZip(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 2 {
		return FAULT, nil
	}
	x2 := AssertStackItem(e.evaluationStack.Pop()).GetBigInteger()
	x1 := AssertStackItem(e.evaluationStack.Pop()).GetBigInteger()
	err := pushData(e, BigIntZip(x1, x2, e.opCode))
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func opBoolZip(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 2 {
		return FAULT, nil
	}
	x2 := AssertStackItem(e.evaluationStack.Pop()).GetBoolean()
	x1 := AssertStackItem(e.evaluationStack.Pop()).GetBoolean()
	err := pushData(e, BoolZip(x1, x2, e.opCode))
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func opBigIntComp(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 2 {
		return FAULT, nil
	}
	x2 := AssertStackItem(e.evaluationStack.Pop()).GetBigInteger()
	x1 := AssertStackItem(e.evaluationStack.Pop()).GetBigInteger()
	err := pushData(e, BigIntMultiComp(x1, x2, e.opCode))
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func opWithIn(e *ExecutionEngine) (VMState, error) {
	if e.evaluationStack.Count() < 3 {
		return FAULT, nil
	}
	b := AssertStackItem(e.evaluationStack.Pop()).GetBigInteger()
	a := AssertStackItem(e.evaluationStack.Pop()).GetBigInteger()
	x := AssertStackItem(e.evaluationStack.Pop()).GetBigInteger()
	err := pushData(e, WithInOp(x, a, b))
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}
