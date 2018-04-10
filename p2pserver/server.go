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

package p2pserver

import (
	types "github.com/Ontology/p2pserver/common"
	P2Pnet "github.com/Ontology/p2pserver/net"
	"github.com/Ontology/p2pserver/peer"
)

//P2P represent the net interface of p2p package
type P2P interface {
	Start()
	Halt()
	Connect(addr string, isConsensus bool) error
	GetVersion() uint32
	GetPort() uint16
	GetConsensusPort() uint16
	GetId() uint64
	GetTime() int64
	GetState() uint32
	GetServices() uint64
	GetNeighborAddrs() ([]types.PeerAddr, uint64)
	GetConnectionCnt() uint32
	IsPeerEstablished(p *peer.Peer) bool
	Send(p *peer.Peer, data []byte, isConsensus bool) error
	GetMsgChan(isConsensus bool) chan types.MsgPayload
	GetPeerFromAddr(addr string) *peer.Peer
}

//NewNetServer return the net object in p2p
func NewNetServer(p *peer.Peer) P2P {

	n := &P2Pnet.NetServer{
		Self:            p,
		PeerSyncAddress: make(map[string]*peer.Peer),
		PeerConsAddress: make(map[string]*peer.Peer),
		SyncChan:        make(chan types.MsgPayload, types.CHAN_CAPABILITY),
		ConsChan:        make(chan types.MsgPayload, types.CHAN_CAPABILITY),
	}

	p.AttachSyncChan(n.SyncChan)
	p.AttachConsChan(n.ConsChan)
	return n
}
