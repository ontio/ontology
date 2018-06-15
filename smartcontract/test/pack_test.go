package test

import (
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/store/statestore"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/ontio/ontology/vm/neovm"
	"os"
	"testing"
)

func TestPackCrash(t *testing.T) {
	// define a leaf
	byteCode := []byte{byte(neovm.PUSH0)}

	// build tree using array packing
	for i := 0; i < 10000; i++ {
		byteCode = append(byteCode, []byte{
			byte(neovm.DUP),
			byte(neovm.PUSH2),
			byte(neovm.PACK),
		}...)
	}
	// compare trees
	byteCode = append(byteCode, []byte{
		byte(neovm.DUP),
		byte(neovm.EQUAL),
	}...)
	// setup VM
	dbFile := "test"
	os.RemoveAll(dbFile)
	testLevelDB, err := leveldbstore.NewLevelDBStore(dbFile)
	if err != nil {
		panic(err)
	}
	store := statestore.NewMemDatabase()
	testBatch := statestore.NewStateStoreBatch(store, testLevelDB)
	config := &smartcontract.Config{
		Time:   10,
		Height: 10,
		Tx:     &types.Transaction{},
	}
	cache := storage.NewCloneCache(testBatch)
	sc := smartcontract.SmartContract{
		Config:     config,
		Gas:        200,
		CloneCache: cache,
	}
	engine, err := sc.NewExecuteEngine(byteCode)
	if err != nil {
		panic(err)
		// cause the VM to hang forever
		_, err = engine.Invoke()
		if err != nil {
		}
		panic(err)
	}
}
