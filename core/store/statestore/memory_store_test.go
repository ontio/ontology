package statestore

import (
	"github.com/ontio/ontology/core/states"
	com "github.com/ontio/ontology/core/store/common"
	"testing"
)

func TestMemoryStore(t *testing.T) {
	memStore := NewMemDatabase()

	prefix := byte(com.ST_STORAGE)
	key := []byte("foo")
	value := &states.StorageItem{Value: []byte("bar")}
	memStore.Put(prefix, key, value, com.Changed)

	v := memStore.Get(prefix, key)
	storeItem := v.Value.(*states.StorageItem)
	if string(storeItem.Value) != string(value.Value) {
		t.Errorf("Get value:%s != %s", storeItem.Value, value.Value)
		return
	}

	key1 := []byte("foo1")
	value1 := &states.StorageItem{Value: []byte("bar1")}
	memStore.Put(prefix, key1, value1, com.None)

	set := memStore.GetChangeSet()
	if len(set) >= 2 {
		t.Errorf("GetChangeSet len:%d error. Shoule = 1", len(set))
		return
	}

	v = set[string(append([]byte{prefix}, key...))]
	storeItem = v.Value.(*states.StorageItem)
	if v == nil || string(storeItem.Value) != string(value.Value) {
		t.Errorf("GetChangeSet error, key:%s value:%s != %s", key, storeItem.Value, value.Value)
		return
	}

	memStore.Delete(prefix, key1)

	v = memStore.Get(prefix, key1)
	if v.State != com.Deleted {
		t.Errorf("State of key:%s != Deleted", key1)
		return
	}
}
