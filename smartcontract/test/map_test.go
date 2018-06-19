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

package test

import (
	"fmt"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/store/statestore"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/ontio/ontology/vm/neovm"
	"os"
	"testing"
)

func TestMap(t *testing.T) {
	byteCode := []byte{
		byte(neovm.NEWMAP),
		byte(neovm.DUP),   // dup map
		byte(neovm.PUSH0), // key (index)
		byte(neovm.SETITEM),

		byte(neovm.DUP),   // dup map
		byte(neovm.PUSH0), // key (index)
		byte(neovm.PUSH1), // value (newItem)
		byte(neovm.SETITEM),
	}

	// pick a value out
	byteCode = append(byteCode,
		[]byte{ // extract element
			byte(neovm.DUP),   // dup map (items)
			byte(neovm.PUSH0), // key (index)

			byte(neovm.PICKITEM),
			byte(neovm.JMPIF), // dup map (items)
			0x04, 0x00,        // skip a drop?
			byte(neovm.DROP),
			byte(neovm.DROP),
		}...)

	// count faults vs successful executions
	N := 10240000
	faults := 0

	dbFile := "/tmp/test"
	os.RemoveAll(dbFile)
	testLevelDB, err := leveldbstore.NewLevelDBStore(dbFile)
	if err != nil {
		panic(err)
	}

	for n := 0; n < N; n++ {
		// Setup Execution Enviroment
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
			Gas:        100,
			CloneCache: cache,
		}
		engine, err := sc.NewExecuteEngine(byteCode)
		if err != nil {
			panic(err) // Run Code
			_, err = engine.Invoke()
			if err != nil {
				faults += 1
			}
		}
	}
	fmt.Println("Ran code", N, "times, experienced", faults, "faults")
}
