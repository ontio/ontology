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

package statestore

import (
	"sort"
	"strings"

	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
)

type OverlayDB struct {
	store  common.StateStore
	memory map[string]states.StateValue
}

func NewOverlayDB(store common.StateStore) *OverlayDB {
	return &OverlayDB{
		store:  store,
		memory: make(map[string]states.StateValue),
	}
}

func (self *OverlayDB) Find(prefix common.DataEntryPrefix, key []byte) ([]*common.StateItem, error) {
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
	for k, v := range self.memory {
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

func (self *OverlayDB) TryAdd(prefix common.DataEntryPrefix, key []byte, value states.StateValue) {
	pkey := append([]byte{byte(prefix)}, key...)
	self.memory[string(pkey)] = value
}

func (self *OverlayDB) TryGet(prefix common.DataEntryPrefix, key []byte) (states.StateValue, error) {
	pkey := append([]byte{byte(prefix)}, key...)
	if state, ok := self.memory[string(pkey)]; ok {
		return state, nil
	}
	return self.store.TryGet(prefix, key)
}

func (self *OverlayDB) TryDelete(prefix common.DataEntryPrefix, key []byte) {
	pkey := append([]byte{byte(prefix)}, key...)
	self.memory[string(pkey)] = nil
}

func (self *OverlayDB) CommitTo() error {
	for k, v := range self.memory {
		pkey := []byte(k)
		self.store.TryAdd(common.DataEntryPrefix(pkey[0]), pkey[1:], v)
	}
	return nil
}
