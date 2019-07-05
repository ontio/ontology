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
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/p2pserver"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/link"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/peer"
)

type MockMsg struct {
	from uint64
	msg  types.Message
}

type MockPeer struct {
	Local *peer.Peer
	Net   *MockNetwork
	Lgr   *ledger.Ledger

	remoteHeights map[uint64]uint32
	syncers       map[common2.ShardID]*p2pserver.BlockSyncMgr
	ConsensusPid  *actor.PID

	MsgChan chan *MockMsg
	Started bool // peer state
}

func NewPeer(lgr *ledger.Ledger) *MockPeer {
	p := &MockPeer{
		Local:         &peer.Peer{},
		Net:           MockNet,
		Lgr:           lgr,
		remoteHeights: make(map[uint64]uint32),
		syncers:       make(map[common2.ShardID]*p2pserver.BlockSyncMgr),
		MsgChan:       make(chan *MockMsg, 1000),
	}
	p.Local.Link = link.NewLink()
	heights := make(map[common2.ShardID]*types.HeightInfo)
	heights[lgr.ShardID] = &types.HeightInfo{
		Height: lgr.GetCurrentBlockHeight(),
	}
	p.Local.UpdateInfo(time.Now(), 1, 1, 20338, rand.Uint64(), 1, heights, "1")
	return p
}

func (peer *MockPeer) Register() {
	peer.Net.RegisterPeer(peer)
}

func (peer *MockPeer) AddBlockSyncer(shardID common2.ShardID, blockSyncer *p2pserver.BlockSyncMgr) {
	peer.syncers[shardID] = blockSyncer
}

func (peer *MockPeer) SetConsensusPid(t *testing.T, pid *actor.PID) {
	if pid == nil {
		t.Fatalf("set mock peer consensus nil pid")
	}
	peer.ConsensusPid = pid
}

func (peer *MockPeer) Start() {
	peer.Started = true
	go func() {
		for {
			select {
			case msg := <-peer.MsgChan:
				switch msg.msg.CmdType() {
				case common.PING_TYPE:
					// update peer height, response pong
					peer.handlePingMsg(msg.from, msg.msg)
				case common.PONG_TYPE:
					// update peer height
					peer.handlePongMsg(msg.from, msg.msg)
				case common.GET_HEADERS_TYPE:
					// get header from ledger
					peer.handleGetHeadersReq(msg.from, msg.msg)
				case common.HEADERS_TYPE:
					// append headers to syncer
					peer.handleHeaders(msg.from, msg.msg)
				case common.INV_TYPE:
					// handle inventory message
					peer.handleInvMsg(msg.from, msg.msg)
				case common.CONSENSUS_TYPE:
					// handle consensus message
					peer.handleConsensusMsg(msg.from, msg.msg)
				case common.GET_DATA_TYPE:
					// handle block/tx req from peer
					peer.handleGetDataReq(msg.from, msg.msg)
				case common.BLOCK_TYPE:
					// handle block msg from perr
					peer.handleBlock(msg.from, msg.msg)
				default:
					panic(fmt.Sprintf("peer %d, not handle msg type: %s", peer.Local.GetID(), msg.msg.CmdType()))
				}
			}
		}
	}()
}

func (peer *MockPeer) Send(p *peer.Peer, msg types.Message, isConsensus bool) error {
	peer.Net.Broadcast(peer.Local.GetID(), msg)
	return nil
}

func (peer *MockPeer) ReachMinConnection() bool {
	return true
}

func (peer *MockPeer) GetNode(id uint64) *peer.Peer {
	return peer.Net.GetPeer(id).Local
}

func (peer *MockPeer) SetHeight(id uint64, h uint32) {
	peer.remoteHeights[id] = h
}

func (peer *MockPeer) PingTo(peers []*peer.Peer) {
	heights := make(map[common2.ShardID]*types.HeightInfo)
	heights[peer.Lgr.ShardID] = &types.HeightInfo{
		Height: peer.Lgr.GetCurrentBlockHeight(),
	}

	pingMsg := msgpack.NewPingMsg(heights)
	peer.Net.Broadcast(peer.Local.GetID(), pingMsg)
}

func (peer *MockPeer) Connected(newPeer uint64) {
	p := peer.GetNode(newPeer)
	if p != nil {
		p.SetState(common.ESTABLISH)
	}
	for _, syncer := range peer.syncers {
		log.Infof("peer %d connected with %d", peer.Local.GetID(), newPeer)
		syncer.OnAddNode(newPeer)
	}
}

func (peer *MockPeer) Disconnected(failedPeer uint64) {
	for _, syncer := range peer.syncers {
		syncer.OnDelNode(failedPeer)
	}
}

func (peer *MockPeer) Receive(fromPeer uint64, msg types.Message) {
	if !peer.Started {
		log.Errorf("peer %d, not started to receive msg from %d", peer.Local.GetID(), fromPeer)
	}
	peer.MsgChan <- &MockMsg{
		from: fromPeer,
		msg:  msg,
	}
}
