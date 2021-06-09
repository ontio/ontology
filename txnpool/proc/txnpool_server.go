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

// Package proc provides functions for handle messages from
// consensus/ledger/net/http/validators
package proc

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	tx "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	httpcom "github.com/ontio/ontology/http/base/common"
	msgpack "github.com/ontio/ontology/p2pserver/message/msg_pack"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	params "github.com/ontio/ontology/smartcontract/service/native/global_params"
	nutils "github.com/ontio/ontology/smartcontract/service/native/utils"
	tc "github.com/ontio/ontology/txnpool/common"
	"github.com/ontio/ontology/validator/stateful"
	"github.com/ontio/ontology/validator/stateless"
	"github.com/ontio/ontology/validator/types"
)

type serverPendingTx struct {
	tx     *tx.Transaction   // Pending tx
	sender tc.SenderType     // Indicate which sender tx is from
	ch     chan *tc.TxResult // channel to send tx result
}

// TXPoolServer contains all api to external modules
type TXPoolServer struct {
	mu                    sync.RWMutex                        // Sync mutex
	wg                    sync.WaitGroup                      // Worker sync
	worker                *txPoolWorker                       // Worker pool
	txPool                *tc.TXPool                          // The tx pool that holds the valid transaction
	allPendingTxs         map[common.Uint256]*serverPendingTx // The txs that server is processing
	actor                 *actor.PID
	Net                   p2p.P2P
	slots                 chan struct{} // The limited slots for the new transaction
	height                uint32        // The current block height
	gasPrice              uint64        // Gas price to enforce for acceptance into the pool
	disablePreExec        bool          // Disbale PreExecute a transaction
	disableBroadcastNetTx bool          // Disable broadcast tx from network
}

// NewTxPoolServer creates a new tx pool server to schedule workers to
// handle and filter inbound transactions from the network, http, and consensus.
func NewTxPoolServer(disablePreExec, disableBroadcastNetTx bool) *TXPoolServer {
	s := &TXPoolServer{}
	s.init(disablePreExec, disableBroadcastNetTx)
	return s
}

// getGlobalGasPrice returns a global gas price
func getGlobalGasPrice() (uint64, error) {
	mutable, err := httpcom.NewNativeInvokeTransaction(0, 0, nutils.ParamContractAddress, 0, "getGlobalParam", []interface{}{[]interface{}{"gasPrice"}})
	if err != nil {
		return 0, fmt.Errorf("NewNativeInvokeTransaction error:%s", err)
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return 0, err
	}
	result, err := ledger.DefLedger.PreExecuteContract(tx)
	if err != nil {
		return 0, fmt.Errorf("PreExecuteContract failed %v", err)
	}

	queriedParams := new(params.Params)
	data, err := hex.DecodeString(result.Result.(string))
	if err != nil {
		return 0, fmt.Errorf("decode result error %v", err)
	}

	err = queriedParams.Deserialization(common.NewZeroCopySource([]byte(data)))
	if err != nil {
		return 0, fmt.Errorf("deserialize result error %v", err)
	}
	_, param := queriedParams.GetParam("gasPrice")
	if param.Value == "" {
		return 0, fmt.Errorf("failed to get param for gasPrice")
	}

	gasPrice, err := strconv.ParseUint(param.Value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse uint %v", err)
	}

	return gasPrice, nil
}

// getGasPriceConfig returns the bigger one between global and cmd configured
func getGasPriceConfig() uint64 {
	globalGasPrice, err := getGlobalGasPrice()
	if err != nil {
		log.Info(err)
		return 0
	}

	if globalGasPrice < config.DefConfig.Common.GasPrice {
		return config.DefConfig.Common.GasPrice
	}
	return globalGasPrice
}

// init initializes the server with the configured settings
func (s *TXPoolServer) init(disablePreExec, disableBroadcastNetTx bool) {
	// Initial txnPool
	s.txPool = &tc.TXPool{}
	s.txPool.Init()
	s.allPendingTxs = make(map[common.Uint256]*serverPendingTx)

	s.slots = make(chan struct{}, tc.MAX_LIMITATION)
	for i := 0; i < tc.MAX_LIMITATION; i++ {
		s.slots <- struct{}{}
	}

	s.gasPrice = getGasPriceConfig()
	log.Infof("tx pool: the current local gas price is %d", s.gasPrice)

	s.disablePreExec = disablePreExec
	s.disableBroadcastNetTx = disableBroadcastNetTx
	// Create the given concurrent workers
	s.wg.Add(1)
	s.worker = NewTxPoolWoker(s)
	go s.worker.start()
}

// getPendingListSize return the length of the pending tx list.
func (s *TXPoolServer) getPendingListSize() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.allPendingTxs)
}

func (s *TXPoolServer) getHeight() uint32 {
	return atomic.LoadUint32(&s.height)
}

func (s *TXPoolServer) setHeight(height uint32) {
	if height == 0 {
		return
	}
	atomic.StoreUint32(&s.height, height)
}

// getGasPrice returns the current gas price enforced by the transaction pool
func (s *TXPoolServer) getGasPrice() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.gasPrice
}

// removePendingTx removes a transaction from the pending list
// when it is handled. And if the submitter of the valid transaction
// is from http, broadcast it to the network. Meanwhile, check if it
// is in the block from consensus.
func (s *TXPoolServer) removePendingTx(hash common.Uint256, err errors.ErrCode) {
	s.mu.Lock()

	pt, ok := s.allPendingTxs[hash]
	if !ok {
		s.mu.Unlock()
		return
	}

	if err == errors.ErrNoError && ((pt.sender == tc.HttpSender) ||
		(pt.sender == tc.NetSender && !s.disableBroadcastNetTx)) {
		if s.Net != nil {
			msg := msgpack.NewTxn(pt.tx)
			go s.Net.Broadcast(msg)
		}
	}

	replyTxResult(pt.ch, hash, err, err.Error())

	delete(s.allPendingTxs, hash)

	if len(s.allPendingTxs) < tc.MAX_LIMITATION {
		select {
		case s.slots <- struct{}{}:
		default:
			log.Debug("removePendingTx: slots is full")
		}
	}

	s.mu.Unlock()
}

// setPendingTx adds a transaction to the pending list, if the
// transaction is already in the pending list, just return false.
func (s *TXPoolServer) setPendingTx(tx *tx.Transaction, sender tc.SenderType, txResultCh chan *tc.TxResult) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ok := s.allPendingTxs[tx.Hash()]; ok != nil {
		log.Debugf("setPendingTx: transaction %x already in the verifying process", tx.Hash())
		return false
	}

	pt := &serverPendingTx{
		tx:     tx,
		sender: sender,
		ch:     txResultCh,
	}

	s.allPendingTxs[tx.Hash()] = pt
	return true
}

// assignTxToWorker assigns a new transaction to a worker by LB
func (s *TXPoolServer) assignTxToWorker(tx *tx.Transaction, sender tc.SenderType, txResultCh chan *tc.TxResult) bool {
	if ok := s.setPendingTx(tx, sender, txResultCh); !ok {
		replyTxResult(txResultCh, tx.Hash(), errors.ErrDuplicateInput, "duplicated transaction input detected")
		return false
	}

	// Add the rcvTxn to the worker
	s.worker.rcvTXCh <- tx
	return true
}

// GetPID returns an actor pid with the actor type, If the type
// doesn't exist, return nil.
func (s *TXPoolServer) GetPID() *actor.PID {
	return s.actor
}

// registers an actor with the actor type and pid.
func (s *TXPoolServer) RegisterActor(pid *actor.PID) {
	s.actor = pid
}

// Stop stops server and workers.
func (s *TXPoolServer) Stop() {
	if s.actor != nil {
		s.actor.Stop()
	}
	//Stop worker
	s.worker.stop()
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

// getTxCount returns current tx count, including pending and verified
func (s *TXPoolServer) getTxCount() []uint32 {
	ret := make([]uint32, 0)
	ret = append(ret, uint32(s.txPool.GetTransactionCount()))
	ret = append(ret, uint32(s.getPendingListSize()))
	return ret
}

// getTxHashList returns a currently pending tx hash list
func (s *TXPoolServer) getTxHashList() []common.Uint256 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	txHashPool := s.txPool.GetTransactionHashList()
	ret := make([]common.Uint256, 0, len(s.allPendingTxs)+len(txHashPool))
	existedTxHash := make(map[common.Uint256]bool)
	for _, hash := range txHashPool {
		ret = append(ret, hash)
		existedTxHash[hash] = true
	}
	for _, v := range s.allPendingTxs {
		hash := v.tx.Hash()
		if !existedTxHash[hash] {
			ret = append(ret, hash)
			existedTxHash[hash] = true
		}
	}
	return ret
}

// cleanTransactionList cleans the txs in the block from the ledger
func (s *TXPoolServer) cleanTransactionList(txs []*tx.Transaction, height uint32) {
	s.txPool.CleanTransactionList(txs)

	// Check whether to update the gas price and remove txs below the threshold
	if height%tc.UPDATE_FREQUENCY == 0 {
		gasPrice := getGasPriceConfig()
		s.mu.Lock()
		oldGasPrice := s.gasPrice
		s.gasPrice = gasPrice
		s.mu.Unlock()
		if oldGasPrice != gasPrice {
			log.Infof("Transaction pool price threshold updated from %d to %d",
				oldGasPrice, gasPrice)
		}

		if oldGasPrice < gasPrice {
			s.txPool.RemoveTxsBelowGasPrice(gasPrice)
		}
	}
	// Cleanup tx pool
	if !s.disablePreExec {
		remain := s.txPool.Remain()
		for _, t := range remain {
			if ok, _ := preExecCheck(t); !ok {
				log.Debugf("cleanTransactionList: preExecCheck tx %x failed", t.Hash())
				continue
			}
			s.reVerifyStateful(t, tc.NilSender)
		}
	}
}

// delTransaction deletes a transaction in the tx pool.
func (s *TXPoolServer) delTransaction(t *tx.Transaction) {
	s.txPool.DelTxList(t)
}

// adds a valid transaction to the tx pool.
func (s *TXPoolServer) addTxList(txEntry *tc.TXEntry) bool {
	ret := s.txPool.AddTxList(txEntry)
	return ret
}

// getTxStatusReq returns a transaction's status with the transaction hash.
func (s *TXPoolServer) getTxStatusReq(hash common.Uint256) *tc.TxStatus {
	if ret := s.worker.GetTxStatus(hash); ret != nil {
		return ret
	}

	return s.txPool.GetTxStatus(hash)
}

// getTransactionCount returns the tx size of the transaction pool.
func (s *TXPoolServer) getTransactionCount() int {
	return s.txPool.GetTransactionCount()
}

// reVerifyStateful re-verify a transaction's stateful data.
func (s *TXPoolServer) reVerifyStateful(tx *tx.Transaction, sender tc.SenderType) {
	if ok := s.setPendingTx(tx, sender, nil); !ok {
		return
	}

	// Add the rcvTxn to the worker
	s.worker.stfTxCh <- tx
}

// verifies the block from consensus.
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

	processedTxs := make([]*tc.VerifyTxResult, len(req.Txs))

	// Check whether a tx's gas price is lower than the required, if yes, just return error
	txs := make(map[common.Uint256]*tx.Transaction, len(req.Txs))
	for _, t := range req.Txs {
		if t.GasPrice < s.gasPrice {
			entry := &tc.VerifyTxResult{
				Height:  req.Height,
				Tx:      t,
				ErrCode: errors.ErrGasPrice,
			}
			processedTxs = append(processedTxs, entry)
			sender.Tell(&tc.VerifyBlockRsp{TxnPool: processedTxs})
			return
		}
		// Check whether double spent
		if _, ok := txs[t.Hash()]; ok {
			entry := &tc.VerifyTxResult{
				Height:  req.Height,
				Tx:      t,
				ErrCode: errors.ErrDoubleSpend,
			}
			processedTxs = append(processedTxs, entry)
			sender.Tell(&tc.VerifyBlockRsp{TxnPool: processedTxs})
			return
		}
		txs[t.Hash()] = t
	}

	checkBlkResult := s.txPool.GetUnverifiedTxs(req.Txs, req.Height)

	if len(checkBlkResult.UnverifiedTxs) > 0 {
		ch := make(chan *types.CheckResponse, len(checkBlkResult.UnverifiedTxs))
		validator := stateless.NewValidatorPool(5)
		for _, t := range checkBlkResult.UnverifiedTxs {
			validator.SubmitVerifyTask(t, ch)
		}
		for i := 0; i < len(checkBlkResult.UnverifiedTxs); i++ {
			response := <-ch
			if response.ErrCode != errors.ErrNoError {
				processedTxs = append(processedTxs, &tc.VerifyTxResult{
					Height:  req.Height,
					Tx:      txs[response.Hash],
					ErrCode: response.ErrCode,
				})
				sender.Tell(&tc.VerifyBlockRsp{TxnPool: processedTxs})
				return
			}
		}
	}

	lenStateFul := len(checkBlkResult.UnverifiedTxs) + len(checkBlkResult.OldTxs)
	if lenStateFul > 0 {
		currHeight := ledger.DefLedger.GetCurrentBlockHeight()
		for currHeight < req.Height {
			// wait ledger sync up
			log.Warnf("ledger need sync up for tx verification, curr height: %d, expected:%d", currHeight, req.Height)
			time.Sleep(time.Second)
			currHeight = ledger.DefLedger.GetCurrentBlockHeight()
		}

		ch := make(chan *types.CheckResponse, lenStateFul)
		validator := stateful.NewValidatorPool(1)
		for _, tx := range checkBlkResult.UnverifiedTxs {
			validator.SubmitVerifyTask(tx, ch)
		}
		for _, tx := range checkBlkResult.OldTxs {
			validator.SubmitVerifyTask(tx, ch)
		}
		for i := 0; i < lenStateFul; i++ {
			resp := <-ch
			processedTxs = append(processedTxs, &tc.VerifyTxResult{
				Height:  resp.Height,
				Tx:      txs[resp.Hash],
				ErrCode: resp.ErrCode,
			})
			if resp.ErrCode != errors.ErrNoError {
				sender.Tell(&tc.VerifyBlockRsp{TxnPool: processedTxs})
				return
			}
		}
	}

	processedTxs = append(processedTxs, checkBlkResult.VerifiedTxs...)
	sender.Tell(&tc.VerifyBlockRsp{TxnPool: processedTxs})
}
