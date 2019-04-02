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
	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestSerialize(t *testing.T) {

	bsValue, err := VmValueFromBytes([]byte("test"))
	fmt.Println(common.ToHexString([]byte("test")))
	assert.Equal(t, err, nil)
	boolValue := VmValueFromBool(true)

	bigin := new(big.Int)
	bigin.SetInt64(int64(1000))
	biginValue, err := VmValueFromBigInt(bigin)
	assert.Equal(t, err, nil)

	uint64Value := VmValueFromUint64(uint64(100))
	s := NewStructValue()
	s.Append(bsValue)
	s.Append(boolValue)
	s.Append(biginValue)
	s.Append(uint64Value)
	structValue := VmValueFromStructVal(s)
	sink := new(common.ZeroCopySink)
	structValue.Serialize(sink)
	fmt.Println(common.ToHexString(sink.Bytes()))
	assert.Equal(t, common.ToHexString(sink.Bytes()), "810400047465737401010202e803020164")

	structValueStr, err := structValue.ConvertNeoVmValueHexString()
	fmt.Println("structValueStr:", structValueStr)

	source := common.NewZeroCopySource(sink.Bytes())
	vs := VmValue{}
	vs.Deserialize(source)
	assert.Equal(t, structValue, vs)

	arr := NewArrayValue()
	arr.Append(bsValue)
	arr.Append(boolValue)
	arr.Append(biginValue)
	arr.Append(uint64Value)
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
	s := NewStructValue()
	s.Append(bsValue)
	s.Append(boolValue)
	s.Append(biginValue)
	s.Append(uint64Value)
	s.Append(VmValueFromMapValue(m))
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
	s := NewStructValue()
	s.Append(bsValue)
	s.Append(boolValue)
	s.Append(biginValue)
	s.Append(uint64Value)
	structValue := VmValueFromStructVal(s)

	s2 := NewStructValue()
	s2.Append(bsValue)
	s2.Append(boolValue)
	s2.Append(biginValue)
	s2.Append(uint64Value)
	structValue2 := VmValueFromStructVal(s2)
	res := structValue.Equals(structValue2)
	assert.Equal(t, res, true)
}
