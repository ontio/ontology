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
	"reflect"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/net/protocol"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology-crypto/keypair"
)

var netServerPid *actor.PID

var node protocol.Noder

type NetServer struct{}

type GetNodeVersionReq struct {
}
type GetNodeVersionRsp struct {
	Version uint32
}

type GetConnectionCntReq struct {
}
type GetConnectionCntRsp struct {
	Cnt uint
}

type GetNodeIdReq struct {
}
type GetNodeIdRsp struct {
	Id uint64
}

type GetNodePortReq struct {
}
type GetNodePortRsp struct {
	Port uint16
}

type GetConsensusPortReq struct {
}
type GetConsensusPortRsp struct {
	Port uint16
}

type GetConnectionStateReq struct {
}
type GetConnectionStateRsp struct {
	State uint32
}

type GetNodeTimeReq struct {
}
type GetNodeTimeRsp struct {
	Time int64
}

type GetNodeTypeReq struct {
}
type GetNodeTypeRsp struct {
	NodeType uint64
}

type GetRelayStateReq struct {
}
type GetRelayStateRsp struct {
	Relay bool
}

type GetNeighborAddrsReq struct {
}
type GetNeighborAddrsRsp struct {
	Addrs []protocol.NodeAddr
	Count uint64
}

type TransmitConsensusMsgReq struct {
	Target keypair.PublicKey
	Msg    []byte
}

func (state *NetServer) Receive(context actor.Context) {
	switch context.Message().(type) {
	case *actor.Restarting:
		log.Warn("p2p actor restarting")
	case *actor.Stopping:
		log.Warn("p2p actor stopping")
	case *actor.Stopped:
		log.Warn("p2p actor stopped")
	case *actor.Started:
		log.Warn("p2p actor started")
	case *actor.Restart:
		log.Warn("p2p actor restart")
	case *GetNodeVersionReq:
		version := node.Version()
		context.Sender().Request(&GetNodeVersionRsp{Version: version}, context.Self())
	case *GetConnectionCntReq:
		connectionCnt := node.GetConnectionCnt()
		context.Sender().Request(&GetConnectionCntRsp{Cnt: connectionCnt}, context.Self())
	case *GetNodePortReq:
		nodePort := node.GetPort()
		context.Sender().Request(&GetNodePortRsp{Port: nodePort}, context.Self())
	case *GetConsensusPortReq:
		conPort := node.GetPort()
		context.Sender().Request(&GetConsensusPortRsp{Port: conPort}, context.Self())
	case *GetNodeIdReq:
		id := node.GetID()
		context.Sender().Request(&GetNodeIdRsp{Id: id}, context.Self())
	case *GetConnectionStateReq:
		state := node.GetState()
		context.Sender().Request(&GetConnectionStateRsp{State: state}, context.Self())
	case *GetNodeTimeReq:
		time := node.GetTime()
		context.Sender().Request(&GetNodeTimeRsp{Time: time}, context.Self())
	case *GetNodeTypeReq:
		nodeType := node.Services()
		context.Sender().Request(&GetNodeTypeRsp{NodeType: nodeType}, context.Self())
	case *GetRelayStateReq:
		relay := node.GetRelay()
		context.Sender().Request(&GetRelayStateRsp{Relay: relay}, context.Self())
	case *GetNeighborAddrsReq:
		addrs, count := node.GetNeighborAddrs()
		context.Sender().Request(&GetNeighborAddrsRsp{Addrs: addrs, Count: count}, context.Self())
	case *TransmitConsensusMsgReq:
		req := context.Message().(*TransmitConsensusMsgReq)
		for _, peer := range node.GetNeighborNoder() {
			if keypair.ComparePublicKey(req.Target, peer.GetPubKey()) {
				peer.Tx(req.Msg)
			}
		}
	default:
		err := node.Xmit(context.Message())
		if nil != err {
			log.Error("Error Xmit message ", err.Error(), reflect.TypeOf(context.Message()))
		}
	}
}

func InitNetServer(netNode protocol.Noder) (*actor.PID, error) {
	props := actor.FromProducer(func() actor.Actor { return &NetServer{} })
	netServerPid, err := actor.SpawnNamed(props, "net_server")
	if err != nil {
		return nil, err
	}
	node = netNode
	return netServerPid, err
}
