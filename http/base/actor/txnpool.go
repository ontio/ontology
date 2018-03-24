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
	"errors"
	. "github.com/Ontology/common"
	"github.com/Ontology/core/types"
	. "github.com/Ontology/errors"
	"github.com/ontio/ontology-eventbus/actor"
	tc "github.com/Ontology/txnpool/common"
	"time"
	"github.com/Ontology/common/log"
)

var txnPid *actor.PID
var txnPoolPid *actor.PID

func SetTxPid(actr *actor.PID) {
	txnPid = actr
}
func SetTxnPoolPid(actr *actor.PID) {
	txnPoolPid = actr
}
func AppendTxToPool(txn *types.Transaction) ErrCode {
	future := txnPid.RequestFuture(txn, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return ErrUnknown
	}
	if result, ok := result.(*tc.TxRsp); !ok {
		return ErrUnknown
	} else if result.Hash != txn.Hash() {
		return ErrUnknown
	} else {
		return result.ErrCode
	}
}

func GetTxsFromPool(byCount bool) (map[Uint256]*types.Transaction, Fixed64) {
	future := txnPoolPid.RequestFuture(&tc.GetTxnPoolReq{ByCount: byCount}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, 0
	}
	txpool, ok := result.(*tc.GetTxnPoolRsp)
	if !ok {
		return nil, 0
	}
	txMap := make(map[Uint256]*types.Transaction)
	var networkFeeSum Fixed64
	for _, v := range txpool.TxnPool {
		txMap[v.Tx.Hash()] = v.Tx
		networkFeeSum += v.Fee
	}
	return txMap, networkFeeSum

}

func GetTxFromPool(hash Uint256) (tc.TXEntry, error) {

	future := txnPid.RequestFuture(&tc.GetTxnReq{hash}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return tc.TXEntry{}, err
	}
	txn, ok := result.(*tc.GetTxnRsp)
	if !ok {
		return tc.TXEntry{}, errors.New("fail")
	}
	if txn == nil {
		return tc.TXEntry{}, errors.New("fail")
	}

	future = txnPid.RequestFuture(&tc.GetTxnStatusReq{hash}, ReqTimeout*time.Second)
	result, err = future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return tc.TXEntry{}, err
	}
	txStatus, ok := result.(*tc.GetTxnStatusRsp)
	if !ok {
		return tc.TXEntry{}, errors.New("fail")
	}
	txnEntry := tc.TXEntry{txn.Txn, 0, txStatus.TxStatus}
	return txnEntry, nil
}

func GetTxnCnt() ([]uint64, error) {
	future := txnPid.RequestFuture(&tc.GetTxnStats{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return []uint64{}, err
	}
	txnCnt, ok := result.(*tc.GetTxnStatsRsp)
	if !ok {
		return []uint64{}, errors.New("fail")
	}
	return txnCnt.Count, nil
}
