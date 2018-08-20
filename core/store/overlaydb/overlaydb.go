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

package overlaydb

import (
	"github.com/ontio/ontology/core/store/common"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type OverlayDB struct {
	store      common.PersistStore
	memdb      *MemDB
	keyScratch []byte
	dbErr      error
}

func NewOverlayDB(store common.PersistStore) *OverlayDB {
	return &OverlayDB{
		store: store,
		memdb: NewMemDB(0),
	}
}

func (self *OverlayDB) Reset() {
	self.memdb.Reset()
}

func makePrefixedKey(dst []byte, prefix byte, key []byte) []byte {
	dst = ensureBuffer(dst, len(key)+1)
	dst[0] = prefix
	copy(dst[1:], key)
	return dst
}

func ensureBuffer(b []byte, n int) []byte {
	if cap(b) < n {
		return make([]byte, n)
	}
	return b[:n]
}

// if key is deleted, value == nil
func (self *OverlayDB) Get(prefix byte, key []byte) (value []byte, err error) {
	var unknown bool
	self.keyScratch = makePrefixedKey(self.keyScratch, prefix, key)
	value, unknown = self.memdb.Get(self.keyScratch)
	if unknown == false {
		return value, nil
	}

	value, err = self.store.Get(self.keyScratch)
	if err != nil {
		if err == common.ErrNotFound {
			return nil, nil
		}
		self.dbErr = err
		return nil, err
	}

	return
}

func (self *OverlayDB) Put(prefix byte, key []byte, value []byte) {
	self.keyScratch = makePrefixedKey(self.keyScratch, prefix, key)
	self.memdb.Put(self.keyScratch, value)
}

func (self *OverlayDB) Delete(prefix byte, key []byte) {
	self.keyScratch = makePrefixedKey(self.keyScratch, prefix, key)
	self.memdb.Delete(self.keyScratch)
}

func (self *OverlayDB) CommitTo() error {
	self.memdb.ForEach(func(key, val []byte) {
		if len(val) == 0 {
			self.store.BatchDelete(key)
		} else {
			self.store.BatchPut(key, val)
		}
	})
	return nil
}

type KeyOrigin byte

const (
	FromMem  KeyOrigin = iota
	FromBack           = iota
	FromBoth           = iota
)

type Iter struct {
	backend     common.StoreIterator
	memdb       common.StoreIterator
	key, value  []byte
	keyOrigin   KeyOrigin
	nextMemEnd  bool
	nextBackEnd bool
	cmp         comparer.BasicComparer
}

func (iter *Iter) First() bool {
	f := iter.first()
	if f == false {
		return false
	}
	for len(iter.value) == 0 {
		if iter.next() == false {
			return false
		}
	}

	return true
}

func (iter *Iter) first() bool {
	var bkey, bval, mkey, mval []byte
	back := iter.backend.First()
	mem := iter.memdb.First()
	if back {
		bkey = iter.backend.Key()
		bval = iter.backend.Value()
		if mem == false {
			iter.key = bkey
			iter.value = bval
			iter.keyOrigin = FromBack
			return true
		}
		mkey = iter.memdb.Key()
		mval = iter.memdb.Value()
		cmp := iter.cmp.Compare(mkey, bkey)
		if cmp < 1 {
			iter.key = mkey
			iter.value = mval
			if cmp == 0 {
				iter.keyOrigin = FromBoth
			} else {
				iter.keyOrigin = FromMem
			}
		} else {
			iter.key = bkey
			iter.value = bval
			iter.keyOrigin = FromBack
		}
		return true
	} else {
		if mem {
			iter.key = iter.memdb.Key()
			iter.value = iter.memdb.Value()
			iter.keyOrigin = FromMem
			return true
		}
		return false
	}
}

func (iter *Iter) Key() []byte {
	return iter.key
}

func (iter *Iter) Value() []byte {
	return iter.value
}

func (iter *Iter) Next() bool {
	f := iter.next()
	if f == false {
		return false
	}

	for len(iter.value) == 0 {
		if iter.next() == false {
			return false
		}
	}

	return true
}

func (iter *Iter) next() bool {
	if (iter.keyOrigin == FromMem || iter.keyOrigin == FromBoth) && iter.nextMemEnd == false {
		iter.nextMemEnd = !iter.memdb.Next()
	}
	if (iter.keyOrigin == FromBack || iter.keyOrigin == FromBoth) && iter.nextBackEnd == false {
		iter.nextBackEnd = !iter.backend.Next()
	}
	if iter.nextBackEnd {
		if iter.nextMemEnd {
			iter.key = nil
			iter.value = nil
			return false
		} else {
			iter.key = iter.memdb.Key()
			iter.value = iter.memdb.Value()
			iter.keyOrigin = FromMem
		}
	} else {
		if iter.nextMemEnd {
			iter.key = iter.backend.Key()
			iter.value = iter.backend.Value()
			iter.keyOrigin = FromBack
		} else {
			bkey := iter.backend.Key()
			mkey := iter.memdb.Key()
			cmp := iter.cmp.Compare(mkey, bkey)
			switch cmp {
			case -1:
				iter.key = mkey
				iter.value = iter.memdb.Value()
				iter.keyOrigin = FromMem
			case 0:
				iter.key = mkey
				iter.value = iter.memdb.Value()
				iter.keyOrigin = FromBoth
			case 1:
				iter.key = bkey
				iter.value = iter.backend.Value()
				iter.keyOrigin = FromBack
			default:
				panic("unreachable")
			}
		}
	}

	return true
}

func (iter *Iter) Release() {
	iter.memdb.Release()
	iter.backend.Release()
}

func (self *OverlayDB) NewIterator(prefix byte, key []byte) common.StoreIterator {
	pkey := make([]byte, len(key)+1)
	pkey[0] = prefix
	copy(pkey[1:], key)

	prefixRange := util.BytesPrefix(pkey)
	backIter := self.store.NewIterator(pkey)
	memIter := self.memdb.NewIterator(prefixRange)

	return &Iter{
		backend: backIter,
		memdb:   memIter,
		cmp:     comparer.DefaultComparer,
	}

}
