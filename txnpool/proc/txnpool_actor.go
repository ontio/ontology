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
	"reflect"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	tx "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/events/message"
	hComm "github.com/ontio/ontology/http/base/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	tc "github.com/ontio/ontology/txnpool/common"
)

// creates an actor to handle the transaction-based messages from network and http
func NewTxPoolService(s *TXPoolServer) *TxPoolService {
	a := &TxPoolService{server: s}
	return a
}

// NewTxPoolActor creates an actor to handle the messages from the consensus
func NewTxPoolActor(s *TXPoolServer) *TxPoolActor {
	a := &TxPoolActor{server: s}
	return a
}

// isBalanceEnough checks if the tranactor has enough to cover gas cost
func isBalanceEnough(address common.Address, gas uint64) bool {
	balance, _, err := hComm.GetContractBalance(0, []common.Address{utils.OngContractAddress}, address, false)
	if err != nil {
		log.Debugf("failed to get contract balance %s err %v", address.ToHexString(), err)
		return false
	}
	return balance[0] >= gas
}

func replyTxResult(txResultCh chan *tc.TxResult, hash common.Uint256, err errors.ErrCode, desc string) {
	if txResultCh != nil {
		result := &tc.TxResult{
			Err:  err,
			Hash: hash,
			Desc: desc,
		}

		txResultCh <- result
	}
}

// preExecCheck checks whether preExec pass
func preExecCheck(txn *tx.Transaction) (bool, string) {
	result, err := ledger.DefLedger.PreExecuteContract(txn)
	if err != nil {
		log.Debugf("preExecCheck: failed to preExecuteContract tx %x err %v",
			txn.Hash(), err)
	}
	if txn.GasLimit < result.Gas {
		log.Debugf("preExecCheck: transaction's gasLimit %d is less than preExec gasLimit %d",
			txn.GasLimit, result.Gas)
		return false, fmt.Sprintf("transaction's gasLimit %d is less than preExec gasLimit %d",
			txn.GasLimit, result.Gas)
	}
	gas, overflow := common.SafeMul(txn.GasPrice, result.Gas)
	if overflow {
		log.Debugf("preExecCheck: gasPrice %d preExec gasLimit %d overflow",
			txn.GasPrice, result.Gas)
		return false, fmt.Sprintf("gasPrice %d preExec gasLimit %d overflow",
			txn.GasPrice, result.Gas)
	}
	if !isBalanceEnough(txn.Payer, gas) {
		log.Debugf("preExecCheck: transactor %s has no balance enough to cover gas cost %d",
			txn.Payer.ToHexString(), gas)
		return false, fmt.Sprintf("transactor %s has no balance enough to cover gas cost %d",
			txn.Payer.ToHexString(), gas)
	}
	return true, ""
}

// TxnActor: Handle the low priority msg from P2P and API
type TxPoolService struct {
	server *TXPoolServer
}

// handles a transaction from network and http
func (ta *TxPoolService) handleTransaction(sender tc.SenderType, txn *tx.Transaction, txResultCh chan *tc.TxResult) {
	if len(txn.ToArray()) > tc.MAX_TX_SIZE {
		log.Debugf("handleTransaction: reject a transaction due to size over 1M")
		replyTxResult(txResultCh, txn.Hash(), errors.ErrUnknown, "size is over 1M")
		return
	}

	if ta.server.getTransaction(txn.Hash()) != nil {
		log.Debugf("handleTransaction: transaction %x already in the txn pool", txn.Hash())

		replyTxResult(txResultCh, txn.Hash(), errors.ErrDuplicateInput,
			fmt.Sprintf("transaction %x is already in the tx pool", txn.Hash()))
		return
	}

	if ta.server.getTransactionCount() >= tc.MAX_CAPACITY {
		log.Debugf("handleTransaction: transaction pool is full for tx %x", txn.Hash())

		replyTxResult(txResultCh, txn.Hash(), errors.ErrTxPoolFull, "transaction pool is full")
		return
	}

	if _, overflow := common.SafeMul(txn.GasLimit, txn.GasPrice); overflow {
		replyTxResult(txResultCh, txn.Hash(), errors.ErrUnknown,
			fmt.Sprintf("gasLimit %d * gasPrice %d overflow", txn.GasLimit, txn.GasPrice))
		return
	}

	gasLimitConfig := config.DefConfig.Common.MinGasLimit
	gasPriceConfig := ta.server.getGasPrice()
	if txn.GasLimit < gasLimitConfig || txn.GasPrice < gasPriceConfig {
		log.Debugf("handleTransaction: invalid gasLimit %v, gasPrice %v", txn.GasLimit, txn.GasPrice)
		replyTxResult(txResultCh, txn.Hash(), errors.ErrUnknown,
			fmt.Sprintf("Please input gasLimit >= %d and gasPrice >= %d", gasLimitConfig, gasPriceConfig))
		return
	}

	if txn.TxType == tx.Deploy && txn.GasLimit < neovm.CONTRACT_CREATE_GAS {
		log.Debugf("handleTransaction: deploy tx invalid gasLimit %v, gasPrice %v", txn.GasLimit, txn.GasPrice)
		replyTxResult(txResultCh, txn.Hash(), errors.ErrUnknown,
			fmt.Sprintf("Deploy tx gaslimit should >= %d", neovm.CONTRACT_CREATE_GAS))
		return
	}

	if txn.IsEipTx() {
		if txn.GasLimit > config.DefConfig.Common.ETHTxGasLimit {
			replyTxResult(txResultCh, txn.Hash(), errors.ErrUnknown, "EIP155 tx gaslimit exceed ")
			return
		}

		eiptx, err := txn.GetEIP155Tx()
		if err != nil {
			log.Errorf("handleTransaction GetEIP155Tx failed:%s", err.Error())
			replyTxResult(txResultCh, txn.Hash(), errors.ErrUnknown, "Invalid EIP155 transaction format ")
			return
		}

		currentNonce := ta.server.CurrentNonce(txn.Payer)
		if eiptx.Nonce() < currentNonce {
			replyTxResult(txResultCh, txn.Hash(), errors.ErrUnknown,
				fmt.Sprintf("handleTransaction lower nonce:%d ,current nonce:%d", eiptx.Nonce(), currentNonce))
			return
		}

		balance, err := tc.GetOngBalance(txn.Payer)
		if err != nil {
			log.Errorf("GetOngBalance failed:%s", err.Error())
			replyTxResult(txResultCh, txn.Hash(), errors.ErrUnknown,
				fmt.Sprintf("GetOngBalance failed:%s", err.Error()))
			return
		}

		if balance.Cmp(txn.Cost()) < 0 {
			replyTxResult(txResultCh, txn.Hash(), errors.ErrUnknown,
				fmt.Sprintf("not enough ong balance for %s - has:%d - want:%d", txn.Payer.ToHexString(), balance, txn.Cost()))
			return
		}
	}

	if !ta.server.disablePreExec {
		if ok, desc := preExecCheck(txn); !ok {
			log.Debugf("handleTransaction: preExecCheck tx %x failed", txn.Hash())
			replyTxResult(txResultCh, txn.Hash(), errors.ErrUnknown, desc)
			return
		}
		log.Debugf("handleTransaction: preExecCheck tx %x passed", txn.Hash())
	}
	<-ta.server.slots
	ta.server.startTxVerify(txn, sender, txResultCh)
}

func (ta *TxPoolService) GetTransaction(hash common.Uint256) *tx.Transaction {
	res := ta.server.getTransaction(hash)
	return res
}

func (ta *TxPoolService) GetTransactionStatus(hash common.Uint256) *tc.TxStatus {
	res := ta.server.getTxStatusReq(hash)
	return res
}

func (ta *TxPoolService) GetTxAmount() []uint32 {
	return ta.server.getTxCount()
}

func (ta *TxPoolService) GetTxList() []common.Uint256 {
	return ta.server.getTxHashList()
}

func (ta *TxPoolService) AppendTransaction(sender tc.SenderType, txn *tx.Transaction) *tc.TxResult {
	ch := make(chan *tc.TxResult, 1)
	ta.handleTransaction(sender, txn, ch)
	return <-ch
}

func (ta *TxPoolService) AppendTransactionAsync(sender tc.SenderType, txn *tx.Transaction) {
	ta.handleTransaction(sender, txn, nil)
}

// TxnPoolActor: Handle the high priority request from Consensus
type TxPoolActor struct {
	server *TXPoolServer
}

// Receive implements the actor interface
func (tpa *TxPoolActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("txpool actor started and be ready to receive txPool msg")

	case *actor.Stopping:
		log.Warn("txpool actor stopping")

	case *actor.Restarting:
		log.Warn("txpool actor Restarting")

	case *tc.GetTxnPoolReq:
		sender := context.Sender()

		log.Debugf("txpool actor receives getting tx pool req from %v", sender)

		res := tpa.server.getTxPool(msg.ByCount, msg.Height)
		if sender != nil {
			sender.Request(&tc.GetTxnPoolRsp{TxnPool: res}, context.Self())
		}

	case *tc.VerifyBlockReq:
		sender := context.Sender()

		log.Debugf("txpool actor receives verifying block req from %v", sender)

		tpa.server.verifyBlock(msg, sender)

	case *message.SaveBlockCompleteMsg:
		sender := context.Sender()

		log.Debugf("txpool actor receives block complete event from %v", sender)

		if msg.Block != nil {
			tpa.server.cleanTransactionList(msg.Block.Transactions, msg.Block.Header.Height)
		}

	default:
		log.Debugf("txpool actor: unknown msg %v type %v", msg, reflect.TypeOf(msg))
	}
}
