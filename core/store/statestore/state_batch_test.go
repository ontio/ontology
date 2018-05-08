package statestore

import (
	"fmt"
	"github.com/ontio/ontology/core/states"
	com "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"testing"
)

var (
	testBatch   *StateBatch
	testLevelDB *leveldbstore.LevelDBStore
)

func TestMain(m *testing.M) {
	dbFile := "test"
	var err error
	testLevelDB, err = leveldbstore.NewLevelDBStore(dbFile)
	if err != nil {
		fmt.Printf("NewLevelDBStore:%s error:%s", dbFile, err)
		return
	}
	testBatch = NewStateStoreBatch(NewMemDatabase(), testLevelDB)
	m.Run()
	testLevelDB.Close()
	os.RemoveAll(dbFile)
	os.RemoveAll("ActorLog")
}

func TestStateBatch_TryGetOrAdd(t *testing.T) {
	prefix := com.ST_STORAGE
	key := []byte("foo")
	value := &states.StorageItem{Value: []byte("bar")}

	err := testBatch.TryGetOrAdd(prefix, key, value)
	if err != nil {
		t.Errorf("TestStateBatch_TryGetOrAdd TryGetOrAdd error:%s", err)
		return
	}

	v, err := testBatch.TryGet(prefix, key)
	if err != nil {
		t.Errorf("TestStateBatch_TryGetOrAdd TryGet error:%s", err)
		return
	}

	storeItem := v.Value.(*states.StorageItem)
	if string(storeItem.Value) != string(value.Value) {
		t.Errorf("TestStateBatch_TryGetOrAdd value:%s != %s", storeItem.Value, value.Value)
		return
	}
}

func TestStateBatch_TryAdd(t *testing.T) {
	prefix := com.ST_STORAGE
	key := []byte("foo1")
	value := &states.StorageItem{Value: []byte("bar1")}

	err := testBatch.TryGetOrAdd(prefix, key, value)
	if err != nil {
		t.Errorf("TestStateBatch_TryGetOrAdd TryGetOrAdd error:%s", err)
		return
	}

	v, err := testBatch.TryGet(prefix, key)
	if err != nil {
		t.Errorf("TestStateBatch_TryGetOrAdd TryGet error:%s", err)
		return
	}

	storeItem := v.Value.(*states.StorageItem)
	if string(storeItem.Value) != string(value.Value) {
		t.Errorf("TestStateBatch_TryGetOrAdd value:%s != %s", storeItem.Value, value.Value)
		return
	}
}

func TestStateBatch_TryGetAndChange(t *testing.T) {
	prefix := com.ST_STORAGE
	key := []byte("foo")
	value := &states.StorageItem{Value: []byte("bar")}

	err := testBatch.TryGetOrAdd(prefix, key, value)
	if err != nil {
		t.Errorf("TestStateBatch_TryGetOrAdd TryGetOrAdd error:%s", err)
		return
	}

	key1 := []byte("foo1")
	value1 := &states.StorageItem{Value: []byte("bar1")}

	err = testBatch.TryGetOrAdd(prefix, key1, value1)
	if err != nil {
		t.Errorf("TestStateBatch_TryGetOrAdd TryGetOrAdd error:%s", err)
		return
	}

	testBatch.TryDelete(prefix, key)

	v, err := testBatch.TryGetAndChange(prefix, key)
	if err != nil {
		t.Errorf("TryGetAndChange error:%s", err)
		return
	}

	if v != nil {
		t.Errorf("TryGetAndChange error key:%s should nil", key)
		return
	}

	v1, err := testBatch.TryGetAndChange(prefix, key1)
	if err != nil {
		t.Errorf("TryGetAndChange error:%s", err)
		return
	}

	storeItem := v1.(*states.StorageItem)
	if string(storeItem.Value) != string(value1.Value) {
		t.Errorf("TestStateBatch_TryGetOrAdd value:%s != %s", storeItem.Value, value.Value)
		return
	}
}

func TestStateBatch_CommitTo(t *testing.T) {
	prefix := com.ST_STORAGE
	key := []byte("foo1")
	value := &states.StorageItem{Value: []byte("bar1")}

	err := testBatch.TryGetOrAdd(prefix, key, value)
	if err != nil {
		t.Errorf("TestStateBatch_TryGetOrAdd TryGetOrAdd error:%s", err)
		return
	}

	testLevelDB.NewBatch()
	err = testBatch.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error:%s", err)
		return
	}

	err = testLevelDB.BatchCommit()
	if err != nil {
		t.Errorf("BatchCommit error:%s", err)
		return
	}

	data, err := testLevelDB.Get(append([]byte{byte(prefix)}, key...))
	if err != nil && err != leveldb.ErrNotFound {
		t.Errorf("testLevelDB.Get error:%s", err)
		return
	}

	item, err := getStateObject(prefix, data)
	if err != nil {
		t.Errorf("TestStateBatch_TryGetOrAdd getStateObject eror:%s", err)
		return
	}

	v := item.(*states.StorageItem)
	if string(v.Value) != string(value.Value) {
		t.Errorf("TestStateBatch_TryGetOrAdd value:%s != %s", v.Value, value.Value)
		return
	}
}
