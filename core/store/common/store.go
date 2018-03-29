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

package common

import (
	"github.com/Ontology/common"
	"github.com/Ontology/core/states"
	"github.com/Ontology/smartcontract/event"
)

type StoreIterator interface {
	Next() bool
	Prev() bool
	First() bool
	Last() bool
	Seek(key []byte) bool
	Key() []byte
	Value() []byte
	Release()
}

type PersistStore interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Delete(key []byte) error
	NewBatch()
	BatchPut(key []byte, value []byte)
	BatchDelete(key []byte)
	BatchCommit() error
	Close() error
	NewIterator(prefix []byte) StoreIterator
}

type StateStore interface {
	TryAdd(prefix DataEntryPrefix, key []byte, value states.StateValue, trie bool)
	TryGetOrAdd(prefix DataEntryPrefix, key []byte, value states.StateValue, trie bool) error
	TryGet(prefix DataEntryPrefix, key []byte) (*StateItem, error)
	TryGetAndChange(prefix DataEntryPrefix, key []byte, trie bool) (states.StateValue, error)
	TryDelete(prefix DataEntryPrefix, key []byte)
	Find(prefix DataEntryPrefix, key []byte) ([]*StateItem, error)
}

type MemoryCacheStore interface {
	Put(prefix byte, key []byte, value states.StateValue, state ItemState, trie bool)
	Get(prefix byte, key []byte) *StateItem
	Delete(prefix byte, key []byte)
	GetChangeSet() map[string]*StateItem
	Change(prefix byte, key []byte, trie bool)
}

type EventStore interface {
	SaveEventNotifyByTx(txHash common.Uint256, notifies []*event.NotifyEventInfo) error
	SaveEventNotifyByBlock(height uint32, txHashs []common.Uint256) error
	GetEventNotifyByTx(txHash common.Uint256) ([]*event.NotifyEventInfo, error)
	CommitTo() error
}

type ItemState byte

const (
	None ItemState = iota
	Changed
	Deleted
)

type StateItem struct {
	Key   string
	Value states.StateValue
	State ItemState
	Trie  bool
}

func (e *StateItem) copy() *StateItem {
	c := *e
	return &c
}
