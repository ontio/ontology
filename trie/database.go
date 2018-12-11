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

package trie

import (
	"sync"
)

type Database struct {
	db map[string][]byte
	lock sync.RWMutex
}

func NewMemDatabase() *Database {
	return &Database {
		db: make(map[string][]byte),
	}
}

func (db *Database) Put(key []byte, value []byte) error{
	db.lock.Lock()
	defer db.lock.Unlock()
	db.db[string(key)] = CopyBytes(value)
	return nil
}

func (db *Database) Has(key []byte) (bool, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	_, ok := db.db[string(key)]
	return ok, nil
}

func (db *Database) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if entry, ok := db.db[string(key)]; ok {
		return CopyBytes(entry), nil
	}
	return nil, nil
}

func (db *Database) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	delete(db.db, string(key))
	return nil
}
