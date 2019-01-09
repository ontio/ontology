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

// this implementation is based on goleveldb "https://github.com/syndtr/goleveldb"
// the semantics of Delete(key) operation is changed to Put(key, nil) to support overlaydb

package overlaydb

import (
	"math/rand"

	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Common errors.
var (
	ErrNotFound     = errors.ErrNotFound
	ErrIterReleased = errors.New("leveldb/memdb: iterator released")
)

const tMaxHeight = 12

type dbIter struct {
	util.BasicReleaser
	p          *MemDB
	slice      *util.Range
	node       int
	forward    bool
	key, value []byte
	err        error
}

func (i *dbIter) fill(checkStart, checkLimit bool) bool {
	if i.node != 0 {
		n := i.p.nodeData[i.node]
		m := n + i.p.nodeData[i.node+nKey]
		i.key = i.p.kvData[n:m]
		if i.slice != nil {
			switch {
			case checkLimit && i.slice.Limit != nil && i.p.cmp.Compare(i.key, i.slice.Limit) >= 0:
				fallthrough
			case checkStart && i.slice.Start != nil && i.p.cmp.Compare(i.key, i.slice.Start) < 0:
				i.node = 0
				goto bail
			}
		}
		i.value = i.p.kvData[m : m+i.p.nodeData[i.node+nVal]]
		return true
	}
bail:
	i.key = nil
	i.value = nil
	return false
}

func (i *dbIter) Valid() bool {
	return i.node != 0
}

func (i *dbIter) First() bool {
	if i.Released() {
		i.err = ErrIterReleased
		return false
	}

	i.forward = true
	if i.slice != nil && i.slice.Start != nil {
		i.node, _ = i.p.findGE(i.slice.Start, false)
	} else {
		i.node = i.p.nodeData[nNext]
	}
	return i.fill(false, true)
}

func (i *dbIter) Last() bool {
	if i.Released() {
		i.err = ErrIterReleased
		return false
	}

	i.forward = false
	if i.slice != nil && i.slice.Limit != nil {
		i.node = i.p.findLT(i.slice.Limit)
	} else {
		i.node = i.p.findLast()
	}
	return i.fill(true, false)
}

func (i *dbIter) Seek(key []byte) bool {
	if i.Released() {
		i.err = ErrIterReleased
		return false
	}

	i.forward = true
	if i.slice != nil && i.slice.Start != nil && i.p.cmp.Compare(key, i.slice.Start) < 0 {
		key = i.slice.Start
	}
	i.node, _ = i.p.findGE(key, false)
	return i.fill(false, true)
}

func (i *dbIter) Next() bool {
	if i.Released() {
		i.err = ErrIterReleased
		return false
	}

	if i.node == 0 {
		if !i.forward {
			return i.First()
		}
		return false
	}
	i.forward = true
	i.node = i.p.nodeData[i.node+nNext]
	return i.fill(false, true)
}

func (i *dbIter) Prev() bool {
	if i.Released() {
		i.err = ErrIterReleased
		return false
	}

	if i.node == 0 {
		if i.forward {
			return i.Last()
		}
		return false
	}
	i.forward = false
	i.node = i.p.findLT(i.key)
	return i.fill(true, false)
}

func (i *dbIter) Key() []byte {
	return i.key
}

func (i *dbIter) Value() []byte {
	return i.value
}

func (i *dbIter) Error() error { return i.err }

func (i *dbIter) Release() {
	if !i.Released() {
		i.p = nil
		i.node = 0
		i.key = nil
		i.value = nil
		i.BasicReleaser.Release()
	}
}

const (
	nKV = iota
	nKey
	nVal
	nHeight
	nNext
)

// MemDB is an in-memdb key/value database.
type MemDB struct {
	cmp comparer.BasicComparer
	rnd *rand.Rand

	kvData []byte
	// Node data:
	// [0]         : KV offset
	// [1]         : Key length
	// [2]         : Value length
	// [3]         : Height
	// [3..height] : Next nodes
	nodeData  []int
	prevNode  [tMaxHeight]int
	maxHeight int
	n         int
	kvSize    int
}

func (p *MemDB) randHeight() (h int) {
	const branching = 4
	h = 1
	for h < tMaxHeight && p.rnd.Int()%branching == 0 {
		h++
	}
	return
}

// Must hold RW-lock if prev == true, as it use shared prevNode slice.
func (p *MemDB) findGE(key []byte, prev bool) (int, bool) {
	node := 0
	h := p.maxHeight - 1
	for {
		next := p.nodeData[node+nNext+h]
		cmp := 1
		if next != 0 {
			o := p.nodeData[next]
			cmp = p.cmp.Compare(p.kvData[o:o+p.nodeData[next+nKey]], key)
		}
		if cmp < 0 {
			// Keep searching in this list
			node = next
		} else {
			if prev {
				p.prevNode[h] = node
			} else if cmp == 0 {
				return next, true
			}
			if h == 0 {
				return next, cmp == 0
			}
			h--
		}
	}
}

func (p *MemDB) findLT(key []byte) int {
	node := 0
	h := p.maxHeight - 1
	for {
		next := p.nodeData[node+nNext+h]
		o := p.nodeData[next]
		if next == 0 || p.cmp.Compare(p.kvData[o:o+p.nodeData[next+nKey]], key) >= 0 {
			if h == 0 {
				break
			}
			h--
		} else {
			node = next
		}
	}
	return node
}

func (p *MemDB) findLast() int {
	node := 0
	h := p.maxHeight - 1
	for {
		next := p.nodeData[node+nNext+h]
		if next == 0 {
			if h == 0 {
				break
			}
			h--
		} else {
			node = next
		}
	}
	return node
}

// Put sets the value for the given key. It overwrites any previous value
// for that key; a MemDB is not a multi-map.
//
// It is safe to modify the contents of the arguments after Put returns.
func (p *MemDB) Put(key []byte, value []byte) {
	if node, exact := p.findGE(key, true); exact {
		if len(value) != 0 {
			kvOffset := len(p.kvData)
			p.kvData = append(p.kvData, key...)
			p.kvData = append(p.kvData, value...)
			p.nodeData[node] = kvOffset
		}
		m := p.nodeData[node+nVal]
		p.nodeData[node+nVal] = len(value)
		p.kvSize += len(value) - m
		return
	}

	h := p.randHeight()
	if h > p.maxHeight {
		for i := p.maxHeight; i < h; i++ {
			p.prevNode[i] = 0
		}
		p.maxHeight = h
	}

	kvOffset := len(p.kvData)
	p.kvData = append(p.kvData, key...)
	p.kvData = append(p.kvData, value...)
	// Node
	node := len(p.nodeData)
	p.nodeData = append(p.nodeData, kvOffset, len(key), len(value), h)
	for i, n := range p.prevNode[:h] {
		m := n + nNext + i
		p.nodeData = append(p.nodeData, p.nodeData[m])
		p.nodeData[m] = node
	}

	p.kvSize += len(key) + len(value)
	p.n++
	return
}

// Delete deletes the value for the given key.
//
// It is safe to modify the contents of the arguments after Delete returns.
func (p *MemDB) Delete(key []byte) {
	p.Put(key, nil)
}

// Get gets the value for the given key. It returns unkown == true if the
// MemDB does not contain the key. It returns nil, false if MemDB has deleted the key
//
// The caller should not modify the contents of the returned slice, but
// it is safe to modify the contents of the argument after Get returns.
func (p *MemDB) Get(key []byte) (value []byte, unkown bool) {
	if node, exact := p.findGE(key, false); exact {
		valen := p.nodeData[node+nVal]
		if valen != 0 {
			o := p.nodeData[node] + p.nodeData[node+nKey]
			value = p.kvData[o : o+valen]
		}
	} else {
		unkown = true
	}
	return
}

// Find finds key/value pair whose key is greater than or equal to the
// given key. It returns ErrNotFound if the table doesn't contain
// such pair.
//
// The caller should not modify the contents of the returned slice, but
// it is safe to modify the contents of the argument after Find returns.
func (p *MemDB) Find(key []byte) (rkey, value []byte, err error) {
	if node, _ := p.findGE(key, false); node != 0 {
		n := p.nodeData[node]
		m := n + p.nodeData[node+nKey]
		rkey = p.kvData[n:m]
		valen := p.nodeData[node+nVal]
		if valen != 0 {
			value = p.kvData[m : m+valen]
		}
	} else {
		err = ErrNotFound
	}
	return
}

func (p *MemDB) ForEach(f func(key, val []byte)) {
	for node := p.nodeData[nNext]; node != 0; node = p.nodeData[node+nNext] {
		n := p.nodeData[node]
		m := n + p.nodeData[node+nKey]
		key := p.kvData[n:m]
		val := p.kvData[m : m+p.nodeData[node+nVal]]
		f(key, val)
	}

}

// NewIterator returns an iterator of the MemDB.
// The returned iterator is not safe for concurrent use, but it is safe to use
// multiple iterators concurrently, with each in a dedicated goroutine.
// It is also safe to use an iterator concurrently with modifying its
// underlying MemDB. However, the resultant key/value pairs are not guaranteed
// to be a consistent snapshot of the MemDB at a particular point in time.
//
// Slice allows slicing the iterator to only contains keys in the given
// range. A nil Range.Start is treated as a key before all keys in the
// MemDB. And a nil Range.Limit is treated as a key after all keys in
// the MemDB.
//
// The iterator must be released after use, by calling Release method.
//
// Also read Iterator documentation of the leveldb/iterator package.
func (p *MemDB) NewIterator(slice *util.Range) iterator.Iterator {
	return &dbIter{p: p, slice: slice}
}

// Capacity returns keys/values buffer capacity.
func (p *MemDB) Capacity() int {
	return cap(p.kvData)
}

// Size returns sum of keys and values length. Note that deleted
// key/value will not be accounted for, but it will still consume
// the buffer, since the buffer is append only.
func (p *MemDB) Size() int {
	return p.kvSize
}

// Free returns keys/values free buffer before need to grow.
func (p *MemDB) Free() int {
	return cap(p.kvData) - len(p.kvData)
}

// Len returns the number of entries in the MemDB.
func (p *MemDB) Len() int {
	return p.n
}

// Reset resets the MemDB to initial empty state. Allows reuse the buffer.
func (p *MemDB) Reset() {
	p.rnd = rand.New(rand.NewSource(0xdeadbeef))
	p.maxHeight = 1
	p.n = 0
	p.kvSize = 0
	p.kvData = p.kvData[:0]
	p.nodeData = p.nodeData[:nNext+tMaxHeight]
	p.nodeData[nKV] = 0
	p.nodeData[nKey] = 0
	p.nodeData[nVal] = 0
	p.nodeData[nHeight] = tMaxHeight
	for n := 0; n < tMaxHeight; n++ {
		p.nodeData[nNext+n] = 0
		p.prevNode[n] = 0
	}
}

// NewMemDB creates a new initialized in-memdb key/value MemDB. The capacity
// is the initial key/value buffer capacity. The capacity is advisory,
// not enforced.
//
// This MemDB is append-only, deleting an entry would remove entry node but not
// reclaim KV buffer.
//
// The returned MemDB instance is safe for concurrent use.
func NewMemDB(capacity int, kvNum int) *MemDB {
	p := &MemDB{
		cmp:       comparer.DefaultComparer,
		rnd:       rand.New(rand.NewSource(0xdeadbeef)),
		maxHeight: 1,
		kvData:    make([]byte, 0, capacity),
		nodeData:  make([]int, 4+tMaxHeight, (4+tMaxHeight)*(1+kvNum)),
	}
	p.nodeData[nHeight] = tMaxHeight
	return p
}
