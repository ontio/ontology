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
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/vm/neovm"
)

type ProgramBuilder struct {
	sink *common.ZeroCopySink
}

func (self *ProgramBuilder) PushOpCode(op neovm.OpCode) *ProgramBuilder {
	self.sink.WriteByte(byte(op))
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
		self.sink.WriteByte(byte(len(data)) + byte(neovm.PUSHBYTES1) - 1)
	} else if len(data) < 0x100 {
		self.sink.WriteByte(byte(neovm.PUSHDATA1))
		self.sink.WriteUint8(uint8(len(data)))
	} else if len(data) < 0x10000 {
		self.sink.WriteByte(byte(neovm.PUSHDATA2))
		self.sink.WriteUint16(uint16(len(data)))
	} else {
		self.sink.WriteByte(byte(neovm.PUSHDATA4))
		self.sink.WriteUint32(uint32(len(data)))
	}
	self.sink.WriteBytes(data)

	return self
}

func (self *ProgramBuilder) Finish() []byte {
	return self.sink.Bytes()
}

func NewProgramBuilder() ProgramBuilder {
	return ProgramBuilder{sink: common.NewZeroCopySink(nil)}
}

func ProgramFromPubKey(pubkey keypair.PublicKey) []byte {
	sink := common.ZeroCopySink{}
	EncodeSinglePubKeyProgramInto(&sink, pubkey)
	return sink.Bytes()
}

func EncodeSinglePubKeyProgramInto(sink *common.ZeroCopySink, pubkey keypair.PublicKey) {
	builder := ProgramBuilder{sink: sink}

	builder.PushPubKey(pubkey).PushOpCode(neovm.CHECKSIG)
}

func EncodeMultiPubKeyProgramInto(sink *common.ZeroCopySink, pubkeys []keypair.PublicKey, m int) error {
	n := len(pubkeys)
	if !(1 <= m && m <= n && n > 1 && n <= constants.MULTI_SIG_MAX_PUBKEY_SIZE) {
		return errors.New("wrong multi-sig param")
	}

	pubkeys = keypair.SortPublicKeys(pubkeys)

	builder := ProgramBuilder{sink: sink}
	builder.PushNum(uint16(m))
	for _, pubkey := range pubkeys {
		key := keypair.SerializePublicKey(pubkey)
		builder.PushBytes(key)
	}

	builder.PushNum(uint16(len(pubkeys)))
	builder.PushOpCode(neovm.CHECKMULTISIG)
	return nil
}

func ProgramFromMultiPubKey(pubkeys []keypair.PublicKey, m int) ([]byte, error) {
	sink := common.ZeroCopySink{}
	err := EncodeMultiPubKeyProgramInto(&sink, pubkeys, m)
	return sink.Bytes(), err
}

func ProgramFromParams(sigs [][]byte) []byte {
	builder := NewProgramBuilder()
	for _, sig := range sigs {
		builder.PushBytes(sig)
	}

	return builder.Finish()
}

func EncodeParamProgramInto(sink *common.ZeroCopySink, sigs [][]byte) {
	builder := ProgramBuilder{sink: sink}
	for _, sig := range sigs {
		builder.PushBytes(sig)
	}
}

type ProgramInfo struct {
	PubKeys []keypair.PublicKey
	M       uint16
}

type programParser struct {
	source *common.ZeroCopySource
}

func newProgramParser(prog []byte) *programParser {
	return &programParser{source: common.NewZeroCopySource(prog)}
}

func (self *programParser) ReadOpCode() (neovm.OpCode, error) {
	code, eof := self.source.NextByte()
	if eof {
		return neovm.OpCode(code), io.ErrUnexpectedEOF
	}
	return neovm.OpCode(code), nil
}

func (self *programParser) PeekOpCode() (neovm.OpCode, error) {
	code, err := self.ReadOpCode()
	if err == nil {
		self.source.BackUp(1)
	}
	return code, err
}

func (self *programParser) ExpectEOF() error {
	if self.source.Len() != 0 {
		return fmt.Errorf("expected eof, but remains %d bytes", self.source.Len())
	}
	return nil
}

func (self *programParser) IsEOF() bool {
	return self.source.Len() == 0
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
	var eof bool
	if code == neovm.PUSHDATA4 {
		var temp uint32
		temp, eof = self.source.NextUint32()
		keylen = uint64(temp)
	} else if code == neovm.PUSHDATA2 {
		var temp uint16
		temp, eof = self.source.NextUint16()
		keylen = uint64(temp)
	} else if code == neovm.PUSHDATA1 {
		var temp uint8
		temp, eof = self.source.NextUint8()
		keylen = uint64(temp)
	} else if byte(code) <= byte(neovm.PUSHBYTES75) && byte(code) >= byte(neovm.PUSHBYTES1) {
		keylen = uint64(code) - uint64(neovm.PUSHBYTES1) + 1
	} else {
		err = fmt.Errorf("unexpected opcode: %d", byte(code))
	}
	if eof {
		err = io.ErrUnexpectedEOF
	}
	if err != nil {
		return nil, err
	}

	buf, eof := self.source.NextBytes(keylen)
	if eof {
		return nil, io.ErrUnexpectedEOF
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
		parser := programParser{source: common.NewZeroCopySource(program[:len(program)-1])}
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
		parser := programParser{source: common.NewZeroCopySource(program)}
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

		if !(1 <= m && int64(m) <= n && n > 1 && n <= constants.MULTI_SIG_MAX_PUBKEY_SIZE) {
			return info, errors.New("wrong multi-sig param")
		}
		info.M = m

		return info, nil
	}

	return info, errors.New("unsupported program")
}

// note output has reference of input `program`
func GetParamInfo(program []byte) ([][]byte, error) {
	parser := programParser{source: common.NewZeroCopySource(program)}

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
