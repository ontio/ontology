package neovm

import (
	"github.com/Ontology/vm/neovm/types"
)

type Element interface {
	GetStackItem() types.StackItemInterface
	GetExecutionContext() *ExecutionContext
}

type RandomAccessStack struct {
	e []Element
}

func NewRandAccessStack() *RandomAccessStack {
	var ras RandomAccessStack
	ras.e = make([]Element, 0)
	return &ras
}

func (r *RandomAccessStack) Count() int {
	return len(r.e)
}

func (r *RandomAccessStack) Insert(index int, t Element) {
	if t == nil {
		return
	}
	l := len(r.e)
	if index > l {
		return
	}
	var array []Element
	index = l - index
	array = append(array, r.e[:index]...)
	array = append(array, t)
	array = append(array, r.e[index:]...)
	r.e = array
}

func (r *RandomAccessStack) Peek(index int) Element {
	l := len(r.e)
	if index >= l {
		return nil
	}
	index = l - index
	return r.e[index - 1]
}

func (r *RandomAccessStack) Remove(index int) Element {
	l := len(r.e)
	if index >= l {
		return nil
	}
	index = l - index
	e := r.e[index - 1]
	var si []Element
	si = append(r.e[:index - 1], r.e[index:]...)
	r.e = si
	return e
}

func (r *RandomAccessStack) Set(index int, t Element) {
	l := len(r.e)
	if index >= l {
		return
	}
	r.e[index] = t
}

func (r *RandomAccessStack) Push(t Element) {
	r.Insert(0, t)
}

func (r *RandomAccessStack) Pop() Element {
	return r.Remove(0)
}

func (r *RandomAccessStack) Swap(i, j int) {
	l := len(r.e)
	r.e[l - i - 1], r.e[l - j - 1] = r.e[l - j - 1], r.e[l - i - 1]
}
