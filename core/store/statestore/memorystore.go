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
	"github.com/Ontology/core/states"
	."github.com/Ontology/core/store/common"
)

type MemoryStore struct {
	memory map[string]*StateItem
}

func NewMemDatabase() *MemoryStore {
	return &MemoryStore{
		memory: make(map[string]*StateItem),
	}
}

func (db *MemoryStore) Put(prefix byte, key []byte, value states.IStateValue, state ItemState, trie bool) {
	db.memory[string(append([]byte{prefix}, key...))] = &StateItem{
		Key: string(key),
		Value: value,
		State: state,
		Trie: trie,
	}
}

func (db *MemoryStore) Get(prefix byte, key []byte) *StateItem {
	if entry, ok := db.memory[string(append([]byte{prefix}, key...))]; ok {
		return entry
	}
	return nil
}

func (db *MemoryStore) Delete(prefix byte, key []byte) {
	if v, ok := db.memory[string(append([]byte{prefix}, key...))]; ok {
		v.State = Deleted
	} else {
		db.memory[string(append([]byte{prefix}, key...))] = &StateItem{
			Key: string(key),
			State: Deleted,
		}
	}

}

func (db *MemoryStore) Change(prefix byte, key []byte, trie bool) {
	db.memory[string(append([]byte{prefix}, key...))].State = Changed
	db.memory[string(append([]byte{prefix}, key...))].Trie = trie
}

func (db *MemoryStore) GetChangeSet() map[string]*StateItem {
	m := make(map[string]*StateItem)
	for k, v := range db.memory {
		if v.State != None {
			m[k] = v
		}
	}
	return m
}



