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

package memory

import (
	"encoding/binary"
	"errors"
	"reflect"

	"github.com/ontio/ontology/vm/wasmvm/util"
)

type PType int

const (
	PInt8 PType = iota
	PInt16
	PInt32
	PInt64
	PFloat32
	PFloat64
	PString
	PStruct
	PUnkown
)

const (
	//VM_NIL_POINTER = math.MaxInt64
	VM_NIL_POINTER = 0
)

type TypeLength struct {
	Ptype  PType
	Length int
}

type VMmemory struct {
	Memory          []byte
	AllocedMemIdex  int
	PointedMemIndex int
	ParamIndex      int //args analyze pointer
	MemPoints       map[uint64]*TypeLength
}

//Alloc memory for base types, return the address in memory
func (vm *VMmemory) Malloc(size int) (int, error) {
	if vm.Memory == nil || len(vm.Memory) == 0 {
		return 0, errors.New("memory is not initialized")
	}
	if vm.AllocedMemIdex+size+1 > len(vm.Memory) {
		return 0, errors.New("memory out of bound")
	}

	if vm.AllocedMemIdex+size+1 > vm.PointedMemIndex {
		return 0, errors.New("memory out of bound")
	}

	offset := vm.AllocedMemIdex + 1
	vm.AllocedMemIdex += size

	return offset, nil
}

//Alloc memory for pointer types, return the address in memory
func (vm *VMmemory) MallocPointer(size int, p_type PType) (int, error) {
	if vm.Memory == nil || len(vm.Memory) == 0 {
		return 0, errors.New("memory is not initialized")
	}
	if vm.PointedMemIndex+size > len(vm.Memory) {
		return 0, errors.New("memory out of bound")
	}

	offset := vm.PointedMemIndex + 1
	vm.PointedMemIndex += size
	//save the point and length
	vm.MemPoints[uint64(offset)] = &TypeLength{Ptype: p_type, Length: size}
	return offset, nil
}

func (vm *VMmemory) copyMemAndGetIdx(b []byte, p_type PType) (int, error) {
	idx, err := vm.MallocPointer(len(b), p_type)
	if err != nil {
		return 0, err
	}
	copy(vm.Memory[idx:idx+len(b)], b)

	return idx, nil
}

//return pointed memory size
func (vm *VMmemory) GetPointerMemSize(addr uint64) int {
	//nil case
	if addr == uint64(VM_NIL_POINTER) {
		return 0
	}

	v, ok := vm.MemPoints[addr]
	if ok {
		return v.Length
	} else {
		return 0
	}
}

//return pointed memory
//when wasm returns a pointer, call this function to get the pointed memory
func (vm *VMmemory) GetPointerMemory(addr uint64) ([]byte, error) {
	//nil case
	if addr == uint64(VM_NIL_POINTER) {
		return nil, nil
	}

	length := vm.GetPointerMemSize(addr)
	if length == 0 {
		return nil,nil
	}

	if int(addr)+length > len(vm.Memory) {
		return nil, errors.New("memory out of bound")
	} else {
		return vm.Memory[int(addr) : int(addr)+length], nil
	}
}

//set pointer types into memory, return address of memory
func (vm *VMmemory) SetPointerMemory(val interface{}) (int, error) {

	////nil case
	if val == nil {
		return VM_NIL_POINTER, nil
	}

	switch reflect.TypeOf(val).Kind() {
	case reflect.String:
		b := []byte(val.(string))
		return vm.copyMemAndGetIdx(b, PString)
	case reflect.Array, reflect.Struct, reflect.Ptr:
		//todo  implement
		return 0, nil
	case reflect.Slice:
		switch val.(type) {
		case []byte:
			return vm.copyMemAndGetIdx(val.([]byte), PString)

		case []int:
			intBytes := make([]byte, len(val.([]int))*4)
			for i, v := range val.([]int) {
				tmp := make([]byte, 4)
				binary.LittleEndian.PutUint32(tmp, uint32(v))
				copy(intBytes[i*4:(i+1)*4], tmp)
			}
			return vm.copyMemAndGetIdx(intBytes, PInt32)
		case []int64:
			intBytes := make([]byte, len(val.([]int64))*8)
			for i, v := range val.([]int64) {
				tmp := make([]byte, 8)
				binary.LittleEndian.PutUint64(tmp, uint64(v))
				copy(intBytes[i*8:(i+1)*8], tmp)
			}
			return vm.copyMemAndGetIdx(intBytes, PInt64)

		case []float32:
			floatBytes := make([]byte, len(val.([]float32))*4)
			for i, v := range val.([]float32) {
				tmp := util.Float32ToBytes(v)
				copy(floatBytes[i*4:(i+1)*4], tmp)
			}
			return vm.copyMemAndGetIdx(floatBytes, PFloat32)

		case []float64:
			floatBytes := make([]byte, len(val.([]float64))*4)
			for i, v := range val.([]float64) {
				tmp := util.Float64ToBytes(v)
				copy(floatBytes[i*8:(i+1)*8], tmp)
			}
			return vm.copyMemAndGetIdx(floatBytes, PFloat64)

		case []string:
			sbytes := make([]byte,len(val.([]string))*4)  //address is 4 bytes
			for i,s := range val.([]string) {
				idx,err := vm.SetPointerMemory(s)
				if err != nil{
					return 0,err
				}
				tmp := make([]byte,4)
				binary.LittleEndian.PutUint32(tmp,uint32(idx))
				copy(sbytes[i*4:(i+1)*4], tmp)
			}
			return vm.copyMemAndGetIdx(sbytes, PInt32)

		case [][]byte:
			bbytes := make([]byte,len(val.([][]byte))*4)  //address is 4 bytes
			for i,b := range val.([][]byte) {
				idx,err := vm.SetPointerMemory(b)
				if err != nil{
					return 0,err
				}
				tmp := make([]byte,4)
				binary.LittleEndian.PutUint32(tmp,uint32(idx))
				copy(bbytes[i*4:(i+1)*4], tmp)
			}
			return vm.copyMemAndGetIdx(bbytes, PInt32)


		default:
			return 0, errors.New("Not supported slice type")
		}

	default:
		return 0, errors.New("not supported type")
	}

}

//set struct into memory , return address of memory
func (vm *VMmemory) SetStructMemory(val interface{}) (int, error) {

	if reflect.TypeOf(val).Kind() != reflect.Struct {
		return 0, errors.New("SetStructMemory :input is not a struct")
	}
	valref := reflect.ValueOf(val)
	//var totalsize = 0
	var index = 0
	for i := 0; i < valref.NumField(); i++ {
		field := valref.Field(i)

		//nested struct case
		if reflect.TypeOf(field.Type()).Kind() == reflect.Struct {
			idx, err := vm.SetStructMemory(field)
			if err != nil {
				return 0, err
			} else {
				if i == 0 && index == 0 {
					index = idx
				}
			}
		} else {
			var fieldVal interface{}
			//todo how to determine the value is int or int64
			var idx int
			var err error
			switch field.Kind() {
			case reflect.Int, reflect.Int32, reflect.Uint, reflect.Uint32:
				fieldVal = int(field.Int())
				idx, err = vm.SetMemory(fieldVal)
			case reflect.Int64, reflect.Uint64:
				fieldVal = field.Int()
				idx, err = vm.SetMemory(fieldVal)
			case reflect.Float32, reflect.Float64:
				fieldVal = field.Float()
				idx, err = vm.SetMemory(fieldVal)
			case reflect.String:
				fieldVal = field.String()
				tmp, err := vm.SetPointerMemory(fieldVal)
				if err != nil {
					return 0, err
				}
				//add the point address to memory
				idx, err = vm.SetMemory(tmp)

			case reflect.Slice:
				//fieldVal = field.Interface()
				//TODO note the struct field MUST be public
				tmp, err := vm.SetPointerMemory(fieldVal)
				if err != nil {
					return 0, err
				}
				//add the point address to memory
				idx, err = vm.SetMemory(tmp)
			}

			if err != nil {
				return 0, err
			} else {
				if i == 0 && index == 0 {
					index = idx
				}
			}
		}
	}
	return index, nil

}

//set base types into memory, return address of memory
func (vm *VMmemory) SetMemory(val interface{}) (int, error) {

	switch val.(type) {
	case string: //use SetPointerMemory for string
		return vm.SetPointerMemory(val.(string))
	case int:
		tmp := make([]byte, 4)
		binary.LittleEndian.PutUint32(tmp, uint32(val.(int)))
		idx, err := vm.Malloc(len(tmp))
		if err != nil {
			return 0, err
		}
		copy(vm.Memory[idx:idx+len(tmp)], tmp)
		return idx, nil
	case int64:
		tmp := make([]byte, 8)
		binary.LittleEndian.PutUint64(tmp, uint64(val.(int64)))
		idx, err := vm.Malloc(len(tmp))
		if err != nil {
			return 0, err
		}
		copy(vm.Memory[idx:idx+len(tmp)], tmp)
		return idx, nil
	case float32:
		tmp := util.Float32ToBytes(val.(float32))

		idx, err := vm.Malloc(len(tmp))
		if err != nil {
			return 0, err
		}
		copy(vm.Memory[idx:idx+len(tmp)], tmp)
		return idx, nil
	case float64:
		tmp := util.Float64ToBytes(val.(float64))
		idx, err := vm.Malloc(len(tmp))
		if err != nil {
			return 0, err
		}
		copy(vm.Memory[idx:idx+len(tmp)], tmp)
		return idx, nil

	default:
		return 0, errors.New("not supported type")
	}
}
