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

package req

import (
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	p2pcommon "github.com/ontio/ontology/p2pserver/common"
	tc "github.com/ontio/ontology/txnpool/common"
)

const txnPoolReqTimeout = p2pcommon.ACTOR_TIMEOUT * time.Second

var txnPoolPid *actor.PID

func SetTxnPoolPid(txnPid *actor.PID) {
	txnPoolPid = txnPid
}

//add txn to txnpool
func AddTransaction(transaction *types.Transaction) {
	txReq := &tc.TxReq{
		Tx:     transaction,
		Sender: tc.NetSender,
	}
	txnPoolPid.Tell(txReq)
}

//get all txns
func GetTxnPool(byCount bool) ([]*tc.TXEntry, error) {
	future := txnPoolPid.RequestFuture(&tc.GetTxnPoolReq{ByCount: byCount}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("net_server GetTxnPool ERROR: "), err)
		return nil, err
	}
	return result.(tc.GetTxnPoolRsp).TxnPool, nil
}

//get txn according to hash
func GetTransaction(hash common.Uint256) (*types.Transaction, error) {
	future := txnPoolPid.RequestFuture(&tc.GetTxnReq{Hash: hash}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("net_server GetTransaction ERROR: "), err)
		return nil, err
	}
	return result.(tc.GetTxnRsp).Txn, nil
}

//check whether txn in txnpool
func CheckTransaction(hash common.Uint256) (bool, error) {
	future := txnPoolPid.RequestFuture(&tc.CheckTxnReq{Hash: hash}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("net_server CheckTransaction ERROR: "), err)
		return false, err
	}
	return result.(tc.CheckTxnRsp).Ok, nil
}

//get tx status according to hash
func GetTransactionStatus(hash common.Uint256) ([]*tc.TXAttr, error) {
	future := txnPoolPid.RequestFuture(&tc.GetTxnStatusReq{Hash: hash}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("net_server GetTransactionStatus ERROR: "), err)
		return nil, err
	}
	return result.(tc.GetTxnStatusRsp).TxStatus, nil
}

//get pending txn by count
func GetPendingTxn(byCount bool) ([]*types.Transaction, error) {
	future := txnPoolPid.RequestFuture(&tc.GetPendingTxnReq{ByCount: byCount}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("net_server GetPendingTxn ERROR: "), err)
		return nil, err
	}
	return result.(tc.GetPendingTxnRsp).Txs, nil
}

//get veritfy block result from txnpool
func VerifyBlock(height uint32, txs []*types.Transaction) ([]*tc.VerifyTxResult, error) {
	future := txnPoolPid.RequestFuture(&tc.VerifyBlockReq{Height: height, Txs: txs}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("net_server VerifyBlock ERROR: "), err)
		return nil, err
	}
	return result.(tc.VerifyBlockRsp).TxnPool, nil
}

//get txn stats according to hash
func GetTransactionStats(hash common.Uint256) ([]uint64, error) {
	future := txnPoolPid.RequestFuture(&tc.GetTxnStats{}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("net_server GetTransactionStats ERROR: "), err)
		return nil, err
	}
	return result.(tc.GetTxnStatsRsp).Count, nil
}
