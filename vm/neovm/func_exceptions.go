package neovm

func opThrow(e *ExecutionEngine) (VMState, error) {
	return FAULT, nil
}

func opThrowIfNot(e *ExecutionEngine) (VMState, error) {
	if !PopBoolean(e) {
		return FAULT, nil
	}
	return NONE, nil
}
