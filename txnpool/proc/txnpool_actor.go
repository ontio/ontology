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
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	tx "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events/message"
	hComm "github.com/ontio/ontology/http/base/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	tc "github.com/ontio/ontology/txnpool/common"
	"github.com/ontio/ontology/validator/types"
)

// NewTxPoolActor creates an actor to handle the messages from the consensus
func NewTxPoolActor(s *TXPoolServer) *TxPoolActor {
	a := &TxPoolActor{}
	a.setServer(s)
	return a
}

// NewVerifyRspActor creates an actor to handle the verified result from validators
func NewVerifyRspActor(s *TXPoolServer) *VerifyRspActor {
	a := &VerifyRspActor{}
	a.setServer(s)
	return a
}

// isBalanceEnough checks if the tranactor has enough to cover gas cost
func isBalanceEnough(address common.Address, gas uint64) bool {
	balance, err := hComm.GetContractBalance(0, utils.OngContractAddress, address)
	if err != nil {
		log.Debugf("failed to get contract balance %s err %v",
			address.ToHexString(), err)
		return false
	}
	return balance >= gas
}

// preExecCheck checks whether preExec pass
func preExecCheck(lgr *ledger.Ledger, txn *tx.Transaction) (bool, string) {
	result, err := lgr.PreExecuteContract(txn)
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

	case *tc.GetPendingTxnReq:
		sender := context.Sender()

		log.Debugf("txpool actor receives getting pedning tx req from %v", sender)

		res := tpa.server.getPendingTxs(msg.ByCount)
		if sender != nil {
			sender.Request(&tc.GetPendingTxnRsp{Txs: res}, context.Self())
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

func (tpa *TxPoolActor) setServer(s *TXPoolServer) {
	tpa.server = s
}

// VerifyRspActor: Handle the response from the validators
type VerifyRspActor struct {
	server *TXPoolServer
}

// Receive implements the actor interface
func (vpa *VerifyRspActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("txpool-verify actor: started and be ready to receive validator's msg, shard %d", vpa.server.shardID)

	case *actor.Stopping:
		log.Warn("txpool-verify actor: stopping")

	case *actor.Restarting:
		log.Warn("txpool-verify actor: Restarting")

	case *types.RegisterValidator:
		log.Debugf("txpool-verify actor:: validator %v connected, shard %d", msg.Sender, vpa.server.shardID)
		vpa.server.registerValidator(msg)

	case *types.UnRegisterValidator:
		log.Debugf("txpool-verify actor:: validator %d:%v disconnected", msg.Type, msg.Id)

		vpa.server.unRegisterValidator(msg.Type, msg.Id)

	case *types.CheckResponse:
		log.Debug("txpool-verify actor:: Receives verify rsp message")

		vpa.server.assignRspToWorker(msg)

	default:
		log.Debugf("txpool-verify actor:Unknown msg %v type %v", msg, reflect.TypeOf(msg))
	}
}

func (vpa *VerifyRspActor) setServer(s *TXPoolServer) {
	vpa.server = s
}
