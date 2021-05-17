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
	"sync"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	tx "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	tc "github.com/ontio/ontology/txnpool/common"
	"github.com/ontio/ontology/validator/stateful"
	"github.com/ontio/ontology/validator/stateless"
	"github.com/ontio/ontology/validator/types"
)

// pendingTx contains the transaction, the time of starting verifying,
// the cache of check request, the verified result and retry mechanism
type pendingTx struct {
	tx              *tx.Transaction // That is unverified or on the verifying process
	valTime         time.Time       // The start time
	passedStateless bool
	passedStateful  bool
	CheckHeight     uint32
}

func (self *pendingTx) GetTxAttr() []*tc.TXAttr {
	var res []*tc.TXAttr
	if self.passedStateless {
		res = append(res,
			&tc.TXAttr{
				Height:  0,
				Type:    types.Stateless,
				ErrCode: errors.ErrNoError,
			})
	}
	if self.passedStateful {
		res = append(res,
			&tc.TXAttr{
				Height:  self.CheckHeight,
				Type:    types.Stateful,
				ErrCode: errors.ErrNoError,
			})
	}

	return res
}

// txPoolWorker handles the tasks scheduled by server
type txPoolWorker struct {
	mu            sync.RWMutex
	rcvTXCh       chan *tx.Transaction          // The channel of receive transaction
	stfTxCh       chan *tx.Transaction          // The channel of txs to be re-verified stateful
	rspCh         chan *types.CheckResponse     // The channel of verified response
	server        *TXPoolServer                 // The txn pool server pointer
	stopCh        chan bool                     // stop routine
	pendingTxList map[common.Uint256]*pendingTx // The transaction on the verifying process
	stateless     *stateless.ValidatorPool
	stateful      *stateful.ValidatorPool
}

func NewTxPoolWoker(s *TXPoolServer) *txPoolWorker {
	worker := &txPoolWorker{}
	worker.rcvTXCh = make(chan *tx.Transaction, tc.MAX_PENDING_TXN)
	worker.stfTxCh = make(chan *tx.Transaction, tc.MAX_PENDING_TXN)
	worker.pendingTxList = make(map[common.Uint256]*pendingTx)
	worker.rspCh = make(chan *types.CheckResponse, tc.MAX_PENDING_TXN)
	worker.stopCh = make(chan bool)
	worker.server = s
	worker.stateless = stateless.NewValidatorPool(2)
	worker.stateful = stateful.NewValidatorPool(1)

	return worker
}

// GetTxStatus returns the status in the pending list with the transaction hash
func (worker *txPoolWorker) GetTxStatus(hash common.Uint256) *tc.TxStatus {
	worker.mu.RLock()
	defer worker.mu.RUnlock()

	pt, ok := worker.pendingTxList[hash]
	if !ok {
		return nil
	}

	txStatus := &tc.TxStatus{
		Hash:  hash,
		Attrs: pt.GetTxAttr(),
	}
	return txStatus
}

// handleRsp handles the verified response from the validator and if
// the tx is valid, add it to the tx pool, or remove it from the pending
// list
func (worker *txPoolWorker) handleRsp(rsp *types.CheckResponse) {
	worker.mu.Lock()
	defer worker.mu.Unlock()

	pt, ok := worker.pendingTxList[rsp.Hash]
	if !ok {
		return
	}
	if rsp.ErrCode != errors.ErrNoError {
		//Verify fail
		log.Debugf("handleRsp: validator %d transaction %x invalid: %s", rsp.Type, rsp.Hash, rsp.ErrCode.Error())
		delete(worker.pendingTxList, rsp.Hash)
		worker.server.removePendingTx(rsp.Hash, rsp.ErrCode)
		if pt.tx.TxType == tx.EIP155 {
			worker.server.pendingNonces.setIfLower(pt.tx.Payer, uint64(pt.tx.Nonce))
			worker.server.pendingEipTxs[pt.tx.Payer].txs.Remove(uint64(pt.tx.Nonce))
		}
		return
	}

	if rsp.Type == types.Stateful && rsp.Height < worker.server.getHeight() {
		// If validator's height is less than the required one, re-validate it.
		worker.startStatefulVerify(pt.tx)
		pt.valTime = time.Now()
		return
	}

	switch rsp.Type {
	case types.Stateful:
		pt.passedStateful = true
		if rsp.Height > pt.CheckHeight {
			pt.CheckHeight = rsp.Height
		}
	case types.Stateless:
		pt.passedStateless = true
	}

	if pt.passedStateless && pt.passedStateful {
		worker.putTxPool(pt)
		delete(worker.pendingTxList, rsp.Hash)
	}
}

func (worker *txPoolWorker) putTxPool(pt *pendingTx) bool {
	txEntry := &tc.TXEntry{
		Tx:    pt.tx,
		Attrs: pt.GetTxAttr(),
	}
	f := worker.server.addTxList(txEntry)
	if f {
		worker.server.removePendingTx(pt.tx.Hash(), errors.ErrNoError)
		//remove from pendingEipTxs
		worker.server.addEIPTxPool(txEntry.Tx)
		worker.server.removeEIPPendingTx(pt.tx)
	}
	return true
}

// verifyTx prepares a check request and sends it to the validators.
func (worker *txPoolWorker) verifyTx(tx *tx.Transaction) {
	if tx := worker.server.getTransaction(tx.Hash()); tx != nil {
		log.Debugf("verifyTx: transaction %x already in the txn pool", tx.Hash())
		worker.server.removePendingTx(tx.Hash(), errors.ErrDuplicateInput)
		return
	}

	if _, ok := worker.pendingTxList[tx.Hash()]; ok {
		log.Debugf("verifyTx: transaction %x already in the verifying process", tx.Hash())
		return
	}

	// Construct the pending transaction
	pt := &pendingTx{
		tx: tx,
	}
	pt.valTime = time.Now()

	// Add it to the pending transaction list
	worker.mu.Lock()
	worker.pendingTxList[tx.Hash()] = pt
	worker.mu.Unlock()

	// need register tx to pending list first to avoid races caused by too fast verification
	worker.startFullVerify(tx)
}

func (worker *txPoolWorker) startFullVerify(tx *tx.Transaction) {
	worker.stateless.SubmitVerifyTask(tx, worker.rspCh)
	worker.stateful.SubmitVerifyTask(tx, worker.rspCh)
}

func (worker *txPoolWorker) startStatefulVerify(tx *tx.Transaction) {
	worker.stateful.SubmitVerifyTask(tx, worker.rspCh)
}

// verifyStateful prepares a check request and sends it to the
// stateful validator
func (worker *txPoolWorker) verifyStateful(tx *tx.Transaction) {
	// Construct the pending transaction
	pt := &pendingTx{
		tx:      tx,
		valTime: time.Now(),
	}

	// Since the signature has been already verified, mark stateless as true
	pt.passedStateless = true

	// Add it to the pending transaction list
	worker.mu.Lock()
	worker.pendingTxList[tx.Hash()] = pt
	worker.mu.Unlock()

	worker.startStatefulVerify(tx)
}

// Start is the main event loop.
func (worker *txPoolWorker) start() {
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
		case rsp, ok := <-worker.rspCh:
			if ok {
				worker.handleRsp(rsp)
			}
		}
	}
}

// stop closes/releases channels
func (worker *txPoolWorker) stop() {
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
