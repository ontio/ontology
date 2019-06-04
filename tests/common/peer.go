package TestCommon

import (
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/p2pserver"
	"github.com/ontio/ontology/p2pserver/common"
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

	msgChan chan *MockMsg
}

func NewPeer(lgr *ledger.Ledger) *MockPeer {
	return &MockPeer{
		Net:           MockNet,
		Lgr:           lgr,
		remoteHeights: make(map[uint64]uint32),
		syncers:       make(map[common2.ShardID]*p2pserver.BlockSyncMgr),
		msgChan:       make(chan *MockMsg, 1000),
	}
}

func (peer *MockPeer) Register() {
	peer.Net.RegisterPeer(peer)
}

func (peer *MockPeer) AddBlockSyncer(shardID common2.ShardID, blockSyncer *p2pserver.BlockSyncMgr) {
	peer.syncers[shardID] = blockSyncer
}

func (peer *MockPeer) Start() {
	go func() {
		for {
			select {
			case msg := <-peer.msgChan:
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
				case common.GET_DATA_TYPE:
					// handle block/tx req from peer
					peer.handleGetDataReq(msg.from, msg.msg)
				case common.BLOCK_TYPE:
					// handle block msg from perr
					peer.handleBlock(msg.from, msg.msg)
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
	heights := make(map[uint64]uint32)
	heights[peer.Lgr.ShardID.ToUint64()] = peer.Lgr.GetCurrentBlockHeight()

	pingMsg := msgpack.NewPingMsg(heights)
	peer.Net.Broadcast(peer.Local.GetID(), pingMsg)
}

func (peer *MockPeer) Connected(newPeer uint64) {
	for _, syncer := range peer.syncers {
		syncer.OnAddNode(newPeer)
	}
}

func (peer *MockPeer) Disconnected(failedPeer uint64) {
	for _, syncer := range peer.syncers {
		syncer.OnDelNode(failedPeer)
	}
}

func (peer *MockPeer) Receive(fromPeer uint64, msg types.Message) {
	peer.msgChan <- &MockMsg{
		from: fromPeer,
		msg:  msg,
	}
}
