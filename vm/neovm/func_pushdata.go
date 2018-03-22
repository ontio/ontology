/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package neovm

func opPushData(e *ExecutionEngine) (VMState, error) {
	data := getPushData(e)
	PushData(e, data)
	return NONE, nil
}

func getPushData(e *ExecutionEngine) interface{} {
	var data interface{}
	if e.opCode >= PUSHBYTES1 && e.opCode <= PUSHBYTES75 {
		data = e.context.OpReader.ReadBytes(int(e.opCode))
	}
	switch e.opCode {
	case PUSH0:
		data = int8(0)
	case PUSHDATA1:
		d, _ := e.context.OpReader.ReadByte()
		data = e.context.OpReader.ReadBytes(int(d))
	case PUSHDATA2:
		data = e.context.OpReader.ReadBytes(int(e.context.OpReader.ReadUint16()))
	case PUSHDATA4:
		i := int(e.context.OpReader.ReadInt32())
		data = e.context.OpReader.ReadBytes(i)
	case PUSHM1, PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH6, PUSH7, PUSH8, PUSH9, PUSH10, PUSH11, PUSH12, PUSH13, PUSH14, PUSH15, PUSH16:
		data = int8(e.opCode - PUSH1 + 1)
	}

	return data
}
