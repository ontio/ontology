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
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events/message"
	common2 "github.com/ontio/ontology/txnpool/common"
	"sync"
	"testing"
)

type MockTxPool struct {
	Lock          sync.Mutex
	Name          string
	ShardID       common.ShardID
	Pid           *actor.PID
	PendingTxList []*types.Transaction
}

func NewTxnPool(t *testing.T, name string, shardID common.ShardID) *MockTxPool {
	return &MockTxPool{
		Name:          name,
		ShardID:       shardID,
		PendingTxList: make([]*types.Transaction, 0),
	}
}

func (this *MockTxPool) GetPID(t *testing.T) *actor.PID {
	if this.Pid == nil {
		t.Fatalf("get pid of uninitialized txpool")
	}
	return this.Pid
}

func (this *MockTxPool) Start(t *testing.T) {
	props := actor.FromProducer(func() actor.Actor { return this })
	p2pPid, err := actor.SpawnNamed(props, "mock_txnpool"+chainmgr.GetShardName(this.ShardID)+this.Name)
	if err != nil {
		t.Fatalf("init txnpool actor %s failed: %s", this.Name, err)
	}
	this.Pid = p2pPid
}

func (this *MockTxPool) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case common2.GetPendingTxnReq:
		if sender := ctx.Sender(); sender != nil {
			sender.Request(&common2.GetPendingTxnRsp{Txs: this.GetPeningTxs()}, ctx.Self())
		}

	case message.SaveBlockCompleteMsg:
		if msg.Block != nil {
			this.CleanTx(msg.Block.Transactions)
		}
	}
}

func (this *MockTxPool) GetPeningTxs() []*types.Transaction {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	return this.PendingTxList
}

func (this *MockTxPool) CleanTx(txs []*types.Transaction) {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	this.PendingTxList = make([]*types.Transaction, 0)
}
