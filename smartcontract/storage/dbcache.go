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
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
	"sort"
	"strings"
)

// CloneCache is smart contract execute cache, it contain transaction cache and block cache
// When smart contract execute finish, need to commit transaction cache to block cache
type CloneCache struct {
	ReadSet []string
	Cache   map[string]states.StateValue
	store   common.StateStore
}

// NewCloneCache return a new contract cache
func NewCloneCache(store common.StateStore) *CloneCache {
	return &CloneCache{
		Cache: make(map[string]states.StateValue),
		store: store,
	}
}

func (self *CloneCache) Reset() {
	self.ReadSet = self.ReadSet[:0]
	self.Cache = make(map[string]states.StateValue)
}

// Commit current transaction cache to block cache
func (self *CloneCache) Commit() {
	for k, v := range self.Cache {
		key := []byte(k)
		if v == nil {
			self.store.TryDelete(common.DataEntryPrefix(key[0]), key[1:])
		} else {
			self.store.TryAdd(common.DataEntryPrefix(key[0]), key[1:], v)
		}
	}
}

// Add item to cache
func (self *CloneCache) Add(prefix common.DataEntryPrefix, key []byte, value states.StateValue) {
	k := string(append([]byte{byte(prefix)}, key...))
	self.Cache[k] = value
}

// GetOrAdd item
// If item has existed, return it
// Else add it to cache
func (self *CloneCache) GetOrAdd(prefix common.DataEntryPrefix, key []byte, value states.StateValue) (states.StateValue, error) {
	item, err := self.Get(prefix, key)
	if err != nil {
		return nil, err
	}
	if item != nil {
		return item, nil
	}

	k := string(append([]byte{byte(prefix)}, key...))
	self.Cache[k] = value
	return value, nil
}

// Get item by key
func (self *CloneCache) Get(prefix common.DataEntryPrefix, key []byte) (states.StateValue, error) {
	k := string(append([]byte{byte(prefix)}, key...))
	if v, ok := self.Cache[k]; ok {
		return v, nil
	}

	self.ReadSet = append(self.ReadSet, k)
	item, err := self.store.TryGet(prefix, key)
	if item == nil || err != nil {
		return nil, err
	}
	return item, err
}

// Delete item from cache
func (self *CloneCache) Delete(prefix common.DataEntryPrefix, key []byte) {
	k := string(append([]byte{byte(prefix)}, key...))
	self.Cache[k] = nil
}

func (self *CloneCache) Find(prefix common.DataEntryPrefix, key []byte) ([]*common.StateItem, error) {
	stats, err := self.store.Find(prefix, key)
	if err != nil {
		return nil, err
	}
	var sts []*common.StateItem
	pkey := append([]byte{byte(prefix)}, key...)

	index := make(map[string]int)
	for i, v := range stats {
		index[v.Key] = i
	}

	deleted := make([]int, 0)
	for k, v := range self.Cache {
		if strings.HasPrefix(k, string(pkey)) {
			if v == nil { // deleted but in inner db, need remove
				if i, ok := index[k]; ok {
					deleted = append(deleted, i)
				}
			} else {
				if i, ok := index[k]; ok {
					sts[i] = &common.StateItem{Key: k, Value: v}
				} else {
					sts = append(sts, &common.StateItem{Key: k, Value: v})
				}
			}
		}
	}

	sort.Ints(deleted)
	for i := len(deleted) - 1; i >= 0; i-- {
		sts = append(sts[:deleted[i]], sts[deleted[i]+1:]...)
	}

	return sts, nil
}
