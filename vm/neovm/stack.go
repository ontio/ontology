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
