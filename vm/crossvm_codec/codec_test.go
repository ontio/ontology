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
package crossvm_codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func TestDeserializeInput(t *testing.T) {

	fmt.Println("===test single byte array===")
	bf := bytes.NewBuffer(nil)
	bf.WriteByte(byte(0))
	bf.WriteByte(ByteArrayType)

	s := []byte("helloworld")
	length := len(s)
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(length))
	bf.Write(tmp)
	bf.Write(s)

	list, err := DeserializeInput(bf.Bytes())
	assert.Nil(t, err)
	assert.NotNil(t, list)

	assert.Equal(t, len(list), 1)
	assert.True(t, string(list[0].([]byte)) == "helloworld")

	fmt.Println("===test single Address===")

	bf = bytes.NewBuffer(nil)
	bf.WriteByte(byte(0))
	bf.WriteByte(AddressType)
	addr, _ := common.AddressFromBase58("AY5hDhn2z8ND6F4JF9rQV1a4SDUT4aUr88")
	bf.Write(addr[:])

	list, err = DeserializeInput(bf.Bytes())
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, len(list), 1)
	tmpaddr := list[0].(common.Address)
	assert.Equal(t, tmpaddr.ToBase58(), "AY5hDhn2z8ND6F4JF9rQV1a4SDUT4aUr88")

	fmt.Println("===test single boolean===")
	bf = bytes.NewBuffer(nil)
	bf.WriteByte(byte(0))
	bf.WriteByte(BooleanType)
	bf.WriteByte(byte(1))

	list, err = DeserializeInput(bf.Bytes())
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, len(list), 1)
	assert.True(t, list[0].(bool))

	fmt.Println("===test single Int===")
	bf = bytes.NewBuffer(nil)
	bf.WriteByte(byte(0))
	bf.WriteByte(IntType)
	bi := big.NewInt(int64(100000))
	bs := common.BigIntToNeoBytes(bi)
	l := len(bs)
	tmpbytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpbytes, uint32(l))
	bf.Write(tmpbytes)
	bf.Write(bs)

	list, err = DeserializeInput(bf.Bytes())
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, len(list), 1)
	assert.Equal(t, list[0].(*big.Int).Int64(), int64(100000))

	fmt.Println("===test single Int 0===")
	bf = bytes.NewBuffer(nil)
	bf.WriteByte(byte(0))
	bf.WriteByte(IntType)
	bi = big.NewInt(int64(0))
	bs = common.BigIntToNeoBytes(bi)
	l = len(bs)
	tmpbytes = make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpbytes, uint32(l))
	bf.Write(tmpbytes)
	bf.Write(bs)

	list, err = DeserializeInput(bf.Bytes())
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, len(list), 1)
	assert.Equal(t, list[0].(*big.Int).Int64(), int64(0))

	fmt.Println("===test minus Int===")
	bf = bytes.NewBuffer(nil)
	bf.WriteByte(byte(0))
	bf.WriteByte(IntType)
	bi = big.NewInt(int64(-100))
	bs = common.BigIntToNeoBytes(bi)
	l = len(bs)
	tmpbytes = make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpbytes, uint32(l))
	bf.Write(tmpbytes)
	bf.Write(bs)

	list, err = DeserializeInput(bf.Bytes())
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, len(list), 1)
	assert.Equal(t, list[0].(*big.Int).Int64(), int64(-100))

	fmt.Println("===test single H256===")
	bf = bytes.NewBuffer(nil)
	bf.WriteByte(byte(0))
	bf.WriteByte(H256Type)

	var h256 common.Uint256
	for i := 0; i < 32; i++ {
		h256[i] = byte(i)
	}
	bf.Write(h256[:])

	list, err = DeserializeInput(bf.Bytes())
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, len(list), 1)
	assert.Equal(t, list[0].(common.Uint256), h256)

	fmt.Println("===test 2 simple elements list===")
	bf = bytes.NewBuffer(nil)
	bf.WriteByte(byte(0))
	bf.WriteByte(ListType)
	tmpbytes = make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpbytes, uint32(2))
	bf.Write(tmpbytes)

	//1st byte array
	bf.WriteByte(ByteArrayType)
	s = []byte("helloworld")
	length = len(s)
	tmp = make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(length))
	bf.Write(tmp)
	bf.Write(s)

	bf.WriteByte(AddressType)
	addr, _ = common.AddressFromBase58("AY5hDhn2z8ND6F4JF9rQV1a4SDUT4aUr88")
	bf.Write(addr[:])

	list, err = DeserializeInput(bf.Bytes())
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, len(list), 1)
	sublist := list[0].([]interface{})

	assert.Equal(t, len(sublist), 2)
	assert.Equal(t, string(sublist[0].([]byte)), "helloworld")
	tmpaddr = sublist[1].(common.Address)
	assert.Equal(t, tmpaddr.ToBase58(), "AY5hDhn2z8ND6F4JF9rQV1a4SDUT4aUr88")

	fmt.Println("===test nested list===")
	bf = bytes.NewBuffer(nil)
	bf.WriteByte(byte(0))
	bf.WriteByte(ListType)
	tmpbytes = make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpbytes, uint32(2))
	bf.Write(tmpbytes)

	bf.WriteByte(ListType)
	tmpbytes = make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpbytes, uint32(2))
	bf.Write(tmpbytes)

	bf.WriteByte(IntType)
	bi = big.NewInt(int64(10000000000))

	bs = common.BigIntToNeoBytes(bi)
	bf.Write(uint32ToLittleEndiaBytes(uint32(len(bs))))
	bf.Write(bs)

	bf.WriteByte(IntType)
	bi = big.NewInt(int64(-100))

	bs = common.BigIntToNeoBytes(bi)
	bf.Write(uint32ToLittleEndiaBytes(uint32(len(bs))))
	bf.Write(bs)

	bf.WriteByte(ListType)
	tmpbytes = make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpbytes, uint32(2))
	bf.Write(tmpbytes)

	bf.WriteByte(ByteArrayType)
	s = []byte("helloworld")
	length = len(s)
	tmp = make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(length))
	bf.Write(tmp)
	bf.Write(s)

	bf.WriteByte(ByteArrayType)
	s = []byte("nested list")
	length = len(s)
	tmp = make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(length))
	bf.Write(tmp)
	bf.Write(s)

	list, err = DeserializeInput(bf.Bytes())
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, len(list), 1)
	sublist = list[0].([]interface{})

	assert.Equal(t, len(sublist), 2)
	ssublist1 := sublist[0].([]interface{})
	ssublist2 := sublist[1].([]interface{})
	assert.Equal(t, len(ssublist1), 2)
	assert.Equal(t, len(ssublist2), 2)

	assert.Equal(t, ssublist1[0].(*big.Int).Int64(), int64(10000000000))
	assert.Equal(t, ssublist1[1].(*big.Int).Int64(), int64(-100))

	assert.Equal(t, string(ssublist2[0].([]byte)), "helloworld")
	assert.Equal(t, string(ssublist2[1].([]byte)), "nested list")

}

func uint32ToLittleEndiaBytes(i uint32) []byte {
	tmpbs := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpbs, i)
	return tmpbs
}
