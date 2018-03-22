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

// Package mpsc provides an efficient implementation of a multi-producer, single-consumer lock-free queue.
//
// The Push function is safe to call from multiple goroutines. The Pop and Empty APIs must only be
// called from a single, consumer goroutine.
//
package mpsc

// This implementation is based on http://www.1024cores.net/home/lock-free-algorithms/queues/non-intrusive-mpsc-node-based-queue

import (
	"sync/atomic"
	"unsafe"
)

type node struct {
	next *node
	val  interface{}
}

type Queue struct {
	head, tail *node
}

func New() *Queue {
	q := &Queue{}
	stub := &node{}
	q.head = stub
	q.tail = stub
	return q
}

// Push adds x to the back of the queue.
//
// Push can be safely called from multiple goroutines
func (q *Queue) Push(x interface{}) {
	n := &node{val: x}
	// current producer acquires head node
	prev := (*node)(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.head)), unsafe.Pointer(n)))

	// release node to consumer
	prev.next = n
}

// Pop removes the item from the front of the queue or nil if the queue is empty
//
// Pop must be called from a single, consumer goroutine
func (q *Queue) Pop() interface{} {
	tail := q.tail
	next := (*node)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&tail.next)))) // acquire
	if next != nil {
		q.tail = next
		v := next.val
		next.val = nil
		return v
	}
	return nil
}

// Empty returns true if the queue is empty
//
// Empty must be called from a single, consumer goroutine
func (q *Queue) Empty() bool {
	tail := q.tail
	next := (*node)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&tail.next))))
	return next == nil
}
