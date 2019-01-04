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
	"sort"

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

func ConvertNeoVmTypeHexString(item interface{}) (interface{}, error) {
	var count int
	return convertNeoVmTypeHexString(item, &count)
}

func convertNeoVmTypeHexString(item interface{}, count *int) (interface{}, error) {
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

func Dump(item interface{}) (string, error) {
	var count int
	return dump(item, &count)
}
func dump(item interface{}, count *int) (string, error) {
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
	case types.ByteArray:
		bs, err := v.GetByteArray()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("bytes(hex:%x)", bs), nil
	case types.Integer:
		b, err := v.GetBigInteger()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("int(%d)", b), nil
	case types.Array:
		arr, err := v.GetArray()
		if err != nil {
			return "", nil
		}
		data := ""
		for _, v := range arr {
			s, err := Dump(v)
			if err != nil {
				return "", err
			}
			data += s + ", "
		}
		return fmt.Sprintf("array[%d]{%s}", len(arr), data), nil
	case types.Map:
		m, err := v.GetMap()
		if err != nil {
			return "", err
		}
		var unsortKey []string
		for k, _ := range m {
			s, err := Dump(k)
			if err != nil {
				return "", err
			}
			unsortKey = append(unsortKey, s)
		}
		sort.Strings(unsortKey)
		data := ""
		for _, key := range unsortKey {

			data += fmt.Sprintf("%x: %s,", key)
		}
	}
	return "", nil
}
