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

package actor

import (
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	"github.com/Ontology/errors"
	. "github.com/Ontology/txnpool/common"
	"github.com/ontio/ontology-eventbus/actor"
	"time"
)

var txnPoolPid *actor.PID

func SetTxnPoolPid(txnPid *actor.PID) {
	txnPoolPid = txnPid
}

func AddTransaction(transaction *types.Transaction) {
	txnPoolPid.Tell(transaction)
}

func GetTxnPool(byCount bool) ([]*TXEntry, error) {
	future := txnPoolPid.RequestFuture(&GetTxnPoolReq{ByCount: byCount}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetTxnPoolRsp).TxnPool, nil
}

func GetTransaction(hash common.Uint256) (*types.Transaction, error) {
	future := txnPoolPid.RequestFuture(&GetTxnReq{Hash: hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetTxnRsp).Txn, nil
}

func CheckTransaction(hash common.Uint256) (bool, error) {
	future := txnPoolPid.RequestFuture(&CheckTxnReq{Hash: hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return false, err
	}
	return result.(CheckTxnRsp).Ok, nil
}

func GetTransactionStatus(hash common.Uint256) ([]*TXAttr, error) {
	future := txnPoolPid.RequestFuture(&GetTxnStatusReq{Hash: hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetTxnStatusRsp).TxStatus, nil
}

func GetPendingTxn(byCount bool) ([]*types.Transaction, error) {
	future := txnPoolPid.RequestFuture(&GetPendingTxnReq{ByCount: byCount}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetPendingTxnRsp).Txs, nil
}

func VerifyBlock(height uint32, txs []*types.Transaction) ([]*VerifyTxResult, error) {
	future := txnPoolPid.RequestFuture(&VerifyBlockReq{Height: height, Txs: txs}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(VerifyBlockRsp).TxnPool, nil
}

func GetTransactionStats(hash common.Uint256) (*[]uint64, error) {
	future := txnPoolPid.RequestFuture(&GetTxnStats{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetTxnStatsRsp).Count, nil
}
