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
var shardId common.ShardID

func SetTxnPoolPid(txnPid *actor.PID) {
	txnPoolPid = txnPid
}

func SetShardId(id common.ShardID) {
	shardId = id
}

//add txn to txnpool
func AddTransaction(transaction *types.Transaction) {
	if txnPoolPid == nil {
		log.Error("[p2p]net_server AddTransaction(): txnpool pid is nil")
		return
	}
	if transaction.ShardID != shardId.ToUint64() {
		log.Error("[p2p]net_server AddTransaction(): shardId unmatch")
		return
	}
	txReq := &tc.TxReq{
		Tx:         transaction,
		Sender:     tc.NetSender,
		TxResultCh: nil,
	}
	txnPoolPid.Tell(txReq)
}

//get txn according to hash
func GetTransaction(hash common.Uint256) (*types.Transaction, error) {
	if txnPoolPid == nil {
		log.Warn("[p2p]net_server tx pool pid is nil")
		return nil, errors.NewErr("[p2p]net_server tx pool pid is nil")
	}
	future := txnPoolPid.RequestFuture(&tc.GetTxnReq{Hash: hash}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		log.Warnf("[p2p]net_server GetTransaction error: %v\n", err)
		return nil, err
	}
	return result.(tc.GetTxnRsp).Txn, nil
}
