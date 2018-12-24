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
		a, err := item.GetArray()
		if err != nil {
			return FAULT, err
		}
		PushData(e, len(a))
	} else {
		b, err := item.GetByteArray()
		if err != nil {
			return FAULT, err
		}
		PushData(e, len(b))
	}

	return NONE, nil
}

func opPack(e *ExecutionEngine) (VMState, error) {
	size, err := PopInt(e)
	if err != nil {
		return FAULT, err
	}
	var items []types.StackItems
	for i := 0; i < size; i++ {
		items = append(items, PopStackItem(e))
	}
	PushData(e, items)
	return NONE, nil
}

func opUnpack(e *ExecutionEngine) (VMState, error) {
	arr, err := PopArray(e)
	if err != nil {
		return FAULT, err
	}
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
		bi, _ := index.GetBigInteger()
		i := int(bi.Int64())
		a, _ := items.GetArray()
		PushData(e, a[i])
	case *types.Struct:
		bi, _ := index.GetBigInteger()
		i := int(bi.Int64())
		s, _ := items.GetStruct()
		PushData(e, s[i])
	case *types.Map:
		PushData(e, items.(*types.Map).TryGetValue(index))
	case *types.ByteArray:
		bi, _ := index.GetBigInteger()
		i := int(bi.Int64())
		a, _ := items.GetByteArray()
		PushData(e, a[i])
	}

	return NONE, nil
}

func opSetItem(e *ExecutionEngine) (VMState, error) {
	newItem := PopStackItem(e)
	if value, ok := newItem.(*types.Struct); ok {
		var err error
		newItem, err = value.Clone()
		if err != nil {
			return FAULT, err
		}
	}

	index := PopStackItem(e)
	item := PopStackItem(e)

	switch item.(type) {
	case *types.Map:
		m := item.(*types.Map)
		m.Add(index, newItem)
	case *types.Array:
		items, _ := item.GetArray()
		bi, _ := index.GetBigInteger()
		i := int(bi.Int64())
		items[i] = newItem
	case *types.Struct:
		items, _ := item.GetStruct()
		bi, _ := index.GetBigInteger()
		i := int(bi.Int64())
		items[i] = newItem
	}

	return NONE, nil
}

func opNewArray(e *ExecutionEngine) (VMState, error) {
	count, _ := PopInt(e)
	var items []types.StackItems
	for i := 0; i < count; i++ {
		items = append(items, types.NewBoolean(false))
	}
	PushData(e, types.NewArray(items))
	return NONE, nil
}

func opNewStruct(e *ExecutionEngine) (VMState, error) {
	count, _ := PopBigInt(e)
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
		var err error
		newItem, err = value.Clone()
		if err != nil {
			return FAULT, err
		}
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
	itemArr, _ := PopArray(e)
	for i, j := 0, len(itemArr)-1; i < j; i, j = i+1, j-1 {
		itemArr[i], itemArr[j] = itemArr[j], itemArr[i]
	}
	return NONE, nil
}

func opRemove(e *ExecutionEngine) (VMState, error) {
	index := PopStackItem(e)
	item := PopStackItem(e)

	switch item.(type) {
	case *types.Map:
		m := item.(*types.Map)
		m.Remove(index)
	case *types.Array:
		m, err := item.GetArray()
		if err != nil {
			return FAULT, errors.NewErr("[opRemove]get Array error!")
		}

		i, err := index.GetBigInteger()
		if err != nil {
			return FAULT, errors.NewErr("[opRemove] index not a interger!")
		}

		if i.Sign() < 0 {
			return FAULT, errors.NewErr("[opRemove] index out of bound!")
		}

		len_t := big.NewInt(int64(len(m)))
		if len_t.Cmp(i) <= 0 {
			return FAULT, errors.NewErr("[opRemove] index out of bound!")
		}

		ii := i.Int64()
		item.(*types.Array).RemoveAt(int(ii) + 1)
	default:
		return FAULT, errors.NewErr("Not a supported remove type")
	}

	return NONE, nil
}

func opHasKey(e *ExecutionEngine) (VMState, error) {
	key := PopStackItem(e)
	item := PopStackItem(e)

	switch item.(type) {
	case *types.Map:
		v := item.(*types.Map).TryGetValue(key)

		ok := false
		if v != nil {
			ok = true
		}

		PushData(e, ok)
	default:
		return FAULT, errors.NewErr("Not a supported haskey type")
	}
	return NONE, nil
}

func opKeys(e *ExecutionEngine) (VMState, error) {
	item := PopStackItem(e)
	switch item.(type) {
	case *types.Map:
		keys, err := item.(*types.Map).GetMapSortedKey()
		if err != nil {
			return FAULT, err
		}

		PushData(e, types.NewArray(keys))
	default:
		return FAULT, errors.NewErr("Not a supported keys type")
	}
	return NONE, nil
}

func opValues(e *ExecutionEngine) (VMState, error) {
	item := PopStackItem(e)
	switch item.(type) {
	case *types.Map:
		mapitem, err := item.GetMap()
		if err != nil {
			return FAULT, err
		}

		values := make([]types.StackItems, len(mapitem))
		keys, err := item.(*types.Map).GetMapSortedKey()
		if err != nil {
			return FAULT, err
		}

		for j, v := range keys {
			values[j] = mapitem[v]
		}

		PushData(e, types.NewArray(values))
	default:
		return FAULT, errors.NewErr("Not a supported values type")
	}
	return NONE, nil
}
