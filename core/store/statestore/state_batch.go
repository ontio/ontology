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
	"bytes"
	"fmt"

	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
	"github.com/syndtr/goleveldb/leveldb"
)

type StateBatch struct {
	store       common.PersistStore
	memoryStore common.MemoryCacheStore
}

func NewStateStoreBatch(memoryStore common.MemoryCacheStore, store common.PersistStore) *StateBatch {
	return &StateBatch{
		store:       store,
		memoryStore: memoryStore,
	}
}

func (self *StateBatch) Find(prefix common.DataEntryPrefix, key []byte) ([]*common.StateItem, error) {
	var states []*common.StateItem
	iter := self.store.NewIterator(append([]byte{byte(prefix)}, key...))
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		state, err := getStateObject(prefix, value)
		if err != nil {
			return nil, err
		}
		states = append(states, &common.StateItem{Key: string(key[1:]), Value: state})
	}
	return states, nil
}

func (self *StateBatch) TryAdd(prefix common.DataEntryPrefix, key []byte, value states.StateValue) {
	self.setStateObject(byte(prefix), key, value, common.Changed)
}

func (self *StateBatch) TryGetOrAdd(prefix common.DataEntryPrefix, key []byte, value states.StateValue) error {
	state := self.memoryStore.Get(byte(prefix), key)
	if state != nil {
		if state.State == common.Deleted {
			self.setStateObject(byte(prefix), key, value, common.Changed)
			return nil
		}
		return nil
	}
	item, err := self.store.Get(append([]byte{byte(prefix)}, key...))
	if err != nil && err != leveldb.ErrNotFound {
		return err
	}
	if item != nil {
		return nil
	}
	self.setStateObject(byte(prefix), key, value, common.Changed)
	return nil
}

func (self *StateBatch) TryGet(prefix common.DataEntryPrefix, key []byte) (*common.StateItem, error) {
	state := self.memoryStore.Get(byte(prefix), key)
	if state != nil {
		if state.State == common.Deleted {
			return nil, nil
		}
		return state, nil
	}
	enc, err := self.store.Get(append([]byte{byte(prefix)}, key...))
	if err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	if enc == nil {
		return nil, nil
	}
	stateVal, err := getStateObject(prefix, enc)
	if err != nil {
		return nil, err
	}
	self.setStateObject(byte(prefix), key, stateVal, common.None)
	return &common.StateItem{Key: string(append([]byte{byte(prefix)}, key...)), Value: stateVal, State: common.None}, nil
}

func (self *StateBatch) TryGetAndChange(prefix common.DataEntryPrefix, key []byte) (states.StateValue, error) {
	state := self.memoryStore.Get(byte(prefix), key)
	if state != nil {
		if state.State == common.Deleted {
			return nil, nil
		} else if state.State == common.None {
			state.State = common.Changed
		}
		return state.Value, nil
	}
	k := append([]byte{byte(prefix)}, key...)
	enc, err := self.store.Get(k)
	if err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	if enc == nil {
		return nil, nil
	}

	val, err := getStateObject(prefix, enc)
	if err != nil {
		return nil, err
	}
	self.setStateObject(byte(prefix), key, val, common.Changed)
	return val, nil
}

func (self *StateBatch) TryDelete(prefix common.DataEntryPrefix, key []byte) {
	self.memoryStore.Delete(byte(prefix), key)
}

func (self *StateBatch) CommitTo() error {
	for k, v := range self.memoryStore.GetChangeSet() {
		if v.State == common.Deleted {
			self.store.BatchDelete([]byte(k))
		} else {
			data := new(bytes.Buffer)
			err := v.Value.Serialize(data)
			if err != nil {
				return fmt.Errorf("error: key %v, value:%v", k, v.Value)
			}
			self.store.BatchPut([]byte(k), data.Bytes())
		}
	}
	return nil
}

func (self *StateBatch) setStateObject(prefix byte, key []byte, value states.StateValue, state common.ItemState) {
	self.memoryStore.Put(prefix, key, value, state)
}

func getStateObject(prefix common.DataEntryPrefix, enc []byte) (states.StateValue, error) {
	reader := bytes.NewBuffer(enc)
	switch prefix {
	case common.ST_BOOKKEEPER:
		bookkeeper := new(payload.Bookkeeper)
		if err := bookkeeper.Deserialize(reader); err != nil {
			return nil, err
		}
		return bookkeeper, nil
	case common.ST_CONTRACT:
		contract := new(payload.DeployCode)
		if err := contract.Deserialize(reader); err != nil {
			return nil, err
		}
		return contract, nil
	case common.ST_STORAGE:
		storage := new(states.StorageItem)
		if err := storage.Deserialize(reader); err != nil {
			return nil, err
		}
		return storage, nil
	default:
		panic("[getStateObject] invalid state type!")
	}
}
