package actor

import (
	"time"
	"github.com/Ontology/eventbus/actor"
	. "github.com/Ontology/net/protocol"
	ac "github.com/Ontology/net/actor"
	"errors"
)

var netServerPid *actor.PID

func SetNetServerActor(actr *actor.PID) {
	netServerPid = actr
}

func Xmit(msg interface{}) error {
	future := netServerPid.RequestFuture(msg, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return errors.New("fail")
	}
	return nil
}

func GetConnectionCnt() (uint, error) {
	future := netServerPid.RequestFuture(&ac.GetConnectionCntReq{}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return 0, err
	}
	r, ok := result.(*ac.GetConnectionCntRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.Cnt, nil
}

func GetNeighborAddrs() ([]NodeAddr, uint64) {
	future := netServerPid.RequestFuture(nil, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return nil, 0
	}
	return nil, 0
}

func GetConnectionState() (uint32, error) {
	future := netServerPid.RequestFuture(&ac.GetConnectionStateReq{}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return 0, err
	}
	r, ok := result.(*ac.GetConnectionStateRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.State, nil
}

func GetNodeTime() (int64, error) {
	future := netServerPid.RequestFuture(&ac.GetNodeTimeReq{}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return 0, err
	}
	r, ok := result.(*ac.GetNodeTimeRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.Time, nil
}

func GetNodePort() (uint16, error) {
	future := netServerPid.RequestFuture(&ac.GetNodePortReq{}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return 0, err
	}
	r, ok := result.(*ac.GetNodePortRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.Port, nil
}

func GetID() (uint64, error) {
	future := netServerPid.RequestFuture(&ac.GetNodeIdReq{}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return 0, err
	}
	r, ok := result.(*ac.GetNodeIdRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.Id, nil
}

func GetRelayState() (bool, error) {
	future := netServerPid.RequestFuture(&ac.GetRelayStateReq{}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return false, err
	}
	r, ok := result.(*ac.GetRelayStateRsp)
	if !ok {
		return false, errors.New("fail")
	}
	return r.Relay, nil
}

func GetNodeVersion() (uint32, error) {
	future := netServerPid.RequestFuture(&ac.GetNodeVersionReq{}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return 0, err
	}
	r, ok := result.(*ac.GetNodeVersionRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.Version, nil
}

func GetNodeType() (uint64, error) {
	future := netServerPid.RequestFuture(&ac.GetNodeTypeReq{}, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return 0, err
	}
	r, ok := result.(*ac.GetNodeTypeRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.NodeType, nil
}


