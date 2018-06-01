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
	"errors"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/event"
)

var ErrNotFound = errors.New("not found")

//Store iterator for iterate store
type StoreIterator interface {
	Next() bool           //Next item. If item available return true, otherwise return false
	Prev() bool           //previous item. If item available return true, otherwise return false
	First() bool          //First item. If item available return true, otherwise return false
	Last() bool           //Last item. If item available return true, otherwise return false
	Seek(key []byte) bool //Seek key. If item available return true, otherwise return false
	Key() []byte          //Return the current item key
	Value() []byte        //Return the current item value
	Release()             //Close iterator
}

//PersistStore of ledger
type PersistStore interface {
	Put(key []byte, value []byte) error      //Put the key-value pair to store
	Get(key []byte) ([]byte, error)          //Get the value if key in store
	Has(key []byte) (bool, error)            //Whether the key is exist in store
	Delete(key []byte) error                 //Delete the key in store
	NewBatch()                               //Start commit batch
	BatchPut(key []byte, value []byte)       //Put a key-value pair to batch
	BatchDelete(key []byte)                  //Delete the key in batch
	BatchCommit() error                      //Commit batch to store
	Close() error                            //Close store
	NewIterator(prefix []byte) StoreIterator //Return the iterator of store
}

//StateStore save result of smart contract execution, before commit to store
type StateStore interface {
	//Add key-value pair to store
	TryAdd(prefix DataEntryPrefix, key []byte, value states.StateValue)
	//Get key from state store, if not exist, add it to store
	TryGetOrAdd(prefix DataEntryPrefix, key []byte, value states.StateValue) error
	//Get key from state store
	TryGet(prefix DataEntryPrefix, key []byte) (*StateItem, error)
	//Delete key in store
	TryDelete(prefix DataEntryPrefix, key []byte)
	//iterator key in store
	Find(prefix DataEntryPrefix, key []byte) ([]*StateItem, error)
}

//MemoryCacheStore
type MemoryCacheStore interface {
	//Put the key-value pair to store
	Put(prefix byte, key []byte, value states.StateValue, state ItemState)
	//Get the value if key in store
	Get(prefix byte, key []byte) *StateItem
	//Delete the key in store
	Delete(prefix byte, key []byte)
	//Get all updated key-value set
	GetChangeSet() map[string]*StateItem
	// Get all key-value in store
	Find() []*StateItem
}

//EventStore save event notify
type EventStore interface {
	//SaveEventNotifyByTx save event notify gen by smart contract execution
	SaveEventNotifyByTx(txHash common.Uint256, notify *event.ExecuteNotify) error
	//Save transaction hashes which have event notify gen
	SaveEventNotifyByBlock(height uint32, txHashs []common.Uint256) error
	//GetEventNotifyByTx return event notify by transaction hash
	GetEventNotifyByTx(txHash common.Uint256) (*event.ExecuteNotify, error)
	//Commit event notify to store
	CommitTo() error
}

//State item type
type ItemState byte

//Status of item
const (
	None    ItemState = iota //no change
	Changed                  //which was be mark delete
	Deleted                  //which wad be mark delete
)

//State item struct
type StateItem struct {
	Key   string            //State key
	Value states.StateValue //State value
	State ItemState         //Status
	Trie  bool              //no use
}

func (e *StateItem) Copy() *StateItem {
	c := *e
	return &c
}
