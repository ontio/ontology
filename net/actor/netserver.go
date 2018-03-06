package actor

import (
	"github.com/ONTID/eventbus/actor"
	"github.com/Ontology/net/protocol"
	"github.com/Ontology/common/log"
)

var NetServerPid *actor.PID

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

type GetSynPortReq struct {
}
type GetSynPortRsp struct {
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

func (state *NetServer) Receive(context actor.Context) {
	switch context.Message().(type) {
	case *GetNodeVersionReq:
		version := node.Version()
		context.Sender().Request(&GetNodeVersionRsp{Version: version}, context.Self())
	case *GetConnectionCntReq:
		connectionCnt := node.GetConnectionCnt()
		context.Sender().Request(&GetConnectionCntRsp{Cnt: connectionCnt}, context.Self())
	case *GetSynPortReq:
		synPort := node.GetPort()
		context.Sender().Request(&GetSynPortRsp{Port: synPort}, context.Self())
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
	default:
		err := node.Xmit(context.Message())
		if nil != err {
			log.Error("Error Xmit message ", err.Error())
		}
	}
}

func init() {
	props := actor.FromProducer(func() actor.Actor { return &NetServer{} })
	NetServerPid = actor.Spawn(props)
}

func SetNode(netNode protocol.Noder){
	node = netNode
}
