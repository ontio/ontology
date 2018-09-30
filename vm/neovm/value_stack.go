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
	"github.com/ontio/ontology/vm/neovm/errors"
)

const initialStackCap = 64 // to avoid reallocation

type ValueStack struct {
	data []types.VmValue
	limit int
}

func NewValueStack(limit int) *ValueStack {
	return &ValueStack{
		data: make([]types.VmValue, 0, initialStackCap),
		limit:limit,
	}
}

func (self *ValueStack) Count() int {
	return len(self.data)
}

func (self *ValueStack) Insert(index int, t types.VmValue) error {
	l := len(self.data)
	if l >= self.limit {
		return errors.ERR_OVER_LIMIT_STACK
	}
	if index > l || index < 0 {
		return errors.ERR_INDEX_OUT_OF_BOUND
	}
	index = l - index
	self.data = append(self.data, self.data[l-1])
	copy(self.data[index+1:l], self.data[index:])
	self.data[index] = t
	return nil
}

func (self *ValueStack) Peek(index int) (value types.VmValue, err error) {
	l := len(self.data)
	if index >= l {
		err = errors.ERR_INDEX_OUT_OF_BOUND
		return
	}
	index = l - index
	value =  self.data[index-1]
	return
}

func (self *ValueStack) Remove(index int) (value types.VmValue, err error) {
	l := len(self.data)
	if index >= l || index < 0 {
		err = errors.ERR_INDEX_OUT_OF_BOUND
		return
	}
	index = l - index
	value = self.data[index-1]
	self.data = append(self.data[:index-1], self.data[index:]...)
	return
}

func (self *ValueStack) Set(index int, t types.VmValue) error {
	l := len(self.data)
	if index >= l || index < 0 {
		return errors.ERR_INDEX_OUT_OF_BOUND
	}
	self.data[index] = t
	return nil
}

func (self *ValueStack) Push(t types.VmValue) error {
	if len(self.data) >= self.limit {
		return errors.ERR_OVER_STACK_LEN
	}
	self.data = append(self.data, t)
	return nil
}

func (self *ValueStack) Pop() (value types.VmValue, err error)  {
	length := len(self.data)
	if length == 0 {
		err = errors.ERR_INDEX_OUT_OF_BOUND
		return
	}
	value = self.data[length-1]
	self.data = self.data[:length-1]
	return
}

func (self *ValueStack) Swap(i, j int) error {
	l := len(self.data)
	if i >= l || i < 0 {
		return errors.ERR_INDEX_OUT_OF_BOUND
	}
	if j >= l || j < 0 {
		return errors.ERR_INDEX_OUT_OF_BOUND
	}
	self.data[l-i-1], self.data[l-j-1] = self.data[l-j-1], self.data[l-i-1]

	return nil
}

func (self *ValueStack) CopyTo(stack *ValueStack) error {
	if len(self.data) + len(stack.data) > stack.limit {
		return errors.ERR_OVER_STACK_LEN
	}
	stack.data = append(stack.data, self.data...)
	return nil
}
