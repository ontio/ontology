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
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/store/common"
)

type StateItem struct {
	Prefix common.DataEntryPrefix
	Key    string
	Value  states.IStateValue
	State  common.ItemState
}

type Memory map[string]*StateItem

type CloneCache struct {
	Memory Memory
	Store  common.IStateStore
}

func NewCloneCache(store common.IStateStore) *CloneCache {
	return &CloneCache{
		Memory: make(Memory),
		Store:  store,
	}
}

func (cloneCache *CloneCache) Commit() {
	for _, v := range cloneCache.Memory {
		if v.State == common.Deleted {
			cloneCache.Store.TryDelete(v.Prefix, []byte(v.Key))
		} else if v.State == common.Changed {
			cloneCache.Store.TryAdd(v.Prefix, []byte(v.Key), v.Value, true)
		}
	}
}

func (cloneCache *CloneCache) Add(prefix common.DataEntryPrefix, key []byte, value states.IStateValue) {
	cloneCache.Memory[string(append([]byte{byte(prefix)}, key...))] = &StateItem{
		Prefix: prefix,
		Key:    string(key),
		Value:  value,
		State:  common.Changed,
	}
}

func (cloneCache *CloneCache) GetOrAdd(prefix common.DataEntryPrefix, key []byte, value states.IStateValue) (states.IStateValue, error) {
	if v, ok := cloneCache.Memory[string(append([]byte{byte(prefix)}, key...))]; ok {
		if v.State == common.Deleted {
			cloneCache.Memory[string(append([]byte{byte(prefix)}, key...))] = &StateItem{Prefix: prefix, Key: string(key), Value: value, State: common.Changed}
			return value, nil
		}
		return v.Value, nil
	}
	item, err := cloneCache.Store.TryGet(prefix, key)
	if err != nil {
		return nil, err
	}
	if item != nil && item.State != common.Deleted {
		return item.Value, nil
	}
	cloneCache.Memory[string(append([]byte{byte(prefix)}, key...))] = &StateItem{Prefix: prefix, Key: string(key), Value: value, State: common.Changed}
	return value, nil
}

func (cloneCache *CloneCache) Get(prefix common.DataEntryPrefix, key []byte) (states.IStateValue, error) {
	if v, ok := cloneCache.Memory[string(append([]byte{byte(prefix)}, key...))]; ok {
		if v.State == common.Deleted {
			return nil, nil
		}
		return v.Value, nil
	}
	item, err := cloneCache.Store.TryGet(prefix, key)
	if err != nil {
		return nil, err
	}
	if item == nil || item.State == common.Deleted {
		return nil, nil
	}
	return item.Value, nil
}

func (cloneCache *CloneCache) Delete(prefix common.DataEntryPrefix, key []byte) {
	if v, ok := cloneCache.Memory[string(append([]byte{byte(prefix)}, key...))]; ok {
		v.State = common.Deleted
	} else {
		cloneCache.Memory[string(append([]byte{byte(prefix)}, key...))] = &StateItem{
			Prefix: prefix,
			Key:    string(key),
			State:  common.Deleted,
		}
	}
}
