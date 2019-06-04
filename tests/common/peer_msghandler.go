package TestCommon

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/message/utils"
	"github.com/ontio/ontology/core/ledger"
)

func (peer *MockPeer) handlePingMsg(from uint64, msg types.Message) {
	ping := msg.(*types.Ping)
	if ping == nil {
		log.Errorf("invalid ping msg from %d", from)
		return
	}
	selfShardId := peer.Lgr.ShardID
	peer.SetHeight(from, ping.Height[selfShardId.ToUint64()])

	h := make(map[uint64]uint32)
	h[peer.Lgr.ShardID.ToUint64()] = peer.Lgr.GetCurrentBlockHeight()
	pong := msgpack.NewPongMsg(h)
	peer.Net.Broadcast(from, pong)
}

func (peer *MockPeer) handlePongMsg(from uint64, msg types.Message) {
	pong := msg.(*types.Pong)
	if pong == nil {
		log.Errorf("invalid pong msg from %d", from)
		return
	}
	selfShardId := peer.Lgr.ShardID
	peer.SetHeight(from, pong.Height[selfShardId.ToUint64()])
}

func (peer *MockPeer) handleGetHeadersReq(from uint64, msg types.Message) {
	headersReq := msg.(*types.HeadersReq)
	if headersReq.ShardID != peer.Lgr.ShardID.ToUint64() {
		log.Errorf("invalid getHeaderReq from shard %d", headersReq.ShardID)
		return
	}

	startHash := headersReq.HashStart
	stopHash := headersReq.HashEnd
	shardId, err := common.NewShardID(headersReq.ShardID)
	if err != nil {
		log.Warnf("get headers in HeadersReqHandle error: %s,shardID:%s", err.Error(), headersReq.ShardID)
		return
	}
	headers, err := utils.GetHeadersFromHash(shardId, startHash, stopHash)
	if err != nil {
		log.Warnf("get headers in HeadersReqHandle error: %s,startHash:%s,stopHash:%s", err.Error(), startHash.ToHexString(), stopHash.ToHexString())
		return
	}
	hdrsRsp := msgpack.NewHeaders(headers)
	peer.Net.Broadcast(from, hdrsRsp)
}

func (peer *MockPeer) handleHeaders(from uint64, msg types.Message) {
	blkHdrs := msg.(*types.BlkHeader)
	if blkHdrs == nil {
		log.Errorf("invalid hdrs rsp from %d", from)
		return
	}
	if len(blkHdrs.BlkHdr) == 0 {
		return
	}

	shardID := common.NewShardIDUnchecked(blkHdrs.BlkHdr[0].ShardID)
	if syncer, present := peer.syncers[shardID]; present {
		syncer.OnHeaderReceive(from, blkHdrs.BlkHdr)
	}
}

func (peer *MockPeer) handleInvMsg(from uint64, msg types.Message) {
	rsp := msg.(*types.Inv)
	if rsp == nil {
		log.Errorf("invalid inv msg frm %d", from)
		return
	}
	log.Infof("received inv msg from %d", from)
}

func (peer *MockPeer) handleGetDataReq(from uint64, msg types.Message) {
	req := msg.(*types.DataReq)
	if req == nil {
		log.Errorf("invalid data req from: %d", from)
		return
	}
	shardId, err := common.NewShardID(req.ShardID)
	if err != nil {
		log.Error("data req with invalid shard %d: %s", req.ShardID, err)
		return
	}
	if shardId != peer.Lgr.ShardID {
		log.Infof("data req to shard %d, local shard %d", req.ShardID, peer.Lgr.ShardID)
		return
	}
	if req.DataType != common.BLOCK {
		return
	}
	hash := req.Hash
	var merkleRoot common.Uint256
	block, err := peer.Lgr.GetBlockByHash(hash)
	if err != nil || block == nil || block.Header == nil {
		log.Debug("[p2p]can't get block by hash: ", hash,
			" ,send not found message")
		return
	}
	merkleRoot, err = ledger.DefLedger.GetStateMerkleRoot(block.Header.Height)
	if err != nil {
		log.Debugf("[p2p]failed to get state merkel root at height %v, err %v",
			block.Header.Height, err)
		return
	}
	blkmsg := msgpack.NewBlock(block, merkleRoot)
	peer.Net.Broadcast(from, blkmsg)
}

func (peer *MockPeer) handleBlock(from uint64, msg types.Message) {
	blk := msg.(*types.Block)
	if blk == nil {
		log.Errorf("invalid blk msg from %d", from)
		return
	}

	shardID := common.NewShardIDUnchecked(blk.Blk.Header.ShardID)
	if syncer, present := peer.syncers[shardID]; present {
		syncer.OnBlockReceive(from, 100, blk.Blk, blk.MerkleRoot)
	}
}
