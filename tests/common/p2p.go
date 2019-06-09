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

package TestCommon

import (
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr"
	"testing"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"reflect"
)

type MockP2PActor struct {
	Name    string
	ShardID common.ShardID
	Pid     *actor.PID
	Peer    *MockPeer
}

func NewP2PActor(t *testing.T, name string, peer *MockPeer) *MockP2PActor {
	if peer.Lgr == nil {
		t.Fatalf("create p2p actor with nil peer.lgr")
	}
	return &MockP2PActor{
		Name:    name,
		ShardID: peer.Lgr.ShardID,
		Peer:    peer,
	}
}

func (this *MockP2PActor) GetPID(t *testing.T) *actor.PID {
	if this.Pid == nil {
		t.Fatalf("get pid of uninitialized p2pactor")
	}
	return this.Pid
}

func (this *MockP2PActor) Start(t *testing.T) {
	props := actor.FromProducer(func() actor.Actor { return this })
	p2pPid, err := actor.SpawnNamed(props, "mock_p2pactor"+chainmgr.GetShardName(this.ShardID)+this.Name)
	if err != nil {
		t.Fatalf("init p2p actor %s failed: %s", this.Name, err)
	}
	this.Pid = p2pPid
}

func (this *MockP2PActor) Receive(ctx actor.Context) {
	msg := ctx.Message()
	switch msg.(type) {
	case *types.ConsensusPayload:
		consensusPayload := msg.(*types.ConsensusPayload)
		msg := msgpack.NewConsensus(consensusPayload)
		err := this.Peer.Send(nil, msg, false)
		if err != nil {
			log.Errorf("err sending msg")
		}
	default:
		log.Errorf("mock p2p xmit msg %v, type %v", msg, reflect.TypeOf(msg))
	}
}
