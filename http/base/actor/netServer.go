package actor

import (
	"errors"
	"github.com/Ontology/common/log"
	"github.com/Ontology/eventbus/actor"
	ac "github.com/Ontology/p2pserver/actor/server"
	msg "github.com/Ontology/p2pserver/message"
	"time"
)

var netServerPid *actor.PID

func SetP2pPid(actr *actor.PID) {
	netServerPid = actr
}

func Xmit(msg interface{}) error {
	netServerPid.Tell(msg)
	return nil
}

func GetConnectionCnt() (uint32, error) {
	future := netServerPid.RequestFuture(&ac.GetConnectionCntReq{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return 0, err
	}
	r, ok := result.(*ac.GetConnectionCntRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.Cnt, nil
}

func GetNeighborAddrs() ([]msg.PeerAddr, uint64) {
	future := netServerPid.RequestFuture(&ac.GetNeighborAddrsReq{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return nil, 0
	}
	r, ok := result.(*ac.GetNeighborAddrsRsp)
	if !ok {
		return nil, 0
	}
	return r.Addrs, r.Count
}

func GetConnectionState() (uint32, error) {
	future := netServerPid.RequestFuture(&ac.GetConnectionStateReq{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return 0, err
	}
	r, ok := result.(*ac.GetConnectionStateRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.State, nil
}

func GetNodeTime() (int64, error) {
	future := netServerPid.RequestFuture(&ac.GetTimeReq{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return 0, err
	}
	r, ok := result.(*ac.GetTimeRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.Time, nil
}

func GetNodePort() (uint16, error) {
	future := netServerPid.RequestFuture(&ac.GetSyncPortReq{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return 0, err
	}
	r, ok := result.(*ac.GetSyncPortRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.SyncPort, nil
}

func GetID() (uint64, error) {
	future := netServerPid.RequestFuture(&ac.GetIdReq{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return 0, err
	}
	r, ok := result.(*ac.GetIdRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.Id, nil
}

func GetRelayState() (bool, error) {
	future := netServerPid.RequestFuture(&ac.GetRelayStateReq{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return false, err
	}
	r, ok := result.(*ac.GetRelayStateRsp)
	if !ok {
		return false, errors.New("fail")
	}
	return r.Relay, nil
}

func GetVersion() (uint32, error) {
	future := netServerPid.RequestFuture(&ac.GetVersionReq{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return 0, err
	}
	r, ok := result.(*ac.GetVersionRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.Version, nil
}

func GetNodeType() (uint64, error) {
	future := netServerPid.RequestFuture(&ac.GetNodeTypeReq{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return 0, err
	}
	r, ok := result.(*ac.GetNodeTypeRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.NodeType, nil
}
