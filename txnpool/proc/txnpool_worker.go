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

package proc

import (
	"fmt"
	"sync"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	tx "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	vtypes "github.com/ontio/ontology/validator/types"
	ttypes "github.com/ontio/ontology/txnpool/types"
)

// pendingTx contains the transaction, the time of starting verifying,
// the cache of check request, the flag indicating the verified status,
// the verified result and retry mechanism
type pendingTxInfo struct {
	tx            *tx.Transaction        // That is unverified or on the verifying process
	verifyTime    time.Time              // The start time
	req           *vtypes.VerifyTxReq    // Req cache
	flag          uint8                  // For different types of verification
	retries       uint8                  // For resend to validator when time out before verified
	verifyResults []*ttypes.VerifyResult // verified results
}

// txPoolWorker handles the tasks scheduled by server
type txPoolWorker struct {
	mu            sync.RWMutex
	workId        uint8                             // Worker ID
	statelessTxCh chan *tx.Transaction              // The channel of receive transaction
	statefulTxCh  chan *tx.Transaction              // The channel of txs to be re-verified stateful
	verifyTxRspCh chan *vtypes.VerifyTxRsp          // The channel of verified response
	txPoolServer  *TxPoolServer                     // The txn pool server pointer
	timer         *time.Timer                       // The timer of reverifying
	stopCh        chan bool                         // stop routine
	pendingTxs    map[common.Uint256]*pendingTxInfo // The transaction on the verifying process
}

// init initializes the worker with the configured settings
func (self *txPoolWorker) init(workID uint8, s *TxPoolServer) {
	self.statelessTxCh = make(chan *tx.Transaction, ttypes.MAX_PENDING_TXN)
	self.statefulTxCh = make(chan *tx.Transaction, ttypes.MAX_PENDING_TXN)
	self.pendingTxs = make(map[common.Uint256]*pendingTxInfo)
	self.verifyTxRspCh = make(chan *vtypes.VerifyTxRsp, ttypes.MAX_PENDING_TXN)
	self.stopCh = make(chan bool)
	self.workId = workID
	self.txPoolServer = s
}

// GetTxStatus returns the status in the pending list with the transaction hash
func (self *txPoolWorker) getTxVerifyStatus(hash common.Uint256) *ttypes.TxVerifyStatus {
	self.mu.RLock()
	defer self.mu.RUnlock()

	ptx, ok := self.pendingTxs[hash]
	if !ok {
		return nil
	}
	return &ttypes.TxVerifyStatus{hash, ptx.verifyResults}
}

// handleRsp handles the verified response from the validator and if
// the tx is valid, add it to the tx pool, or remove it from the pending
// list
func (self *txPoolWorker) handleValidatorRsp(rsp *vtypes.VerifyTxRsp) {
	if rsp.WorkerId != self.workId {
		return
	}

	self.mu.Lock()
	defer self.mu.Unlock()

	ptx, ok := self.pendingTxs[rsp.Hash]
	if !ok {
		return
	}
	if rsp.ErrCode != errors.ErrNoError {
		//Verify fail
		log.Info(fmt.Sprintf("Validator %d: Transaction %x invalid: %s",
			rsp.VerifyType, rsp.Hash, rsp.ErrCode.Error()))
		delete(self.pendingTxs, rsp.Hash)
		self.txPoolServer.removePendingTx(rsp.Hash, rsp.ErrCode)
		return
	}

	if ptx.flag&(0x1<<rsp.VerifyType) == 0 {
		ptx.flag |= (0x1 << rsp.VerifyType)
		ptx.verifyResults = append(ptx.verifyResults, &ttypes.VerifyResult{rsp.Height, rsp.VerifyType, rsp.ErrCode})
	}

	if ptx.flag&0xf == ttypes.VERIFY_MASK {
		self.txPoolServer.moveTx2Pool(ptx)
		delete(self.pendingTxs, rsp.Hash)
	}
}

/* Check if the transaction need to be sent to validator to verify
 * when time out.
 * Todo: Going through the list will take time if the list is too
 * long, need to change the algorithm later
 */
func (self *txPoolWorker) handleTimeoutEvent() {
	if len(self.pendingTxs) <= 0 {
		return
	}

	/* Go through the pending list, for those unverified txns,
	 * resend them to the validators
	 */
	for k, v := range self.pendingTxs {
		if v.flag&0xf != ttypes.VERIFY_MASK && (time.Now().Sub(v.verifyTime)/time.Second) >=
			ttypes.EXPIRE_INTERVAL {
			if v.retries < ttypes.MAX_RETRIES {
				self.reVerifyStatelessTx(k)
				v.retries++
			} else {
				log.Infof("Retry to verify transaction exhausted %x", k.ToArray())
				self.mu.Lock()
				delete(self.pendingTxs, k)
				self.mu.Unlock()
				self.txPoolServer.removePendingTx(k, errors.ErrRetryExhausted)
			}
		}
	}
}

// verifyTx prepares a check request and sends it to the validators.
func (self *txPoolWorker) verifyStatelessTx(tx *tx.Transaction) {
	if tx := self.txPoolServer.GetTxFromPool(tx.Hash()); tx != nil {
		log.Info(fmt.Sprintf("Transaction %x already in the txn pool",
			tx.Hash()))
		self.txPoolServer.removePendingTx(tx.Hash(), errors.ErrDuplicateInput)
		return
	}

	if _, ok := self.pendingTxs[tx.Hash()]; ok {
		log.Info(fmt.Sprintf("Transaction %x already in the verifying process",
			tx.Hash()))
		return
	}
	// Construct the request and send it to each validator server to verify
	req := &vtypes.VerifyTxReq{
		WorkerId: self.workId,
		Tx:       *tx,
	}

	self.txPoolServer.sender.SendVerifyTxReq(vtypes.Stateless, req)

	// Construct the pending transaction
	ptx := &pendingTxInfo{
		tx:         tx,
		req:        req,
		flag:       0,
		retries:    0,
		verifyTime: time.Now(), // Record the time per a txn
	}
	// Add it to the pending transaction list
	self.mu.Lock()
	self.pendingTxs[tx.Hash()] = ptx
	self.mu.Unlock()
}

// reVerifyTx re-sends a check request to the validators.
func (self *txPoolWorker) reVerifyStatelessTx(txHash common.Uint256) {
	pt, ok := self.pendingTxs[txHash]
	if !ok {
		return
	}

	if pt.flag&0xf != ttypes.VERIFY_MASK {
		self.txPoolServer.sender.SendVerifyTxReq(vtypes.Stateless, pt.req)
	}

	// Update the verifying time
	pt.verifyTime = time.Now()
}

// verifyStateful prepares a check request and sends it to the
// stateful validator
func (self *txPoolWorker) verifyStatefulTx(tx *tx.Transaction) {
	req := &vtypes.VerifyTxReq{
		WorkerId: self.workId,
		Tx:       *tx,
	}

	// Construct the pending transaction
	pendingtx := &pendingTxInfo{
		tx:         tx,
		req:        req,
		retries:    0,
		verifyTime: time.Now(),
	}

	initResult := &ttypes.VerifyResult{
		Height:  0,
		Type:    vtypes.Stateless,
		ErrCode: errors.ErrNoError,
	}

	pendingtx.verifyResults = append(pendingtx.verifyResults, initResult)
	// Since the signature has been already verified, mark stateless as true
	pendingtx.flag |= ttypes.STATELESS_MASK

	// Add it to the pending transaction list
	self.mu.Lock()
	self.pendingTxs[tx.Hash()] = pendingtx
	self.mu.Unlock()

	self.txPoolServer.sender.SendVerifyTxReq(vtypes.Stateless, req)
}

// Start is the main event loop.
func (self *txPoolWorker) start() {
	self.timer = time.NewTimer(time.Second * ttypes.EXPIRE_INTERVAL)
	for {
		select {
		case <-self.stopCh:
			self.txPoolServer.wg.Done()
			return
		case tx, ok := <-self.statelessTxCh:
			if ok {
				// Verify rcvTxn
				self.verifyStatelessTx(tx)
			}
		case tx, ok := <-self.statefulTxCh:
			if ok {
				self.verifyStatefulTx(tx)
			}
		case <-self.timer.C:
			self.handleTimeoutEvent()
			self.timer.Stop()
			self.timer.Reset(time.Second * ttypes.EXPIRE_INTERVAL)
		case rsp, ok := <-self.verifyTxRspCh:
			if ok {
				/* Handle the response from validator, if all of cases
				 * are verified, put it to txnPool
				 */
				self.handleValidatorRsp(rsp)
			}
		}
	}
}

// stop closes/releases channels and stops timer
func (self *txPoolWorker) stop() {
	if self.timer != nil {
		self.timer.Stop()
	}
	if self.statelessTxCh != nil {
		close(self.statelessTxCh)
	}
	if self.statefulTxCh != nil {
		close(self.statefulTxCh)
	}
	if self.verifyTxRspCh != nil {
		close(self.verifyTxRspCh)
	}

	if self.stopCh != nil {
		self.stopCh <- true
		close(self.stopCh)
	}
}
