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
	"github.com/ontio/ontology/validator/types"
)

type txStats struct {
	sync.RWMutex
	count []uint64
}

type TxInfo struct {
	tx   *ctypes.Transaction // Pending tx
	from ttypes.SenderType    // Indicate which sender tx is from
}

type pendingBlock struct {
	mu             sync.RWMutex
	consusActor    *actor.PID                               // Consensus PID
	height         uint32                                   // The block height
	processedTxs   map[common.Uint256]*ttypes.VerifyTxResult // Transaction which has been processed
	unProcessedTxs map[common.Uint256]*ctypes.Transaction   // Transaction which is not processed
}

type roundRobinState struct {
	state map[types.VerifyType]int // Keep the round robin index for each verify type
}

type validators struct {
	sync.RWMutex
	entries map[types.VerifyType][]*types.RegisterValidatorReq // Registered validator container
	state   roundRobinState                                    // For loadbance
}

// TXPoolServer contains all api to external modules
type TXPoolServer struct {
	mu           sync.RWMutex                   // Sync mutex
	wg           sync.WaitGroup                 // Worker sync
	workers      []txPoolWorker                 // Worker pool
	txPool       *tcomn.TXPool                  // The tx pool that holds the valid transaction
	pendingTxs   map[common.Uint256]*TxInfo     // The txs that server is processing
	pendingBlock *pendingBlock                  // The block that server is processing
	actors       map[ttypes.ActorType]*actor.PID // The actors running in the server
	validators   *validators                    // The registered validators
	stats        txStats                        // The transaction statstics
	Slots        chan struct{}                  // The limited slots for the new transaction
}

// NewTxPoolServer creates a new tx pool server to schedule workers to
// handle and filter inbound transactions from the network, http, and consensus.
func NewTxPoolServer(num uint8) *TXPoolServer {
	s := &TXPoolServer{}
	s.init(num)
	return s
}

// init initializes the server with the configured settings
func (self *TXPoolServer) init(num uint8) {
	// Initial txnPool
	self.txPool = &tcomn.TXPool{}
	self.txPool.Init()
	self.pendingTxs = make(map[common.Uint256]*TxInfo)
	self.actors = make(map[ttypes.ActorType]*actor.PID)

	self.validators = &validators{
		entries: make(map[types.VerifyType][]*types.RegisterValidatorReq),
		state: roundRobinState{
			state: make(map[types.VerifyType]int),
		},
	}

	self.pendingBlock = &pendingBlock{
		processedTxs:   make(map[common.Uint256]*ttypes.VerifyTxResult, 0),
		unProcessedTxs: make(map[common.Uint256]*ctypes.Transaction, 0),
	}

	self.stats = txStats{count: make([]uint64, ttypes.MaxStats-1)}

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
func (self *TXPoolServer) updateTxInPendingBlock(hash common.Uint256,
	err errors.ErrCode) {

	// Check if the tx is in pending block, if yes, move it to
	// the verified tx list
	self.pendingBlock.mu.Lock()
	defer self.pendingBlock.mu.Unlock()

	tx, ok := self.pendingBlock.unProcessedTxs[hash]
	if !ok {
		return
	}

	entry := &ttypes.VerifyTxResult{
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
func (self *TXPoolServer) getPendingTxsSize() int {
	self.mu.Lock()
	defer self.mu.Unlock()
	return len(self.pendingTxs)
}

// removePendingTx removes a transaction from the pending list
// when it is handled. And if the submitter of the valid transaction
// is from http, broadcast it to the network. Meanwhile, check if it
// is in the block from consensus.
func (self *TXPoolServer) removePendingTx(hash common.Uint256,
	err errors.ErrCode) {

	self.mu.Lock()

	pt, ok := self.pendingTxs[hash]
	if !ok {
		self.mu.Unlock()
		return
	}

	if err == errors.ErrNoError && pt.from == ttypes.HttpSender {
		pid := self.GetPID(ttypes.NetActor)
		if pid != nil {
			pid.Tell(pt.tx)
		}
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
func (self *TXPoolServer) setPendingTx(tx *ctypes.Transaction,
	from ttypes.SenderType) bool {

	self.mu.Lock()
	defer self.mu.Unlock()
	if t := self.pendingTxs[tx.Hash()]; t != nil {
		log.Info("Transaction already in the verifying process",
			tx.Hash())
		return false
	}

	pt := &TxInfo{
		tx:   tx,
		from: from,
	}

	self.pendingTxs[tx.Hash()] = pt
	return true
}

// assignTxToWorker assigns a new transaction to a worker by LB
func (self *TXPoolServer) AssignTxToWorker(tx *ctypes.Transaction,
	sender ttypes.SenderType) bool {

	if tx == nil {
		return false
	}

	if ok := self.setPendingTx(tx, sender); !ok {
		self.IncreaseStats(ttypes.DuplicateStats)
		return false
	}
	// Add the rcvTxn to the worker
	lb := make(ttypes.LBSlice, len(self.workers))
	for i := 0; i < len(self.workers); i++ {
		entry := ttypes.LB{Size: len(self.workers[i].pendingTxs),
			WorkerID: uint8(i),
		}
		lb[i] = entry
	}
	sort.Sort(lb)
	self.workers[lb[0].WorkerID].rcvTxCh <- tx
	return true
}

// assignRspToWorker assigns a check response from the validator to
// the correct worker.
func (self *TXPoolServer) AssignVerifyRspToWorker(rsp *types.VerifyTxRsp) bool {

	if rsp == nil {
		return false
	}

	if rsp.WorkerId >= 0 && rsp.WorkerId < uint8(len(self.workers)) {
		self.workers[rsp.WorkerId].rspCh <- rsp
	}

	if rsp.ErrCode == errors.ErrNoError {
		self.IncreaseStats(ttypes.SuccessStats)
	} else {
		self.IncreaseStats(ttypes.FailureStats)
		if rsp.Type == types.Stateless {
			self.IncreaseStats(ttypes.SigErrStats)
		} else {
			self.IncreaseStats(ttypes.StateErrStats)
		}
	}
	return true
}

// GetPID returns an actor pid with the actor type, If the type
// doesn't exist, return nil.
func (self *TXPoolServer) GetPID(actor ttypes.ActorType) *actor.PID {
	if actor < ttypes.TxActor || actor >= ttypes.MaxActor {
		return nil
	}

	return self.actors[actor]
}

// RegisterActor registers an actor with the actor type and pid.
func (self *TXPoolServer) RegisterActor(actor ttypes.ActorType, pid *actor.PID) {
	self.actors[actor] = pid
}

// UnRegisterActor cancels the actor with the actor type.
func (self *TXPoolServer) UnRegisterActor(actor ttypes.ActorType) {
	delete(self.actors, actor)
}

// registerValidator registers a validator to verify a transaction.
func (self *TXPoolServer) RegisterValidator(v *types.RegisterValidatorReq) {
	self.validators.Lock()
	defer self.validators.Unlock()

	_, ok := self.validators.entries[v.Type]

	if !ok {
		self.validators.entries[v.Type] = make([]*types.RegisterValidatorReq, 0, 1)
	}
	self.validators.entries[v.Type] = append(self.validators.entries[v.Type], v)
}

// unRegisterValidator cancels a validator with the verify type and id.
func (self *TXPoolServer) UnRegisterValidator(verifyType types.VerifyType,
	id string) {

	self.validators.Lock()
	defer self.validators.Unlock()

	tmpSlice, ok := self.validators.entries[verifyType]
	if !ok {
		log.Error("No validator on check type:%d\n", verifyType)
		return
	}

	for i, v := range tmpSlice {
		if v.Id == id {
			self.validators.entries[verifyType] =
				append(tmpSlice[0:i], tmpSlice[i+1:]...)
			if v.Sender != nil {
				v.Sender.Tell(&types.UnRegisterValidatorRsp{Id: id, Type: verifyType})
			}
			if len(self.validators.entries[verifyType]) == 0 {
				delete(self.validators.entries, verifyType)
			}
		}
	}
}

// getNextValidatorPIDs returns the next pids to verify the transaction using
// roundRobin LB.
func (self *TXPoolServer) getNextValidatorPIDs() []*actor.PID {
	self.validators.Lock()
	defer self.validators.Unlock()

	if len(self.validators.entries) == 0 {
		return nil
	}

	ret := make([]*actor.PID, 0, len(self.validators.entries))
	for k, v := range self.validators.entries {
		lastIdx := self.validators.state.state[k]
		next := (lastIdx + 1) % len(v)
		self.validators.state.state[k] = next
		ret = append(ret, v[next].Sender)
	}
	return ret
}

// getNextValidatorPID returns the next pid with the verify type using roundRobin LB
func (self *TXPoolServer) getNextValidatorPID(key types.VerifyType) *actor.PID {
	self.validators.Lock()
	defer self.validators.Unlock()

	length := len(self.validators.entries[key])
	if length == 0 {
		return nil
	}

	entries := self.validators.entries[key]
	lastIdx := self.validators.state.state[key]
	next := (lastIdx + 1) % length
	self.validators.state.state[key] = next
	return entries[next].Sender
}

// Stop stops server and workers.
func (self *TXPoolServer) Stop() {
	for _, v := range self.actors {
		v.Stop()
	}
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
func (self *TXPoolServer) GetTransactionFromPool(hash common.Uint256) *ctypes.Transaction {
	return self.txPool.GetTransaction(hash)
}

// getTxPool returns a tx list for consensus.
func (self *TXPoolServer) GetTxEntrysFromPool(byCount bool, height uint32) []*ttypes.TxEntry {
	avlTxList, oldTxList := self.txPool.GetTxPool(byCount, height)

	for _, t := range oldTxList {
		self.deleteTransactionFromPool(t)
		self.reVerifyStateful(t, ttypes.NilSender)
	}

	return avlTxList
}

// getPendingTxs returns a currently pending tx list
func (self *TXPoolServer) GetPendingTxs(byCount bool) []*ctypes.Transaction {
	self.mu.RLock()
	defer self.mu.RUnlock()
	ret := make([]*ctypes.Transaction, 0, len(self.pendingTxs))
	for _, v := range self.pendingTxs {
		ret = append(ret, v.tx)
	}
	return ret
}

// cleanTransactionList cleans the txs in the block from the ledger
func (self *TXPoolServer) RemoveTransactionsFromPool(txs []*ctypes.Transaction) error {
	return self.txPool.RemoveTransactions(txs)
}

// delTransaction deletes a transaction in the tx pool.
func (self *TXPoolServer) deleteTransactionFromPool(t *ctypes.Transaction) {
	self.txPool.DeleteTransaction(t)
}

// addTxList adds a valid transaction to the tx pool.
func (self *TXPoolServer) appendTxEntry2Pool(txEntry *ttypes.TxEntry) bool {
	ret := self.txPool.AppendTxEntry(txEntry)
	if !ret {
		self.IncreaseStats(ttypes.DuplicateStats)
	}
	return ret
}

// increaseStats increases the count with the stats type
func (self *TXPoolServer) IncreaseStats(v ttypes.TxnStatsType) {
	self.stats.Lock()
	defer self.stats.Unlock()
	self.stats.count[v-1]++
}

// getStats returns the transaction statistics
func (self *TXPoolServer) GetStats() []uint64 {
	self.stats.RLock()
	defer self.stats.RUnlock()
	ret := make([]uint64, 0, len(self.stats.count))
	for _, v := range self.stats.count {
		ret = append(ret, v)
	}
	return ret
}

// checkTx checks whether a transaction is in the pending list or
// the transacton pool
func (self *TXPoolServer) IsContainTx(hash common.Uint256) bool {
	// Check if the tx is in pending list
	self.mu.RLock()
	if ok := self.pendingTxs[hash]; ok != nil {
		self.mu.RUnlock()
		return true
	}
	self.mu.RUnlock()

	// Check if the tx is in txn pool
	if res := self.txPool.GetTransaction(hash); res != nil {
		return true
	}

	return false
}

// getTxStatusReq returns a transaction's status with the transaction hash.
func (self *TXPoolServer) GetTxStatus(hash common.Uint256) *ttypes.TxStatus {
	for i := 0; i < len(self.workers); i++ {
		ret := self.workers[i].GetTxStatus(hash)
		if ret != nil {
			return ret
		}
	}

	return self.txPool.GetTxStatus(hash)
}

// getTransactionCount returns the tx size of the transaction pool.
func (self *TXPoolServer) GetTransactionCountFromPool() int {
	return self.txPool.GetTransactionCount()
}

// reVerifyStateful re-verify a transaction's stateful data.
func (self *TXPoolServer) reVerifyStateful(tx *ctypes.Transaction, sender ttypes.SenderType) {
	if ok := self.setPendingTx(tx, sender); !ok {
		self.IncreaseStats(ttypes.DuplicateStats)
		return
	}

	// Add the rcvTxn to the worker
	lb := make(ttypes.LBSlice, len(self.workers))
	for i := 0; i < len(self.workers); i++ {
		entry := ttypes.LB{Size: len(self.workers[i].pendingTxs),
			WorkerID: uint8(i),
		}
		lb[i] = entry
	}

	sort.Sort(lb)
	self.workers[lb[0].WorkerID].stfTxCh <- tx
}

// sendBlkResult2Consensus sends the result of verifying block to  consensus
func (self *TXPoolServer) sendVerifyBlkResult2Consensus() {
	rsp := &ttypes.VerifyBlockRsp{
		TxResults: make([]*ttypes.VerifyTxResult,
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
func (self *TXPoolServer) HandleVerifyBlockReq(req *ttypes.VerifyBlockReq, consusActor *actor.PID) {
	if req == nil || len(req.Txs) == 0 {
		return
	}

	self.pendingBlock.mu.Lock()
	defer self.pendingBlock.mu.Unlock()

	self.pendingBlock.consusActor = consusActor
	self.pendingBlock.height = req.Height
	self.pendingBlock.processedTxs = make(map[common.Uint256]*ttypes.VerifyTxResult, len(req.Txs))
	self.pendingBlock.unProcessedTxs = make(map[common.Uint256]*ctypes.Transaction, 0)

	checkBlkResult := self.txPool.GetVerifyBlkResult(req.Txs, req.Height)

	for _, t := range checkBlkResult.UnVerifiedTxs {
		self.AssignTxToWorker(t, ttypes.NilSender)
		self.pendingBlock.unProcessedTxs[t.Hash()] = t
	}

	for _, t := range checkBlkResult.ReVerifyTxs {
		self.reVerifyStateful(t, ttypes.NilSender)
		self.pendingBlock.unProcessedTxs[t.Hash()] = t
	}

	for _, t := range checkBlkResult.VerifiedTxs {
		self.pendingBlock.processedTxs[t.Tx.Hash()] = t
	}

	/* If all the txs in the blocks are verified, send response
	 * to the consensus directly
	 */
	if len(self.pendingBlock.unProcessedTxs) == 0 {
		self.sendVerifyBlkResult2Consensus()
	}
}
