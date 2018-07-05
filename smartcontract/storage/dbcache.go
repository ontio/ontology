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
)

// CloneCache is smart contract execute cache, it contain transaction cache and block cache
// When smart contract execute finish, need to commit transaction cache to block cache
type CloneCache struct {
	ReadSet []string
	Cache   map[string]states.StateValue
	Store   common.StateStore
}

// NewCloneCache return a new contract cache
func NewCloneCache(store common.StateStore) *CloneCache {
	return &CloneCache{
		Cache: make(map[string]states.StateValue),
		Store: store,
	}
}

// Commit current transaction cache to block cache
func (cloneCache *CloneCache) Commit() {
	for k, v := range cloneCache.Cache {
		key := []byte(k)
		if v == nil {
			cloneCache.Store.TryDelete(common.DataEntryPrefix(key[0]), key[1:])
		} else {
			cloneCache.Store.TryAdd(common.DataEntryPrefix(key[0]), key[1:], v)
		}
	}
}

// Add item to cache
func (cloneCache *CloneCache) Add(prefix common.DataEntryPrefix, key []byte, value states.StateValue) {
	k := string(append([]byte{byte(prefix)}, key...))
	cloneCache.Cache[k] = value
}

// GetOrAdd item
// If item has existed, return it
// Else add it to cache
func (cloneCache *CloneCache) GetOrAdd(prefix common.DataEntryPrefix, key []byte, value states.StateValue) (states.StateValue, error) {
	item, err := cloneCache.Get(prefix, key)
	if err != nil {
		return nil, err
	}
	if item != nil {
		return item, nil
	}

	k := string(append([]byte{byte(prefix)}, key...))
	cloneCache.Cache[k] = value
	return value, nil
}

// Get item by key
func (cloneCache *CloneCache) Get(prefix common.DataEntryPrefix, key []byte) (states.StateValue, error) {
	k := string(append([]byte{byte(prefix)}, key...))
	if v, ok := cloneCache.Cache[k]; ok {
		return v, nil
	}

	cloneCache.ReadSet = append(cloneCache.ReadSet, k)
	item, err := cloneCache.Store.TryGet(prefix, key)
	if item == nil || err != nil {
		return nil, err
	}
	return item.Value, err
}

// Delete item from cache
func (cloneCache *CloneCache) Delete(prefix common.DataEntryPrefix, key []byte) {
	k := string(append([]byte{byte(prefix)}, key...))
	cloneCache.Cache[k] = nil
}
