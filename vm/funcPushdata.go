package vm

import (
)

func opPushData(e *ExecutionEngine) (VMState,error) {
	data,err := getPushData(e)
	if err != nil {
		return FAULT,err
	}

	e.Stack.Push(NewStackItem(data))
	return NONE,nil
}

func getPushData(e *ExecutionEngine) (interface{},error) {
	var data interface{}

	if  e.opcode >= OP_PUSHBYTES1 && e.opcode <= OP_PUSHBYTES75 {
		data = e.opReader.ReadBytes(int(e.opcode))
	}

	switch e.opcode {
	case OP_0:
		data = new([]byte)
		break
	case OP_PUSHDATA1:
		d,_ := e.opReader.ReadByte()
		data =  e.opReader.ReadBytes(int(d))
		break
	case OP_PUSHDATA2:
		data =  e.opReader.ReadBytes(int(e.opReader.ReadUint16()))
		break
	case OP_PUSHDATA4:
		data =  e.opReader.ReadBytes(int(e.opReader.ReadInt32()))
		break
	case OP_1NEGATE,OP_1,OP_2,OP_3,OP_4,OP_5,OP_6,OP_7,OP_8,OP_9,OP_10,OP_11,OP_12,OP_13,OP_14,OP_15,OP_16:
		data =  int8(e.opcode - OP_1 + 1)
		break
	}
	return data,nil
}
