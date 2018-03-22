package server

import (
	"github.com/Ontology/common/log"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/p2pserver"
	"github.com/Ontology/p2pserver/protocol"
	"reflect"
)

var netServerPid *actor.PID

var server p2pserver.P2pServer

type NetServer struct{}

type StartServerReq struct {
	StartSync bool
}
type StartServerRsp struct {
	Error error
}

type StopServerReq struct {
}
type StopServerRsp struct {
	Error error
}

type GetVersionReq struct {
}
type GetVersionRsp struct {
	Version uint32
}

type GetConnectionCntReq struct {
}
type GetConnectionCntRsp struct {
	Cnt uint
}

type GetIdReq struct {
}
type GetIdRsp struct {
	Id uint64
}

type GetSyncPortReq struct {
}
type GetSyncPortRsp struct {
	SyncPort uint16
}

type GetConsPortReq struct {
}
type GetConsPortRsp struct {
	ConsPort uint16
}

type GetPortReq struct {
}
type GetPortRsp struct {
	SyncPort uint16
	ConsPort uint16
}

type GetConnectionStateReq struct {
}
type GetConnectionStateRsp struct {
	State uint32
}

type GetTimeReq struct {
}
type GetTimeRsp struct {
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
	Addrs []protocol.PeerAddr
	Count uint64
}

type IsSyncingReq struct {
}
type IsSyncingRsp struct {
	IsSyncing bool
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
	case *StartServerReq:
		err := StartServer(context.Message().(StartServerReq).StartSync)
		context.Sender().Request(&StartServerRsp{Error: err}, context.Self())
	case *StopServerReq:
		err := StopServer()
		context.Sender().Request(&StopServerRsp{Error: err}, context.Self())
	case *IsSyncingReq:
		isSyncing := IsSyncing()
		context.Sender().Request(&IsSyncingRsp{IsSyncing: isSyncing}, context.Self())
	case *GetPortReq:
		syncPort, consPort := GetPort()
		context.Sender().Request(&GetPortRsp{SyncPort: syncPort, ConsPort: consPort}, context.Self())
	case *GetVersionReq:
		version := server.GetVersion()
		context.Sender().Request(&GetVersionRsp{Version: version}, context.Self())
	case *GetConnectionCntReq:
		connectionCnt := server.GetConnectionCnt()
		context.Sender().Request(&GetConnectionCntRsp{Cnt: uint(connectionCnt)}, context.Self())
	case *GetSyncPortReq:
		syncPort, _ := GetPort()
		context.Sender().Request(&GetSyncPortRsp{SyncPort: syncPort}, context.Self())
	case *GetConsPortReq:
		_, conPort := GetPort()
		context.Sender().Request(&GetConsPortRsp{ConsPort: conPort}, context.Self())
	case *GetIdReq:
		id := server.GetId()
		context.Sender().Request(&GetIdRsp{Id: id}, context.Self())
	case *GetConnectionStateReq:
		state := server.GetState()
		context.Sender().Request(&GetConnectionStateRsp{State: state}, context.Self())
	case *GetTimeReq:
		time := server.GetTime()
		context.Sender().Request(&GetTimeRsp{Time: time}, context.Self())
	case *GetNodeTypeReq: //this function will be deleted
		nodeType := server.Services()
		context.Sender().Request(&GetNodeTypeRsp{NodeType: nodeType}, context.Self())
	// case *GetRelayStateReq: //this function will be deleted
	// 	relay := act.server.GetRelay()
	// 	context.Sender().Request(&GetRelayStateRsp{Relay: relay}, context.Self())
	case *GetNeighborAddrsReq:
		addrs, count := server.GetNeighborAddrs()
		context.Sender().Request(&GetNeighborAddrsRsp{Addrs: addrs, Count: count}, context.Self())
	default:
		err := server.Xmit(context.Message())
		if nil != err {
			log.Error("Error Xmit message ", err.Error(), reflect.TypeOf(context.Message()))
		}
	}
}

func InitNetServer(p2p p2pserver.P2pServer) (*actor.PID, error) {
	props := actor.FromProducer(func() actor.Actor { return &NetServer{} })
	netServerPid, err := actor.SpawnNamed(props, "net_server")
	if err != nil {
		return nil, err
	}
	server = p2p
	return netServerPid, err
}

func StartServer(startSync bool) error {
	var err error
	if startSync {
		//TODO

		err = nil
	} else {
		//TODO

		err = nil
	}
	return err
}

func StopServer() error {
	var err error
	err = nil
	//TODO

	return err
}

func IsSyncing() bool {
	var isSyncing bool
	//TODO

	return isSyncing
}

func GetPort() (uint16, uint16) {
	return 0, 0
}
