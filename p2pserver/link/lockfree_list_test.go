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
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -race

func TestLockfreeList_Push(t *testing.T) {
	list := &LockFreeList{}
	N := 200
	wg := &sync.WaitGroup{}
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < N; j++ {
				list.Push([]byte{1})
			}
		}()
	}

	wg.Wait()
	owned, _ := list.Take()
	for i := 0; i < N*N; i++ {
		buf := owned.Pop()
		assert.Equal(t, buf, []byte{1})
	}
}

func TestLockfreeList_Take(t *testing.T) {
	list := &LockFreeList{}
	N := 200
	wg := &sync.WaitGroup{}
	wg.Add(N + 1)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < N; j++ {
				list.Push([]byte{1})
			}
		}()
	}

	go func() {
		defer wg.Done()
		for i := 0; i < N*N; i++ {
			owned, _ := list.Take()
			for buf := owned.Pop(); buf != nil; buf = owned.Pop() {
				assert.Equal(t, buf, []byte{1})
			}
		}
	}()

	wg.Wait()
}

func TestLockfreeList_ConcurrentTake(t *testing.T) {
	list := &LockFreeList{}
	N := 200
	wg := &sync.WaitGroup{}
	wg.Add(N * 2)
	pushed := uint32(0)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < N; j++ {
				if list.Push([]byte{1}) {
					atomic.AddUint32(&pushed, 1)
				}
			}
		}()
	}

	poped := uint32(0)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()

			owned, _ := list.Take()
			for buf := owned.Pop(); buf != nil; buf = owned.Pop() {
				assert.Equal(t, buf, []byte{1})
				atomic.AddUint32(&poped, 1)
			}
		}()
	}

	wg.Wait()
	owned, _ := list.Take()
	for buf := owned.Pop(); buf != nil; buf = owned.Pop() {
		assert.Equal(t, buf, []byte{1})
		atomic.AddUint32(&poped, 1)
	}

	assert.Equal(t, pushed, poped)
}

func TestLockfreeList_TakeAndSeal(t *testing.T) {
	list := &LockFreeList{}
	N := 200
	wg := &sync.WaitGroup{}
	wg.Add(N + 1)
	pushed := uint32(0)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < N; j++ {
				if list.Push([]byte{1}) {
					atomic.AddUint32(&pushed, 1)
				}
			}
		}()
	}

	poped := uint32(0)
	go func() {
		defer wg.Done()

		owned := list.TakeAndSeal()
		for buf := owned.Pop(); buf != nil; buf = owned.Pop() {
			assert.Equal(t, buf, []byte{1})
			poped += 1
		}
	}()

	wg.Wait()

	assert.Equal(t, pushed, poped)
}

func BenchmarkLockfreeList_Push(b *testing.B) {
	list := &LockFreeList{}
	G := 10
	wg := &sync.WaitGroup{}
	wg.Add(G)
	for g := 0; g < G; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < b.N; i++ {
				list.Push([]byte{1})
			}
		}()
	}
	wg.Wait()
}

type LockedList struct {
	sync.Mutex
	list [][]byte
}

func BenchmarkLockedList_Push(b *testing.B) {
	list := &LockedList{}
	G := 10
	wg := &sync.WaitGroup{}
	wg.Add(G)
	for g := 0; g < G; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < b.N; i++ {
				list.Lock()
				list.list = append(list.list, []byte{1})
				list.Unlock()
			}
		}()
	}
	wg.Wait()
}

func BenchmarkArray_SinglePush(b *testing.B) {
	var list [][]byte
	for i := 0; i < b.N; i++ {
		list = append(list, []byte{1})
	}
}

func BenchmarkPreAllocArray_SinglePush(b *testing.B) {
	list := make([][]byte, 0, b.N)
	for i := 0; i < b.N; i++ {
		list = append(list, []byte{1})
	}
}

func BenchmarkLockfreeList_SinglePush(b *testing.B) {
	list := &LockFreeList{}
	for i := 0; i < b.N; i++ {
		list.Push([]byte{1})
	}
}
