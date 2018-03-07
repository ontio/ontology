package actor

import (
	"github.com/Ontology/common/log"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/net/protocol"
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

func (state *NetServer) Receive(context actor.Context) {
	switch context.Message().(type) {
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
		conPort := node.GetConsensusPort()
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
		context.Sender().Request(&GetNeighborAddrsRsp{Addrs: addrs, Count:count}, context.Self())
	default:
		err := node.Xmit(context.Message())
		if nil != err {
			log.Error("Error Xmit message ", err.Error())
		}
	}
}

func InitNetServer(netNode protocol.Noder) (*actor.PID, error){
	props := actor.FromProducer(func() actor.Actor { return &NetServer{} })
	netServerPid, err := actor.SpawnNamed(props, "net_server")
	if err != nil {
		return nil, err
	}
	node = netNode
	return netServerPid, err
}
