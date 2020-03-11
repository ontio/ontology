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

package types

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func buildStruct(item []VmValue) (*StructValue, error) {
	s := NewStructValue()
	for _, val := range item {
		err := s.Append(val)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

func buildArray(item []VmValue) *ArrayValue {
	arr := NewArrayValue()
	for _, val := range item {
		arr.Append(val)
	}
	return arr
}

func TestSerialize(t *testing.T) {

	bsValue, err := VmValueFromBytes([]byte("test"))
	assert.Equal(t, err, nil)

	boolValue := VmValueFromBool(true)

	bigin := new(big.Int)
	bigin.SetInt64(int64(1000))
	biginValue, err := VmValueFromBigInt(bigin)
	assert.Equal(t, err, nil)

	uint64Value := VmValueFromUint64(uint64(100))

	s, err := buildStruct([]VmValue{bsValue, boolValue, biginValue, uint64Value})
	assert.Nil(t, err)
	structValue := VmValueFromStructVal(s)
	sink := new(common.ZeroCopySink)
	structValue.Serialize(sink)
	assert.Equal(t, common.ToHexString(sink.Bytes()), "810400047465737401010202e803020164")

	source := common.NewZeroCopySource(sink.Bytes())
	vs := VmValue{}
	vs.Deserialize(source)
	assert.Equal(t, structValue, vs)

	arr := buildArray([]VmValue{bsValue, boolValue, biginValue, uint64Value})
	sinkArr := new(common.ZeroCopySink)
	arrValue := VmValueFromArrayVal(arr)
	arrValue.Serialize(sinkArr)
	assert.Equal(t, common.ToHexString(sinkArr.Bytes()), "800400047465737401010202e803020164")

	arrValue2 := VmValue{}
	source = common.NewZeroCopySource(sinkArr.Bytes())
	arrValue2.Deserialize(source)
	assert.Equal(t, arrValue2, arrValue)

	m := NewMapValue()

	m.Set(bsValue, arrValue)
	m.Set(biginValue, structValue)
	m.Set(uint64Value, boolValue)

	mValue := VmValueFromMapValue(m)
	sinkMap := new(common.ZeroCopySink)
	mValue.Serialize(sinkMap)
	assert.Equal(t, "82030201640101000474657374800400047465737401010202e8030201640202e803810400047465737401010202e803020164", common.ToHexString(sinkMap.Bytes()))

	arr = NewArrayValue()
	b, _ := new(big.Int).SetString("9923372036854775807", 10)
	bi, err := VmValueFromBigInt(b)
	assert.Nil(t, err)
	arr.Append(bi)
	val_arr := VmValueFromArrayVal(arr)
	res_t, err := val_arr.ConvertNeoVmValueHexString()
	assert.Nil(t, err)
	fmt.Println("res_t:", res_t)
	assert.Equal(t, "ffffc58e4ae6b68900", res_t.([]interface{})[0])

}

func TestVmValue_ConvertNeoVmValueHexString(t *testing.T) {
	bs := make([]byte, 0, 64*1024)
	for i := 0; i < 64*1024; i++ {
		bs = append(bs, byte(1))
	}
	vm, err := VmValueFromBytes(bs)
	assert.Nil(t, err)
	_, err = vm.ConvertNeoVmValueHexString()
	assert.Nil(t, err)

	bs2 := make([]byte, 0, 64*1024)
	for i := 0; i < 64*1024+1; i++ {
		bs2 = append(bs2, byte(1))
	}
	vm2, err := VmValueFromBytes(bs2)
	assert.Nil(t, err)
	_, err = vm2.ConvertNeoVmValueHexString()
	assert.NotNil(t, err)
}

func TestStructValue_Clone(t *testing.T) {
	bsValue, err := VmValueFromBytes([]byte("test"))
	assert.Equal(t, err, nil)
	boolValue := VmValueFromBool(true)

	bigin := new(big.Int)
	bigin.SetInt64(int64(1000))
	biginValue, err := VmValueFromBigInt(bigin)
	assert.Equal(t, err, nil)

	uint64Value := VmValueFromUint64(uint64(100))

	m := NewMapValue()
	m.Set(bsValue, bsValue)
	s, err := buildStruct([]VmValue{bsValue, boolValue, biginValue, uint64Value, VmValueFromMapValue(m)})
	assert.Nil(t, err)
	s2, _ := s.Clone()
	structValue := VmValueFromStructVal(s)
	m2 := s2.Data[s2.Len()-1]
	mm2, _ := m2.AsMapValue()
	mm2.Set(bsValue, uint64Value)
	structValue2 := VmValueFromStructVal(s2)
	assert.Equal(t, structValue.Equals(structValue2), true)
}

func TestVmValue_Equals(t *testing.T) {
	bsValue, err := VmValueFromBytes([]byte("test"))
	assert.Equal(t, err, nil)
	boolValue := VmValueFromBool(true)

	bigin := new(big.Int)
	bigin.SetInt64(int64(1000))
	biginValue, err := VmValueFromBigInt(bigin)
	assert.Equal(t, err, nil)

	uint64Value := VmValueFromUint64(uint64(100))
	s, err := buildStruct([]VmValue{bsValue, boolValue, biginValue, uint64Value})
	assert.Nil(t, err)
	structValue := VmValueFromStructVal(s)

	s2, err := buildStruct([]VmValue{bsValue, boolValue, biginValue, uint64Value})
	assert.Nil(t, err)
	structValue2 := VmValueFromStructVal(s2)
	res := structValue.Equals(structValue2)
	assert.True(t, res)

	m := NewMapValue()
	m3 := VmValueFromMapValue(m)
	m2 := NewMapValue()
	assert.False(t, m3.Equals(VmValueFromMapValue(m2)))
	assert.True(t, m3.Equals(m3))

	arr := VmValueFromArrayVal(NewArrayValue())
	arr2 := VmValueFromArrayVal(NewArrayValue())
	assert.False(t, arr.Equals(arr2))

	intero := VmValueFromInteropValue(NewInteropValue(nil))
	intero2 := VmValueFromInteropValue(NewInteropValue(nil))
	assert.False(t, intero.Equals(intero2))

	_, err = intero.AsInteropValue()
	assert.Nil(t, err)
	_, err = arr.AsInteropValue()
	assert.NotNil(t, err)
}

func TestVmValue_BuildParamToNative(t *testing.T) {
	inte, err := VmValueFromBigInt(new(big.Int).SetUint64(math.MaxUint64))
	assert.Nil(t, err)
	boo := VmValueFromBool(false)
	bs, err := VmValueFromBytes([]byte("hello"))
	assert.Nil(t, err)

	stru, err := buildStruct([]VmValue{inte, boo, bs})
	assert.Nil(t, err)
	arr := NewArrayValue()
	s := VmValueFromStructVal(stru)
	r, _ := s.AsBool()
	assert.True(t, r)
	arr.Append(s)

	res := VmValueFromArrayVal(arr)

	_, err = res.AsBool()
	assert.NotNil(t, err)

	sink := common.NewZeroCopySink(nil)
	err = res.BuildParamToNative(sink)
	assert.Nil(t, err)
	assert.Equal(t, "010109ffffffffffffffff00000568656c6c6f", common.ToHexString(sink.Bytes()))
}

func TestVmValueFromUint64(t *testing.T) {
	val := VmValueFromUint64(math.MaxUint64)
	assert.Equal(t, val.valType, bigintType)
}

func TestVmValue_Deserialize(t *testing.T) {
	b, _ := new(big.Int).SetString("9923372036854775807", 10)
	val_b, err := VmValueFromBigInt(b)
	assert.Nil(t, err)
	m := NewMapValue()
	bs, err := VmValueFromBytes([]byte("key"))
	assert.Nil(t, err)
	m.Set(bs, val_b)

	val_m := VmValueFromMapValue(m)
	sink := common.NewZeroCopySink(nil)
	val_m.Serialize(sink)
	assert.Equal(t, "820100036b65790209ffffc58e4ae6b68900", common.ToHexString(sink.Bytes()))

	val_m2 := VmValueFromMapValue(nil)
	bss, err := common.HexToBytes("820100036b65790209ffffc58e4ae6b68900")
	assert.Nil(t, err)
	source := common.NewZeroCopySource(bss)
	val_m2.Deserialize(source)
	assert.Equal(t, val_m, val_m2)
}

func TestVmValue_DeserializeArray(t *testing.T) {
	val := VmValue{}
	sink := common.NewZeroCopySink(nil)
	for i := 0; i < MAX_COUNT; i++ {
		sink.WriteByte(arrayType)
		sink.WriteVarUint(1)
	}
	sink.WriteByte(boolType)
	sink.WriteBool(true)
	source := common.NewZeroCopySource(sink.Bytes())
	err := val.Deserialize(source)
	assert.Nil(t, err)

	val2 := VmValue{}
	sink2 := common.NewZeroCopySink(nil)
	for i := 0; i < MAX_COUNT+1; i++ {
		sink2.WriteByte(arrayType)
		sink2.WriteVarUint(1)
	}
	sink2.WriteByte(boolType)
	sink2.WriteBool(true)
	source2 := common.NewZeroCopySource(sink2.Bytes())
	err = val2.Deserialize(source2)
	assert.NotNil(t, err)
}

func TestVmValue_DeserializeMap(t *testing.T) {
	val := VmValue{}
	sink := common.NewZeroCopySink(nil)
	sink.WriteByte(mapType)
	sink.WriteVarUint(1)
	for i := 0; i < MAX_COUNT-1; i++ {
		sink.WriteByte(bytearrayType)
		sink.WriteVarUint(1)
		sink.WriteByte(1)
		sink.WriteByte(mapType)
		sink.WriteVarUint(1)
	}
	sink.WriteByte(bytearrayType)
	sink.WriteVarUint(1)
	sink.WriteByte(1)

	sink.WriteByte(bytearrayType)
	sink.WriteVarUint(1)
	sink.WriteByte(1)

	source := common.NewZeroCopySource(sink.Bytes())
	err := val.Deserialize(source)
	assert.Nil(t, err)

	val2 := VmValue{}
	sink2 := common.NewZeroCopySink(nil)
	sink2.WriteByte(mapType)
	sink2.WriteVarUint(1)
	for i := 0; i < MAX_COUNT; i++ {
		sink2.WriteByte(bytearrayType)
		sink2.WriteVarUint(1)
		sink2.WriteByte(1)
		sink2.WriteByte(mapType)
		sink2.WriteVarUint(1)
	}
	sink2.WriteByte(boolType)
	sink2.WriteBool(true)
	sink2.WriteByte(boolType)
	sink2.WriteBool(true)

	source2 := common.NewZeroCopySource(sink2.Bytes())
	err = val2.Deserialize(source2)
	assert.Equal(t, "vmvalue depth over the uplimit", err.Error())
}

func TestVmValue_DeserializeStruct(t *testing.T) {
	sink := common.NewZeroCopySink(nil)

	for i := 0; i < MAX_COUNT-1; i++ {
		sink.WriteByte(structType)
		sink.WriteVarUint(1)
	}
	sink.WriteByte(boolType)
	sink.WriteBool(true)

	source := common.NewZeroCopySource(sink.Bytes())
	val := VmValue{}
	err := val.Deserialize(source)
	assert.Nil(t, err)

	val2 := VmValue{}
	sink2 := common.NewZeroCopySink(nil)
	for i := 0; i < MAX_COUNT+1; i++ {
		sink2.WriteByte(structType)
		sink2.WriteVarUint(1)
	}
	sink2.WriteByte(boolType)
	sink2.WriteBool(true)

	source2 := common.NewZeroCopySource(sink2.Bytes())
	err = val2.Deserialize(source2)
	assert.Equal(t, "vmvalue depth over the uplimit", err.Error())
}

func TestVmValue_AsBool(t *testing.T) {
	b, _ := new(big.Int).SetString("9923372036854775807", 10)
	val, err := VmValueFromBigInt(b)
	assert.Nil(t, err)
	res, err := val.AsBool()
	assert.Nil(t, err)
	assert.Equal(t, true, res)
	//9223372036854775807
	bb, _ := new(big.Int).SetString("9223372036854775807", 10)
	val, err = VmValueFromBigInt(bb)
	assert.Nil(t, err)
	in, err := val.AsInt64()
	assert.Equal(t, in, int64(9223372036854775807))

	val, err = VmValueFromBytes([]byte("hello"))
	assert.Nil(t, err)
	res, err = val.AsBool()
	assert.Nil(t, err)
	assert.Equal(t, true, res)

	m := NewMapValue()
	val = VmValueFromMapValue(m)
	res, err = val.AsBool()
	assert.Nil(t, err)
	assert.Equal(t, true, res)

	val_bs, err := VmValueFromBytes(common.BigIntToNeoBytes(b))
	assert.Nil(t, err)
	val_b, err := VmValueFromBigInt(b)
	assert.Nil(t, err)
	in_val, _ := val_bs.AsIntValue()
	in_val2, _ := val_b.AsIntValue()
	assert.Equal(t, in_val, in_val2)

	_, err = val.AsIntValue()
	assert.NotNil(t, err)
	res, _ = val.AsBool()
	assert.True(t, res)
}

func TestVmValueFromInteropValue(t *testing.T) {
	u, _ := common.Uint256FromHexString("a00000000000000000000a000000000000000000000000000000000000000000")
	val_u := NewInteropValue(&u)
	vmVal_u := VmValueFromInteropValue(val_u)

	u2, _ := common.Uint256FromHexString("a00000000000000000000a000000000000000000000000000000000000000001")

	val_u2 := NewInteropValue(&u2)
	vmVal_u2 := VmValueFromInteropValue(val_u2)
	assert.False(t, vmVal_u.Equals(vmVal_u2))
}

func TestVmValue_CircularRefAndDepthDetection(t *testing.T) {
	a := NewArrayValue()
	aVal := VmValueFromArrayVal(a)

	b := NewArrayValue()
	b.Append(aVal)
	bVal := VmValueFromArrayVal(b)

	abool, err := aVal.CircularRefAndDepthDetection()
	assert.Nil(t, err)
	assert.False(t, abool)

	bbool, err := bVal.CircularRefAndDepthDetection()
	assert.Nil(t, err)
	assert.False(t, bbool)
}

func TestVmValue_CircularRefAndDepthDetection2(t *testing.T) {
	ba1, err := VmValueFromBytes([]byte{1, 2, 3})
	assert.Nil(t, err)
	ba2, err := VmValueFromBytes([]byte{4, 5, 6})
	assert.Nil(t, err)

	bf := VmValueFromBool(false)
	bt := VmValueFromBool(true)

	checkVal(t, ba1)
	checkVal(t, ba2)
	checkVal(t, bf)
	checkVal(t, bt)

	array := buildArray([]VmValue{ba1, ba2, bf, bt})
	arrayVal := VmValueFromArrayVal(array)
	checkVal(t, arrayVal)

	stru, err := buildStruct([]VmValue{ba1, ba2, bf, bt})
	assert.Nil(t, err)
	struVal := VmValueFromStructVal(stru)
	checkVal(t, struVal)

	array.Append(struVal)
	arrayVal = VmValueFromArrayVal(array)
	checkVal(t, arrayVal)

	err = stru.Append(arrayVal)
	assert.Nil(t, err)
	checkVal(t, VmValueFromStructVal(stru))

	map1 := NewMapValue()
	map1Val := VmValueFromMapValue(map1)
	checkVal(t, map1Val)

	map1.Set(arrayVal, bf)
	checkVal(t, VmValueFromMapValue(map1))

	stru2 := NewStructValue()
	array2 := NewArrayValue()

	stru2.Append(VmValueFromArrayVal(array2))
	array2.Append(VmValueFromStructVal(stru2))

	stru2.Append(VmValueFromArrayVal(array2))
	array2.Append(VmValueFromStructVal(stru2))

	arrayVal = VmValueFromArrayVal(array2)
	boo, err := arrayVal.CircularRefAndDepthDetection()
	assert.Nil(t, err)
	assert.True(t, boo)
}

func checkVal(t *testing.T, value VmValue) {
	boo, err := value.CircularRefAndDepthDetection()
	assert.Nil(t, err)
	assert.False(t, boo)
}
