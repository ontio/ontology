package statestore

import (
	"bytes"
	"fmt"
	"github.com/Ontology/core/payload"
	. "github.com/Ontology/core/states"
	. "github.com/Ontology/core/store/common"
	"github.com/syndtr/goleveldb/leveldb"
)

type StateBatch struct {
	store       IStore
	memoryStore IMemoryStore
}

func NewStateStoreBatch(memoryStore IMemoryStore, store IStore) *StateBatch {
	return &StateBatch{
		store:       store,
		memoryStore: memoryStore,
	}
}

func (self *StateBatch) Find(prefix DataEntryPrefix, key []byte) ([]*StateItem, error) {
	var states []*StateItem
	iter := self.store.NewIterator(append([]byte{byte(prefix)}, key...))
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		state, err := getStateObject(prefix, value)
		if err != nil {
			return nil, err
		}
		states = append(states, &StateItem{Key: string(key[1:]), Value: state})
	}
	return states, nil
}

func (self *StateBatch) TryAdd(prefix DataEntryPrefix, key []byte, value IStateValue, trie bool) {
	self.setStateObject(byte(prefix), key, value, Changed, trie)
}

func (self *StateBatch) TryGetOrAdd(prefix DataEntryPrefix, key []byte, value IStateValue, trie bool) error {
	state := self.memoryStore.Get(byte(prefix), key)
	if state != nil {
		if state.State == Deleted {
			self.setStateObject(byte(prefix), key, value, Changed, trie)
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
	self.setStateObject(byte(prefix), key, value, Changed, trie)
	return nil
}

func (self *StateBatch) TryGet(prefix DataEntryPrefix, key []byte) (*StateItem, error) {
	state := self.memoryStore.Get(byte(prefix), key)
	if state != nil {
		if state.State == Deleted {
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
	self.setStateObject(byte(prefix), key, stateVal, None, false)
	return &StateItem{Key: string(append([]byte{byte(prefix)}, key...)), Value: stateVal, State: None}, nil
}

func (self *StateBatch) TryGetAndChange(prefix DataEntryPrefix, key []byte, trie bool) (IStateValue, error) {
	state := self.memoryStore.Get(byte(prefix), key)
	if state != nil {
		if state.State == Deleted {
			return nil, nil
		} else if state.State == None {
			state.State = Changed
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
	self.setStateObject(byte(prefix), key, val, Changed, trie)
	return val, nil
}

func (self *StateBatch) TryDelete(prefix DataEntryPrefix, key []byte) {
	self.memoryStore.Delete(byte(prefix), key)
}

func (self *StateBatch) CommitTo() error {
	for k, v := range self.memoryStore.GetChangeSet() {
		if v.State == Deleted {
			if err := self.store.BatchDelete([]byte(k)); err != nil {
				return err
			}
		} else {
			data := new(bytes.Buffer)
			err := v.Value.Serialize(data)
			if err != nil {
				return fmt.Errorf("error: key %v, value:%v", k, v.Value)
			}
			if err = self.store.BatchPut([]byte(k), data.Bytes()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *StateBatch) Change(prefix byte, key []byte, trie bool) {
	this.memoryStore.Change(prefix, key, trie)
}

func (self *StateBatch) setStateObject(prefix byte, key []byte, value IStateValue, state ItemState, trie bool) {
	self.memoryStore.Put(prefix, key, value, state, trie)
}

func getStateObject(prefix DataEntryPrefix, enc []byte) (IStateValue, error) {
	reader := bytes.NewBuffer(enc)
	switch prefix {
	case ST_BookKeeper:
		bookKeeper := new(payload.BookKeeper)
		if err := bookKeeper.Deserialize(reader); err != nil {
			return nil, err
		}
		return bookKeeper, nil
	case ST_Contract:
		contract := new(payload.DeployCode)
		if err := contract.Deserialize(reader); err != nil {
			return nil, err
		}
		return contract, nil
	case ST_Storage:
		storage := new(StorageItem)
		if err := storage.Deserialize(reader); err != nil {
			return nil, err
		}
		return storage, nil
	default:
		panic("[getStateObject] invalid state type!")
	}
}
