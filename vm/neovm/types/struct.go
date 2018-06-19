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

type Struct struct {
	_array []StackItems
}

func NewStruct(value []StackItems) *Struct {
	var this Struct
	this._array = value
	return &this
}

func (this *Struct) Equals(other StackItems) bool {
	return reflect.DeepEqual(this, other)
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

func (s *Struct) Clone() StackItems {
	var arr []StackItems
	for _, v := range s._array {
		if value, ok := v.(*Struct); ok {
			arr = append(arr, value.Clone())
		} else {
			arr = append(arr, v)
		}
	}
	return &Struct{arr}
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
