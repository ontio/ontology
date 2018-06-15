package test

import (
	"crypto/rand"
	"fmt"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/store/statestore"
	"github.com/ontio/ontology/core/types"
	. "github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/storage"
	"os"
	"testing"
)

func TestRandomCodeCrash(t *testing.T) {
	dbFile := "test"
	log.InitLog(4)
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

	for i := 1; i < 100; i++ {
		fmt.Print("test round ", i)
		code := make([]byte, i)
		for j := 0; j < 100000; j++ {
			rand.Read(code)

			cache := storage.NewCloneCache(testBatch)
			sc := SmartContract{
				Config:     config,
				Gas:        10000,
				CloneCache: cache,
			}
			engine, _ := sc.NewExecuteEngine(code)
			engine.Invoke()
		}
	}
}
