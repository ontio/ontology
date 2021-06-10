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

// Package actor privides communication with other actor
package actor

import (
	"errors"
	"fmt"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	ontErrors "github.com/ontio/ontology/errors"
	tcomn "github.com/ontio/ontology/txnpool/common"
)

var txnPoolPid *actor.PID
var DisableSyncVerifyTx = false
var txPoolService tcomn.TxPoolService

func SetTxPoolService(pool tcomn.TxPoolService) {
	txPoolService = pool
}

func SetTxnPoolPid(actr *actor.PID) {
	txnPoolPid = actr
}

//append transaction to pool to txpool actor
func AppendTxToPool(txn *types.Transaction) (ontErrors.ErrCode, string) {
	if DisableSyncVerifyTx {
		txPoolService.AppendTransactionAsync(tcomn.HttpSender, txn)
		return ontErrors.ErrNoError, ""
	}
	//add Pre Execute Contract
	_, err := PreExecuteContract(txn)
	if err != nil {
		return ontErrors.ErrUnknown, err.Error()
	}
	msg := txPoolService.AppendTransaction(tcomn.HttpSender, txn)
	fmt.Println("===AppendTxToPool")
	return msg.Err, msg.Desc
}

//GetTxsFromPool from txpool actor
func GetTxsFromPool(byCount bool) map[common.Uint256]*types.Transaction {
	future := txnPoolPid.RequestFuture(&tcomn.GetTxnPoolReq{ByCount: byCount}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return nil
	}
	txpool, ok := result.(*tcomn.GetTxnPoolRsp)
	if !ok {
		return nil
	}
	txMap := make(map[common.Uint256]*types.Transaction)
	for _, v := range txpool.TxnPool {
		txMap[v.Tx.Hash()] = v.Tx
	}
	return txMap
}

//GetTxFromPool from txpool actor
func GetTxFromPool(hash common.Uint256) (tcomn.TXEntry, error) {
	txn := txPoolService.GetTransaction(hash)
	if txn == nil {
		return tcomn.TXEntry{}, errors.New("fail")
	}

	status := txPoolService.GetTransactionStatus(hash)
	return tcomn.TXEntry{Tx: txn, Attrs: status.Attrs}, nil
}

//GetTxnCount from txpool actor
func GetTxnCount() []uint32 {
	return txPoolService.GetTxAmount()
}

//GetTxnHashList from txpool actor
func GetTxnHashList() []common.Uint256 {
	return txPoolService.GetTxList()
}
