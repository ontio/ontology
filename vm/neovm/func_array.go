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
	index := PopStackItem(e)
	items := PopStackItem(e)

	switch items.(type) {
	case *types.Array:
		i := int(index.GetBigInteger().Int64())
		if i < 0 || i >= len(items.GetArray()) {
			return FAULT, errors.NewErr("opPickItem invalid array.")
		}
		PushData(e, items.GetArray()[i])
	case *types.Struct:
		i := int(index.GetBigInteger().Int64())
		if i < 0 || i >= len(items.GetStruct()) {
			return FAULT, errors.NewErr("opPickItem invalid array.")
		}
		PushData(e, items.GetStruct()[i])
	case *types.Map:
		value := items.(*types.Map).TryGetValue(index)
		//TODO should return a nil type when not exist?
		if value == nil {
			return FAULT, errors.NewErr("opPickItem map element not exist.")
		}
		PushData(e, value)

	default:
		return FAULT, errors.NewErr("opPickItem unknown item type.")
	}

	return NONE, nil
}

func opSetItem(e *ExecutionEngine) (VMState, error) {
	newItem := PopStackItem(e)
	if value, ok := newItem.(*types.Struct); ok {
		newItem = value.Clone()
	}

	index := PopStackItem(e)
	item := PopStackItem(e)

	switch item.(type) {
	case *types.Map:
		item.GetMap()[index] = newItem
	case *types.Array:
		items := item.GetArray()
		i := int(index.GetBigInteger().Int64())
		if i < 0 || i >= len(items) {
			return FAULT, errors.NewErr("opSetItem invalid array.")
		}
		items[i] = newItem
	case *types.Struct:
		items := item.GetStruct()
		i := int(index.GetBigInteger().Int64())
		if i < 0 || i >= len(items) {
			return FAULT, errors.NewErr("opSetItem invalid array.")
		}
		items[i] = newItem
	default:
		return FAULT, errors.NewErr("opSetItem unknown item type.")
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
	items := PopStackItem(e)
	if item, ok := items.(*types.Array); ok {
		item.Add(newItem)
	}
	if item, ok := items.(*types.Struct); ok {
		item.Add(newItem)
	}
	return NONE, nil
}

func opReverse(e *ExecutionEngine) (VMState, error) {
	itemArr := PopArray(e)
	for i, j := 0, len(itemArr)-1; i < j; i, j = i+1, j-1 {
		itemArr[i], itemArr[j] = itemArr[j], itemArr[i]
	}
	return NONE, nil
}
