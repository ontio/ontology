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

package storage

import (
	"bytes"

	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// CacheDB is smart contract execute cache, it contain transaction cache and block cache
// When smart contract execute finish, need to commit transaction cache to block cache
type CacheDB struct {
	memdb      *overlaydb.MemDB
	backend    *overlaydb.OverlayDB
	keyScratch []byte
}

// NewCacheDB return a new contract cache
func NewCacheDB(store *overlaydb.OverlayDB) *CacheDB {
	return &CacheDB{
		backend: store,
		memdb:   overlaydb.NewMemDB(0),
	}
}

func ensureBuffer(b []byte, n int) []byte {
	if cap(b) < n {
		return make([]byte, n)
	}
	return b[:n]
}

func makePrefixedKey(dst []byte, prefix byte, key []byte) []byte {
	dst = ensureBuffer(dst, len(key)+1)
	dst[0] = prefix
	copy(dst[1:], key)
	return dst
}

// Commit current transaction cache to block cache
func (self *CacheDB) Commit() {
	self.memdb.ForEach(func(key, val []byte) {
		if len(val) == 0 {
			self.backend.Delete(key)
		} else {
			self.backend.Put(key, val)
		}
	})
}

// Add item to cache
func (self *CacheDB) Add(prefix common.DataEntryPrefix, key []byte, value states.StateValue) {
	data := new(bytes.Buffer)
	// todo: remove error or return to caller
	err := value.Serialize(data)
	if err != nil {
		panic(err)
	}

	self.keyScratch = makePrefixedKey(self.keyScratch, byte(prefix), key)
	self.memdb.Put(self.keyScratch, data.Bytes())
}

// GetOrAdd item
// If item has existed, return it
// Else add it to cache
func (self *CacheDB) GetOrAdd(prefix common.DataEntryPrefix, key []byte, value states.StateValue) (states.StateValue, error) {
	val, err := self.Get(prefix, key)
	if err != nil {
		return nil, err
	}
	if val != nil {
		return val, nil
	}

	self.Add(prefix, key, value)

	return value, nil
}

func (self *CacheDB) GetContractCode(addr common2.Address) ([]byte, error) {
	dep, err := self.Get(common.ST_CONTRACT, addr[:])
	if err != nil {
		return nil, err
	}
	if dep == nil {
		return nil, nil
	}
	return dep.(*payload.DeployCode).Code, nil
}

// Get item by key
func (self *CacheDB) Get(prefix common.DataEntryPrefix, key []byte) (states.StateValue, error) {
	self.keyScratch = makePrefixedKey(self.keyScratch, byte(prefix), key)
	value, unknown := self.memdb.Get(self.keyScratch)
	if unknown {
		v, err := self.backend.Get(self.keyScratch)
		if err != nil {
			return nil, err
		}
		value = v
	}

	if len(value) == 0 {
		return nil, nil
	}

	return decodeStateObject(prefix, value)
}

func decodeStateObject(prefix common.DataEntryPrefix, enc []byte) (states.StateValue, error) {
	reader := bytes.NewBuffer(enc)
	switch prefix {
	case common.ST_BOOKKEEPER:
		bookkeeper := new(payload.Bookkeeper)
		if err := bookkeeper.Deserialize(reader); err != nil {
			return nil, err
		}
		return bookkeeper, nil
	case common.ST_CONTRACT:
		contract := new(payload.DeployCode)
		if err := contract.Deserialize(reader); err != nil {
			return nil, err
		}
		return contract, nil
	case common.ST_STORAGE:
		storage := new(states.StorageItem)
		if err := storage.Deserialize(reader); err != nil {
			return nil, err
		}
		return storage, nil
	default:
		panic("[decodeStateObject] invalid state type!")
	}
}

// Delete item from cache
func (self *CacheDB) Delete(prefix common.DataEntryPrefix, key []byte) {
	self.keyScratch = makePrefixedKey(self.keyScratch, byte(prefix), key)
	self.memdb.Delete(self.keyScratch)
}

func (self *CacheDB) NewIterator(prefix common.DataEntryPrefix, key []byte) common.StoreIterator {
	pkey := make([]byte, 1+len(key))
	pkey[0] = byte(prefix)
	copy(pkey[1:], key)
	prefixRange := util.BytesPrefix(pkey)
	backIter := self.backend.NewIterator(pkey)
	memIter := self.memdb.NewIterator(prefixRange)

	return overlaydb.NewJoinIter(memIter, backIter)
}
