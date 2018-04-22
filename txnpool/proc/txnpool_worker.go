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
	tx      *tx.Transaction    // That is unverified or on the verifying process
	valTime time.Time          // The start time
	req     *vtypes.VerifyTxReq // Req cache
	flag    uint8              // For different types of verification
	retries uint8              // For resend to validator when time out before verified
	ret     []*ttypes.VerifyResult       // verified results
}

// txPoolWorker handles the tasks scheduled by server
type txPoolWorker struct {
	mu         sync.RWMutex
	workId     uint8                             // Worker ID
	rcvTxCh    chan *tx.Transaction              // The channel of receive transaction
	stfTxCh    chan *tx.Transaction              // The channel of txs to be re-verified stateful
	rspCh      chan *vtypes.VerifyTxRsp           // The channel of verified response
	server     *TXPoolServer                     // The txn pool server pointer
	timer      *time.Timer                       // The timer of reverifying
	stopCh     chan bool                         // stop routine
	pendingTxs map[common.Uint256]*pendingTxInfo // The transaction on the verifying process
}

// init initializes the worker with the configured settings
func (worker *txPoolWorker) init(workID uint8, s *TXPoolServer) {
	worker.rcvTxCh = make(chan *tx.Transaction, ttypes.MAX_PENDING_TXN)
	worker.stfTxCh = make(chan *tx.Transaction, ttypes.MAX_PENDING_TXN)
	worker.pendingTxs = make(map[common.Uint256]*pendingTxInfo)
	worker.rspCh = make(chan *vtypes.VerifyTxRsp, ttypes.MAX_PENDING_TXN)
	worker.stopCh = make(chan bool)
	worker.workId = workID
	worker.server = s
}

// GetTxStatus returns the status in the pending list with the transaction hash
func (worker *txPoolWorker) GetTxStatus(hash common.Uint256) *ttypes.TxStatus {
	worker.mu.RLock()
	defer worker.mu.RUnlock()

	pt, ok := worker.pendingTxs[hash]
	if !ok {
		return nil
	}

	txStatus := &ttypes.TxStatus{
		Hash:  hash,
		Attrs: pt.ret,
	}
	return txStatus
}

// handleRsp handles the verified response from the validator and if
// the tx is valid, add it to the tx pool, or remove it from the pending
// list
func (worker *txPoolWorker) handleRsp(rsp *vtypes.VerifyTxRsp) {
	if rsp.WorkerId != worker.workId {
		return
	}

	worker.mu.Lock()
	defer worker.mu.Unlock()

	pt, ok := worker.pendingTxs[rsp.Hash]
	if !ok {
		return
	}
	if rsp.ErrCode != errors.ErrNoError {
		//Verify fail
		log.Info(fmt.Sprintf("Validator %d: Transaction %x invalid: %s",
			rsp.Type, rsp.Hash, rsp.ErrCode.Error()))
		delete(worker.pendingTxs, rsp.Hash)
		worker.server.removePendingTx(rsp.Hash, rsp.ErrCode)
		return
	}

	if pt.flag&(0x1<<rsp.Type) == 0 {
		retAttr := &ttypes.VerifyResult{
			Height:  rsp.Height,
			Type:    rsp.Type,
			ErrCode: rsp.ErrCode,
		}
		pt.flag |= (0x1 << rsp.Type)
		pt.ret = append(pt.ret, retAttr)
	}

	if pt.flag&0xf == ttypes.VERIFY_MASK {
		worker.putTxPool(pt)
		delete(worker.pendingTxs, rsp.Hash)
	}
}

/* Check if the transaction need to be sent to validator to verify
 * when time out.
 * Todo: Going through the list will take time if the list is too
 * long, need to change the algorithm later
 */
func (worker *txPoolWorker) handleTimeoutEvent() {
	if len(worker.pendingTxs) <= 0 {
		return
	}

	/* Go through the pending list, for those unverified txns,
	 * resend them to the validators
	 */
	for k, v := range worker.pendingTxs {
		if v.flag&0xf != ttypes.VERIFY_MASK && (time.Now().Sub(v.valTime)/time.Second) >=
			ttypes.EXPIRE_INTERVAL {
			if v.retries < ttypes.MAX_RETRIES {
				worker.reVerifyTx(k)
				v.retries++
			} else {
				log.Infof("Retry to verify transaction exhausted %x", k.ToArray())
				worker.mu.Lock()
				delete(worker.pendingTxs, k)
				worker.mu.Unlock()
				worker.server.removePendingTx(k, errors.ErrRetryExhausted)
			}
		}
	}
}

// putTxPool adds a valid transaction to the tx pool and removes it from
// the pending list.
func (worker *txPoolWorker) putTxPool(pt *pendingTxInfo) bool {
	txEntry := &ttypes.TxEntry{
		Tx:    pt.tx,
		Attrs: pt.ret,
		Fee:   pt.tx.GetTotalFee(),
	}
	worker.server.appendTxEntry2Pool(txEntry)
	worker.server.removePendingTx(pt.tx.Hash(), errors.ErrNoError)
	return true
}

// verifyTx prepares a check request and sends it to the validators.
func (worker *txPoolWorker) verifyTx(tx *tx.Transaction) {
	if tx := worker.server.GetTransactionFromPool(tx.Hash()); tx != nil {
		log.Info(fmt.Sprintf("Transaction %x already in the txn pool",
			tx.Hash()))
		worker.server.removePendingTx(tx.Hash(), errors.ErrDuplicateInput)
		return
	}

	if _, ok := worker.pendingTxs[tx.Hash()]; ok {
		log.Info(fmt.Sprintf("Transaction %x already in the verifying process",
			tx.Hash()))
		return
	}
	// Construct the request and send it to each validator server to verify
	req := &vtypes.VerifyTxReq{
		WorkerId: worker.workId,
		Tx:       *tx,
	}

	worker.sendStatelessVerifyTxReq(req)

	// Construct the pending transaction
	pt := &pendingTxInfo{
		tx:      tx,
		req:     req,
		flag:    0,
		retries: 0,
	}
	// Add it to the pending transaction list
	worker.mu.Lock()
	worker.pendingTxs[tx.Hash()] = pt
	worker.mu.Unlock()
	// Record the time per a txn
	pt.valTime = time.Now()
}

// reVerifyTx re-sends a check request to the validators.
func (worker *txPoolWorker) reVerifyTx(txHash common.Uint256) {
	pt, ok := worker.pendingTxs[txHash]
	if !ok {
		return
	}

	if pt.flag&0xf != ttypes.VERIFY_MASK {
		worker.sendStatelessVerifyTxReq(pt.req)
	}

	// Update the verifying time
	pt.valTime = time.Now()
}

// sendReq2Validator sends a check request to the validators
func (worker *txPoolWorker) sendStatelessVerifyTxReq(req *vtypes.VerifyTxReq) bool {
	rspPid := worker.server.GetPID(ttypes.VerifyRspActor)
	if rspPid == nil {
		log.Info("VerifyRspActor not exist")
		return false
	}

	pids := worker.server.getNextValidatorPIDs()
	if pids == nil {
		return false
	}
	for _, pid := range pids {
		pid.Request(req, rspPid)
	}

	return true
}

// sendReq2StatefulV sends a check request to the stateful validator
func (worker *txPoolWorker) sendStatefulVerifyTxReq(req *vtypes.VerifyTxReq) {
	rspPid := worker.server.GetPID(ttypes.VerifyRspActor)
	if rspPid == nil {
		log.Info("VerifyRspActor not exist")
		return
	}

	pid := worker.server.getNextValidatorPID(vtypes.Statefull)
	log.Info("worker send tx to the stateful")
	if pid == nil {
		return
	}

	pid.Request(req, rspPid)

}

// verifyStateful prepares a check request and sends it to the
// stateful validator
func (worker *txPoolWorker) verifyStateful(tx *tx.Transaction) {
	req := &vtypes.VerifyTxReq{
		WorkerId: worker.workId,
		Tx:       *tx,
	}

	// Construct the pending transaction
	pt := &pendingTxInfo{
		tx:      tx,
		req:     req,
		retries: 0,
		valTime: time.Now(),
	}

	retAttr := &ttypes.VerifyResult{
		Height:  0,
		Type:    vtypes.Stateless,
		ErrCode: errors.ErrNoError,
	}

	pt.ret = append(pt.ret, retAttr)
	// Since the signature has been already verified, mark stateless as true
	pt.flag |= ttypes.STATELESS_MASK

	// Add it to the pending transaction list
	worker.mu.Lock()
	worker.pendingTxs[tx.Hash()] = pt
	worker.mu.Unlock()

	worker.sendStatefulVerifyTxReq(req)
}

// Start is the main event loop.
func (worker *txPoolWorker) start() {
	worker.timer = time.NewTimer(time.Second * ttypes.EXPIRE_INTERVAL)
	for {
		select {
		case <-worker.stopCh:
			worker.server.wg.Done()
			return
		case rcvTx, ok := <-worker.rcvTxCh:
			if ok {
				// Verify rcvTxn
				worker.verifyTx(rcvTx)
			}
		case stfTx, ok := <-worker.stfTxCh:
			if ok {
				worker.verifyStateful(stfTx)
			}
		case <-worker.timer.C:
			worker.handleTimeoutEvent()
			worker.timer.Stop()
			worker.timer.Reset(time.Second * ttypes.EXPIRE_INTERVAL)
		case rsp, ok := <-worker.rspCh:
			if ok {
				/* Handle the response from validator, if all of cases
				 * are verified, put it to txnPool
				 */
				worker.handleRsp(rsp)
			}
		}
	}
}

// stop closes/releases channels and stops timer
func (worker *txPoolWorker) stop() {
	if worker.timer != nil {
		worker.timer.Stop()
	}
	if worker.rcvTxCh != nil {
		close(worker.rcvTxCh)
	}
	if worker.stfTxCh != nil {
		close(worker.stfTxCh)
	}
	if worker.rspCh != nil {
		close(worker.rspCh)
	}

	if worker.stopCh != nil {
		worker.stopCh <- true
		close(worker.stopCh)
	}
}
