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

package txnpool

import (
	"reflect"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	tx "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	tc "github.com/ontio/ontology/txnpool/common"
	"github.com/ontio/ontology/txnpool/proc"
)

// TxnActor: Handle the low priority msg from P2P and API
type TxActor struct {
	poolMgr *TxnPoolManager
}

// NewTxActor creates an actor to handle the transaction-based messages from
// network and http
func NewTxActor(s *TxnPoolManager) *TxActor {
	a := &TxActor{}
	a.setServer(s)
	return a
}

// handleTransaction handles a transaction from network and http
func (ta *TxActor) handleTransaction(sender tc.SenderType, self *actor.PID,
	txn *tx.Transaction, txResultCh chan *tc.TxResult) {
	shardID, err := tx.NewShardID(txn.ShardID)
	if err != nil {
		if (sender == tc.HttpSender || sender == tc.ShardSender) && txResultCh != nil {
			proc.ReplyTxResult(txResultCh, txn.Hash(), errors.ErrUnknown, "invalid shardID in tx")
		}
		return
	}
	server := ta.poolMgr.GetTxnPoolServer(shardID)
	if server == nil {
		if (sender == tc.HttpSender || sender == tc.ShardSender) && txResultCh != nil {
			proc.ReplyTxResult(txResultCh, txn.Hash(), errors.ErrUnknown, "txn processor not started")
		}
		return
	}

	errCode, errDesc := server.HandleTransaction(sender, txn, txResultCh)
	if errCode != 0 && (sender == tc.HttpSender || sender == tc.ShardSender) && txResultCh != nil {
		proc.ReplyTxResult(txResultCh, txn.Hash(), errCode, errDesc)
	}
}

// Receive implements the actor interface
func (ta *TxActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("txpool-tx actor started and be ready to receive tx msg")

	case *actor.Stopping:
		log.Warn("txpool-tx actor stopping")

	case *actor.Restarting:
		log.Warn("txpool-tx actor restarting")

	case *tc.TxReq:
		sender := msg.Sender

		log.Debugf("txpool-tx actor receives tx from %v ", sender.Sender())

		ta.handleTransaction(sender, context.Self(), msg.Tx, msg.TxResultCh)

	case *tc.GetTxnReq:
		sender := context.Sender()

		log.Debugf("txpool-tx actor receives getting tx req from %v", sender)
		var res *tx.Transaction
		if server := ta.poolMgr.GetTxnPoolServer(ta.poolMgr.ShardID); server != nil {
			res = server.GetTransaction(msg.Hash)
		}
		if sender != nil {
			sender.Request(&tc.GetTxnRsp{Txn: res},
				context.Self())
		}

	case *tc.GetTxnStats:
		sender := context.Sender()

		log.Debugf("txpool-tx actor receives getting tx stats from %v", sender)
		var res []uint64
		if server := ta.poolMgr.GetTxnPoolServer(ta.poolMgr.ShardID); server != nil {
			res = server.GetStats()
		}
		if sender != nil {
			sender.Request(&tc.GetTxnStatsRsp{Count: res},
				context.Self())
		}

	case *tc.CheckTxnReq:
		sender := context.Sender()

		log.Debugf("txpool-tx actor receives checking tx req from %v", sender)
		var res bool
		if server := ta.poolMgr.GetTxnPoolServer(ta.poolMgr.ShardID); server != nil {
			res = server.CheckTx(msg.Hash)
		}
		if sender != nil {
			sender.Request(&tc.CheckTxnRsp{Ok: res},
				context.Self())
		}

	case *tc.GetTxnStatusReq:
		sender := context.Sender()

		log.Debugf("txpool-tx actor receives getting tx status req from %v", sender)
		var res *tc.TxStatus
		if server := ta.poolMgr.GetTxnPoolServer(ta.poolMgr.ShardID); server != nil {
			res = server.GetTxStatusReq(msg.Hash)
		}
		if sender != nil {
			if res == nil {
				sender.Request(&tc.GetTxnStatusRsp{Hash: msg.Hash,
					TxStatus: nil}, context.Self())
			} else {
				sender.Request(&tc.GetTxnStatusRsp{Hash: res.Hash,
					TxStatus: res.Attrs}, context.Self())
			}
		}

	case *tc.GetTxnCountReq:
		sender := context.Sender()

		log.Debugf("txpool-tx actor receives getting tx count req from %v", sender)
		var res []uint32
		if server := ta.poolMgr.GetTxnPoolServer(ta.poolMgr.ShardID); server != nil {
			res = server.GetTxCount()
		}
		if sender != nil {
			sender.Request(&tc.GetTxnCountRsp{Count: res}, context.Self())
		}

	default:
		log.Debugf("txpool-tx actor: unknown msg %v type %v", msg, reflect.TypeOf(msg))
	}
}

func (ta *TxActor) setServer(s *TxnPoolManager) {
	ta.poolMgr = s
}
