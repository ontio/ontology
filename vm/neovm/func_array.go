package neovm

import (
	"github.com/Ontology/vm/neovm/types"
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
	items := NewStackItems()
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
		Push(e, NewStackItem(arr[i]))
	}
	PushData(e, l)
	return NONE, nil
}

func opPickItem(e *ExecutionEngine) (VMState, error) {
	index := PopInt(e)
	items := PopArray(e)
	PushData(e, items[index])
	return NONE, nil
}

func opSetItem(e *ExecutionEngine) (VMState, error) {
	newItem := PopStackItem(e)
	index := PopInt(e)
	itemArr := PopStackItem(e)
	if _, ok := itemArr.(*types.Array); ok {
		items := itemArr.GetArray()
		items[index] = newItem
	} else {
		items := itemArr.GetByteArray()
		items[index] = newItem.GetByteArray()[0]
	}
	return NONE, nil
}

func opNewArray(e *ExecutionEngine) (VMState, error) {
	count := PopInt(e)
	items := NewStackItems();
	for i := 0; i < count; i++ {
		items = append(items, types.NewBoolean(false))
	}
	PushData(e, items)
	return NONE, nil
}


