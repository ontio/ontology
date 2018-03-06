package actor

import (
	"time"
	"github.com/Ontology/eventbus/actor"
	. "github.com/Ontology/net/protocol"
	"errors"
)

var p2pPid *actor.PID

func SetP2pActor(actr *actor.PID) {
	p2pPid = actr
}

func Xmit(msg interface{}) error {
	future := p2pPid.RequestFuture(msg, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return errors.New("fail")
	}
	return nil
}

func GetConnectionCnt() (uint, error) {
	future := p2pPid.RequestFuture(nil, 10*time.Second)
	result, err := future.Result()
	if err != nil {
		return 0, err
	}
	r, ok := result.(uint)
	if ok {
		return 0, errors.New("")
	}
	return r, nil
}

func GetNeighborAddrs() ([]NodeAddr, uint64) {
	future := p2pPid.RequestFuture(nil, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return nil, 0
	}
	return nil, 0
}

func GetState() (uint32, error) {
	future := p2pPid.RequestFuture(nil, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return 0, nil
	}
	return 0, nil
}

func GetTime() (int64, error) {
	future := p2pPid.RequestFuture(nil, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return 0, nil
	}
	return 0, nil
}

func GetPort() (uint16, error) {
	future := p2pPid.RequestFuture(nil, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return 0, nil
	}
	return 0, nil
}

func GetID() (uint64, error) {
	future := p2pPid.RequestFuture(nil, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return 0, nil
	}
	return 0, nil
}

func GetVersion() (uint32, error) {
	future := p2pPid.RequestFuture(nil, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return 0, nil
	}
	return 0, nil
}

func Services() (uint64, error) {
	future := p2pPid.RequestFuture(nil, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return 0, nil
	}
	return 0, nil
}

func GetRelay() (bool, error) {
	future := p2pPid.RequestFuture(nil, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return false, nil
	}
	return false, nil
}

func GetHeight() (uint64, error) {
	future := p2pPid.RequestFuture(nil, 10*time.Second)
	_, err := future.Result()
	if err != nil {
		return 0, nil
	}
	return 0, nil
}

