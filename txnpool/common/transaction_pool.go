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

// Package common provides constants, common types for other packages
package common

import (
	"sort"
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
	Attrs []*TXAttr          // the result from each validator
}

type VerifiedTx struct {
	Tx             *types.Transaction // transaction which has been verified
	VerifiedHeight uint32
	Nonce          uint64
}

func (self *VerifiedTx) IsVerfiyExpired(height uint32) bool {
	return self.VerifiedHeight < height
}

func (self *VerifiedTx) GetAttrs() []*TXAttr {
	return []*TXAttr{
		{
			Height:  0,
			Type:    vt.Stateless,
			ErrCode: errors.ErrNoError,
		}, {
			Height:  self.VerifiedHeight,
			Type:    vt.Stateful,
			ErrCode: errors.ErrNoError,
		},
	}
}

type UserNonceInfo struct {
	Height uint32
	Nonce  uint64
}

// TXPool contains all currently valid transactions. Transactions
// enter the pool when they are valid from the network,
// consensus or submitted. They exit the pool when they are included
// in the ledger.
type TXPool struct {
	sync.RWMutex
	validTxMap            map[common.Uint256]*VerifiedTx    // Transactions which have been verified
	eipTxPool             map[common.Address]*txSortedMap   // The tx pool that holds the valid transaction
	userLatestEiptxHeight map[common.Address]*UserNonceInfo // record last block height user commit eiptx
}

func NewTxPool() *TXPool {
	return &TXPool{
		validTxMap:            make(map[common.Uint256]*VerifiedTx),
		eipTxPool:             make(map[common.Address]*txSortedMap),
		userLatestEiptxHeight: make(map[common.Address]*UserNonceInfo),
	}
}

func (s *TXPool) CleanStaledEIPTx(height uint32) {
	s.Lock()
	defer s.Unlock()
	if len(s.validTxMap) > MAX_LIMITATION {
		for addr, v := range s.userLatestEiptxHeight {
			if height >= v.Height+EIPTX_EXPIRATION_BLOCKS {
				if list := s.eipTxPool[addr]; list != nil {
					for _, txn := range list.items {
						delete(s.validTxMap, txn.Hash())
					}

					delete(s.eipTxPool, addr)
				}
				delete(s.userLatestEiptxHeight, addr)
			}
		}
	}
}

// get next nonce from txpool
func (s *TXPool) NextNonce(addr common.Address) uint64 {
	s.RLock()
	defer s.RUnlock()
	list := s.eipTxPool[addr]
	if list == nil {
		return 0
	}

	//if 1st tx nonce in eiptxpool is not 0,need to check whether it equals ledger nonce
	//otherwise return the ledgerNonce
	heading := list.Heading()
	if heading != nil && len(heading) > 0 {
		headNonce := heading[0].Nonce
		if headNonce > 0 && uint64(headNonce) != s.userLatestEiptxHeight[addr].Nonce {
			return s.userLatestEiptxHeight[addr].Nonce
		}

		return uint64(heading[len(heading)-1].Nonce + 1)
	}
	return 0

}

func (s *TXPool) getTxListByAddr(addr common.Address) *txSortedMap {
	if _, ok := s.eipTxPool[addr]; !ok {
		s.eipTxPool[addr] = newTxSortedMap()
	}

	return s.eipTxPool[addr]
}

func (s *TXPool) addEIPTxPool(trans *types.Transaction) (replaced *types.Transaction, err errors.ErrCode) {
	list := s.getTxListByAddr(trans.Payer)

	// does the same nonce exist?
	old := list.Get(uint64(trans.Nonce))
	if old == nil {
		s.eipTxPool[trans.Payer].Put(trans)
		return nil, errors.ErrNoError
	}

	if trans.GasPrice > old.GasPrice*101/100 {
		log.Infof("replace transaction %s with lower gas fee", old.Hash().ToHexString())
		s.eipTxPool[trans.Payer].Put(trans)
		return old, errors.ErrNoError
	}

	return nil, errors.ErrSameNonceExist
}

// AddTxList adds a valid transaction to the transaction pool. If the
// transaction is already in the pool, just return false. Parameter
// txEntry includes transaction, fee, and verified information(height,
// validator, error code).
func (tp *TXPool) AddTxList(txEntry *VerifiedTx) errors.ErrCode {
	tp.Lock()
	defer tp.Unlock()
	txHash := txEntry.Tx.Hash()
	if txEntry.Tx.IsEipTx() {
		//check the new tx nonce should not be greater than latest nonce + 1000
		if uint64(txEntry.Tx.Nonce) >= txEntry.Nonce+EIPTX_NONCE_MAX_GAP {
			return errors.ErrETHTxNonceToobig
		}

		repalced, code := tp.addEIPTxPool(txEntry.Tx)
		if repalced != nil {
			delete(tp.validTxMap, repalced.Hash())
		}
		if !code.Success() {
			return code
		}
		if tp.userLatestEiptxHeight[txEntry.Tx.Payer] == nil {
			tp.userLatestEiptxHeight[txEntry.Tx.Payer] = &UserNonceInfo{
				Height: txEntry.VerifiedHeight,
				Nonce:  txEntry.Nonce,
			}
		}
	}

	if _, ok := tp.validTxMap[txHash]; ok {
		ShowTraceLog("AddTxList: transaction %x is already in the pool", txHash)
		return errors.ErrDuplicatedTx
	}

	tp.validTxMap[txHash] = txEntry
	return errors.ErrNoError
}

//clean the EIP txpool and eip pending txpool under the tx nonce
func (s *TXPool) cleanCompletedEipTxPool(txs []*types.Transaction, height uint32) []*types.Transaction {
	var cleaned []*types.Transaction
	for _, tx := range txs {
		if tx.IsEipTx() {
			if _, ok := s.eipTxPool[tx.Payer]; ok {
				cleaned = append(cleaned, s.eipTxPool[tx.Payer].Forward(uint64(tx.Nonce+1))...)
				if s.eipTxPool[tx.Payer].Len() == 0 {
					delete(s.eipTxPool, tx.Payer)
					delete(s.userLatestEiptxHeight, tx.Payer)
				} else {
					s.userLatestEiptxHeight[tx.Payer] = &UserNonceInfo{
						Height: height,
						Nonce:  uint64(tx.Nonce) + 1,
					}
				}
			}
		}
	}
	return cleaned
}

// cleans the transaction list included in the ledger.
func (tp *TXPool) CleanCompletedTransactionList(txs []*types.Transaction, height uint32) {
	cleaned := 0
	txsNum := len(txs)
	tp.Lock()
	defer tp.Unlock()
	cleanedEips := tp.cleanCompletedEipTxPool(txs, height)
	txs = append(txs, cleanedEips...)
	for _, tx := range txs {
		if _, ok := tp.validTxMap[tx.Hash()]; ok {
			delete(tp.validTxMap, tx.Hash())
			cleaned++
			ShowTraceLog("transaction cleaned: %s", tx.Hash().ToHexString())
		}
	}

	log.Infof("clean txes: total %d, cleaned %d, remains %d in TxPool", txsNum, cleaned, len(tp.validTxMap))
}

func (tp *TXPool) selectSortEIP155WithLock(eiptxs []Transactions) []*VerifiedTx {
	idx := make([]int, len(eiptxs))
	total := 0
	for _, eips := range eiptxs {
		total += len(eips)
	}
	count := 0
	ret := make([]*VerifiedTx, 0, total)

	for count < total {
		roundMaxGasIdx := 0
		var roundMaxGas uint64

		for i, curIdx := range idx {
			if curIdx >= len(eiptxs[i]) {
				continue
			}

			if eiptxs[i][curIdx].GasPrice >= roundMaxGas {
				roundMaxGasIdx = i
				roundMaxGas = eiptxs[i][curIdx].GasPrice
			}
		}

		vtxn := tp.validTxMap[eiptxs[roundMaxGasIdx][idx[roundMaxGasIdx]].Hash()]
		if vtxn != nil {
			ret = append(ret, vtxn)
		} else {
			log.Errorf("eip tx %s not in tx list, impossible!", eiptxs[roundMaxGasIdx][idx[roundMaxGasIdx]].Hash().ToHexString())
		}

		idx[roundMaxGasIdx]++
		count++
	}

	return ret
}

// gets the transaction lists from the pool for the consensus,
// if the byCount is marked, return the configured number at most; if the
// the byCount is not marked, return all of the current transaction pool.
func (tp *TXPool) GetTxPool(byCount bool, height uint32) ([]*VerifiedTx, []*types.Transaction) {
	tp.RLock()

	eiplst := make([]Transactions, 0, len(tp.eipTxPool))
	for _, list := range tp.eipTxPool {
		// group by account
		curEipTxs := list.Heading()
		eiplst = append(eiplst, curEipTxs)
	}
	eipTxs := tp.selectSortEIP155WithLock(eiplst)

	orderByFeeList := make([]*VerifiedTx, 0, len(tp.validTxMap))
	for _, txEntry := range tp.validTxMap {
		if !txEntry.Tx.IsEipTx() {
			orderByFeeList = append(orderByFeeList, txEntry)
		}
	}

	tp.RUnlock()
	//this make EIP155 > other tx type
	//for EIP155 case:
	//if payer is same , order by nonce 0,1,2...
	//otherwise , order by gas price
	sort.Sort(OrderByNetWorkFee(orderByFeeList))
	orderByFeeList = append(eipTxs, orderByFeeList...)

	count := int(config.DefConfig.Consensus.MaxTxInBlock)
	if count <= 0 {
		byCount = false
	}
	if len(orderByFeeList) < count || !byCount {
		count = len(orderByFeeList)
	}

	validList := make([]*VerifiedTx, 0, count)
	oldTxList := make([]*types.Transaction, 0)
	for _, txEntry := range orderByFeeList {
		if txEntry.IsVerfiyExpired(height) {
			oldTxList = append(oldTxList, txEntry.Tx)
			continue
		}
		if len(validList) < count {
			validList = append(validList, txEntry)
		}
	}

	tp.Lock()
	for _, tx := range oldTxList {
		delete(tp.validTxMap, tx.Hash())
		if tx.IsEipTx() {
			removed := tp.eipTxPool[tx.Payer].Remove(uint64(tx.Nonce))
			if !removed {
				log.Errorf("transaction not in eip pool: %s, impossible", tx.Hash().ToHexString())
			}
		}

		ShowTraceLog("remove expired tx: %s from pool", tx.Hash().ToHexString())
	}
	tp.Unlock()

	return validList, oldTxList
}

// GetTransaction returns a transaction if it is contained in the pool
// and nil otherwise.
func (tp *TXPool) GetTransaction(hash common.Uint256) *types.Transaction {
	tp.RLock()
	defer tp.RUnlock()
	if tx := tp.validTxMap[hash]; tx == nil {
		return nil
	}
	return tp.validTxMap[hash].Tx
}

// GetTxStatus returns a transaction status if it is contained in the pool
// and nil otherwise.
func (tp *TXPool) GetTxStatus(hash common.Uint256) *TxStatus {
	tp.RLock()
	defer tp.RUnlock()
	txEntry, ok := tp.validTxMap[hash]
	if !ok {
		return nil
	}
	ret := &TxStatus{
		Hash:  hash,
		Attrs: txEntry.GetAttrs(),
	}
	return ret
}

// GetTransactionCount returns the tx number of the pool.
func (tp *TXPool) GetTransactionCount() int {
	tp.RLock()
	defer tp.RUnlock()
	return len(tp.validTxMap)
}

// GetTransactionCount returns the tx number of the pool.
func (tp *TXPool) GetTransactionHashList() []common.Uint256 {
	tp.RLock()
	defer tp.RUnlock()
	ret := make([]common.Uint256, 0, len(tp.validTxMap))
	for txHash := range tp.validTxMap {
		ret = append(ret, txHash)
	}
	return ret
}

// checks the tx list in the block from consensus,
// and returns verified tx list, unverified tx list, and
// the tx list to be re-verified
func (tp *TXPool) GetUnverifiedTxs(txs []*types.Transaction, height uint32) *CheckBlkResult {
	tp.Lock()
	defer tp.Unlock()
	res := &CheckBlkResult{
		VerifiedTxs:   make([]*VerifyTxResult, 0, len(txs)),
		UnverifiedTxs: make([]*types.Transaction, 0),
		OldTxs:        make([]*types.Transaction, 0),
	}
	for _, tx := range txs {
		txEntry := tp.validTxMap[tx.Hash()]
		if txEntry == nil {
			res.UnverifiedTxs = append(res.UnverifiedTxs, tx)
			continue
		}

		if !txEntry.IsVerfiyExpired(height) {
			// note: can not remove from tx pool since it is verified in another validator and will not be add back to pool
			res.OldTxs = append(res.OldTxs, txEntry.Tx)
			continue
		}

		res.VerifiedTxs = append(res.VerifiedTxs, &VerifyTxResult{
			Height:  txEntry.VerifiedHeight,
			Tx:      txEntry.Tx,
			ErrCode: errors.ErrNoError,
		})
	}

	return res
}

// RemoveTxsBelowGasPrice drops all transactions below the gas price
func (tp *TXPool) RemoveTxsBelowGasPrice(gasPrice uint64) {
	tp.Lock()
	defer tp.Unlock()
	for _, txEntry := range tp.validTxMap {
		tx := txEntry.Tx
		if tx.GasPrice < gasPrice {
			delete(tp.validTxMap, tx.Hash())
			if tx.IsEipTx() {
				tp.eipTxPool[tx.Payer].Remove(uint64(tx.Nonce))
			}
			ShowTraceLog("tx %s cleaned because of lower gas: %d, want: %d", tx.Hash().ToHexString(), txEntry.Tx.GasPrice, gasPrice)
		}
	}
}

// returns the remaining tx list to cleanup
func (tp *TXPool) Remain() []*types.Transaction {
	tp.Lock()
	defer tp.Unlock()

	tp.eipTxPool = make(map[common.Address]*txSortedMap) // clean all eip tx
	txList := make([]*types.Transaction, 0, len(tp.validTxMap))
	for _, txEntry := range tp.validTxMap {
		txList = append(txList, txEntry.Tx)
		delete(tp.validTxMap, txEntry.Tx.Hash())
		ShowTraceLog("pool remain: remove tx: %s from pool", txEntry.Tx.Hash().ToHexString())
	}

	return txList
}
