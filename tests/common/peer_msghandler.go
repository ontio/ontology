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
	"errors"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	types2 "github.com/ontio/ontology/core/types"
	common2 "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/message/types"
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
	peer.Net.Send(peer.Local.GetID(), from, pong)
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
	headers, err := GetHeadersFromHash(peer.Lgr, startHash, stopHash)
	if err != nil {
		log.Warnf("get headers in HeadersReqHandle error: %s,startHash:%s,stopHash:%s", err.Error(), startHash.ToHexString(), stopHash.ToHexString())
		return
	}
	log.Infof("get headers req: %v", headersReq)
	log.Infof("get headers rsp: %v", headers)
	hdrsRsp := msgpack.NewHeaders(headers)
	peer.Net.Send(peer.Local.GetID(), from, hdrsRsp)
}

func (peer *MockPeer) handleHeaders(from uint64, msg types.Message) {
	blkHdrs := msg.(*types.RawBlockHeader)
	if blkHdrs == nil {
		log.Errorf("invalid hdrs rsp from %d", from)
		return
	}
	if len(blkHdrs.BlkHdr) == 0 {
		return
	}

	hdrs := make([]*types2.Header, 0)
	for _, rawHdr := range blkHdrs.BlkHdr {
		hdr := &types2.Header{}
		hdr.Deserialization(common.NewZeroCopySource(rawHdr.Payload))
		hdrs = append(hdrs, hdr)
	}

	shardID := common.NewShardIDUnchecked(hdrs[0].ShardID)
	if syncer, present := peer.syncers[shardID]; present {
		log.Infof("receives headers")
		syncer.OnHeaderReceive(from, hdrs)
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
	merkleRoot, err = peer.Lgr.GetStateMerkleRoot(block.Header.Height)
	if err != nil {
		log.Debugf("[p2p]failed to get state merkel root at height %v, err %v",
			block.Header.Height, err)
		return
	}
	blkmsg := msgpack.NewBlock(block, merkleRoot)
	peer.Net.Send(peer.Local.GetID(), from, blkmsg)
}

func (peer *MockPeer) handleBlock(from uint64, msg types.Message) {
	blk := msg.(*types.Block)
	if blk == nil {
		log.Errorf("invalid blk msg from %d", from)
		return
	}

	log.Infof("peer %d received block %d", peer.Local.GetID(), blk.Blk.Header.Height)
	shardID := common.NewShardIDUnchecked(blk.Blk.Header.ShardID)
	if syncer, present := peer.syncers[shardID]; present {
		syncer.OnBlockReceive(from, 100, blk.Blk, blk.MerkleRoot)
	}
}

///////////////////////

//get blk hdrs from starthash to stophash
func GetHeadersFromHash(lgr *ledger.Ledger, startHash common.Uint256, stopHash common.Uint256) ([]*types2.RawHeader, error) {
	var count uint32 = 0
	headers := []*types2.RawHeader{}
	var startHeight uint32
	var stopHeight uint32
	curHeight := lgr.GetCurrentHeaderHeight()
	if startHash == common.UINT256_EMPTY {
		if stopHash == common.UINT256_EMPTY {
			if curHeight > common2.MAX_BLK_HDR_CNT {
				count = common2.MAX_BLK_HDR_CNT
			} else {
				count = curHeight
			}
		} else {
			bkStop, err := lgr.GetRawHeaderByHash(stopHash)
			if err != nil || bkStop == nil {
				return nil, err
			}
			stopHeight = bkStop.Height
			count = curHeight - stopHeight
			if count > common2.MAX_BLK_HDR_CNT {
				count = common2.MAX_BLK_HDR_CNT
			}
		}
	} else {
		bkStart, err := lgr.GetRawHeaderByHash(startHash)
		if err != nil || bkStart == nil {
			return nil, err
		}
		startHeight = bkStart.Height
		if stopHash != common.UINT256_EMPTY {
			bkStop, err := lgr.GetRawHeaderByHash(stopHash)
			if err != nil || bkStop == nil {
				return nil, err
			}
			stopHeight = bkStop.Height

			// avoid unsigned integer underflow
			if startHeight < stopHeight {
				return nil, errors.New("[p2p]do not have header to send")
			}
			count = startHeight - stopHeight

			if count >= common2.MAX_BLK_HDR_CNT {
				count = common2.MAX_BLK_HDR_CNT
				stopHeight = startHeight - common2.MAX_BLK_HDR_CNT
			}
		} else {

			if startHeight > common2.MAX_BLK_HDR_CNT {
				count = common2.MAX_BLK_HDR_CNT
			} else {
				count = startHeight
			}
		}
	}

	var i uint32
	for i = 1; i <= count; i++ {
		hash := lgr.GetBlockHash(stopHeight + i)
		hd, err := lgr.GetRawHeaderByHash(hash)
		if err != nil {
			log.Debugf("[p2p]net_server GetBlockWithHeight failed with err=%s, hash=%x,height=%d,shardID:%v,\n", err.Error(), hash, stopHeight+i, lgr.ShardID)
			return nil, err
		}
		headers = append(headers, hd)
	}

	return headers, nil
}
