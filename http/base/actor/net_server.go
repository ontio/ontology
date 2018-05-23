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
	"errors"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	ac "github.com/ontio/ontology/p2pserver/actor/server"
	"github.com/ontio/ontology/p2pserver/common"
)

var netServerPid *actor.PID

func SetNetServerPID(actr *actor.PID) {
	netServerPid = actr
}

func Xmit(msg interface{}) error {
	if netServerPid == nil {
		return nil
	}
	netServerPid.Tell(msg)
	return nil
}

func GetConnectionCnt() (uint32, error) {
	if netServerPid == nil {
		return 1, nil
	}
	future := netServerPid.RequestFuture(&ac.GetConnectionCntReq{}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return 0, err
	}
	r, ok := result.(*ac.GetConnectionCntRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.Cnt, nil
}

func GetNeighborAddrs() []common.PeerAddr {
	if netServerPid == nil {
		return []common.PeerAddr{}
	}
	future := netServerPid.RequestFuture(&ac.GetNeighborAddrsReq{}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return nil
	}
	r, ok := result.(*ac.GetNeighborAddrsRsp)
	if !ok {
		return nil
	}
	return r.Addrs
}

func GetConnectionState() (uint32, error) {
	if netServerPid == nil {
		return 0, nil
	}
	future := netServerPid.RequestFuture(&ac.GetConnectionStateReq{}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return 0, err
	}
	r, ok := result.(*ac.GetConnectionStateRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.State, nil
}

func GetNodeTime() (int64, error) {
	if netServerPid == nil {
		return 0, nil
	}
	future := netServerPid.RequestFuture(&ac.GetTimeReq{}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return 0, err
	}
	r, ok := result.(*ac.GetTimeRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.Time, nil
}

func GetNodePort() (uint16, error) {
	if netServerPid == nil {
		return 0, nil
	}
	future := netServerPid.RequestFuture(&ac.GetPortReq{}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return 0, err
	}
	r, ok := result.(*ac.GetPortRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.SyncPort, nil
}

func GetID() (uint64, error) {
	if netServerPid == nil {
		return 0, nil
	}
	future := netServerPid.RequestFuture(&ac.GetIdReq{}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return 0, err
	}
	r, ok := result.(*ac.GetIdRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.Id, nil
}

func GetRelayState() (bool, error) {
	if netServerPid == nil {
		return false, nil
	}
	future := netServerPid.RequestFuture(&ac.GetRelayStateReq{}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return false, err
	}
	r, ok := result.(*ac.GetRelayStateRsp)
	if !ok {
		return false, errors.New("fail")
	}
	return r.Relay, nil
}

func GetVersion() (uint32, error) {
	if netServerPid == nil {
		return 0, nil
	}
	future := netServerPid.RequestFuture(&ac.GetVersionReq{}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return 0, err
	}
	r, ok := result.(*ac.GetVersionRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.Version, nil
}

func GetNodeType() (uint64, error) {
	if netServerPid == nil {
		return 0, nil
	}
	future := netServerPid.RequestFuture(&ac.GetNodeTypeReq{}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return 0, err
	}
	r, ok := result.(*ac.GetNodeTypeRsp)
	if !ok {
		return 0, errors.New("fail")
	}
	return r.NodeType, nil
}
