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
	tx "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	tc "github.com/ontio/ontology/txnpool/common"
	"github.com/ontio/ontology/validator/types"
)

type txStats struct {
	sync.RWMutex
	count []uint64
}

type serverPendingTx struct {
	tx     *tx.Transaction // Pending tx
	sender tc.SenderType   // Indicate which sender tx is from
}

type pendingBlock struct {
	mu             sync.RWMutex
	sender         *actor.PID                            // Consensus PID
	height         uint32                                // The block height
	processedTxs   map[common.Uint256]*tc.VerifyTxResult // Transaction which has been processed
	unProcessedTxs map[common.Uint256]*tx.Transaction    // Transaction which is not processed
}

type roundRobinState struct {
	state map[types.VerifyType]int // Keep the round robin index for each verify type
}

type registerValidators struct {
	sync.RWMutex
	entries map[types.VerifyType][]*types.RegisterValidator // Registered validator container
	state   roundRobinState                                 // For loadbance
}

// TXPoolServer contains all api to external modules
type TXPoolServer struct {
	mu            sync.RWMutex                        // Sync mutex
	wg            sync.WaitGroup                      // Worker sync
	workers       []txPoolWorker                      // Worker pool
	txPool        *tc.TXPool                          // The tx pool that holds the valid transaction
	allPendingTxs map[common.Uint256]*serverPendingTx // The txs that server is processing
	pendingBlock  *pendingBlock                       // The block that server is processing
	actors        map[tc.ActorType]*actor.PID         // The actors running in the server
	validators    *registerValidators                 // The registered validators
	stats         txStats                             // The transaction statstics
	slots         chan struct{}                       // The limited slots for the new transaction
	height        uint32                              // The current block height
}

// NewTxPoolServer creates a new tx pool server to schedule workers to
// handle and filter inbound transactions from the network, http, and consensus.
func NewTxPoolServer(num uint8) *TXPoolServer {
	s := &TXPoolServer{}
	s.init(num)
	return s
}

// init initializes the server with the configured settings
func (s *TXPoolServer) init(num uint8) {
	// Initial txnPool
	s.txPool = &tc.TXPool{}
	s.txPool.Init()
	s.allPendingTxs = make(map[common.Uint256]*serverPendingTx)
	s.actors = make(map[tc.ActorType]*actor.PID)

	s.validators = &registerValidators{
		entries: make(map[types.VerifyType][]*types.RegisterValidator),
		state: roundRobinState{
			state: make(map[types.VerifyType]int),
		},
	}

	s.pendingBlock = &pendingBlock{
		processedTxs:   make(map[common.Uint256]*tc.VerifyTxResult, 0),
		unProcessedTxs: make(map[common.Uint256]*tx.Transaction, 0),
	}

	s.stats = txStats{count: make([]uint64, tc.MaxStats-1)}

	s.slots = make(chan struct{}, tc.MAX_LIMITATION)
	for i := 0; i < tc.MAX_LIMITATION; i++ {
		s.slots <- struct{}{}
	}

	// Create the given concurrent workers
	s.workers = make([]txPoolWorker, num)
	// Initial and start the workers
	var i uint8
	for i = 0; i < num; i++ {
		s.wg.Add(1)
		s.workers[i].init(i, s)
		go s.workers[i].start()
	}
}

// checkPendingBlockOk checks whether a block from consensus is verified.
// If some transaction is invalid, return the result directly at once, no
// need to wait for verifying the complete block.
func (s *TXPoolServer) checkPendingBlockOk(hash common.Uint256,
	err errors.ErrCode) {

	// Check if the tx is in pending block, if yes, move it to
	// the verified tx list
	s.pendingBlock.mu.Lock()
	defer s.pendingBlock.mu.Unlock()

	tx, ok := s.pendingBlock.unProcessedTxs[hash]
	if !ok {
		return
	}

	entry := &tc.VerifyTxResult{
		Height:  s.pendingBlock.height,
		Tx:      tx,
		ErrCode: err,
	}

	s.pendingBlock.processedTxs[hash] = entry
	delete(s.pendingBlock.unProcessedTxs, hash)

	// if the tx is invalid, send the response at once
	if err != errors.ErrNoError ||
		len(s.pendingBlock.unProcessedTxs) == 0 {
		s.sendBlkResult2Consensus()
	}
}

// getPendingListSize return the length of the pending tx list.
func (s *TXPoolServer) getPendingListSize() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.allPendingTxs)
}

// getHeight return current block height
func (s *TXPoolServer) getHeight() uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.height
}

// setHeight set current block height
func (s *TXPoolServer) setHeight(height uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.height = height
}

// removePendingTx removes a transaction from the pending list
// when it is handled. And if the submitter of the valid transaction
// is from http, broadcast it to the network. Meanwhile, check if it
// is in the block from consensus.
func (s *TXPoolServer) removePendingTx(hash common.Uint256,
	err errors.ErrCode) {

	s.mu.Lock()

	pt, ok := s.allPendingTxs[hash]
	if !ok {
		s.mu.Unlock()
		return
	}

	if err == errors.ErrNoError && pt.sender == tc.HttpSender {
		pid := s.GetPID(tc.NetActor)
		if pid != nil {
			pid.Tell(pt.tx)
		}
	}

	delete(s.allPendingTxs, hash)

	if len(s.allPendingTxs) < tc.MAX_LIMITATION {
		select {
		case s.slots <- struct{}{}:
		default:
			log.Debug("removePendingTx: slots is full")
		}
	}

	s.mu.Unlock()

	// Check if the tx is in the pending block and
	// the pending block is verified
	s.checkPendingBlockOk(hash, err)
}

// setPendingTx adds a transaction to the pending list, if the
// transaction is already in the pending list, just return false.
func (s *TXPoolServer) setPendingTx(tx *tx.Transaction,
	sender tc.SenderType) bool {

	s.mu.Lock()
	defer s.mu.Unlock()
	if ok := s.allPendingTxs[tx.Hash()]; ok != nil {
		log.Infof("setPendingTx: transaction %x already in the verifying process",
			tx.Hash())
		return false
	}

	pt := &serverPendingTx{
		tx:     tx,
		sender: sender,
	}

	s.allPendingTxs[tx.Hash()] = pt
	return true
}

// assignTxToWorker assigns a new transaction to a worker by LB
func (s *TXPoolServer) assignTxToWorker(tx *tx.Transaction,
	sender tc.SenderType) bool {

	if tx == nil {
		return false
	}

	if ok := s.setPendingTx(tx, sender); !ok {
		s.increaseStats(tc.DuplicateStats)
		return false
	}
	// Add the rcvTxn to the worker
	lb := make(tc.LBSlice, len(s.workers))
	for i := 0; i < len(s.workers); i++ {
		entry := tc.LB{Size: len(s.workers[i].rcvTXCh) +
			len(s.workers[i].pendingTxList),
			WorkerID: uint8(i),
		}
		lb[i] = entry
	}
	sort.Sort(lb)
	s.workers[lb[0].WorkerID].rcvTXCh <- tx
	return true
}

// assignRspToWorker assigns a check response from the validator to
// the correct worker.
func (s *TXPoolServer) assignRspToWorker(rsp *types.CheckResponse) bool {

	if rsp == nil {
		return false
	}

	if rsp.WorkerId >= 0 && rsp.WorkerId < uint8(len(s.workers)) {
		s.workers[rsp.WorkerId].rspCh <- rsp
	}

	if rsp.ErrCode == errors.ErrNoError {
		s.increaseStats(tc.SuccessStats)
	} else {
		s.increaseStats(tc.FailureStats)
		if rsp.Type == types.Stateless {
			s.increaseStats(tc.SigErrStats)
		} else {
			s.increaseStats(tc.StateErrStats)
		}
	}
	return true
}

// GetPID returns an actor pid with the actor type, If the type
// doesn't exist, return nil.
func (s *TXPoolServer) GetPID(actor tc.ActorType) *actor.PID {
	if actor < tc.TxActor || actor >= tc.MaxActor {
		return nil
	}

	return s.actors[actor]
}

// RegisterActor registers an actor with the actor type and pid.
func (s *TXPoolServer) RegisterActor(actor tc.ActorType, pid *actor.PID) {
	s.actors[actor] = pid
}

// UnRegisterActor cancels the actor with the actor type.
func (s *TXPoolServer) UnRegisterActor(actor tc.ActorType) {
	delete(s.actors, actor)
}

// registerValidator registers a validator to verify a transaction.
func (s *TXPoolServer) registerValidator(v *types.RegisterValidator) {
	s.validators.Lock()
	defer s.validators.Unlock()

	_, ok := s.validators.entries[v.Type]

	if !ok {
		s.validators.entries[v.Type] = make([]*types.RegisterValidator, 0, 1)
	}
	s.validators.entries[v.Type] = append(s.validators.entries[v.Type], v)
}

// unRegisterValidator cancels a validator with the verify type and id.
func (s *TXPoolServer) unRegisterValidator(checkType types.VerifyType,
	id string) {

	s.validators.Lock()
	defer s.validators.Unlock()

	tmpSlice, ok := s.validators.entries[checkType]
	if !ok {
		log.Errorf("unRegisterValidator: validator not found with type:%d, id:%s",
			checkType, id)
		return
	}

	for i, v := range tmpSlice {
		if v.Id == id {
			s.validators.entries[checkType] =
				append(tmpSlice[0:i], tmpSlice[i+1:]...)
			if v.Sender != nil {
				v.Sender.Tell(&types.UnRegisterAck{Id: id, Type: checkType})
			}
			if len(s.validators.entries[checkType]) == 0 {
				delete(s.validators.entries, checkType)
			}
		}
	}
}

// getNextValidatorPIDs returns the next pids to verify the transaction using
// roundRobin LB.
func (s *TXPoolServer) getNextValidatorPIDs() []*actor.PID {
	s.validators.Lock()
	defer s.validators.Unlock()

	if len(s.validators.entries) == 0 {
		return nil
	}

	ret := make([]*actor.PID, 0, len(s.validators.entries))
	for k, v := range s.validators.entries {
		lastIdx := s.validators.state.state[k]
		next := (lastIdx + 1) % len(v)
		s.validators.state.state[k] = next
		ret = append(ret, v[next].Sender)
	}
	return ret
}

// getNextValidatorPID returns the next pid with the verify type using roundRobin LB
func (s *TXPoolServer) getNextValidatorPID(key types.VerifyType) *actor.PID {
	s.validators.Lock()
	defer s.validators.Unlock()

	length := len(s.validators.entries[key])
	if length == 0 {
		return nil
	}

	entries := s.validators.entries[key]
	lastIdx := s.validators.state.state[key]
	next := (lastIdx + 1) % length
	s.validators.state.state[key] = next
	return entries[next].Sender
}

// Stop stops server and workers.
func (s *TXPoolServer) Stop() {
	for _, v := range s.actors {
		v.Stop()
	}
	//Stop worker
	for i := 0; i < len(s.workers); i++ {
		s.workers[i].stop()
	}
	s.wg.Wait()

	if s.slots != nil {
		close(s.slots)
	}
}

// getTransaction returns a transaction with the transaction hash.
func (s *TXPoolServer) getTransaction(hash common.Uint256) *tx.Transaction {
	return s.txPool.GetTransaction(hash)
}

// getTxPool returns a tx list for consensus.
func (s *TXPoolServer) getTxPool(byCount bool, height uint32) []*tc.TXEntry {
	s.setHeight(height)

	avlTxList, oldTxList := s.txPool.GetTxPool(byCount, height)

	for _, t := range oldTxList {
		s.delTransaction(t)
		s.reVerifyStateful(t, tc.NilSender)
	}

	return avlTxList
}

// getPendingTxs returns a currently pending tx list
func (s *TXPoolServer) getPendingTxs(byCount bool) []*tx.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ret := make([]*tx.Transaction, 0, len(s.allPendingTxs))
	for _, v := range s.allPendingTxs {
		ret = append(ret, v.tx)
	}
	return ret
}

// cleanTransactionList cleans the txs in the block from the ledger
func (s *TXPoolServer) cleanTransactionList(txs []*tx.Transaction) error {
	return s.txPool.CleanTransactionList(txs)
}

// delTransaction deletes a transaction in the tx pool.
func (s *TXPoolServer) delTransaction(t *tx.Transaction) {
	s.txPool.DelTxList(t)
}

// addTxList adds a valid transaction to the tx pool.
func (s *TXPoolServer) addTxList(txEntry *tc.TXEntry) bool {
	ret := s.txPool.AddTxList(txEntry)
	if !ret {
		s.increaseStats(tc.DuplicateStats)
	}
	return ret
}

// increaseStats increases the count with the stats type
func (s *TXPoolServer) increaseStats(v tc.TxnStatsType) {
	s.stats.Lock()
	defer s.stats.Unlock()
	s.stats.count[v-1]++
}

// getStats returns the transaction statistics
func (s *TXPoolServer) getStats() []uint64 {
	s.stats.RLock()
	defer s.stats.RUnlock()
	ret := make([]uint64, 0, len(s.stats.count))
	for _, v := range s.stats.count {
		ret = append(ret, v)
	}
	return ret
}

// checkTx checks whether a transaction is in the pending list or
// the transacton pool
func (s *TXPoolServer) checkTx(hash common.Uint256) bool {
	// Check if the tx is in pending list
	s.mu.RLock()
	if ok := s.allPendingTxs[hash]; ok != nil {
		s.mu.RUnlock()
		return true
	}
	s.mu.RUnlock()

	// Check if the tx is in txn pool
	if res := s.txPool.GetTransaction(hash); res != nil {
		return true
	}

	return false
}

// getTxStatusReq returns a transaction's status with the transaction hash.
func (s *TXPoolServer) getTxStatusReq(hash common.Uint256) *tc.TxStatus {
	for i := 0; i < len(s.workers); i++ {
		ret := s.workers[i].GetTxStatus(hash)
		if ret != nil {
			return ret
		}
	}

	return s.txPool.GetTxStatus(hash)
}

// getTransactionCount returns the tx size of the transaction pool.
func (s *TXPoolServer) getTransactionCount() int {
	return s.txPool.GetTransactionCount()
}

// reVerifyStateful re-verify a transaction's stateful data.
func (s *TXPoolServer) reVerifyStateful(tx *tx.Transaction, sender tc.SenderType) {
	if ok := s.setPendingTx(tx, sender); !ok {
		s.increaseStats(tc.DuplicateStats)
		return
	}

	// Add the rcvTxn to the worker
	lb := make(tc.LBSlice, len(s.workers))
	for i := 0; i < len(s.workers); i++ {
		entry := tc.LB{Size: len(s.workers[i].stfTxCh) +
			len(s.workers[i].pendingTxList),
			WorkerID: uint8(i),
		}
		lb[i] = entry
	}

	sort.Sort(lb)
	s.workers[lb[0].WorkerID].stfTxCh <- tx
}

// sendBlkResult2Consensus sends the result of verifying block to  consensus
func (s *TXPoolServer) sendBlkResult2Consensus() {
	rsp := &tc.VerifyBlockRsp{
		TxnPool: make([]*tc.VerifyTxResult,
			0, len(s.pendingBlock.processedTxs)),
	}
	for _, v := range s.pendingBlock.processedTxs {
		rsp.TxnPool = append(rsp.TxnPool, v)
	}

	if s.pendingBlock.sender != nil {
		s.pendingBlock.sender.Tell(rsp)
	}

	// Clear the processedTxs for the next block verify req
	for k := range s.pendingBlock.processedTxs {
		delete(s.pendingBlock.processedTxs, k)
	}
}

// verifyBlock verifies the block from consensus.
// There are three cases to handle.
// 1, for those unverified txs, assign them to the available worker;
// 2, for those verified txs whose height >= block's height, nothing to do;
// 3, for those verified txs whose height < block's height, re-verify their
// stateful data.
func (s *TXPoolServer) verifyBlock(req *tc.VerifyBlockReq, sender *actor.PID) {
	if req == nil || len(req.Txs) == 0 {
		return
	}

	s.setHeight(req.Height)
	s.pendingBlock.mu.Lock()
	defer s.pendingBlock.mu.Unlock()

	s.pendingBlock.sender = sender
	s.pendingBlock.height = req.Height
	s.pendingBlock.processedTxs = make(map[common.Uint256]*tc.VerifyTxResult, len(req.Txs))
	s.pendingBlock.unProcessedTxs = make(map[common.Uint256]*tx.Transaction, 0)

	checkBlkResult := s.txPool.GetUnverifiedTxs(req.Txs, req.Height)

	for _, t := range checkBlkResult.UnverifiedTxs {
		s.assignTxToWorker(t, tc.NilSender)
		s.pendingBlock.unProcessedTxs[t.Hash()] = t
	}

	for _, t := range checkBlkResult.OldTxs {
		s.reVerifyStateful(t, tc.NilSender)
		s.pendingBlock.unProcessedTxs[t.Hash()] = t
	}

	for _, t := range checkBlkResult.VerifiedTxs {
		s.pendingBlock.processedTxs[t.Tx.Hash()] = t
	}

	/* If all the txs in the blocks are verified, send response
	 * to the consensus directly
	 */
	if len(s.pendingBlock.unProcessedTxs) == 0 {
		s.sendBlkResult2Consensus()
	}
}
