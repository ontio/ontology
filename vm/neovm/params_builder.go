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

import (
	"bytes"
	"encoding/binary"
	"math/big"

	"github.com/ontio/ontology/common"
)

type ParamsBuilder struct {
	buffer *bytes.Buffer
}

func NewParamsBuilder(buffer *bytes.Buffer) *ParamsBuilder {
	return &ParamsBuilder{buffer}
}

func (p *ParamsBuilder) Emit(op OpCode) {
	p.buffer.WriteByte(byte(op))
}

func (p *ParamsBuilder) EmitPushBool(data bool) {
	if data {
		p.Emit(PUSHT)
		return
	}
	p.Emit(PUSHF)
}

func (p *ParamsBuilder) EmitPushInteger(data *big.Int) {
	if data.Int64() == -1 {
		p.Emit(PUSHM1)
		return
	}
	if data.Int64() == 0 {
		p.Emit(PUSH0)
		return
	}
	if data.Int64() > 0 && data.Int64() < 16 {
		p.Emit(OpCode((int(PUSH1) - 1 + int(data.Int64()))))
		return
	}

	bytes := common.BigIntToNeoBytes(data)
	p.EmitPushByteArray(bytes)
}

func (p *ParamsBuilder) EmitPushByteArray(data []byte) {
	l := len(data)
	if l < int(PUSHBYTES75) {
		p.buffer.WriteByte(byte(l))
	} else if l < 0x100 {
		p.Emit(PUSHDATA1)
		p.buffer.WriteByte(byte(l))
	} else if l < 0x10000 {
		p.Emit(PUSHDATA2)
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(l))
		p.buffer.Write(b)
	} else {
		p.Emit(PUSHDATA4)
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(l))
		p.buffer.Write(b)
	}
	p.buffer.Write(data)
}

func (p *ParamsBuilder) EmitPushCall(codeHash []byte) {
	p.Emit(APPCALL)
	p.buffer.Write(codeHash)
}

func (p *ParamsBuilder) ToArray() []byte {
	return p.buffer.Bytes()
}
