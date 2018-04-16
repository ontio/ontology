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
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
)

type MemoryStore struct {
	memory map[string]*common.StateItem
}

func NewMemDatabase() *MemoryStore {
	return &MemoryStore{
		memory: make(map[string]*common.StateItem),
	}
}

func (db *MemoryStore) Put(prefix byte, key []byte, value states.StateValue, state common.ItemState) {
	db.memory[string(append([]byte{prefix}, key...))] = &common.StateItem{
		Key:   string(key),
		Value: value,
		State: state,
	}
}

func (db *MemoryStore) Get(prefix byte, key []byte) *common.StateItem {
	if entry, ok := db.memory[string(append([]byte{prefix}, key...))]; ok {
		return entry
	}
	return nil
}

func (db *MemoryStore) Delete(prefix byte, key []byte) {
	if v, ok := db.memory[string(append([]byte{prefix}, key...))]; ok {
		v.State = common.Deleted
	} else {
		db.memory[string(append([]byte{prefix}, key...))] = &common.StateItem{
			Key:   string(key),
			State: common.Deleted,
		}
	}

}

func (db *MemoryStore) Find() []*common.StateItem {
	var memory []*common.StateItem
	for _, v := range db.memory {
		memory = append(memory, v)
	}
	return memory
}

func (db *MemoryStore) GetChangeSet() map[string]*common.StateItem {
	m := make(map[string]*common.StateItem)
	for k, v := range db.memory {
		if v.State != common.None {
			m[k] = v
		}
	}
	return m
}
