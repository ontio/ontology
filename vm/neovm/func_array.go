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

package neovm

import (
	"math/big"

	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/vm/neovm/types"
	"fmt"
)

func opArraySize(e *ExecutionEngine) (VMState, error) {
	item := PopStackItem(e)
	if _, ok := item.(*types.Array); ok {
		PushData(e, len(item.GetArray()))
	} else {
		PushData(e, len(item.GetByteArray()))
	}

	return NONE, nil
}

func opPack(e *ExecutionEngine) (VMState, error) {
	size := PopInt(e)
	var items []types.StackItems
	for i := 0; i < size; i++ {
		items = append(items, PopStackItem(e))
	}
	PushData(e, items)
	return NONE, nil
}

func opUnpack(e *ExecutionEngine) (VMState, error) {
	arr := PopArray(e)
	l := len(arr)
	for i := l - 1; i >= 0; i-- {
		Push(e, arr[i])
	}
	PushData(e, l)
	return NONE, nil
}

func opPickItem(e *ExecutionEngine) (VMState, error) {
	//index := PopInt(e)
	//items := PopArray(e)
	//PushData(e, items[index])
	//return NONE, nil

	key := PopStackItem(e)
	item := PopStackItem(e)

	switch item.(type) {
	case *types.Array:
		index := int(key.(*types.Integer).GetBigInteger().Int64())
		//arr := item.(*types.Array).GetArray()
		//arr := item.GetArray()
		if index < 0 || index >= len(item.GetArray()) {
			return NONE, errors.NewErr("invalid array.")
		}
		//PushData(e, arr[index])
		PushData(e, item.GetArray()[index])
	case *types.Map:
		mp := item.(*types.Map)
		v := mp.TryGetValue(key)
		if v == nil {
			return NONE, errors.NewErr("invalid map element.")
		}
		PushData(e, v)
	default:
		return NONE, errors.NewErr("invalid item type.")
	}
	return NONE, nil
}

func opSetItem(e *ExecutionEngine) (VMState, error) {
	//newItem := PopStackItem(e)
	//if value, ok := newItem.(*types.Struct); ok {
	//	newItem = value.Clone()
	//}
	//index := PopInt(e)
	//items := PopArray(e)
	//items[index] = newItem
	//return NONE, nil

	fmt.Printf("======call setItem====\n")
	newItem := PopStackItem(e)
	if value, ok := newItem.(*types.Struct); ok {
		newItem = value.Clone()
	}

	key := PopStackItem(e)

	item := PopStackItem(e)
	switch item.(type) {
	case *types.Array:
		fmt.Printf("======call setItem type is array====\n")
		index := int(key.(*types.Integer).GetBigInteger().Int64())
		fmt.Printf("=====set item index = %d\n", index)
		//arr := item.(*types.Array).GetArray()
		arr := item.GetArray()
		fmt.Printf("======setItem 1====\n")
		if index < 0 || index >= len(arr) {
			fmt.Printf("======setItem 3====\n")
			return NONE, errors.NewErr("invalid array.")
		}
		fmt.Printf("======setItem 2====\n")
		arr[index] = newItem
		fmt.Printf("======setItem 4====\n")
	case *types.Map:
		fmt.Printf("======call setItem type is Map====\n")
		mp := item.(*types.Map).GetMap()
		mp[key] = newItem
	default:
		return NONE, errors.NewErr("invalid item type.")
	}

	return NONE, nil
}

func opNewArray(e *ExecutionEngine) (VMState, error) {
	count := PopInt(e)
	var items []types.StackItems
	for i := 0; i < count; i++ {
		items = append(items, types.NewBoolean(false))
	}
	PushData(e, types.NewArray(items))
	return NONE, nil
}

func opNewStruct(e *ExecutionEngine) (VMState, error) {
	count := PopBigInt(e)
	var items []types.StackItems
	for i := 0; count.Cmp(big.NewInt(int64(i))) > 0; i++ {
		items = append(items, types.NewBoolean(false))
	}
	PushData(e, types.NewStruct(items))
	return NONE, nil
}

func opNewMap(e *ExecutionEngine) (VMState, error) {
	PushData(e, types.NewMap())
	return NONE, nil
}

func opAppend(e *ExecutionEngine) (VMState, error) {
	newItem := PopStackItem(e)
	if value, ok := newItem.(*types.Struct); ok {
		newItem = value.Clone()
	}
	itemArr := PopArray(e)
	itemArr = append(itemArr, newItem)
	return NONE, nil
}

func opReverse(e *ExecutionEngine) (VMState, error) {
	itemArr := PopArray(e)
	for i, j := 0, len(itemArr)-1; i < j; i, j = i+1, j-1 {
		itemArr[i], itemArr[j] = itemArr[j], itemArr[i]
	}
	return NONE, nil
}
