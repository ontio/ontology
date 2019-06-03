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
	"errors"
	"fmt"
)

//only for debug/testing
func Stringify(item StackItems) (string, error) {
	var count int
	return stringify(item, &count)
}
func stringify(item StackItems, count *int) (string, error) {
	if item == nil {
		return "", nil
	}
	if *count > MAX_COUNT {
		return "", errors.New("over max parameters convert length")
	}
	switch v := item.(type) {
	case *Boolean, *ByteArray, *Integer:
		b, err := item.GetByteArray()
		if err != nil {
			return "", err
		}
		if len(b) == 0 {
			b = []byte{0}
		}
		return fmt.Sprintf("bytes(hex:%x)", b), nil
	case *Array:
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
	case *Map:
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
	case *Struct:
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
func Dump(item StackItems) (string, error) {
	var count int
	return dump(item, &count)
}
func dump(item StackItems, count *int) (string, error) {
	if item == nil {
		return "", nil
	}
	if *count > MAX_COUNT {
		return "", errors.New("over max parameters convert length")
	}
	switch v := item.(type) {
	case *Boolean:
		b, err := v.GetBoolean()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("bool(%v)", b), nil
	case *ByteArray:
		b, err := v.GetByteArray()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("bytes(hex:%x)", b), nil
	case *Integer:
		b, err := v.GetBigInteger()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("int(%d)", b), nil
	case *Array:
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
	case *Map:
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
	case *Struct:
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
