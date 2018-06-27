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
	"math/big"
	"reflect"

	"github.com/ontio/ontology/vm/neovm/interfaces"
)

const (
	MAX_STRCUT_DEPTH = 1024
)

type Struct struct {
	_array []StackItems
}

func NewStruct(value []StackItems) *Struct {
	var this Struct
	this._array = value
	return &this
}

func (this *Struct) Equals(other StackItems) bool {

	if this == other {
		return true
	}

	oa, err := other.GetStruct()
	if err != nil {
		return false
	}

	return reflect.DeepEqual(this._array, oa)
}

func (this *Struct) GetBigInteger() (*big.Int, error) {
	return nil, fmt.Errorf("%s", "Not support struct to integer")
}

func (this *Struct) GetBoolean() (bool, error) {
	return true, nil
}

func (this *Struct) GetByteArray() ([]byte, error) {
	return nil, fmt.Errorf("%s", "Not support struct to byte array")
}

func (this *Struct) GetInterface() (interfaces.Interop, error) {
	return nil, fmt.Errorf("%s", "Not support struct to interface")
}

func (s *Struct) GetArray() ([]StackItems, error) {
	return s._array, nil
}

func (s *Struct) GetStruct() ([]StackItems, error) {
	return s._array, nil
}

func (s *Struct) Clone() (StackItems, error) {
	if checkStructRef(s, make(map[uintptr]bool), 0) {
		return nil, fmt.Errorf("%s", "struct contain self reference or over max depth!")
	}
	return clone(s)
}

func clone(s *Struct) (StackItems, error) {
	var arr []StackItems
	for _, v := range s._array {
		if value, ok := v.(*Struct); ok {
			vc, err := clone(value)
			if err != nil {
				return nil, err
			}
			arr = append(arr, vc)
		} else {
			arr = append(arr, v)
		}
	}
	return &Struct{arr}, nil
}

func checkStructRef(item StackItems, visited map[uintptr]bool, depth int) bool {
	if depth > MAX_STRCUT_DEPTH {
		return true
	}
	switch item.(type) {
	case *Struct:
		p := reflect.ValueOf(item).Pointer()
		if visited[p] {
			return true
		}
		visited[p] = true
		st, _ := item.GetStruct()
		for _, v := range st {
			if checkStructRef(v, visited, depth+1) {
				return true
			}
		}
		delete(visited, p)
		return false
	default:
		return false
	}
}

func (this *Struct) GetMap() (map[StackItems]StackItems, error) {
	return nil, fmt.Errorf("%s", "Not support struct to map")
}

func (this *Struct) Add(item StackItems) {
	this._array = append(this._array, item)
}

func (this *Struct) Count() int {
	return len(this._array)
}
