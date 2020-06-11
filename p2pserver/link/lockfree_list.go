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

package link

import (
	"sync/atomic"
	"unsafe"
)

var sealed = unsafe.Pointer(&innerNode{}) // dummy node

// LockFreeList is a mpmc List support take batch nodes at oneshot
type LockFreeList struct {
	head unsafe.Pointer
}

type OwnedList struct {
	head unsafe.Pointer
}

type innerNode struct {
	next unsafe.Pointer
	data interface{}
}

func (self *LockFreeList) Push(data interface{}) bool {
	node := &innerNode{data: data}
	for {
		head := atomic.LoadPointer(&self.head)
		if head == sealed {
			return false
		}
		node.next = head
		if atomic.CompareAndSwapPointer(&self.head, head, unsafe.Pointer(node)) {
			return true
		}
	}
}

func (self *LockFreeList) Sealed() bool {
	return atomic.LoadPointer(&self.head) == sealed
}

// return list contains appended node and sealed state
func (self *LockFreeList) Take() (*OwnedList, bool) {
	list := atomic.LoadPointer(&self.head)
	for {
		if list == sealed || atomic.CompareAndSwapPointer(&self.head, list, nil) {
			break
		}
		list = atomic.LoadPointer(&self.head)
	}

	return &OwnedList{head: list}, list == sealed
}

func (self *OwnedList) Pop() interface{} {
	head := self.head
	if head == nil || head == sealed {
		return nil
	}

	node := (*innerNode)(head)
	self.head = node.next

	return node.data
}

func (self *LockFreeList) TakeAndSeal() *OwnedList {
	return &OwnedList{head: atomic.SwapPointer(&self.head, sealed)}
}
