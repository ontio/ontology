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
package rocksdbstore

import (
	"fmt"
	"os"
	"testing"
)

var testRocksDB *RocksDBStore

func TestMain(m *testing.M) {
	dbFile := "./test"
	var err error
	testRocksDB, err = NewRocksDBStore(dbFile)
	if err != nil {
		fmt.Printf("NewLevelDBStore error:%s\n", err)
		return
	}
	m.Run()
	testRocksDB.Close()
	os.RemoveAll(dbFile)
	os.RemoveAll("ActorLog")
}

func TestRocksDB(t *testing.T) {
	key := "foo"
	value := "bar"
	err := testRocksDB.Put([]byte(key), []byte(value))
	if err != nil {
		t.Errorf("Put error:%s", err)
		return
	}
	v, err := testRocksDB.Get([]byte(key))
	if err != nil {
		t.Errorf("Get error:%s", err)
		return
	}
	if string(v) != value {
		t.Errorf("Get error %s != %s", v, value)
		return
	}
	err = testRocksDB.Delete([]byte(key))
	if err != nil {
		t.Errorf("Delete error:%s", err)
		return
	}
	ok, err := testRocksDB.Has([]byte(key))
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
	testRocksDB.NewBatch()

	key1 := "foo1"
	value1 := "bar1"
	testRocksDB.BatchPut([]byte(key1), []byte(value1))

	key2 := "foo2"
	value2 := "bar2"
	testRocksDB.BatchPut([]byte(key2), []byte(value2))

	err := testRocksDB.BatchCommit()
	if err != nil {
		t.Errorf("BatchCommit error:%s", err)
		return
	}

	v1, err := testRocksDB.Get([]byte(key1))
	if err != nil {
		t.Errorf("Get error:%s", err)
		return
	}
	if string(v1) != value1 {
		t.Errorf("Get %s != %s", v1, value1)
		return
	}
	v2, err := testRocksDB.Get([]byte(key2))
	if err != nil {
		t.Errorf("Get error:%s", err)
		return
	}
	if string(v2) != value2 {
		t.Errorf("Get %s != %s", v2, value2)
		return
	}
}

func TestIterator(t *testing.T) {
	key := "foo"
	value := "bar"
	err := testRocksDB.Put([]byte(key), []byte(value))
	if err != nil {
		t.Errorf("Put error:%s", err)
		return
	}

	key1 := "foo1"
	value1 := "bar1"
	err = testRocksDB.Put([]byte(key1), []byte(value1))
	if err != nil {
		t.Errorf("Put error:%s", err)
		return
	}

	key2 := "foo11"
	value2 := "bar11"
	err = testRocksDB.Put([]byte(key2), []byte(value2))
	if err != nil {
		t.Errorf("Put error:%s", err)
		return
	}

	kvs := make(map[string]string)
	iter := testRocksDB.NewIterator([]byte("foo1"))
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		kvs[string(key)] = string(value)
		fmt.Printf("TestIterator Key:%s Value:%s\n", key, value)
	}
	iter.Release()

	v := kvs[key]
	if v == value {
		t.Errorf("TestIterator Key:%s value:%s == %s", key, v, value)
		return
	}

	v = kvs[key1]
	if v != value1 {
		t.Errorf("TestIterator Key:%s value:%s != %s", key1, v, value1)
		return
	}

	v = kvs[key2]
	if v != value2 {
		t.Errorf("TestIterator Key:%s value:%s != %s", key2, v, value2)
		return
	}
}
