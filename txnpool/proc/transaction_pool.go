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

// Package common privides constants, common types for other packages
package proc

import (
	"fmt"
	"sync"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	vt "github.com/ontio/ontology/validator/types"
	ttypes "github.com/ontio/ontology/txnpool/types"
)

// TXPool contains all currently valid transactions. Transactions
// enter the pool when they are valid from the network,
// consensus or submitted. They exit the pool when they are included
// in the ledger.
type TxPool struct {
	sync.RWMutex
	txEntrys map[common.Uint256]*ttypes.TxEntry // Transactions which have been verified
}

// Init creates a new transaction pool to gather.
func (self *TxPool) Init() {
	self.Lock()
	defer self.Unlock()
	self.txEntrys = make(map[common.Uint256]*ttypes.TxEntry)
}

// AddTxList adds a valid transaction to the transaction pool. If the
// transaction is already in the pool, just return false. Parameter
// txEntry includes transaction, fee, and verified information(height,
// validator, error code).
func (self *TxPool) appendTxEntry(txEntry *ttypes.TxEntry) bool {
	self.Lock()
	defer self.Unlock()
	txHash := txEntry.Tx.Hash()
	if _, ok := self.txEntrys[txHash]; ok {
		log.Info("Transaction %x already existed in the pool\n", txHash)
		return false
	}

	self.txEntrys[txHash] = txEntry
	return true
}

// CleanTransactionList cleans the transaction list included in the ledger.
func (self *TxPool) removeTransactions(txs []*types.Transaction) error {
	cleaned := 0
	txsNum := len(txs)
	self.Lock()
	defer self.Unlock()
	for _, tx := range txs {
		if tx.TxType == types.BookKeeping {
			txsNum = txsNum - 1
			continue
		}
		if _, ok := self.txEntrys[tx.Hash()]; ok {
			delete(self.txEntrys, tx.Hash())
			cleaned++
		}
	}

	log.Debug(fmt.Sprintf("[cleanTransactionList],transaction %d Requested,%d cleaned, Remains %d in TxPool",
		txsNum, cleaned, len(self.txEntrys)))
	return nil
}

// DelTxList removes a single transaction from the pool.
func (self *TxPool) deleteTransaction(tx *types.Transaction) bool {
	self.Lock()
	defer self.Unlock()
	txHash := tx.Hash()
	if _, ok := self.txEntrys[txHash]; !ok {
		return false
	}
	delete(self.txEntrys, txHash)
	return true
}

// compareTxHeight compares a verifed transaction's height with the next
// block height from consensus. If the height is less than the next block
// height, re-verify it.
func (self *TxPool) compareTxHeight(txEntry *ttypes.TxEntry, height uint32) bool {
	for _, v := range txEntry.VerifyResults {
		if v.Type == vt.Statefull &&
			v.Height < height {
			return false
		}
	}
	return true
}

// GetTxPool gets the transaction lists from the pool for the consensus,
// if the byCount is marked, return the configured number at most; if the
// the byCount is not marked, return all of the current transaction pool.
func (self *TxPool) getTransactions(byCount bool, height uint32) ([]*ttypes.TxEntry,
	[]*types.Transaction) {
	self.RLock()
	defer self.RUnlock()

	count := config.Parameters.MaxTxInBlock
	if count <= 0 {
		byCount = false
	}
	if len(self.txEntrys) < count || !byCount {
		count = len(self.txEntrys)
	}

	var num int
	txList := make([]*ttypes.TxEntry, 0, count)
	oldTxList := make([]*types.Transaction, 0)
	for _, txEntry := range self.txEntrys {
		if !self.compareTxHeight(txEntry, height) {
			oldTxList = append(oldTxList, txEntry.Tx)
			continue
		}
		txList = append(txList, txEntry)
		num++
		if num >= count {
			break
		}
	}

	return txList, oldTxList
}

// GetTransaction returns a transaction if it is contained in the pool
// and nil otherwise.
func (self *TxPool) getTransaction(hash common.Uint256) *types.Transaction {
	self.RLock()
	defer self.RUnlock()
	if tx := self.txEntrys[hash]; tx == nil {
		return nil
	}
	return self.txEntrys[hash].Tx
}

// GetTxStatus returns a transaction status if it is contained in the pool
// and nil otherwise.
func (self *TxPool) getTxVerifyStatus(hash common.Uint256) *ttypes.TxVerifyStatus {
	self.RLock()
	defer self.RUnlock()
	txEntry, ok := self.txEntrys[hash]
	if !ok {
		return nil
	}
	ret := &ttypes.TxVerifyStatus{
		Hash:  hash,
		VerifyResults: txEntry.VerifyResults,
	}
	return ret
}

// GetTransactionCount returns the tx number of the pool.
func (self *TxPool) getTransactionCount() int {
	self.RLock()
	defer self.RUnlock()
	return len(self.txEntrys)
}

// GetUnverifiedTxs checks the tx list in the block from consensus,
// and returns verified tx list, unverified tx list, and
// the tx list to be re-verified
func (self *TxPool) getVerifyBlockResult(txs []*types.Transaction,
	height uint32) *ttypes.BlockVerifyStatus {
	self.Lock()
	defer self.Unlock()
	res := &ttypes.BlockVerifyStatus{
		VerifiedTxs:   make([]*ttypes.TxResult, 0, len(txs)),
		UnVerifiedTxs: make([]*types.Transaction, 0),
		ReVerifyTxs:        make([]*types.Transaction, 0),
	}
	for _, tx := range txs {
		txEntry := self.txEntrys[tx.Hash()]
		if txEntry == nil {
			res.UnVerifiedTxs = append(res.UnVerifiedTxs,
				tx)
			continue
		}

		if !self.compareTxHeight(txEntry, height) {
			delete(self.txEntrys, tx.Hash())
			res.ReVerifyTxs = append(res.ReVerifyTxs, txEntry.Tx)
			continue
		}

		for _, v := range txEntry.VerifyResults {
			if v.Type == vt.Statefull {
				entry := &ttypes.TxResult{
					Tx:      tx,
					Height:  v.Height,
					ErrCode: v.ErrCode,
				}
				res.VerifiedTxs = append(res.VerifiedTxs,
					entry)
				break
			}
		}
	}

	return res
}
