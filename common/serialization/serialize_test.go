package serialization

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"testing"
)

func BenchmarkWriteVarUint(b *testing.B) {
	n := uint64(math.MaxUint64)
	for i := 0; i < b.N; i++ {
		WriteVarUint(ioutil.Discard, n)
	}
}

func BenchmarkWriteVarBytes(b *testing.B) {
	s := []byte{10, 11, 12}
	buf := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		WriteVarBytes(buf, s)
	}
}

func BenchmarkWriteVarString(b *testing.B) {
	s := "jim"
	buf := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		WriteVarString(buf, s)
	}
}

func BenchmarkReadVarUint(b *testing.B) {
	data := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	r := bytes.NewReader(data)
	ReadVarUint(r, 0)
}

func BenchmarkReadVarBytes(b *testing.B) {
	data := []byte{10, 11, 12}
	r := bytes.NewReader(data)
	ReadVarBytes(r)
}

func BenchmarkReadVarString(b *testing.B) {
	data := []byte{10, 11, 12}
	r := bytes.NewReader(data)
	ReadVarString(r)
}

func TestSerialize(t *testing.T) {
	b := new(bytes.Buffer)
	a3 := uint8(100)
	a4 := uint16(65535)
	a5 := uint32(4294967295)
	a6 := uint64(18446744073709551615)
	a7 := uint64(18446744073709551615)
	a8 := []byte{10, 11, 12}
	a9 := "hello onchain."

	WriteVarUint(b, uint64(a3))
	WriteVarUint(b, uint64(a4))
	WriteVarUint(b, uint64(a5))
	WriteVarUint(b, uint64(a6))
	WriteVarUint(b, uint64(a7))
	WriteVarBytes(b, a8)
	WriteVarString(b, a9)

	fmt.Println(ReadVarUint(b, math.MaxUint64))
	fmt.Println(ReadVarUint(b, math.MaxUint64))
	fmt.Println(ReadVarUint(b, math.MaxUint64))
	fmt.Println(ReadVarUint(b, math.MaxUint64))
	fmt.Println(ReadVarUint(b, math.MaxUint32))
	fmt.Println(ReadVarBytes(b))
	fmt.Println(ReadVarString(b))

	fmt.Printf("100 size is %d byte.\t\n", GetVarUintSize(uint64(100)))
	fmt.Printf("65535 size is %d byte.\t\n", GetVarUintSize(uint64(65535)))
	fmt.Printf("4294967295 size is %d byte.\t\n", GetVarUintSize(uint64(4294967295)))
	fmt.Printf("18446744073709551615 size is %d byte.\t\n", GetVarUintSize(uint64(18446744073709551615)))

	b.WriteByte(20)
	b.WriteByte(21)
	b.WriteByte(22)
	fmt.Println(ReadBytes(b, uint64(3)))

}

func TestReadWriteInt(t *testing.T) {
	b3 := new(bytes.Buffer)
	b4 := new(bytes.Buffer)
	b5 := new(bytes.Buffer)
	b6 := new(bytes.Buffer)

	a3 := uint8(100)
	a4 := uint16(65535)
	a5 := uint32(4294967295)
	a6 := uint64(18446744073709551615)

	WriteUint8(b3, a3)
	WriteUint16(b4, a4)
	WriteUint32(b5, a5)
	WriteUint64(b6, a6)

	fmt.Printf("uint8 %x\n", b3.Bytes())
	fmt.Printf("uint16 %x\n", b4.Bytes())
	fmt.Printf("uint32 %x\n", b5.Bytes())
	fmt.Printf("uint63 %x\n", b6.Bytes())

	fmt.Println(ReadUint8(b3))
	fmt.Println(ReadUint16(b4))
	fmt.Println(ReadUint32(b5))
	fmt.Println(ReadUint64(b6))

}
