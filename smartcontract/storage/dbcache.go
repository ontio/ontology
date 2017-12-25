package storage

import (
	"github.com/Ontology/core/store"
	"github.com/Ontology/core/states"
)

type StateItem struct {
	Prefix store.DataEntryPrefix
	Key    string
	Value  states.IStateValue
	State  store.ItemState
}

type Memory map[string]*StateItem

type CloneCache struct {
	Memory Memory
	Store  store.IStateStore
}

func NewCloneCache(store  store.IStateStore) *CloneCache {
	return &CloneCache{
		Memory: make(Memory),
		Store: store,
	}
}

func (cloneCache *CloneCache) Commit() {
	for _, v := range cloneCache.Memory {
		if v.State == store.Deleted {
			cloneCache.Store.TryDelete(v.Prefix, []byte(v.Key))
		} else if v.State == store.Changed {
			cloneCache.Store.TryAdd(v.Prefix, []byte(v.Key), v.Value, true)
		}
	}
}

func (cloneCache *CloneCache) Add(prefix store.DataEntryPrefix, key []byte, value states.IStateValue) {
	cloneCache.Memory[string(append([]byte{byte(prefix)}, key...))] = &StateItem{
		Prefix: prefix,
		Key: string(key),
		Value: value,
		State: store.Changed,
	}
}

func (cloneCache *CloneCache) GetOrAdd(prefix store.DataEntryPrefix, key []byte, value states.IStateValue) (states.IStateValue, error) {
	if v, ok := cloneCache.Memory[string(append([]byte{byte(prefix)}, key...))]; ok {
		return v.Value, nil
	}
	item, err := cloneCache.Store.TryGet(prefix, key)
	if err != nil {
		return nil, err
	}
	if item != nil && item.State != store.Deleted {
		return item.Value, nil
	}
	cloneCache.Memory[string(append([]byte{byte(prefix)}, key...))] = &StateItem{Prefix: prefix, Key: string(key), Value: value, State: store.Changed}
	return value, nil
}

func (cloneCache *CloneCache) Get(prefix store.DataEntryPrefix, key []byte) (states.IStateValue, error) {
	if v, ok := cloneCache.Memory[string(append([]byte{byte(prefix)}, key...))]; ok {
		if v.State == store.Deleted {
			return nil, nil
		}
		return v.Value, nil
	}
	item, err := cloneCache.Store.TryGet(prefix, key)
	if err != nil {
		return nil, err
	}
	return item.Value, nil
}

func (cloneCache *CloneCache) Delete(prefix store.DataEntryPrefix, key []byte) {
	if v, ok := cloneCache.Memory[string(append([]byte{byte(prefix)}, key...))]; ok {
		v.State = store.Deleted
	}else {
		cloneCache.Memory[string(append([]byte{byte(prefix)}, key...))] = &StateItem{
			State: store.Deleted,
		}
	}
}



