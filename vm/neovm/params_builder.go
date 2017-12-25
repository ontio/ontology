package neovm

import (
	"bytes"
	"encoding/binary"
	. "github.com/Ontology/vm/neovm/types"
	"math/big"
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

	bytes := ConvertBigIntegerToBytes(data)
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
