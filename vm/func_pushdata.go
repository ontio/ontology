package vm

func opPushData(e *ExecutionEngine) (VMState, error) {
	data, err := getPushData(e)
	if err != nil {
		return FAULT, err
	}
	err = pushData(e, data)
	if err != nil {
		return FAULT, err
	}
	return NONE, nil
}

func getPushData(e *ExecutionEngine) (interface{}, error) {
	var data interface{}

	if e.opCode >= PUSHBYTES1 && e.opCode <= PUSHBYTES75 {
		data = e.context.OpReader.ReadBytes(int(e.opCode))
	}
	switch e.opCode {
	case PUSH0:
		data = []byte{0}
	case PUSHDATA1:
		d, _ := e.context.OpReader.ReadByte()
		data = e.context.OpReader.ReadBytes(int(d))
	case PUSHDATA2:
		data = e.context.OpReader.ReadBytes(int(e.context.OpReader.ReadUint16()))
	case PUSHDATA4:
		data = e.context.OpReader.ReadBytes(int(e.context.OpReader.ReadInt32()))
	case PUSHM1, PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH6, PUSH7, PUSH8, PUSH9, PUSH10, PUSH11, PUSH12, PUSH13, PUSH14, PUSH15, PUSH16:
		data = int8(e.opCode - PUSH1 + 1)
	}

	return data, nil
}
