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

// Package proc privides functions for handle messages from
// consensus/ledger/net/http/validators
package proc

import (
	"sort"
	"sync"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	ctypes "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	tcomn "github.com/ontio/ontology/txnpool/common"
	ttypes "github.com/ontio/ontology/txnpool/types"
	vtypes "github.com/ontio/ontology/validator/types"
)

type statistics struct {
	sync.RWMutex
	count []uint64
}

type TxInfo struct {
	tx   *ctypes.Transaction // Pending tx
	from ttypes.SenderType   // Indicate which sender tx is from
}

type pendingBlock struct {
	mu             sync.RWMutex
	consusActor    *actor.PID                             // Consensus PID
	height         uint32                                 // The block height
	processedTxs   map[common.Uint256]*ttypes.TxResult    // Transaction which has been processed
	unProcessedTxs map[common.Uint256]*ctypes.Transaction // Transaction which is not processed
}

// TXPoolServer contains all api to external modules
type TxPoolServer struct {
	mu           sync.RWMutex               // Sync mutex
	wg           sync.WaitGroup             // Worker sync
	workers      []txPoolWorker             // Worker pool
	txPool       *TxPool                    // The tx pool that holds the valid transaction
	pendingTxs   map[common.Uint256]*TxInfo // The txs that server is processing
	pendingBlock *pendingBlock              // The block that server is processing
	sender       *tcomn.Sender
	txStatistics statistics    // The transaction statstics
	Slots        chan struct{} // The limited slots for the new transaction
}

// NewTxPoolServer creates a new tx pool server to schedule workers to
// handle and filter inbound transactions from the network, http, and consensus.
func NewTxPoolServer(sender *tcomn.Sender, num uint8) *TxPoolServer {
	s := &TxPoolServer{sender: sender}
	s.init(num)
	return s
}

// init initializes the server with the configured settings
func (self *TxPoolServer) init(num uint8) {
	// Initial txnPool
	self.txPool = &TxPool{}
	self.txPool.Init()
	self.pendingTxs = make(map[common.Uint256]*TxInfo)
	//self.actors = make(map[ttypes.ActorType]*actor.PID)

	self.pendingBlock = &pendingBlock{
		processedTxs:   make(map[common.Uint256]*ttypes.TxResult, 0),
		unProcessedTxs: make(map[common.Uint256]*ctypes.Transaction, 0),
	}

	self.txStatistics = statistics{count: make([]uint64, ttypes.MaxStats-1)}

	self.Slots = make(chan struct{}, ttypes.MAX_LIMITATION)
	for i := 0; i < ttypes.MAX_LIMITATION; i++ {
		self.Slots <- struct{}{}
	}

	// Create the given concurrent workers
	self.workers = make([]txPoolWorker, num)
	// Initial and start the workers
	var i uint8
	for i = 0; i < num; i++ {
		self.wg.Add(1)
		self.workers[i].init(i, self)
		go self.workers[i].start()
	}
}

// checkPendingBlockOk checks whether a block from consensus is verified.
// If some transaction is invalid, return the result directly at once, no
// need to wait for verifying the complete block.
func (self *TxPoolServer) updateTxInPendingBlock(hash common.Uint256,
	err errors.ErrCode) {

	// Check if the tx is in pending block, if yes, move it to
	// the verified tx list
	self.pendingBlock.mu.Lock()
	defer self.pendingBlock.mu.Unlock()

	tx, ok := self.pendingBlock.unProcessedTxs[hash]
	if !ok {
		return
	}

	entry := &ttypes.TxResult{
		Height:  self.pendingBlock.height,
		Tx:      tx,
		ErrCode: err,
	}

	self.pendingBlock.processedTxs[hash] = entry
	delete(self.pendingBlock.unProcessedTxs, hash)

	// if the tx is invalid, send the response at once
	if err != errors.ErrNoError ||
		len(self.pendingBlock.unProcessedTxs) == 0 {
		self.sendVerifyBlkResult2Consensus()
	}
}

// getPendingListSize return the length of the pending tx list.
func (self *TxPoolServer) getPendingTxSize() int {
	self.mu.Lock()
	defer self.mu.Unlock()
	return len(self.pendingTxs)
}

// removePendingTx removes a transaction from the pending list
// when it is handled. And if the submitter of the valid transaction
// is from http, broadcast it to the network. Meanwhile, check if it
// is in the block from consensus.
func (self *TxPoolServer) removePendingTx(hash common.Uint256,
	err errors.ErrCode) {

	self.mu.Lock()

	pt, ok := self.pendingTxs[hash]
	if !ok {
		self.mu.Unlock()
		return
	}

	if err == errors.ErrNoError && pt.from == ttypes.HttpSender {
		self.sender.SendTxToNetActor(pt.tx)
	}

	delete(self.pendingTxs, hash)

	if len(self.pendingTxs) < ttypes.MAX_LIMITATION {
		select {
		case self.Slots <- struct{}{}:
		default:
			log.Debug("slots is full")
		}
	}

	self.mu.Unlock()

	// Check if the tx is in the pending block and
	// the pending block is verified
	self.updateTxInPendingBlock(hash, err)
}

// setPendingTx adds a transaction to the pending list, if the
// transaction is already in the pending list, just return false.
func (self *TxPoolServer) putPendingTx(tx *ctypes.Transaction,
	from ttypes.SenderType) bool {

	self.mu.Lock()
	defer self.mu.Unlock()
	if t := self.pendingTxs[tx.Hash()]; t != nil {
		log.Info("Transaction already in the verifying process",
			tx.Hash())
		return false
	}

	ptx := &TxInfo{
		tx:   tx,
		from: from,
	}

	self.pendingTxs[tx.Hash()] = ptx
	return true
}
func (self *TxPoolServer) getWorkId() uint8 {
	// Add the rcvTxn to the worker
	lb := make(tcomn.LoadBalances, len(self.workers))
	for i := 0; i < len(self.workers); i++ {
		entry := tcomn.LoadBalance{Size: len(self.workers[i].pendingTxs),
			WorkerID: uint8(i),
		}
		lb[i] = entry
	}
	sort.Sort(lb)
	return lb[0].WorkerID
}
func (self *TxPoolServer) PutTransaction(tx *ctypes.Transaction,
	from ttypes.SenderType) bool {
	return self.assignTxToWorker(tx, from)
}

// assignTxToWorker assigns a new transaction to a worker by LB
func (self *TxPoolServer) assignTxToWorker(tx *ctypes.Transaction,
	from ttypes.SenderType) bool {

	if tx == nil {
		return false
	}

	if ok := self.putPendingTx(tx, from); !ok {
		self.Increase(ttypes.Duplicate)
		return false
	}
	id := self.getWorkId()
	self.workers[id].statelessTxCh <- tx
	return true
}

// assignRspToWorker assigns a check response from the validator to
// the correct worker.
func (self *TxPoolServer) PutVerifyTxRsp(rsp *vtypes.VerifyTxRsp) bool {
	if rsp == nil {
		return false
	}

	if rsp.WorkerId >= 0 && rsp.WorkerId < uint8(len(self.workers)) {
		self.workers[rsp.WorkerId].verifyTxRspCh <- rsp
	}

	if rsp.ErrCode == errors.ErrNoError {
		self.Increase(ttypes.Success)
	} else {
		self.Increase(ttypes.Failure)
		if rsp.VerifyType == vtypes.Stateless {
			self.Increase(ttypes.SigErr)
		} else {
			self.Increase(ttypes.StateErr)
		}
	}
	return true
}

// putTxPool adds a valid transaction to the tx pool and removes it from
// the pending list.
func (self *TxPoolServer) moveTx2Pool(pt *pendingTxInfo) bool {
	txEntry := &ttypes.TxEntry{
		Tx:            pt.tx,
		VerifyResults: pt.verifyResults,
		Fee:           pt.tx.GetTotalFee(),
	}
	self.appendTxEntry2Pool(txEntry)
	self.removePendingTx(pt.tx.Hash(), errors.ErrNoError)
	return true
}

// Stop stops server and workers.
func (self *TxPoolServer) Stop() {
	self.sender.Stop()
	//Stop worker
	for i := 0; i < len(self.workers); i++ {
		self.workers[i].stop()
	}
	self.wg.Wait()

	if self.Slots != nil {
		close(self.Slots)
	}
}

// getTransaction returns a transaction with the transaction hash.
func (self *TxPoolServer) GetTxFromPool(hash common.Uint256) *ctypes.Transaction {
	return self.txPool.getTransaction(hash)
}

// getTxPool returns a tx list for consensus.
func (self *TxPoolServer) GetTxEntrysFromPool(byCount bool, height uint32) []*ttypes.TxEntry {
	avlTxList, oldTxList := self.txPool.getTransactions(byCount, height)

	for _, t := range oldTxList {
		self.deleteTransactionFromPool(t)
		self.reVerifyStatefulTx(t, ttypes.NilSender)
	}

	return avlTxList
}

// getPendingTxs returns a currently pending tx list
func (self *TxPoolServer) GetPendingTxs(byCount bool) []*ctypes.Transaction {
	self.mu.RLock()
	defer self.mu.RUnlock()
	ret := make([]*ctypes.Transaction, 0, len(self.pendingTxs))
	for _, v := range self.pendingTxs {
		ret = append(ret, v.tx)
	}
	return ret
}

// cleanTransactionList cleans the txs in the block from the ledger
func (self *TxPoolServer) RemoveTransactionsFromPool(txs []*ctypes.Transaction) error {
	return self.txPool.removeTransactions(txs)
}

// delTransaction deletes a transaction in the tx pool.
func (self *TxPoolServer) deleteTransactionFromPool(t *ctypes.Transaction) {
	self.txPool.deleteTransaction(t)
}

// addTxList adds a valid transaction to the tx pool.
func (self *TxPoolServer) appendTxEntry2Pool(txEntry *ttypes.TxEntry) bool {
	ret := self.txPool.appendTxEntry(txEntry)
	if !ret {
		self.Increase(ttypes.Duplicate)
	}
	return ret
}

// increaseStats increases the count with the stats type
func (self *TxPoolServer) Increase(v ttypes.VerifyResultType) {
	self.txStatistics.Lock()
	defer self.txStatistics.Unlock()
	self.txStatistics.count[v-1]++
}

// getStats returns the transaction statistics
func (self *TxPoolServer) GetVerifyResultStatistics() []uint64 {
	self.txStatistics.RLock()
	defer self.txStatistics.RUnlock()
	ret := make([]uint64, 0, len(self.txStatistics.count))
	for _, v := range self.txStatistics.count {
		ret = append(ret, v)
	}
	return ret
}

// checkTx checks whether a transaction is in the pending list or
// the transacton pool
func (self *TxPoolServer) IsContainTx(hash common.Uint256) bool {
	// Check if the tx is in pending list
	self.mu.RLock()
	if ok := self.pendingTxs[hash]; ok != nil {
		self.mu.RUnlock()
		return true
	}
	self.mu.RUnlock()

	// Check if the tx is in txn pool
	if res := self.txPool.getTransaction(hash); res != nil {
		return true
	}

	return false
}

// getTxStatusReq returns a transaction's status with the transaction hash.
func (self *TxPoolServer) GetTxVerifyStatus(hash common.Uint256) *ttypes.TxVerifyStatus {
	for i := 0; i < len(self.workers); i++ {
		ret := self.workers[i].getTxVerifyStatus(hash)
		if ret != nil {
			return ret
		}
	}

	return self.txPool.getTxVerifyStatus(hash)
}

// getTransactionCount returns the tx size of the transaction pool.
func (self *TxPoolServer) GetTxCountFromPool() int {
	return self.txPool.getTransactionCount()
}

// reVerifyStateful re-verify a transaction's stateful data.
func (self *TxPoolServer) reVerifyStatefulTx(tx *ctypes.Transaction, sender ttypes.SenderType) {
	if ok := self.putPendingTx(tx, sender); !ok {
		self.Increase(ttypes.Duplicate)
		return
	}
	id := self.getWorkId()
	self.workers[id].statefulTxCh <- tx
}

// sendBlkResult2Consensus sends the result of verifying block to  consensus
func (self *TxPoolServer) sendVerifyBlkResult2Consensus() {
	rsp := &ttypes.VerifyBlockRsp{
		TxResults: make([]*ttypes.TxResult,
			0, len(self.pendingBlock.processedTxs)),
	}
	for _, v := range self.pendingBlock.processedTxs {
		rsp.TxResults = append(rsp.TxResults, v)
	}

	if self.pendingBlock.consusActor != nil {
		self.pendingBlock.consusActor.Tell(rsp)
	}

	// Clear the processedTxs for the next block verify req
	for k := range self.pendingBlock.processedTxs {
		delete(self.pendingBlock.processedTxs, k)
	}
}

// verifyBlock verifies the block from consensus.
// There are three cases to handle.
// 1, for those unverified txs, assign them to the available worker;
// 2, for those verified txs whose height >= block's height, nothing to do;
// 3, for those verified txs whose height < block's height, re-verify their
// stateful data.
func (self *TxPoolServer) AddVerifyBlock(height uint32, txs []*ctypes.Transaction, consusActor *actor.PID) {
	if len(txs) == 0 {
		return
	}

	self.pendingBlock.mu.Lock()
	defer self.pendingBlock.mu.Unlock()

	self.pendingBlock.consusActor = consusActor
	self.pendingBlock.height = height
	self.pendingBlock.processedTxs = make(map[common.Uint256]*ttypes.TxResult, len(txs))
	self.pendingBlock.unProcessedTxs = make(map[common.Uint256]*ctypes.Transaction, 0)

	blkResult := self.txPool.getVerifyBlockResult(txs, height)

	for _, t := range blkResult.UnVerifiedTxs {
		self.assignTxToWorker(t, ttypes.NilSender)
		self.pendingBlock.unProcessedTxs[t.Hash()] = t
	}

	for _, t := range blkResult.ReVerifyTxs {
		self.reVerifyStatefulTx(t, ttypes.NilSender)
		self.pendingBlock.unProcessedTxs[t.Hash()] = t
	}

	for _, t := range blkResult.VerifiedTxs {
		self.pendingBlock.processedTxs[t.Tx.Hash()] = t
	}

	/* If all the txs in the blocks are verified, send response
	 * to the consensus directly
	 */
	if len(self.pendingBlock.unProcessedTxs) == 0 {
		self.sendVerifyBlkResult2Consensus()
	}
}
