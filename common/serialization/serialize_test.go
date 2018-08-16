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

package serialization

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"testing"

	"crypto/rand"
	"github.com/stretchr/testify/assert"
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
		WriteString(buf, s)
	}
}

func BenchmarkReadVarUint(b *testing.B) {
	data := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data)
		ReadVarUint(r, 0)
	}
}

func BenchmarkReadVarBytes(b *testing.B) {
	data := []byte{10, 11, 12}
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data)
		ReadVarBytes(r)
	}
}

func BenchmarkReadVarString(b *testing.B) {
	data := []byte{10, 11, 12}
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data)
		ReadString(r)
	}
}

func BenchmarkSerialize(ben *testing.B) {
	a3 := uint8(100)
	a4 := uint16(65535)
	a5 := uint32(4294967295)
	a6 := uint64(18446744073709551615)
	a7 := uint64(18446744073709551615)
	a8 := []byte{10, 11, 12}
	a9 := "hello onchain."
	for i := 0; i < ben.N; i++ {
		b := new(bytes.Buffer)

		WriteVarUint(b, uint64(a3))
		WriteVarUint(b, uint64(a4))
		WriteVarUint(b, uint64(a5))
		WriteVarUint(b, uint64(a6))
		WriteVarUint(b, uint64(a7))
		WriteVarBytes(b, a8)
		WriteString(b, a9)

		ReadVarUint(b, math.MaxUint64)
		ReadVarUint(b, math.MaxUint64)
		ReadVarUint(b, math.MaxUint64)
		ReadVarUint(b, math.MaxUint64)
		ReadVarUint(b, math.MaxUint32)
		ReadVarBytes(b)
		ReadString(b)

		GetVarUintSize(uint64(100))
		GetVarUintSize(uint64(65535))
		GetVarUintSize(uint64(4294967295))
		GetVarUintSize(uint64(18446744073709551615))

		b.WriteByte(20)
		b.WriteByte(21)
		b.WriteByte(22)
		ReadBytes(b, uint64(3))
	}

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
	WriteString(b, a9)

	fmt.Println(ReadVarUint(b, math.MaxUint64))
	fmt.Println(ReadVarUint(b, math.MaxUint64))
	fmt.Println(ReadVarUint(b, math.MaxUint64))
	fmt.Println(ReadVarUint(b, math.MaxUint64))
	fmt.Println(ReadVarUint(b, math.MaxUint32))
	fmt.Println(ReadVarBytes(b))
	fmt.Println(ReadString(b))

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

func TestReadVarBytesMemAllocAttack(t *testing.T) {
	buff := bytes.NewBuffer([]byte{1, 2, 3})
	length := math.MaxInt64
	_, err := byteXReader(buff, uint64(length))
	assert.NotNil(t, err)
}

func TestReadVarBytesRead(t *testing.T) {
	bs := make([]byte, 2048+1)
	for i := 0; i < len(bs); i++ {
		bs[i] = byte(i)
	}
	buff := bytes.NewBuffer(bs)
	read, err := byteXReader(buff, uint64(len(bs)))
	assert.Nil(t, err)
	assert.Equal(t, bs, read)
}

const N = 24829*1 + 1

func BenchmarkBytesXReader(b *testing.B) {
	bs := make([]byte, N)
	rand.Read(bs)
	for i := 0; i < b.N; i++ {
		buff := bytes.NewBuffer(bs)
		byteXReader(buff, uint64(len(bs)))
	}
}
