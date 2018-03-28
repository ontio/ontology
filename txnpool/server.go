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
	"github.com/Ontology/common/log"
	"github.com/Ontology/events"
	"github.com/Ontology/events/message"
	tc "github.com/Ontology/txnpool/common"
	tp "github.com/Ontology/txnpool/proc"
	"github.com/ontio/ontology-eventbus/actor"
)

func startActor(obj interface{}, id string) *actor.PID {
	props := actor.FromProducer(func() actor.Actor {
		return obj.(actor.Actor)
	})

	pid, _ := actor.SpawnNamed(props, id)
	if pid == nil {
		log.Error("Fail to start actor")
		return nil
	}
	return pid
}

func StartTxnPoolServer() *tp.TXPoolServer {
	var s *tp.TXPoolServer

	/* Start txnpool server to receive msgs from p2p,
	 * consensus and valdiators
	 */
	s = tp.NewTxPoolServer(tc.MAXWORKERNUM)

	// Initialize an actor to handle the msgs from valdiators
	rspActor := tp.NewVerifyRspActor(s)
	rspPid := startActor(rspActor, "txVerifyRsp")
	if rspPid == nil {
		log.Error("Fail to start verify rsp actor")
		return nil
	}
	s.RegisterActor(tc.VerifyRspActor, rspPid)

	// Initialize an actor to handle the msgs from consensus
	tpa := tp.NewTxPoolActor(s)
	txPoolPid := startActor(tpa, "txPool")
	if txPoolPid == nil {
		log.Error("Fail to start txnpool actor")
		return nil
	}
	s.RegisterActor(tc.TxPoolActor, txPoolPid)

	// Initialize an actor to handle the msgs from p2p and api
	ta := tp.NewTxActor(s)
	txPid := startActor(ta, "tx")
	if txPid == nil {
		log.Error("Fail to start txn actor")
		return nil
	}
	s.RegisterActor(tc.TxActor, txPid)

	// Subscribe the block complete event
	var sub = events.NewActorSubscriber(txPoolPid)
	sub.Subscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)
	return s
}
