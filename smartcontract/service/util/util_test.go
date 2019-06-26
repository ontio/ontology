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
package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/vm/neovm/types"
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

	fmt.Println("===test single uint32===")
	bf = bytes.NewBuffer(nil)
	bf.WriteByte(byte(0))
	bf.WriteByte(UsizeType)
	tmpbytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpbytes, uint32(100000))
	bf.Write(tmpbytes)

	list, err = DeserializeInput(bf.Bytes())
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, len(list), 1)
	assert.Equal(t, list[0].(uint32), uint32(100000))

	fmt.Println("===test single int64===")
	bf = bytes.NewBuffer(nil)
	bf.WriteByte(byte(0))
	bf.WriteByte(Int64Type)
	tmpbytes = make([]byte, 8)
	tmpbf := bytes.NewBuffer(nil)
	tmpint := int64(1000000000000)

	binary.Write(tmpbf, binary.LittleEndian, &tmpint)

	bf.Write(tmpbf.Bytes())

	list, err = DeserializeInput(bf.Bytes())
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, len(list), 1)
	assert.Equal(t, list[0].(int64), tmpint)

	fmt.Println("===test minus int64===")
	bf = bytes.NewBuffer(nil)
	bf.WriteByte(byte(0))
	bf.WriteByte(Int64Type)
	tmpbf = bytes.NewBuffer(nil)
	tmpint = int64(-100)
	binary.Write(tmpbf, binary.LittleEndian, &tmpint)
	bf.Write(tmpbf.Bytes())

	list, err = DeserializeInput(bf.Bytes())
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, len(list), 1)
	assert.Equal(t, list[0].(int64), tmpint)

	fmt.Println("===test single uint64===")
	bf = bytes.NewBuffer(nil)
	bf.WriteByte(byte(0))
	bf.WriteByte(Uint64Type)

	tmpbytes = make([]byte, 8)
	binary.LittleEndian.PutUint64(tmpbytes, uint64(10000000000))
	bf.Write(tmpbytes)

	list, err = DeserializeInput(bf.Bytes())
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, len(list), 1)
	assert.Equal(t, list[0].(uint64), uint64(10000000000))

	fmt.Println("===test single uint256===")
	bf = bytes.NewBuffer(nil)
	bf.WriteByte(byte(0))
	bf.WriteByte(Uint256Type)

	u256, err := common.Uint256FromHexString("11111111AAAAAAAA8888888877777777")
	bf.Write(u256[:])

	list, err = DeserializeInput(bf.Bytes())
	assert.Nil(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, len(list), 1)
	assert.Equal(t, list[0].(common.Uint256), u256)

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

	bf.WriteByte(Uint64Type)
	tmpbytes = make([]byte, 8)
	binary.LittleEndian.PutUint64(tmpbytes, uint64(10000000000))
	bf.Write(tmpbytes)

	bf.WriteByte(Int64Type)
	tmpbf = bytes.NewBuffer(nil)
	tmpint = int64(-100)
	binary.Write(tmpbf, binary.LittleEndian, &tmpint)
	bf.Write(tmpbf.Bytes())

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

	assert.Equal(t, ssublist1[0].(uint64), uint64(10000000000))
	assert.Equal(t, ssublist1[1].(int64), int64(-100))

	assert.Equal(t, string(ssublist2[0].([]byte)), "helloworld")
	assert.Equal(t, string(ssublist2[1].([]byte)), "nested list")

}

func TestBuildResultFromNeo(t *testing.T) {
	fmt.Println("===Test Bytearray")
	ba := types.NewByteArray([]byte("helloworld"))
	bf := bytes.NewBuffer([]byte{VERSION})
	err := BuildResultFromNeo(ba, bf)
	assert.Nil(t, err)
	bs := bf.Bytes()
	assert.NotNil(t, bs)
	list, err := DeserializeInput(bs)
	assert.Equal(t, string(list[0].([]byte)), "helloworld")

	fmt.Println("===Test Bool")
	bf = bytes.NewBuffer([]byte{VERSION})
	bv := types.NewBoolean(false)
	err = BuildResultFromNeo(bv, bf)
	assert.Nil(t, err)
	bs = bf.Bytes()
	assert.NotNil(t, bs)
	list, err = DeserializeInput(bs)
	assert.False(t, list[0].(bool))

	fmt.Println("===Test Int")
	bf = bytes.NewBuffer([]byte{VERSION})
	iv := types.NewInteger(big.NewInt(int64(10000000)))
	err = BuildResultFromNeo(iv, bf)
	assert.Nil(t, err)
	bs = bf.Bytes()
	assert.NotNil(t, bs)
	list, err = DeserializeInput(bs)
	assert.Equal(t, list[0].(int64), int64(10000000))

	fmt.Println("===Test Array single bytearray")
	bf = bytes.NewBuffer([]byte{VERSION})
	ba = types.NewByteArray([]byte("helloworld"))
	array := types.NewArray([]types.StackItems{ba})
	err = BuildResultFromNeo(array, bf)
	assert.Nil(t, err)
	bs = bf.Bytes()
	assert.NotNil(t, bs)
	list, err = DeserializeInput(bs)
	assert.Equal(t, string(list[0].([]interface{})[0].([]byte)), "helloworld")

	fmt.Println("===Test Array  bytearray ,bool and int")
	bf = bytes.NewBuffer([]byte{VERSION})
	ba = types.NewByteArray([]byte("helloworld"))
	bv = types.NewBoolean(true)
	iv = types.NewInteger(big.NewInt(int64(10000000)))
	array = types.NewArray([]types.StackItems{ba, bv, iv})
	err = BuildResultFromNeo(array, bf)
	assert.Nil(t, err)
	bs = bf.Bytes()
	assert.NotNil(t, bs)
	list, err = DeserializeInput(bs)
	assert.Equal(t, string(list[0].([]interface{})[0].([]byte)), "helloworld")
	assert.True(t, list[0].([]interface{})[1].(bool))
	assert.Equal(t, list[0].([]interface{})[2].(int64), int64(10000000))

	fmt.Println("===Test nested Array ")
	bf = bytes.NewBuffer([]byte{VERSION})
	ba = types.NewByteArray([]byte("helloworld"))
	array1 := types.NewArray([]types.StackItems{ba})
	iv = types.NewInteger(big.NewInt(int64(10000000)))
	array2 := types.NewArray([]types.StackItems{iv})
	array = types.NewArray([]types.StackItems{array1, array2})
	err = BuildResultFromNeo(array, bf)
	assert.Nil(t, err)
	bs = bf.Bytes()
	assert.NotNil(t, bs)
	list, err = DeserializeInput(bs)

	assert.Equal(t, string(list[0].([]interface{})[0].([]interface{})[0].([]byte)), "helloworld")
	assert.Equal(t, list[0].([]interface{})[1].([]interface{})[0].(int64), int64(10000000))

	fmt.Println("===Test unsupport type")
	bf = bytes.NewBuffer([]byte{VERSION})
	mapv := types.NewMap()
	mapv.Add(ba, iv)
	err = BuildResultFromNeo(mapv, bf)
	assert.NotNil(t, err)

}
