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
	"github.com/ontio/ontology/vm/neovm/types"
)

const initStackCap = 16 // to avoid reallocation

type RandomAccessStack struct {
	e []types.StackItems
}

func NewRandAccessStack() *RandomAccessStack {
	var ras RandomAccessStack
	ras.e = make([]types.StackItems, 0, initStackCap)
	return &ras
}

func (r *RandomAccessStack) Count() int {
	return len(r.e)
}

func (r *RandomAccessStack) Insert(index int, t types.StackItems) {
	if t == nil {
		return
	}
	l := len(r.e)
	if index > l {
		return
	}
	index = l - index
	r.e = append(r.e, r.e[l-1])
	copy(r.e[index+1:l], r.e[index:])
	r.e[index] = t
}

func (r *RandomAccessStack) Peek(index int) types.StackItems {
	l := len(r.e)
	if index >= l {
		return nil
	}
	index = l - index
	return r.e[index-1]
}

func (r *RandomAccessStack) Remove(index int) types.StackItems {
	l := len(r.e)
	if index >= l {
		return nil
	}
	index = l - index
	e := r.e[index-1]
	r.e = append(r.e[:index-1], r.e[index:]...)
	return e
}

func (r *RandomAccessStack) Set(index int, t types.StackItems) {
	l := len(r.e)
	if index >= l {
		return
	}
	r.e[index] = t
}

func (r *RandomAccessStack) Push(t types.StackItems) {
	r.e = append(r.e, t)
}

func (r *RandomAccessStack) Pop() types.StackItems {
	var res types.StackItems
	num := len(r.e)
	if num > 0 {
		res = r.e[num-1]
		r.e = r.e[:num-1]
	}
	return res
}

func (r *RandomAccessStack) Swap(i, j int) {
	l := len(r.e)
	r.e[l-i-1], r.e[l-j-1] = r.e[l-j-1], r.e[l-i-1]
}

func (r *RandomAccessStack) CopyTo(stack *RandomAccessStack) {
	stack.e = append(stack.e, r.e...)
}
