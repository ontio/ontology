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

package common

import (
	"testing"
	"time"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"
)

var (
	txn *types.Transaction
)

func init() {
	log.InitLog(log.InfoLog, log.Stdout)

	mutable := &types.MutableTransaction{
		TxType:  types.InvokeNeo,
		Nonce:   uint32(time.Now().Unix()),
		Payload: &payload.InvokeCode{Code: []byte{}},
	}

	txn, _ = mutable.IntoImmutable()
}

func TestTxPool(t *testing.T) {
	txPool := &TXPool{}
	txPool.Init()

	txEntry := &TXEntry{
		Tx:    txn,
		Attrs: []*TXAttr{},
	}

	ret := txPool.AddTxList(txEntry)
	if ret == false {
		t.Error("Failed to add tx to the pool")
		return
	}

	ret = txPool.AddTxList(txEntry)
	if ret == true {
		t.Error("Failed to add tx to the pool")
		return
	}

	txList, oldTxList := txPool.GetTxPool(true, 0)
	for _, v := range txList {
		assert.NotNil(t, v)
	}

	for _, v := range oldTxList {
		assert.NotNil(t, v)
	}

	entry := txPool.GetTransaction(txn.Hash())
	if entry == nil {
		t.Error("Failed to get the transaction")
		return
	}

	assert.Equal(t, txn.Hash(), entry.Hash())

	status := txPool.GetTxStatus(txn.Hash())
	if status == nil {
		t.Error("failed to get the status")
		return
	}

	assert.Equal(t, txn.Hash(), status.Hash)

	count := txPool.GetTransactionCount()
	assert.Equal(t, count, 1)

	txPool.CleanTransactionList([]*types.Transaction{txn})
}
