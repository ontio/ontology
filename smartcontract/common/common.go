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

package common

import (
	"errors"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/vm/neovm/types"
)

// ConvertReturnTypes return neovm stack element value
// According item types convert to hex string value
// Now neovm support type contain: ByteArray/Integer/Boolean/Array/Struct/Interop/StackItems
const (
	MAX_COUNT = 1024
)

func ConvertNeoVmTypeHexString(item types.StackItems) (interface{}, error) {
	var count int
	return convertNeoVmTypeHexString(item, &count)
}

func convertNeoVmTypeHexString(item types.StackItems, count *int) (interface{}, error) {
	if item == nil {
		return nil, nil
	}
	if *count > MAX_COUNT {
		return nil, errors.New("over max parameters convert length")
	}
	switch v := item.(type) {
	case *types.ByteArray:
		arr, _ := v.GetByteArray()
		return common.ToHexString(arr), nil
	case *types.Integer:
		i, _ := v.GetBigInteger()
		if i.Sign() == 0 {
			return common.ToHexString([]byte{0}), nil
		} else {
			return common.ToHexString(common.BigIntToNeoBytes(i)), nil
		}
	case *types.Boolean:
		b, _ := v.GetBoolean()
		if b {
			return common.ToHexString([]byte{1}), nil
		} else {
			return common.ToHexString([]byte{0}), nil
		}
	case *types.Array:
		var arr []interface{}
		ar, _ := v.GetArray()
		for _, val := range ar {
			*count++
			cv, err := convertNeoVmTypeHexString(val, count)
			if err != nil {
				return nil, err
			}
			arr = append(arr, cv)
		}
		return arr, nil
	case *types.Struct:
		var arr []interface{}
		ar, _ := v.GetStruct()
		for _, val := range ar {
			*count++
			cv, err := convertNeoVmTypeHexString(val, count)
			if err != nil {
				return nil, err
			}
			arr = append(arr, cv)
		}
		return arr, nil
	case *types.Interop:
		it, _ := v.GetInterface()
		return common.ToHexString(it.ToArray()), nil
	default:
		log.Error("[ConvertTypes] Invalid Types!")
		return nil, errors.New("[ConvertTypes] Invalid Types!")
	}
}

//only for debug/testing
func Stringify(item types.StackItems) (string, error) {
	var count int
	return stringify(item, &count)
}
func stringify(item types.StackItems, count *int) (string, error) {
	if item == nil {
		return "", nil
	}
	if *count > MAX_COUNT {
		return "", errors.New("over max parameters convert length")
	}
	switch v := item.(type) {
	case *types.Boolean, *types.ByteArray, *types.Integer:
		b, err := item.GetByteArray()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("bytes(hex:%x)", b), nil
	case *types.Array:
		arr, err := v.GetArray()
		if err != nil {
			return "", nil
		}
		data := ""
		for _, v := range arr {
			*count++
			s, err := stringify(v, count)
			if err != nil {
				return "", err
			}
			data += s + ", "
		}
		return fmt.Sprintf("array[%d]{%s}", len(arr), data), nil
	case *types.Map:
		m, err := v.GetMap()
		if err != nil {
			return "", err
		}
		data := ""
		sortedKey, err := v.GetMapSortedKey()
		if err != nil {
			return "", err
		}
		for _, key := range sortedKey {
			value := m[key]
			*count++
			val, err := stringify(value, count)
			if err != nil {
				return "", nil
			}
			data += fmt.Sprintf("%x: %s,", key, val)
		}
		return fmt.Sprintf("map[%d]{%s}", len(m), data), nil
	case *types.Struct:
		s, err := v.GetStruct()
		if err != nil {
			return "", err
		}
		data := ""
		for _, v := range s {
			*count++
			vs, err := stringify(v, count)
			if err != nil {
				return "", nil
			}
			data += vs + ", "
		}
		return fmt.Sprintf("struct[%d]{%s}", len(s), data), nil
	default:
		return "", fmt.Errorf("[Stringify] Invalid Types!")
	}
}

//only for debug/testing
func Dump(item types.StackItems) (string, error) {
	var count int
	return dump(item, &count)
}
func dump(item types.StackItems, count *int) (string, error) {
	if item == nil {
		return "", nil
	}
	if *count > MAX_COUNT {
		return "", errors.New("over max parameters convert length")
	}
	switch v := item.(type) {
	case *types.Boolean:
		b, err := v.GetBoolean()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("bool(%v)", b), nil
	case *types.ByteArray:
		b, err := v.GetByteArray()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("bytes(hex:%x)", b), nil
	case *types.Integer:
		b, err := v.GetBigInteger()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("int(%d)", b), nil
	case *types.Array:
		arr, err := v.GetArray()
		if err != nil {
			return "", nil
		}
		data := ""
		for _, v := range arr {
			*count++
			s, err := dump(v, count)
			if err != nil {
				return "", err
			}
			data += s + ", "
		}
		return fmt.Sprintf("array[%d]{%s}", len(arr), data), nil
	case *types.Map:
		m, err := v.GetMap()
		if err != nil {
			return "", err
		}
		data := ""
		sortedKey, err := v.GetMapSortedKey()
		if err != nil {
			return "", err
		}
		for _, key := range sortedKey {
			value := m[key]
			*count++
			val, err := dump(value, count)
			if err != nil {
				return "", nil
			}
			data += fmt.Sprintf("%x: %s,", key, val)
		}
		return fmt.Sprintf("map[%d]{%s}", len(m), data), nil
	case *types.Struct:
		s, err := v.GetStruct()
		if err != nil {
			return "", err
		}
		data := ""
		for _, v := range s {
			*count++
			vs, err := dump(v, count)
			if err != nil {
				return "", nil
			}
			data += vs + ", "
		}
		return fmt.Sprintf("struct[%d]{%s}", len(s), data), nil
	default:
		return "", fmt.Errorf("[Dump] Invalid Types!")
	}
}
