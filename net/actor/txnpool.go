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
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	ttypes "github.com/ontio/ontology/txnpool/types"
	"github.com/ontio/ontology-eventbus/actor"
)

const txnPoolReqTimeout = 5 * time.Second

var txnPoolPid *actor.PID

func SetTxnPoolPid(txnPid *actor.PID) {
	txnPoolPid = txnPid
}

func AddTransaction(transaction *types.Transaction) {
	txReq := &ttypes.AppendTxReq{
		Tx:     transaction,
		Sender: ttypes.NetSender,
	}
	txnPoolPid.Tell(txReq)
}

func GetTxnPool(byCount bool) ([]*ttypes.TxEntry, error) {
	future := txnPoolPid.RequestFuture(&ttypes.GetTxnPoolReq{ByCount: byCount}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(ttypes.GetTxnPoolRsp).TxnPool, nil
}

func GetTransaction(hash common.Uint256) (*types.Transaction, error) {
	future := txnPoolPid.RequestFuture(&ttypes.GetTxnReq{Hash: hash}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(ttypes.GetTxnRsp).Txn, nil
}

func CheckTransaction(hash common.Uint256) (bool, error) {
	future := txnPoolPid.RequestFuture(&ttypes.CheckTxnReq{Hash: hash}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return false, err
	}
	return result.(ttypes.CheckTxnRsp).Ok, nil
}

func GetTransactionStatus(hash common.Uint256) ([]*ttypes.VerifyResult, error) {
	future := txnPoolPid.RequestFuture(&ttypes.GetTxnStatusReq{Hash: hash}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(ttypes.GetTxnStatusRsp).VerifyResults, nil
}

func GetPendingTxn(byCount bool) ([]*types.Transaction, error) {
	future := txnPoolPid.RequestFuture(&ttypes.GetPendingTxnReq{ByCount: byCount}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(ttypes.GetPendingTxnRsp).Txs, nil
}

func VerifyBlock(height uint32, txs []*types.Transaction) ([]*ttypes.VerifyTxResult, error) {
	future := txnPoolPid.RequestFuture(&ttypes.VerifyBlockReq{Height: height, Txs: txs}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(ttypes.VerifyBlockRsp).TxResults, nil
}

func GetTransactionStats(hash common.Uint256) ([]uint64, error) {
	future := txnPoolPid.RequestFuture(&ttypes.GetTxnStats{}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(ttypes.GetTxnStatsRsp).Count, nil
}
