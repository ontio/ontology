package actor

import (
	"time"
	"github.com/Ontology/eventbus/actor"
	. "github.com/Ontology/p2pserver/protocol"
	ac "github.com/Ontology/p2pserver/actor"
	"errors"
	"github.com/Ontology/common/log"
)

var netServerPid *actor.PID

func SetNetServerPid(actr *actor.PID) {
	netServerPid = actr
}

func Xmit(msg interface{}) error {
	netServerPid.Tell(msg)
	return nil
}

func GetConnectionCnt() (uint, error) {
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

func GetNeighborAddrs() ([]NodeAddr, uint64) {
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
	future := netServerPid.RequestFuture(&ac.GetNodePortReq{}, ReqTimeout*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ErrActorComm, err)
		return 0, err
	}
	r, ok := result.(*ac.GetNodePortRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.Port, nil
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
