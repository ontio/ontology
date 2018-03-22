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
	. "github.com/Ontology/common"
	"github.com/Ontology/trie"
	"github.com/Ontology/core/store/common"
)

const (
	maxPastTries = 12
)

type ITrieStore interface {
	OpenTrie(root Uint256) (ITrie, error)
}

type ITrie interface {
	TryGet(key []byte) ([]byte, error)
	TryUpdate(key, value []byte) error
	TryDelete(key []byte) error
	Hash() Uint256
	CommitTo() (Uint256, error)
}

type cachingDB struct {
	db        common.IStore
	pastTries []*trie.SecureTrie
}

func NewTrieStore(db common.IStore) ITrieStore {
	return &cachingDB{db: db}
}

func (db *cachingDB) OpenTrie(root Uint256) (ITrie, error) {
	for i := len(db.pastTries) - 1; i >= 0; i-- {
		h := db.pastTries[i].Hash()
		if h.CompareTo(root) == 0 {
			return cachedTrie{db.pastTries[i].Copy(), db}, nil
		}
	}
	tr, err := trie.NewSecure(root, db.db)
	if err != nil {
		return nil, err
	}
	return cachedTrie{tr, db}, nil
}

func (db *cachingDB) pushTrie(t *trie.SecureTrie) {
	if len(db.pastTries) > maxPastTries {
		copy(db.pastTries, db.pastTries[1:])
		db.pastTries[len(db.pastTries) - 1] = t
	} else {
		db.pastTries = append(db.pastTries, t)
	}
}

type cachedTrie struct {
	*trie.SecureTrie
	*cachingDB
}

func (c cachedTrie) CommitTo() (Uint256, error) {
	root, err := c.SecureTrie.CommitTo(c.cachingDB.db)
	if err != nil {
		return Uint256{}, err
	}
	c.cachingDB.pushTrie(c.SecureTrie)
	return root, nil
}


