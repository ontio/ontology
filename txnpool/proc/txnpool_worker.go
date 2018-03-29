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
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	tx "github.com/Ontology/core/types"
	"github.com/Ontology/errors"
	tc "github.com/Ontology/txnpool/common"
	"github.com/Ontology/validator/types"
	"sync"
	"time"
)

type pendingTx struct {
	tx      *tx.Transaction // That is unverified or on the verifying process
	worker  *txPoolWorker   // Which worker handles it
	valTime time.Time       // The start time
	req     *types.CheckTx  // Req cache
	flag    uint8           // For different types of verification
	retries uint8           // For resend to validator when time out before verified
	ret     []*tc.TXAttr    // verified results
}

type txPoolWorker struct {
	mu            sync.RWMutex
	workId        uint8                         // Worker ID
	rcvTXCh       chan *tx.Transaction          // The channel of receive transaction
	stfTxCh       chan *tx.Transaction          // The channel of txs to be re-verified stateful
	rspCh         chan *types.CheckResponse     // The channel of verified response
	server        *TXPoolServer                 // The txn pool server pointer
	timer         *time.Timer                   // The timer of reverifying
	stopCh        chan bool                     // stop routine
	pendingTxList map[common.Uint256]*pendingTx // The transaction on the verifying process
}

func (worker *txPoolWorker) init(workID uint8, s *TXPoolServer) {
	worker.rcvTXCh = make(chan *tx.Transaction, tc.MAX_PENDING_TXN)
	worker.stfTxCh = make(chan *tx.Transaction, tc.MAX_PENDING_TXN)
	worker.pendingTxList = make(map[common.Uint256]*pendingTx)
	worker.rspCh = make(chan *types.CheckResponse, tc.MAX_PENDING_TXN)
	worker.stopCh = make(chan bool)
	worker.workId = workID
	worker.server = s
}

func (worker *txPoolWorker) GetTxStatus(hash common.Uint256) *tc.TxStatus {
	worker.mu.RLock()
	defer worker.mu.RUnlock()

	pt, ok := worker.pendingTxList[hash]
	if !ok {
		return nil
	}

	txStatus := &tc.TxStatus{
		Hash:  hash,
		Attrs: pt.ret,
	}
	return txStatus
}

func (worker *txPoolWorker) handleRsp(rsp *types.CheckResponse) {
	if rsp.WorkerId != worker.workId {
		return
	}

	worker.mu.Lock()
	defer worker.mu.Unlock()

	pt, ok := worker.pendingTxList[rsp.Hash]
	if !ok {
		return
	}
	if rsp.ErrCode != errors.ErrNoError {
		//Verify fail
		log.Info(fmt.Sprintf("Validator %d: Transaction %x invalid: %s",
			rsp.Type, rsp.Hash, rsp.ErrCode.Error()))
		delete(worker.pendingTxList, rsp.Hash)
		worker.server.removePendingTx(rsp.Hash, rsp.ErrCode)
		return
	}

	if pt.flag&(0x1<<rsp.Type) == 0 {
		retAttr := &tc.TXAttr{
			Height:  rsp.Height,
			Type:    rsp.Type,
			ErrCode: rsp.ErrCode,
		}
		pt.flag |= (0x1 << rsp.Type)
		pt.ret = append(pt.ret, retAttr)
	}

	if pt.flag&0xf == tc.VERIFY_MASK {
		worker.putTxPool(pt)
		delete(worker.pendingTxList, rsp.Hash)
	}
}

/* Check if the transaction need to be sent to validator to verify
 * when time out.
 * Todo: Going through the list will take time if the list is too
 * long, need to change the algorithm later
 */
func (worker *txPoolWorker) handleTimeoutEvent() {
	if len(worker.pendingTxList) <= 0 {
		return
	}

	/* Go through the pending list, for those unverified txns,
	 * resend them to the validators
	 */
	for k, v := range worker.pendingTxList {
		if v.flag&0xf != tc.VERIFY_MASK && (time.Now().Sub(v.valTime)/time.Second) >=
			tc.EXPIRE_INTERVAL {
			if v.retries < tc.MAX_RETRIES {
				worker.reVerifyTx(k)
				v.retries++
			} else {
				log.Infof("Retry to verify transaction exhausted %x", k.ToArray())
				worker.mu.Lock()
				delete(worker.pendingTxList, k)
				worker.mu.Unlock()
				worker.server.removePendingTx(k, errors.ErrRetryExhausted)
			}
		}
	}
}

func (worker *txPoolWorker) putTxPool(pt *pendingTx) bool {
	txEntry := &tc.TXEntry{
		Tx:    pt.tx,
		Attrs: pt.ret,
		Fee:   pt.tx.GetTotalFee(),
	}
	worker.server.addTxList(txEntry)
	worker.server.removePendingTx(pt.tx.Hash(), errors.ErrNoError)
	return true
}

func (worker *txPoolWorker) verifyTx(tx *tx.Transaction) {
	if tx := worker.server.getTransaction(tx.Hash()); tx != nil {
		log.Info(fmt.Sprintf("Transaction %x already in the txn pool",
			tx.Hash()))
		worker.server.removePendingTx(tx.Hash(), errors.ErrDuplicatedTx)
		return
	}

	if _, ok := worker.pendingTxList[tx.Hash()]; ok {
		log.Info(fmt.Sprintf("Transaction %x already in the verifying process",
			tx.Hash()))
		return
	}
	// Construct the request and send it to each validator server to verify
	req := &types.CheckTx{
		WorkerId: worker.workId,
		Tx:       *tx,
	}

	worker.sendReq2Validator(req)

	// Construct the pending transaction
	pt := &pendingTx{
		tx:      tx,
		worker:  worker,
		req:     req,
		flag:    0,
		retries: 0,
	}
	// Add it to the pending transaction list
	worker.mu.Lock()
	worker.pendingTxList[tx.Hash()] = pt
	worker.mu.Unlock()
	// Record the time per a txn
	pt.valTime = time.Now()
}

func (worker *txPoolWorker) reVerifyTx(txHash common.Uint256) {
	pt, ok := worker.pendingTxList[txHash]
	if !ok {
		return
	}

	if pt.flag&0xf != tc.VERIFY_MASK {
		worker.sendReq2Validator(pt.req)
	}

	// Update the verifying time
	pt.valTime = time.Now()
}

func (worker *txPoolWorker) sendReq2Validator(req *types.CheckTx) (send bool) {
	defer func() {
		if recover() != nil {
			send = false
		}
	}()

	rspPid := worker.server.GetPID(tc.VerifyRspActor)
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

func (worker *txPoolWorker) sendReq2StatefulV(req *types.CheckTx) {
	rspPid := worker.server.GetPID(tc.VerifyRspActor)
	if rspPid == nil {
		log.Info("VerifyRspActor not exist")
		return
	}

	pid := worker.server.getNextValidatorPID(types.Statefull)
	log.Info("worker send tx to the stateful")
	if pid == nil {
		return
	}

	pid.Request(req, rspPid)

}

func (worker *txPoolWorker) verifyStateful(tx *tx.Transaction) {
	req := &types.CheckTx{
		WorkerId: worker.workId,
		Tx:       *tx,
	}

	// Construct the pending transaction
	pt := &pendingTx{
		tx:      tx,
		worker:  worker,
		req:     req,
		retries: 0,
		valTime: time.Now(),
	}

	retAttr := &tc.TXAttr{
		Height:  0,
		Type:    types.Statefull,
		ErrCode: errors.ErrNoError,
	}

	pt.ret = append(pt.ret, retAttr)
	// Since the signature has been already verified, mark stateless as true
	pt.flag |= tc.STATELESS_MASK

	// Add it to the pending transaction list
	worker.mu.Lock()
	worker.pendingTxList[tx.Hash()] = pt
	worker.mu.Unlock()

	worker.sendReq2StatefulV(req)
}

func (worker *txPoolWorker) start() {
	worker.timer = time.NewTimer(time.Second * tc.EXPIRE_INTERVAL)
	for {
		select {
		case <-worker.stopCh:
			worker.server.wg.Done()
			return
		case rcvTx, ok := <-worker.rcvTXCh:
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
			worker.timer.Reset(time.Second * tc.EXPIRE_INTERVAL)
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

func (worker *txPoolWorker) stop() {
	if worker.timer != nil {
		worker.timer.Stop()
	}
	if worker.rcvTXCh != nil {
		close(worker.rcvTXCh)
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
