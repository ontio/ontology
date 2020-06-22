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
package p2p

import (
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/peer"
)

type Context struct {
	sender  *peer.Peer
	net     P2P
	MsgSize uint32
}

func NewContext(sender *peer.Peer, net P2P, msgSize uint32) *Context {
	return &Context{sender, net, msgSize}
}

func (self *Context) Sender() *peer.Peer {
	return self.sender
}

func (self *Context) Network() P2P {
	return self.net
}

type Protocol interface {
	HandlePeerMessage(ctx *Context, msg types.Message)
	HandleSystemMessage(net P2P, msg SystemMessage)
}

type SystemMessage interface {
	systemMessage()
}

type implSystemMessage struct{}

func (self implSystemMessage) systemMessage() {}

type PeerConnected struct {
	Info *peer.PeerInfo
	implSystemMessage
}

type PeerDisConnected struct {
	Info *peer.PeerInfo
	implSystemMessage
}

type NetworkStart struct {
	implSystemMessage
}

type NetworkStop struct {
	implSystemMessage
}

type HostAddrDetected struct {
	implSystemMessage
	ListenAddr string
}
