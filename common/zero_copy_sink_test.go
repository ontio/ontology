package common

import (
	"testing"

	"bytes"
	ser "github.com/ontio/ontology/common/serialization"
)

func BenchmarkSerialize(ben *testing.B) {
	N := 1000
	a3 := uint8(100)
	a4 := uint16(65535)
	a5 := uint32(4294967295)
	a6 := uint64(18446744073709551615)
	a7 := uint64(18446744073709551615)
	a8 := []byte{10, 11, 12}
	a9 := "hello onchain."
	b := new(bytes.Buffer)
	for i := 0; i < ben.N; i++ {
		b.Reset()
		for j := 0; j < N; j++ {
			ser.WriteVarUint(b, uint64(a3))
			ser.WriteVarUint(b, uint64(a4))
			ser.WriteVarUint(b, uint64(a5))
			ser.WriteVarUint(b, uint64(a6))
			ser.WriteVarUint(b, uint64(a7))
			ser.WriteVarBytes(b, a8)
			ser.WriteString(b, a9)

			b.WriteByte(20)
			b.WriteByte(21)
			b.WriteByte(22)
		}
	}
}

func BenchmarkZeroCopySink(ben *testing.B) {
	N := 1000
	a3 := uint8(100)
	a4 := uint16(65535)
	a5 := uint32(4294967295)
	a6 := uint64(18446744073709551615)
	a7 := uint64(18446744073709551615)
	a8 := []byte{10, 11, 12}
	a9 := "hello onchain."
	sink := NewZeroCopySink(nil)
	for i := 0; i < ben.N; i++ {
		sink.Reset()
		for j := 0; j < N; j++ {
			sink.WriteVarUint(uint64(a3))
			sink.WriteVarUint(uint64(a4))
			sink.WriteVarUint(uint64(a5))
			sink.WriteVarUint(uint64(a6))
			sink.WriteVarUint(uint64(a7))
			sink.WriteVarBytes(a8)
			sink.WriteString(a9)
			sink.WriteByte(20)
			sink.WriteByte(21)
			sink.WriteByte(22)
		}
	}

}
