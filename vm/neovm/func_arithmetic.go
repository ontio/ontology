package neovm

func opBigInt(e *ExecutionEngine) (VMState, error) {
	x := PopBigInt(e)
	PushData(e, BigIntOp(x, e.opCode))
	return NONE, nil
}

func opSign(e *ExecutionEngine) (VMState, error) {
	x := PopBigInt(e)
	PushData(e, x.Sign())
	return NONE, nil
}

func opNot(e *ExecutionEngine) (VMState, error) {
	x := PopBoolean(e)
	PushData(e, !x)
	return NONE, nil
}

func opNz(e *ExecutionEngine) (VMState, error) {
	x := PopBigInt(e)
	PushData(e, BigIntComp(x, e.opCode))
	return NONE, nil
}

func opBigIntZip(e *ExecutionEngine) (VMState, error) {
	x2 := PopBigInt(e)
	x1 := PopBigInt(e)
	b := BigIntZip(x1, x2, e.opCode)
	PushData(e, b)
	return NONE, nil
}

func opBoolZip(e *ExecutionEngine) (VMState, error) {
	x2 := PopBoolean(e)
	x1 := PopBoolean(e)
	PushData(e, BoolZip(x1, x2, e.opCode))
	return NONE, nil
}

func opBigIntComp(e *ExecutionEngine) (VMState, error) {
	x2 := PopBigInt(e)
	x1 := PopBigInt(e)
	PushData(e, BigIntMultiComp(x1, x2, e.opCode))
	return NONE, nil
}

func opWithIn(e *ExecutionEngine) (VMState, error) {
	b := PopBigInt(e)
	a := PopBigInt(e)
	c := PopBigInt(e)
	PushData(e, WithInOp(c, a, b))
	return NONE, nil
}
