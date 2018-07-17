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
	"strings"

	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
)

type StateBatch struct {
	store  common.PersistStore
	memory map[string]states.StateValue
	//readCache map[string]states.StateValue
	dbErr error
}

func NewStateStoreBatch(store common.PersistStore) *StateBatch {
	return &StateBatch{
		store:  store,
		memory: make(map[string]states.StateValue),
		//readCache: make(map[string]states.StateValue),
	}
}

func (self *StateBatch) Find(prefix common.DataEntryPrefix, key []byte) ([]*common.StateItem, error) {
	var sts []*common.StateItem
	pkey := append([]byte{byte(prefix)}, key...)
	iter := self.store.NewIterator(pkey)
	defer iter.Release()
	for iter.Next() {
		k := iter.Key()
		kv := k[1:]
		if _, ok := self.memory[string(pkey)]; ok == false {
			value := iter.Value()
			state, err := decodeStateObject(prefix, value)
			if err != nil {
				return nil, err
			}
			sts = append(sts, &common.StateItem{Key: string(kv), Value: state})
		}
	}

	for k, v := range self.memory {
		if v != nil && strings.HasPrefix(k, string(pkey)) {
			sts = append(sts, &common.StateItem{Key: k, Value: v})
		}
	}
	return sts, nil
}

func (self *StateBatch) TryAdd(prefix common.DataEntryPrefix, key []byte, value states.StateValue) {
	pkey := append([]byte{byte(prefix)}, key...)
	self.memory[string(pkey)] = value
	//delete(self.readCache, string(pkey))
}

func (self *StateBatch) TryGetOrAdd(prefix common.DataEntryPrefix, key []byte, value states.StateValue) error {
	val, err := self.TryGet(prefix, key)
	if err != nil {
		return err
	}
	if val == nil {
		self.TryAdd(prefix, key, value)
	}
	return nil
}

func (self *StateBatch) TryGet(prefix common.DataEntryPrefix, key []byte) (states.StateValue, error) {
	pkey := append([]byte{byte(prefix)}, key...)
	if state, ok := self.memory[string(pkey)]; ok {
		return state, nil
	}
	//if state, ok := self.readCache[string(pkey)]; ok {
	//	return state, nil
	//}
	enc, err := self.store.Get(pkey)
	if err != nil {
		if err == common.ErrNotFound {
			return nil, nil
		}
		errs := errors.NewDetailErr(err, errors.ErrNoCode, "[TryGet], store get data failed.")
		self.SetError(errs)
		return nil, errs
	}

	stateVal, err := decodeStateObject(prefix, enc)
	if err != nil {
		return nil, err
	}
	//self.readCache[string(pkey)] = stateVal
	return stateVal, nil
}

func (self *StateBatch) TryDelete(prefix common.DataEntryPrefix, key []byte) {
	pkey := append([]byte{byte(prefix)}, key...)
	self.memory[string(pkey)] = nil
	//delete(self.readCache, string(pkey))
}

func (self *StateBatch) CommitTo() error {
	for k, v := range self.memory {
		if v == nil {
			self.store.BatchDelete([]byte(k))
		} else {
			data := new(bytes.Buffer)
			err := v.Serialize(data)
			if err != nil {
				return fmt.Errorf("error: key %v, value:%v", k, v)
			}
			self.store.BatchPut([]byte(k), data.Bytes())
		}
	}
	return nil
}

func (self *StateBatch) SetError(err error) {
	if self.dbErr == nil {
		self.dbErr = err
	}
}

func (self *StateBatch) Error() error {
	return self.dbErr
}

func decodeStateObject(prefix common.DataEntryPrefix, enc []byte) (states.StateValue, error) {
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
		panic("[decodeStateObject] invalid state type!")
	}
}
