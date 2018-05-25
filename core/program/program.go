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

package program

import (
	"bytes"
	"errors"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/vm/neovm"
)

type ProgramBuilder struct {
	buffer bytes.Buffer
}

func (self *ProgramBuilder) PushOpCode(op neovm.OpCode) *ProgramBuilder {
	self.buffer.WriteByte(byte(op))
	return self
}

func (self *ProgramBuilder) PushPubKey(pubkey keypair.PublicKey) *ProgramBuilder {
	buf := keypair.SerializePublicKey(pubkey)
	return self.PushData(buf)
}

func (self *ProgramBuilder) PushNum(num uint8) *ProgramBuilder {
	if num == 0 {
		return self.PushOpCode(neovm.PUSH0)
	} else if num <= 16 {
		return self.PushOpCode(neovm.OpCode(num - 1 + byte(neovm.PUSH1)))
	}

	return self.PushData([]byte{num})
}

func (self *ProgramBuilder) PushData(data []byte) *ProgramBuilder {
	if data == nil {
		panic("push data error: data is nil")
	}

	if len(data) <= int(neovm.PUSHBYTES75) {
		self.buffer.WriteByte(byte(len(data)) + byte(neovm.PUSHBYTES1) - 1)
	} else if len(data) < 0x100 {
		self.buffer.WriteByte(byte(neovm.PUSHDATA1))
		self.buffer.WriteByte(byte(len(data)))
	} else if len(data) < 0x10000 {
		self.buffer.WriteByte(byte(neovm.PUSHDATA2))
		serialization.WriteUint16(&self.buffer, uint16(len(data)))
	} else {
		self.buffer.WriteByte(byte(neovm.PUSHDATA4))
		serialization.WriteUint32(&self.buffer, uint32(len(data)))
	}
	self.buffer.Write(data)

	return self
}

func (self *ProgramBuilder) Finish() []byte {
	return self.buffer.Bytes()
}

func ProgramFromPubKey(pubkey keypair.PublicKey) []byte {
	builder := ProgramBuilder{}
	return builder.PushPubKey(pubkey).PushOpCode(neovm.CHECKSIG).Finish()
}

func ProgramFromMultiPubKey(pubkeys []keypair.PublicKey, m int) ([]byte, error) {
	n := len(pubkeys)
	if m <= 0 || m > n || n > 24 {
		return nil, errors.New("wrong multi-sig param")
	}
	builder := ProgramBuilder{}
	builder.PushNum(uint8(m))
	for _, pubkey := range pubkeys {
		builder.PushPubKey(pubkey)
	}

	builder.PushNum(uint8(len(pubkeys)))
	builder.PushOpCode(neovm.CHECKMULTISIG)
	return builder.Finish(), nil
}
