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
	"github.com/ontio/ontology/p2pserver/common"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
)

var netServer p2p.P2P

func SetNetServer(p2p p2p.P2P) {
	netServer = p2p
}

//GetConnectionCnt from netSever actor
func GetConnectionCnt() uint32 {
	if netServer == nil {
		return 1
	}

	return netServer.GetConnectionCnt()
}

//GetMaxPeerBlockHeight from netSever actor
func GetMaxPeerBlockHeight() uint64 {
	if netServer == nil {
		return 1
	}
	return netServer.GetMaxPeerBlockHeight()
}

//GetNeighborAddrs from netSever actor
func GetNeighborAddrs() []common.PeerAddr {
	if netServer == nil {
		return []common.PeerAddr{}
	}
	return netServer.GetNeighborAddrs()
}

//GetNodePort from netSever actor
func GetNodePort() uint16 {
	if netServer == nil {
		return 0
	}
	return netServer.GetHostInfo().Port
}

//GetID from netSever actor
func GetID() common.PeerId {
	if netServer == nil {
		return common.PeerId{}
	}
	return netServer.GetID()
}

//GetRelayState from netSever actor
func GetRelayState() bool {
	if netServer == nil {
		return false
	}
	return netServer.GetHostInfo().Relay
}

//GetVersion from netSever actor
func GetVersion() uint32 {
	if netServer == nil {
		return 0
	}
	return netServer.GetHostInfo().Version
}

//GetNodeType from netSever actor
func GetNodeType() uint64 {
	if netServer == nil {
		return 0
	}
	return netServer.GetHostInfo().Services
}
