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

package server

import (
	"reflect"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver"
	"github.com/ontio/ontology/p2pserver/common"
)

type P2PActor struct {
	props  *actor.Props
	server *p2pserver.P2PServer
}

// NewP2PActor creates an actor to handle the messages from
// txnpool and consensus
func NewP2PActor(p2pServer *p2pserver.P2PServer) *P2PActor {
	return &P2PActor{
		server: p2pServer,
	}
}

//start a actor called net_server
func (this *P2PActor) Start() (*actor.PID, error) {
	this.props = actor.FromProducer(func() actor.Actor { return this })
	p2pPid, err := actor.SpawnNamed(this.props, "net_server")
	return p2pPid, err
}

//message handler
func (this *P2PActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Restarting:
		log.Warn("[p2p]actor restarting")
	case *actor.Stopping:
		log.Warn("[p2p]actor stopping")
	case *actor.Stopped:
		log.Warn("[p2p]actor stopped")
	case *actor.Started:
		log.Debug("[p2p]actor started")
	case *actor.Restart:
		log.Warn("[p2p]actor restart")
	case *TransmitConsensusMsgReq:
		this.handleTransmitConsensusMsgReq(ctx, msg)
	case *common.AppendPeerID:
		this.server.OnAddNode(msg.ID)
	case *common.RemovePeerID:
		this.server.OnDelNode(msg.ID)
	case *common.AppendHeaders:
		this.server.OnHeaderReceive(msg.FromID, msg.Headers)
	case *common.AppendBlock:
		this.server.OnBlockReceive(msg.FromID, msg.BlockSize, msg.Block, msg.CCMsg, msg.MerkleRoot)
	default:
		err := this.server.Xmit(ctx.Message())
		if nil != err {
			log.Warn("[p2p]error xmit message ", err.Error(), reflect.TypeOf(ctx.Message()))
		}
	}
}

func (this *P2PActor) handleTransmitConsensusMsgReq(ctx actor.Context, req *TransmitConsensusMsgReq) {
	peer := this.server.GetNetWork().GetPeer(req.Target)
	if peer != nil {
		this.server.Send(peer, req.Msg)
	} else {
		log.Warnf("[p2p]can`t transmit consensus msg:no valid neighbor peer: %d\n", req.Target)
	}
}
