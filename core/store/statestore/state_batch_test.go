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
	"fmt"
	"github.com/ontio/ontology/core/states"
	com "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
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
	if err != nil && err != com.ErrNotFound {
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
