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
	"fmt"
	"reflect"

	"github.com/Ontology/common"
	"github.com/Ontology/vm/neovm/types"
)

type States struct {
	Value interface{}
}

func ConvertTypes(item types.StackItems) (results []States) {
	if item == nil {
		return
	}
	switch v := item.(type) {
	case *types.ByteArray:
		results = append(results, States{common.ToHexString(v.GetByteArray())})
	case *types.Integer:
		if v.GetBigInteger().Sign() == 0 {
			results = append(results, States{common.ToHexString([]byte{0})})
		} else {
			results = append(results, States{common.ToHexString(types.ConvertBigIntegerToBytes(v.GetBigInteger()))})
		}
	case *types.Boolean:
		if v.GetBoolean() {
			results = append(results, States{common.ToHexString([]byte{1})})
		} else {
			results = append(results, States{common.ToHexString([]byte{0})})
		}
	case *types.Array:
		var arr []States
		for _, val := range v.GetArray() {
			arr = append(arr, ConvertTypes(val)...)
		}
		results = append(results, States{arr})
	case *types.Interop:
		results = append(results, States{common.ToHexString(v.GetInterface().ToArray())})
	case types.StackItems:
		ConvertTypes(v)
	default:
		panic(fmt.Sprintf("[ConvertTypes] Invalid Types: %v", reflect.TypeOf(v)))
	}
	return
}

func ConvertReturnTypes(item types.StackItems) (results []interface{}) {
	if item == nil {
		return
	}
	switch v := item.(type) {
	case *types.ByteArray:
		results = append(results, common.ToHexString(v.GetByteArray()))
	case *types.Integer:
		results = append(results, v.GetBigInteger())
	case *types.Boolean:
		results = append(results, v.GetBoolean())
	case *types.Array:
		var arr []interface{}
		for _, val := range v.GetArray() {
			arr = append(arr, ConvertReturnTypes(val)...)
		}
		results = append(results, arr)
	case *types.Interop:
		results = append(results, common.ToHexString(v.GetInterface().ToArray()))
	case types.StackItems:
		ConvertTypes(v)
	default:
		panic("[ConvertTypes] Invalid Types!")
	}
	return
}
