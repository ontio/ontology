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
	"testing"

	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/vm/neovm"
	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	byteCode := []byte{
		byte(neovm.NEWMAP),
		byte(neovm.DUP),   // dup map
		byte(neovm.PUSH0), // key (index)
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
		}...)

	// count faults vs successful executions
	N := 1024
	faults := 0

	//dbFile := "/tmp/test"
	//os.RemoveAll(dbFile)
	//testLevelDB, err := leveldbstore.NewLevelDBStore(dbFile)
	//if err != nil {
	//	panic(err)
	//}

	for n := 0; n < N; n++ {
		// Setup Execution Environment
		//store := statestore.NewMemDatabase()
		//testBatch := statestore.NewStateStoreBatch(store, testLevelDB)
		config := &smartcontract.Config{
			Time:   10,
			Height: 10,
			Tx:     &types.Transaction{},
		}
		sc := smartcontract.SmartContract{
			Config:  config,
			Gas:     100,
			CacheDB: nil,
		}
		engine, err := sc.NewExecuteEngine(byteCode, types.InvokeNeo)

		_, err = engine.Invoke()
		if err != nil {
			fmt.Println("err:", err)
			faults += 1
		}
	}
	assert.Equal(t, faults, 0)

}
