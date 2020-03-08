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
package protocols

import (
	"github.com/ontio/ontology-eventbus/actor"
	core "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/kbucket"
	"github.com/ontio/ontology/p2pserver/message/types"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
)

type Context struct {
	sender  *peer.Peer
	net     p2p.P2P
	pid     *actor.PID
	msgSize uint32
}

func NewContext(sender *peer.Peer, net p2p.P2P, pid *actor.PID, msgSize uint32) *Context {
	return &Context{sender, net, pid, msgSize}
}

func (self *Context) Sender() *peer.Peer {
	return self.sender
}

func (self *Context) Network() p2p.P2P {
	return self.net
}

func (self *Context) ReceivedHeaders(sender kbucket.KadId, headers []*core.Header) {
	pid := self.pid
	if pid != nil {
		input := &common.AppendHeaders{
			FromID:  sender.ToUint64(),
			Headers: headers,
		}
		pid.Tell(input)
	}
}

func (self *Context) ReceivedBlock(sender kbucket.KadId, block *types.Block) {
	pid := self.pid
	if pid != nil {
		input := &common.AppendBlock{
			FromID:     sender.ToUint64(),
			BlockSize:  self.msgSize,
			Block:      block.Blk,
			CCMsg:      block.CCMsg,
			MerkleRoot: block.MerkleRoot,
		}
		pid.Tell(input)
	}
}

type Protocol interface {
	PeerConnected(p *peer.PeerInfo)
	PeerDisConnected(p *peer.PeerInfo)
	HandleMessage(ctx *Context, msg types.Message)
}
