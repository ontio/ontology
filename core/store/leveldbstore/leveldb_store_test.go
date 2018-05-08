package leveldbstore

import (
	"fmt"
	"os"
	"testing"
)

var testLevelDB *LevelDBStore

func TestMain(m *testing.M) {
	dbFile := "./test"
	var err error
	testLevelDB, err = NewLevelDBStore(dbFile)
	if err != nil {
		fmt.Printf("NewLevelDBStore error:%s\n", err)
		return
	}
	m.Run()
	testLevelDB.Close()
	os.RemoveAll(dbFile)
	os.RemoveAll("ActorLog")
}

func TestLevelDB(t *testing.T) {
	key := "foo"
	value := "bar"
	err := testLevelDB.Put([]byte(key), []byte(value))
	if err != nil {
		t.Errorf("Put error:%s", err)
		return
	}
	v, err := testLevelDB.Get([]byte(key))
	if err != nil {
		t.Errorf("Get error:%s", err)
		return
	}
	if string(v) != value {
		t.Errorf("Get error %s != %s", v, value)
		return
	}
	err = testLevelDB.Delete([]byte(key))
	if err != nil {
		t.Errorf("Delete error:%s", err)
		return
	}
	ok, err := testLevelDB.Has([]byte(key))
	if err != nil {
		t.Errorf("Has error:%s", err)
		return
	}
	if ok {
		t.Errorf("Key:%s shoule delete", key)
		return
	}
}

func TestBatch(t *testing.T) {
	testLevelDB.NewBatch()

	key1 := "foo1"
	value1 := "bar1"
	testLevelDB.BatchPut([]byte(key1), []byte(value1))

	key2 := "foo2"
	value2 := "bar2"
	testLevelDB.BatchPut([]byte(key2), []byte(value2))

	err := testLevelDB.BatchCommit()
	if err != nil {
		t.Errorf("BatchCommit error:%s", err)
		return
	}

	v1, err := testLevelDB.Get([]byte(key1))
	if err != nil {
		t.Errorf("Get error:%s", err)
		return
	}
	if string(v1) != value1 {
		t.Errorf("Get %s != %s", v1, value1)
		return
	}
}

func TestIterator(t *testing.T) {
	key := "foo"
	value := "bar"
	err := testLevelDB.Put([]byte(key), []byte(value))
	if err != nil {
		t.Errorf("Put error:%s", err)
		return
	}

	key1 := "foo1"
	value1 := "bar1"
	err = testLevelDB.Put([]byte(key1), []byte(value1))
	if err != nil {
		t.Errorf("Put error:%s", err)
		return
	}

	kvs := make(map[string]string)
	iter := testLevelDB.NewIterator([]byte("fo"))
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		kvs[string(key)] = string(value)
		fmt.Printf("Key:%s Value:%s\n", key, value)
	}
	iter.Release()

	v := kvs[key]
	if v != value {
		t.Errorf("TestIterator Key:%s value:%s != %s", key, v, value)
		return
	}

	v = kvs[key1]
	if v != value1 {
		t.Errorf("TestIterator Key:%s value:%s != %s", key1, v, value1)
		return
	}

}
