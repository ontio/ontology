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
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
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
	return self.PushBytes(buf)
}

func (self *ProgramBuilder) PushNum(num uint16) *ProgramBuilder {
	if num == 0 {
		return self.PushOpCode(neovm.PUSH0)
	} else if num <= 16 {
		return self.PushOpCode(neovm.OpCode(uint8(num) - 1 + uint8(neovm.PUSH1)))
	}

	bint := big.NewInt(int64(num))
	return self.PushBytes(common.BigIntToNeoBytes(bint))
}

func (self *ProgramBuilder) PushBytes(data []byte) *ProgramBuilder {
	if len(data) == 0 {
		panic("push data error: data is nil")
	}

	if len(data) <= int(neovm.PUSHBYTES75)+1-int(neovm.PUSHBYTES1) {
		self.buffer.WriteByte(byte(len(data)) + byte(neovm.PUSHBYTES1) - 1)
	} else if len(data) < 0x100 {
		self.buffer.WriteByte(byte(neovm.PUSHDATA1))
		serialization.WriteUint8(&self.buffer, uint8(len(data)))
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
	if !(1 <= m && m <= n && n <= constants.MULTI_SIG_MAX_PUBKEY_SIZE) {
		return nil, errors.New("wrong multi-sig param")
	}

	list := keypair.NewPublicList(pubkeys)
	sort.Sort(list)

	builder := ProgramBuilder{}
	builder.PushNum(uint16(m))
	for _, pubkey := range list {
		builder.PushBytes(pubkey)
	}

	builder.PushNum(uint16(len(pubkeys)))
	builder.PushOpCode(neovm.CHECKMULTISIG)
	return builder.Finish(), nil
}

func ProgramFromParams(sigs [][]byte) []byte {
	builder := ProgramBuilder{}
	for _, sig := range sigs {
		builder.PushBytes(sig)
	}

	return builder.Finish()
}

type ProgramInfo struct {
	PubKeys []keypair.PublicKey
	M       uint16
}

type programParser struct {
	buffer *bytes.Buffer
}

func newProgramParser(prog []byte) *programParser {
	return &programParser{buffer: bytes.NewBuffer(prog)}
}

func (self *programParser) ReadOpCode() (neovm.OpCode, error) {
	code, err := self.buffer.ReadByte()
	return neovm.OpCode(code), err
}

func (self *programParser) PeekOpCode() (neovm.OpCode, error) {
	code, err := self.ReadOpCode()
	if err == nil {
		self.buffer.UnreadByte()
	}
	return code, err
}

func (self *programParser) ExpectEOF() error {
	if self.buffer.Len() != 0 {
		return fmt.Errorf("expected eof, but remains %d bytes", self.buffer.Len())
	}
	return nil
}

func (self *programParser) IsEOF() bool {
	return self.buffer.Len() == 0
}

func (self *programParser) ReadNum() (uint16, error) {
	code, err := self.PeekOpCode()
	if err != nil {
		return 0, err
	}

	if code == neovm.PUSH0 {
		self.ReadOpCode()
		return 0, nil
	} else if num := int(code) - int(neovm.PUSH1) + 1; 1 <= num && num <= 16 {
		self.ReadOpCode()
		return uint16(num), nil
	}

	buff, err := self.ReadBytes()
	if err != nil {
		return 0, err
	}
	bint := common.BigIntFromNeoBytes(buff)
	num := bint.Int64()
	if num > math.MaxUint16 || num <= 16 {
		return 0, fmt.Errorf("num not in range (16, MaxUint16]: %d", num)
	}

	return uint16(num), nil
}

func (self *programParser) ReadBytes() ([]byte, error) {
	code, err := self.ReadOpCode()
	if err != nil {
		return nil, err
	}

	var keylen uint64
	if code == neovm.PUSHDATA4 {
		var temp uint32
		temp, err = serialization.ReadUint32(self.buffer)
		keylen = uint64(temp)
	} else if code == neovm.PUSHDATA2 {
		var temp uint16
		temp, err = serialization.ReadUint16(self.buffer)
		keylen = uint64(temp)
	} else if code == neovm.PUSHDATA1 {
		var temp uint8
		temp, err = serialization.ReadUint8(self.buffer)
		keylen = uint64(temp)
	} else if byte(code) <= byte(neovm.PUSHBYTES75) && byte(code) >= byte(neovm.PUSHBYTES1) {
		keylen = uint64(code) - uint64(neovm.PUSHBYTES1) + 1
	} else {
		err = fmt.Errorf("unexpected opcode: %d", byte(code))
	}
	if err != nil {
		return nil, err
	}

	buf, err := serialization.ReadBytes(self.buffer, keylen)
	if err != nil {
		return nil, err
	}

	return buf, err
}

func (self *programParser) ReadPubKey() (keypair.PublicKey, error) {
	buf, err := self.ReadBytes()
	if err != nil {
		return nil, err
	}

	pubkey, err := keypair.DeserializePublicKey(buf)
	return pubkey, err
}

func GetProgramInfo(program []byte) (ProgramInfo, error) {
	info := ProgramInfo{}

	if len(program) <= 2 {
		return info, errors.New("wrong program")
	}

	end := program[len(program)-1]
	if end == byte(neovm.CHECKSIG) {
		parser := programParser{buffer: bytes.NewBuffer(program[:len(program)-1])}
		pubkey, err := parser.ReadPubKey()
		if err != nil {
			return info, err
		}
		err = parser.ExpectEOF()
		if err != nil {
			return info, err
		}
		info.PubKeys = append(info.PubKeys, pubkey)
		info.M = 1

		return info, nil
	} else if end == byte(neovm.CHECKMULTISIG) {
		parser := programParser{buffer: bytes.NewBuffer(program)}
		m, err := parser.ReadNum()
		if err != nil {
			return info, err
		}
		for i := 0; i < int(m); i++ {
			key, err := parser.ReadPubKey()
			if err != nil {
				return info, err
			}
			info.PubKeys = append(info.PubKeys, key)
		}
		var buffers [][]byte
		for {
			code, err := parser.PeekOpCode()
			if err != nil {
				return info, err
			}

			if code == neovm.CHECKMULTISIG {
				parser.ReadOpCode()
				break
			} else if code == neovm.PUSH0 {
				parser.ReadOpCode()
				bint := big.NewInt(0)
				buffers = append(buffers, common.BigIntToNeoBytes(bint))
			} else if num := int(code) - int(neovm.PUSH1) + 1; 1 <= num && num <= 16 {
				parser.ReadOpCode()
				bint := big.NewInt(int64(num))
				buffers = append(buffers, common.BigIntToNeoBytes(bint))
			} else {
				buff, err := parser.ReadBytes()
				if err != nil {
					return info, err
				}
				buffers = append(buffers, buff)
			}
		}
		err = parser.ExpectEOF()
		if err != nil {
			return info, err
		}
		if len(buffers) < 1 {
			return info, errors.New("missing pubkey length")
		}
		bint := big.NewInt(0)
		bint.SetBytes(buffers[len(buffers)-1])
		n := bint.Int64()

		for i := 0; i < len(buffers)-1; i++ {
			pubkey, err := keypair.DeserializePublicKey(buffers[i])
			if err != nil {
				return info, err
			}
			info.PubKeys = append(info.PubKeys, pubkey)
		}
		if int64(len(info.PubKeys)) != n {
			return info, fmt.Errorf("number of pubkeys unmarched, expected:%d, got: %d", len(info.PubKeys), n)
		}

		if !(1 <= m && int64(m) <= n && n <= constants.MULTI_SIG_MAX_PUBKEY_SIZE) {
			return info, errors.New("wrong multi-sig param")
		}
		info.M = m

		return info, nil
	}

	return info, errors.New("unsupported program")
}

func GetParamInfo(program []byte) ([][]byte, error) {
	parser := programParser{buffer: bytes.NewBuffer(program)}

	var signatures [][]byte
	for parser.IsEOF() == false {
		sig, err := parser.ReadBytes()
		if err != nil {
			return nil, err
		}
		signatures = append(signatures, sig)
	}

	return signatures, nil
}
