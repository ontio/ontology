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
	"fmt"
	"sync"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	vt "github.com/ontio/ontology/validator/types"
)

type TXAttr struct {
	Height  uint32         // The height in which tx was verified
	Type    vt.VerifyType  // The validator flag: stateless/stateful
	ErrCode errors.ErrCode // Verified result
}

type TXEntry struct {
	Tx    *types.Transaction // transaction which has been verified
	Fee   common.Fixed64     // Total fee per transaction
	Attrs []*TXAttr          // the result from each validator
}

// TXPool contains all currently valid transactions. Transactions
// enter the pool when they are valid from the network,
// consensus or submitted. They exit the pool when they are included
// in the ledger.
type TXPool struct {
	sync.RWMutex
	txList map[common.Uint256]*TXEntry // Transactions which have been verified
}

// Init creates a new transaction pool to gather.
func (tp *TXPool) Init() {
	tp.Lock()
	defer tp.Unlock()
	tp.txList = make(map[common.Uint256]*TXEntry)
}

// AddTxList adds a valid transaction to the transaction pool. If the
// transaction is already in the pool, just return false. Parameter
// txEntry includes transaction, fee, and verified information(height,
// validator, error code).
func (tp *TXPool) AddTxList(txEntry *TXEntry) bool {
	tp.Lock()
	defer tp.Unlock()
	txHash := txEntry.Tx.Hash()
	if _, ok := tp.txList[txHash]; ok {
		log.Info("Transaction %x already existed in the pool\n", txHash)
		return false
	}

	tp.txList[txHash] = txEntry
	return true
}

// CleanTransactionList cleans the transaction list included in the ledger.
func (tp *TXPool) CleanTransactionList(txs []*types.Transaction) error {
	cleaned := 0
	txsNum := len(txs)
	for _, tx := range txs {
		if tx.TxType == types.BookKeeping {
			txsNum = txsNum - 1
			continue
		}
		if tp.DelTxList(tx) {
			cleaned++
		}
	}
	if txsNum != cleaned {
		log.Info(fmt.Sprintf("The Transactions num Unmatched. Expect %d,got %d .\n",
			txsNum, cleaned))
	}
	log.Debug(fmt.Sprintf("[cleanTransactionList],transaction %d Requested,%d cleaned, Remains %d in TxPool",
		txsNum, cleaned, tp.GetTransactionCount()))
	return nil
}

// DelTxList removes a single transaction from the pool.
func (tp *TXPool) DelTxList(tx *types.Transaction) bool {
	tp.Lock()
	defer tp.Unlock()
	txHash := tx.Hash()
	if _, ok := tp.txList[txHash]; !ok {
		return false
	}
	delete(tp.txList, txHash)
	return true
}

// GetTxPool gets the transaction lists from the pool for the consensus,
// if the byCount is marked, return the configured number at most; if the
// the byCount is not marked, return all of the current transaction pool.
func (tp *TXPool) GetTxPool(byCount bool) []*TXEntry {
	tp.RLock()
	defer tp.RUnlock()

	count := config.Parameters.MaxTxInBlock
	if count <= 0 {
		byCount = false
	}
	if len(tp.txList) < count || !byCount {
		count = len(tp.txList)
	}

	var num int
	txList := make([]*TXEntry, count)
	for _, txEntry := range tp.txList {
		txList[num] = txEntry
		num++
		if num >= count {
			break
		}
	}
	return txList
}

// GetTransaction returns a transaction if it is contained in the pool
// and nil otherwise.
func (tp *TXPool) GetTransaction(hash common.Uint256) *types.Transaction {
	tp.RLock()
	defer tp.RUnlock()
	if tx := tp.txList[hash]; tx == nil {
		return nil
	}
	return tp.txList[hash].Tx
}

// GetTxStatus returns a transaction status if it is contained in the pool
// and nil otherwise.
func (tp *TXPool) GetTxStatus(hash common.Uint256) *TxStatus {
	tp.RLock()
	defer tp.RUnlock()
	txEntry, ok := tp.txList[hash]
	if !ok {
		return nil
	}
	ret := &TxStatus{
		Hash:  hash,
		Attrs: txEntry.Attrs,
	}
	return ret
}

// GetTransactionCount returns the tx number of the pool.
func (tp *TXPool) GetTransactionCount() int {
	tp.RLock()
	defer tp.RUnlock()
	return len(tp.txList)
}
