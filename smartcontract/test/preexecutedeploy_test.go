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
	"os"
	"testing"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/store/ledgerstore"
	"github.com/stretchr/testify/assert"
)

func TestPreExecuteContractWasmDeploy(t *testing.T) {
	acct := account.NewAccount("")
	testLedgerStore, err := ledgerstore.NewLedgerStore("test/ledgerfortmp", 0)
	/** file: test_create.wat
		(module
		  (type (;0;) (func))
		  (type (;1;) (func (param i32 i32)))
		  (import "env" "ontio_return" (func (;0;) (type 1)))
		  (func (;1;) (type 0)
			i32.const 0
		    i64.const 2222
			i64.store
			i32.const 0
			i32.const 8
			call 0
			)
		  (memory (;0;) 1)
		  (export "invoke" (func 1)))
	 **/
	code := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x09, 0x02, 0x60, 0x00, 0x00, 0x60, 0x02, 0x7f, 0x7f, 0x00, 0x02, 0x14, 0x01, 0x03, 0x65, 0x6e, 0x76, 0x0c, 0x6f, 0x6e, 0x74, 0x69, 0x6f, 0x5f, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x00, 0x01, 0x03, 0x02, 0x01, 0x00, 0x05, 0x03, 0x01, 0x00, 0x01, 0x07, 0x0a, 0x01, 0x06, 0x69, 0x6e, 0x76, 0x6f, 0x6b, 0x65, 0x00, 0x01, 0x0a, 0x12, 0x01, 0x10, 0x00, 0x41, 0x00, 0x42, 0xae, 0x11, 0x37, 0x03, 0x00, 0x41, 0x00, 0x41, 0x08, 0x10, 0x00, 0x0b}
	mutable, _ := utils.NewDeployCodeTransaction(0, 100000000, code, payload.NEOVM_TYPE, "name", "version",
		"author", "email", "desc")
	_ = utils.SignTransaction(acct, mutable)
	tx, err := mutable.IntoImmutable()
	_, err = testLedgerStore.PreExecuteContract(tx)
	assert.EqualError(t, err, "this code is wasm binary. can not deployed as neo contract")

	mutable, _ = utils.NewDeployCodeTransaction(0, 100000000, code, payload.WASMVM_TYPE, "name", "version",
		"author", "email", "desc")
	_ = utils.SignTransaction(acct, mutable)
	tx, err = mutable.IntoImmutable()
	_, err = testLedgerStore.PreExecuteContract(tx)
	assert.Nil(t, err)

	_ = os.RemoveAll("./test")
}
