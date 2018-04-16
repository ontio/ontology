package statestore

import (
	"fmt"
	"github.com/ontio/ontology/core/states"
	com "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"os"
	"testing"
	//"github.com/syndtr/goleveldb/leveldb"
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

	v , err := testBatch.TryGet(prefix, key)
	if err != nil {
		t.Errorf("TestStateBatch_TryGetOrAdd TryGet error:%s", err)
		return
	}

	storeItem := v.Value.(*states.StorageItem)
	if string(storeItem.Value) != string(value.Value){
		t.Errorf("TestStateBatch_TryGetOrAdd value:%s != %s", storeItem.Value, value.Value)
		return
	}
}
