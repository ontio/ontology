package test

import (
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/store/statestore"
	"github.com/ontio/ontology/core/types"
	. "github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/storage"
	"os"
	"testing"
)

func TestInfiniteLoopCrash(t *testing.T) {
	evilBytecode := []byte(" e\xff\u007f\xffhm\xb7%\xa7AAAAAAAAAAAAAAAC\xef\xed\x04INVERT\x95ve")
	dbFile := "test"
	defer func() {
		os.RemoveAll(dbFile)
	}()
	testLevelDB, err := leveldbstore.NewLevelDBStore(dbFile)
	if err != nil {
		t.Fatal(err)
	}
	store := statestore.NewMemDatabase()
	testBatch := statestore.NewStateStoreBatch(store, testLevelDB)
	config := &Config{
		Time:   10,
		Height: 10,
		Tx:     &types.Transaction{},
	}
	cache := storage.NewCloneCache(testBatch)
	sc := SmartContract{
		Config:     config,
		Gas:        10000,
		CloneCache: cache,
	}
	engine, err := sc.NewExecuteEngine(evilBytecode)
	if err != nil {
		t.Fatal(err)
	}
	_, err = engine.Invoke()
	if err != nil {
		t.Fatal(err)
	}
}
