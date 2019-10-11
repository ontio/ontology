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
	"bytes"
	"fmt"
	"io"
	"math"
	"math/big"
	"reflect"
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/vm/crossvm_codec"
	"github.com/ontio/ontology/vm/neovm/constants"
	"github.com/ontio/ontology/vm/neovm/errors"
)

const (
	bytearrayType byte = 0x00
	boolType      byte = 0x01
	integerType   byte = 0x02
	bigintType    byte = 0x03
	interopType   byte = 0x40
	arrayType     byte = 0x80
	structType    byte = 0x81
	mapType       byte = 0x82
)

const (
	MAX_COUNT         = 1024
	MAX_NOTIFY_LENGTH = 64 * 1024 //64Kb
)

type VmValue struct {
	valType   byte
	integer   int64
	bigInt    *big.Int
	byteArray []byte
	structval *StructValue
	array     *ArrayValue
	mapval    *MapValue
	interop   InteropValue
}

func VmValueFromInt64(val int64) VmValue {
	return VmValue{valType: integerType, integer: val}
}

func VmValueFromBytes(val []byte) (result VmValue, err error) {
	if len(val) > constants.MAX_BYTEARRAY_SIZE {
		err = errors.ERR_OVER_MAX_ITEM_SIZE
		return
	}
	result.valType = bytearrayType
	result.byteArray = val
	return
}

func VmValueFromBool(val bool) VmValue {
	if val {
		return VmValue{valType: boolType, integer: 1}
	} else {
		return VmValue{valType: boolType, integer: 0}
	}
}

func VmValueFromUint64(val uint64) VmValue {
	if val <= math.MaxInt64 {
		return VmValueFromInt64(int64(val))
	}

	b := big.NewInt(0)
	b.SetUint64(val)
	return VmValue{valType: bigintType, bigInt: b}
}

func VmValueFromBigInt(val *big.Int) (result VmValue, err error) {
	value, e := IntValFromBigInt(val)
	if e != nil {
		err = e
		return
	}

	return VmValueFromIntValue(value), nil
}

func VmValueFromArrayVal(array *ArrayValue) VmValue {
	return VmValue{valType: arrayType, array: array}
}

func VmValueFromStructVal(val *StructValue) VmValue {
	return VmValue{valType: structType, structval: val}
}

func VmValueFromInteropValue(val InteropValue) VmValue {
	return VmValue{valType: interopType, interop: val}
}
func VmValueFromMapValue(val *MapValue) VmValue {
	return VmValue{valType: mapType, mapval: val}
}

func NewMapVmValue() VmValue {
	return VmValue{valType: mapType, mapval: NewMapValue()}
}

func VmValueFromIntValue(val IntValue) VmValue {
	if val.isbig {
		return VmValue{valType: bigintType, bigInt: val.bigint}
	} else {
		return VmValue{valType: integerType, integer: val.integer}
	}
}

func (self *VmValue) AsBytes() ([]byte, error) {
	switch self.valType {
	case boolType:
		if self.integer == 0 {
			return []byte{0}, nil
		} else {
			return []byte{1}, nil
		}
	case integerType:
		return common.BigIntToNeoBytes(big.NewInt(self.integer)), nil
	case bigintType:
		return common.BigIntToNeoBytes(self.bigInt), nil
	case bytearrayType:
		return self.byteArray, nil
	case arrayType, mapType, structType, interopType:
		return nil, errors.ERR_BAD_TYPE
	default:
		panic("unreachable!")
	}
}

func (self *VmValue) BuildParamToNative(sink *common.ZeroCopySink) error {
	b, err := self.CircularRefAndDepthDetection()
	if err != nil {
		return err
	}
	if b {
		return fmt.Errorf("runtime serialize: can not serialize circular reference data")
	}
	return self.buildParamToNative(sink)
}

func (self *VmValue) buildParamToNative(sink *common.ZeroCopySink) error {
	switch self.valType {
	case bytearrayType:
		sink.WriteVarBytes(self.byteArray)
	case boolType:
		b, err := self.AsBool()
		if err != nil {
			return err
		}
		sink.WriteBool(b)
	case integerType, bigintType:
		bs, err := self.AsBytes()
		if err != nil {
			return err
		}
		sink.WriteVarBytes(bs)
	case arrayType:
		sink.WriteVarBytes(common.BigIntToNeoBytes(big.NewInt(int64(len(self.array.Data)))))
		for _, v := range self.array.Data {
			err := v.BuildParamToNative(sink)
			if err != nil {
				return err
			}
		}
	case structType:
		for _, v := range self.structval.Data {
			err := v.BuildParamToNative(sink)
			if err != nil {
				return err
			}
		}
	case mapType:
		//TODO
		return errors.ERR_BAD_TYPE
	case interopType:
		return errors.ERR_BAD_TYPE
	default:
		panic("unreachable!")
	}
	return nil
}
func (self *VmValue) ConvertNeoVmValueHexString() (interface{}, error) {
	var count int
	var length int
	res, err := self.convertNeoVmValueHexString(&count, &length)
	if err != nil {
		return nil, err
	}
	if length > MAX_NOTIFY_LENGTH {
		return nil, fmt.Errorf("length over max parameters convert length")
	}
	return res, nil
}

func (self *VmValue) convertNeoVmValueHexString(count *int, length *int) (interface{}, error) {
	if *count > MAX_COUNT {
		return nil, fmt.Errorf("over max parameters convert length")
	}
	if *length > MAX_NOTIFY_LENGTH {
		return nil, fmt.Errorf("length over max parameters convert length")
	}
	switch self.valType {
	case boolType:
		boo, err := self.AsBool()
		if err != nil {
			return nil, err
		}
		*length++
		if boo {
			return common.ToHexString([]byte{1}), nil
		} else {
			return common.ToHexString([]byte{0}), nil
		}
	case bytearrayType:
		*length += len(self.byteArray)
		return common.ToHexString(self.byteArray), nil
	case integerType:
		var bs []byte
		if self.integer == 0 {
			bs = []byte{0}
		} else {
			bs = common.BigIntToNeoBytes(big.NewInt(self.integer))
		}
		*length += len(bs)
		return common.ToHexString(bs), nil
	case bigintType:
		var bs []byte
		if self.bigInt.Sign() == 0 {
			bs = []byte{0}
		} else {
			bs = common.BigIntToNeoBytes(self.bigInt)
		}
		*length += len(bs)
		return common.ToHexString(bs), nil
	case structType:
		var sstr []interface{}
		for i := 0; i < len(self.structval.Data); i++ {
			*count++
			t, err := self.structval.Data[i].convertNeoVmValueHexString(count, length)
			if err != nil {
				return nil, err
			}
			sstr = append(sstr, t)
		}
		return sstr, nil
	case arrayType:
		var sstr []interface{}
		for i := 0; i < len(self.array.Data); i++ {
			*count++
			t, err := self.array.Data[i].convertNeoVmValueHexString(count, length)
			if err != nil {
				return nil, err
			}
			sstr = append(sstr, t)
		}
		return sstr, nil
	case interopType:
		bs := self.interop.Data.ToArray()
		*length += len(bs)
		return common.ToHexString(bs), nil
	default:
		log.Errorf("[ConvertTypes] Invalid Types!, %x", self.valType)
		return nil, fmt.Errorf("[ConvertTypes] Invalid Types!, %x", self.valType)
	}
}
func (self *VmValue) Deserialize(source *common.ZeroCopySource) error {
	return self.deserialize(source, 0)
}

func (self *VmValue) deserialize(source *common.ZeroCopySource, depth int) error {
	if depth > MAX_COUNT {
		return fmt.Errorf("vmvalue depth over the uplimit")
	}
	t, eof := source.NextByte()
	if eof {
		return io.ErrUnexpectedEOF
	}
	switch t {
	case boolType:
		b, irregular, eof := source.NextBool()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		*self = VmValueFromBool(b)
	case bytearrayType:
		data, _, irregular, eof := source.NextVarBytes()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		value, err := VmValueFromBytes(data)
		if err != nil {
			return err
		}
		*self = value
	case integerType:
		data, _, irregular, eof := source.NextVarBytes()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		value, err := VmValueFromBigInt(common.BigIntFromNeoBytes(data))
		if err != nil {
			return err
		}
		*self = value
	case arrayType:
		l, _, irregular, eof := source.NextVarUint()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		arr := new(ArrayValue)
		for i := 0; i < int(l); i++ {
			v := VmValue{}
			err := v.deserialize(source, depth+1)
			if err != nil {
				return err
			}
			err = arr.Append(v)
			if err != nil {
				return err
			}
		}
		*self = VmValueFromArrayVal(arr)
	case mapType:
		l, _, irregular, eof := source.NextVarUint()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		mapValue := NewMapValue()
		for i := 0; i < int(l); i++ {
			keyValue := VmValue{}
			err := keyValue.deserialize(source, depth+1)
			if err != nil {
				return err
			}
			v := VmValue{}
			err = v.deserialize(source, depth+1)
			if err != nil {
				return err
			}
			err = mapValue.Set(keyValue, v)
			if err != nil {
				return err
			}
		}
		*self = VmValueFromMapValue(mapValue)
	case structType:
		l, _, irregular, eof := source.NextVarUint()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		structValue := NewStructValue()
		for i := 0; i < int(l); i++ {
			v := VmValue{}
			err := v.deserialize(source, depth+1)
			if err != nil {
				return err
			}
			err = structValue.Append(v)
			if err != nil {
				return err
			}
		}
		*self = VmValueFromStructVal(structValue)
	default:
		return errors.ERR_BAD_TYPE
	}
	return nil
}

func (self *VmValue) Serialize(sink *common.ZeroCopySink) error {
	b, err := self.CircularRefAndDepthDetection()
	if err != nil {
		return err
	}
	if b {
		return fmt.Errorf("runtime serialize: can not serialize circular reference data")
	}
	switch self.valType {
	case boolType:
		sink.WriteByte(boolType)
		boo, err := self.AsBool()
		if err != nil {
			return err
		}
		sink.WriteBool(boo)
	case bytearrayType:
		sink.WriteByte(bytearrayType)
		sink.WriteVarBytes(self.byteArray)
	case bigintType:
		sink.WriteByte(integerType)
		sink.WriteVarBytes(common.BigIntToNeoBytes(self.bigInt))
	case integerType:
		sink.WriteByte(integerType)
		t := big.NewInt(self.integer)
		sink.WriteVarBytes(common.BigIntToNeoBytes(t))
	case arrayType:
		sink.WriteByte(arrayType)
		sink.WriteVarUint(uint64(len(self.array.Data)))
		for i := 0; i < len(self.array.Data); i++ {
			err := self.array.Data[i].Serialize(sink)
			if err != nil {
				return err
			}

		}
	case mapType:
		sink.WriteByte(mapType)
		sink.WriteVarUint(uint64(len(self.mapval.Data)))
		keys := self.mapval.getMapSortedKey()
		for _, key := range keys {
			val := self.mapval.Data[key]
			keyVal := val[0]
			err = keyVal.Serialize(sink)
			if err != nil {
				return err
			}
			value := val[1]
			err = value.Serialize(sink)
			if err != nil {
				return err
			}
		}
	case structType:
		sink.WriteByte(structType)
		sink.WriteVarUint(uint64(len(self.structval.Data)))
		for _, item := range self.structval.Data {
			err := item.Serialize(sink)
			if err != nil {
				return err
			}
		}
	case interopType:
		return fmt.Errorf("not support type: interopType")
	default:
		panic("unreachable!")
	}
	if sink.Size() > constants.MAX_BYTEARRAY_SIZE {
		return fmt.Errorf("runtime serialize: can not serialize length over the uplimit")
	}
	return nil
}

func (self *VmValue) CircularRefAndDepthDetection() (bool, error) {
	return self.circularRefAndDepthDetection(make(map[uintptr]bool), 0)
}

func (self *VmValue) circularRefAndDepthDetection(visited map[uintptr]bool, depth int) (bool, error) {
	if depth > MAX_STRUCT_DEPTH {
		return true, nil
	}
	switch self.valType {
	case arrayType:
		arr, err := self.AsArrayValue()
		if err != nil {
			return true, err
		}
		if len(arr.Data) == 0 {
			return false, nil
		}
		p := reflect.ValueOf(arr.Data).Pointer()
		if visited[p] {
			return true, nil
		}
		visited[p] = true
		for _, v := range arr.Data {
			return v.circularRefAndDepthDetection(visited, depth+1)
		}
		delete(visited, p)
		return false, nil
	case structType:
		s, err := self.AsStructValue()
		if err != nil {
			return true, err
		}
		if len(s.Data) == 0 {
			return false, nil
		}

		p := reflect.ValueOf(s.Data).Pointer()
		if visited[p] {
			return true, nil
		}
		visited[p] = true

		for _, v := range s.Data {
			return v.circularRefAndDepthDetection(visited, depth+1)
		}

		delete(visited, p)
		return false, nil
	case mapType:
		mp, err := self.AsMapValue()
		if err != nil {
			return true, err
		}
		p := reflect.ValueOf(mp.Data).Pointer()
		if visited[p] {
			return true, nil
		}
		visited[p] = true
		for _, v := range mp.Data {
			return v[1].circularRefAndDepthDetection(visited, depth+1)
		}
		delete(visited, p)
		return false, nil
	default:
		return false, nil
	}
}

func (self *VmValue) AsInt64() (int64, error) {
	val, err := self.AsIntValue()
	if err != nil {
		return 0, err
	}
	if val.isbig {
		if val.bigint.IsInt64() == false {
			return 0, errors.ERR_INTEGER_UNDERFLOW
		}
		return val.bigint.Int64(), nil
	}

	return val.integer, nil
}

// urgly hack: only used in cmp opcode to lift the 32byte limit of integer
func (self *VmValue) AsBigInt() (*big.Int, error) {
	switch self.valType {
	case integerType, boolType:
		return big.NewInt(self.integer), nil
	case bigintType:
		return self.bigInt, nil
	case bytearrayType:
		value := common.BigIntFromNeoBytes(self.byteArray)
		return value, nil
	case arrayType, mapType, structType, interopType:
		return nil, errors.ERR_BAD_TYPE
	default:
		panic("unreachable!")
	}
}

func (self *VmValue) AsIntValue() (IntValue, error) {
	switch self.valType {
	case integerType, boolType:
		return IntValFromInt(self.integer), nil
	case bigintType:
		return IntValFromBigInt(self.bigInt)
	case bytearrayType:
		return IntValFromNeoBytes(self.byteArray)
	case arrayType, mapType, structType, interopType:
		return IntValue{}, errors.ERR_BAD_TYPE
	default:
		panic("unreachable!")
	}
}

func (self *VmValue) AsBool() (bool, error) {
	switch self.valType {
	case integerType, boolType:
		return self.integer != 0, nil
	case bigintType:
		return self.bigInt.Sign() != 0, nil
	case bytearrayType:
		for _, b := range self.byteArray {
			if b != 0 {
				return true, nil
			}
		}
		return false, nil
	case structType, mapType:
		return true, nil
	case arrayType:
		return false, errors.ERR_BAD_TYPE
	case interopType:
		return self.interop != InteropValue{}, nil
	default:
		panic("unreachable!")
	}
}

func (self *VmValue) AsMapValue() (*MapValue, error) {
	switch self.valType {
	case mapType:
		return self.mapval, nil
	default:
		return nil, errors.ERR_BAD_TYPE
	}
}

func (self *VmValue) AsStructValue() (*StructValue, error) {
	switch self.valType {
	case structType:
		return self.structval, nil
	default:
		return nil, errors.ERR_BAD_TYPE
	}
}

func (self *VmValue) AsArrayValue() (*ArrayValue, error) {
	switch self.valType {
	case arrayType:
		return self.array, nil
	default:
		return nil, errors.ERR_BAD_TYPE
	}
}

func (self *VmValue) AsInteropValue() (InteropValue, error) {
	switch self.valType {
	case interopType:
		return self.interop, nil
	default:
		return InteropValue{}, errors.ERR_BAD_TYPE
	}
}

func (self *VmValue) Equals(other VmValue) bool {
	v1, e1 := self.AsBytes()
	v2, e2 := other.AsBytes()
	if e1 == nil && e2 == nil { // both are primitive type
		return bytes.Equal(v1, v2)
	}

	// here more than one are compound type
	if self.valType != other.valType {
		return false
	}

	switch self.valType {
	case mapType:
		return self.mapval == other.mapval
	case structType:
		// todo: fix inconsistence
		return reflect.DeepEqual(self.structval, other.structval)
	case arrayType:
		return self.array == other.array
	case interopType:
		return self.interop.Equals(other.interop)
	default:
		panic("unreachable!")
	}
}

func (self *VmValue) GetType() byte {
	switch self.valType {
	case integerType, bigintType:
		return IntegerType
	default:
		return self.valType
	}
}

func (self *VmValue) GetMapKey() (string, error) {
	val, err := self.AsBytes()
	if err != nil {
		return "", err
	}
	return string(val), nil
}

//only for debug/testing
func (self *VmValue) Stringify() (string, error) {
	b, err := self.CircularRefAndDepthDetection()
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}
	if b {
		return "", fmt.Errorf("error: can not serialize circular reference data")
	}
	return self.stringify(), nil
}
func (self *VmValue) stringify() string {
	switch self.valType {
	case boolType, bytearrayType, bigintType, integerType:
		bs, _ := self.AsBytes()
		if len(bs) == 0 {
			bs = []byte{0}
		}
		return fmt.Sprintf("bytes(hex:%x)", bs)
	case arrayType:
		data := ""
		for _, v := range self.array.Data {
			data += v.stringify() + ", "
		}
		return fmt.Sprintf("array[%d]{%s}", len(self.array.Data), data)
	case mapType:
		keys := self.mapval.getMapSortedKey()
		data := ""
		for _, key := range keys {
			v := self.mapval.Data[key][1]
			data += fmt.Sprintf("%x: %s,", key, v.stringify())
		}
		return fmt.Sprintf("map[%d]{%s}", len(self.mapval.Data), data)
	case interopType:
		ty := reflect.TypeOf(self.interop.Data).String()
		return fmt.Sprintf("interop{type:%s}", ty)
	case structType:
		data := ""
		for _, v := range self.structval.Data {
			data += v.stringify() + ", "
		}
		return fmt.Sprintf("struct[%d]{%s}", len(self.structval.Data), data)
	default:
		panic("unreachable!")
	}
	return ""
}

//only for debug/testing
func (self *VmValue) Dump() string {
	b, err := self.CircularRefAndDepthDetection()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	if b {
		return "error: can not serialize circular reference data"
	}
	return self.dump()
}

func (self *VmValue) dump() string {
	switch self.valType {
	case boolType:
		bs, _ := self.AsBool()
		return fmt.Sprintf("bool(%v)", bs)
	case integerType:
		return fmt.Sprintf("int(%d)", self.integer)
	case bigintType:
		return fmt.Sprintf("bigint(0x%x)", self.bigInt)
	case bytearrayType:
		return fmt.Sprintf("string(\"%s\")", self.byteArray)
	case arrayType:
		data := ""
		for _, v := range self.array.Data {
			data += v.dump() + ", "
		}
		return fmt.Sprintf("array[%d]{%s}", len(self.array.Data), data)
	case mapType:
		var unsortKey []string
		for k := range self.mapval.Data {
			unsortKey = append(unsortKey, k)
		}
		sort.Strings(unsortKey)
		data := ""
		for _, key := range unsortKey {
			v := self.mapval.Data[key][1]
			k := self.mapval.Data[key][0]
			data += fmt.Sprintf("%s: %s,", k.dump(), v.dump())
		}
		return fmt.Sprintf("map[%d]{%s}", len(self.mapval.Data), data)
	case structType:
		data := ""
		for _, v := range self.structval.Data {
			data += v.dump() + ", "
		}
		return fmt.Sprintf("struct[%d]{%s}", len(self.structval.Data), data)
	case interopType:
		return fmt.Sprintf("interop[%x]", self.interop.Data)
	default:
		panic("unreachable!")
	}
	return ""
}

//encode the neovm return vmval
//transform neovm contract result to encoded byte array
func BuildResultFromNeo(item VmValue, bf *common.ZeroCopySink) error {
	if len(bf.Bytes()) > crossvm_codec.MAX_PARAM_LENGTH {
		return fmt.Errorf("parameter buf is too long")
	}
	switch item.valType {
	case bytearrayType:
		bs := item.byteArray
		crossvm_codec.EncodeBytes(bf, bs)
	case integerType:
		val := common.I128FromInt64(item.integer)
		crossvm_codec.EncodeInt128(bf, val)
	case bigintType:
		val := item.bigInt
		err := crossvm_codec.EncodeBigInt(bf, val)
		if err != nil {
			return err
		}
	case boolType:
		val, err := item.AsBool()
		if err != nil {
			return err
		}
		crossvm_codec.EncodeBool(bf, val)
	case arrayType:
		val := item.array
		if val == nil {
			return fmt.Errorf("get array error")
		}
		bf.WriteByte(crossvm_codec.ListType)
		bf.WriteUint32(uint32(len(val.Data)))
		for _, si := range val.Data {
			err := BuildResultFromNeo(si, bf)
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("not a supported return type")
	}
	return nil
}
