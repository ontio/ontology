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

package msgpack

import (
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	ct "github.com/ontio/ontology/core/types"
	msgCommon "github.com/ontio/ontology/p2pserver/common"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	p2pnet "github.com/ontio/ontology/p2pserver/net/protocol"
)

//Peer address package
func NewAddrs(nodeAddrs []msgCommon.PeerAddr) mt.Message {
	log.Trace()
	var addr mt.Addr
	addr.NodeAddrs = nodeAddrs

	return &addr
}

//Peer address request package
func NewAddrReq() mt.Message {
	log.Trace()
	var msg mt.AddrReq
	return &msg
}

///block package
func NewBlock(bk *ct.Block, merkleRoot common.Uint256) mt.Message {
	log.Trace()
	var blk mt.Block
	blk.Blk = bk
	blk.MerkleRoot = merkleRoot

	return &blk
}

//blk hdr package
func NewHeaders(headers []*ct.RawHeader) mt.Message {
	log.Trace()
	var blkHdr mt.RawBlockHeader
	blkHdr.BlkHdr = headers
	return &blkHdr
}

//blk hdr req package
func NewHeadersReq(shardId uint64, curHdrHash common.Uint256) mt.Message {
	log.Trace()
	var h mt.HeadersReq
	h.Len = 1
	h.HashEnd = curHdrHash
	h.ShardID = shardId
	return &h
}

////Consensus info package
func NewConsensus(cp *mt.ConsensusPayload) mt.Message {
	log.Trace()
	var cons mt.Consensus
	cons.Cons = *cp

	return &cons
}

func NewCrossShard(cross *mt.CrossShardPayload) mt.Message {
	log.Trace()
	var crossmsg mt.CrossShard
	crossmsg.Cons = *cross
	return &crossmsg
}

//InvPayload
func NewInvPayload(invType common.InventoryType, msg []common.Uint256) *mt.InvPayload {
	log.Trace()
	return &mt.InvPayload{
		InvType: invType,
		Blk:     msg,
	}
}

//Inv request package
func NewInv(invPayload *mt.InvPayload) mt.Message {
	log.Trace()
	var inv mt.Inv
	inv.P.Blk = invPayload.Blk
	inv.P.InvType = invPayload.InvType

	return &inv
}

//NotFound package
func NewNotFound(hash common.Uint256) mt.Message {
	log.Trace()
	var notFound mt.NotFound
	notFound.Hash = hash

	return &notFound
}

//ping msg package
func NewPingMsg(heights map[uint64]*mt.HeightInfo) *mt.Ping {
	log.Trace()
	var ping mt.Ping
	ping.Height = heights

	return &ping
}

//pong msg package
func NewPongMsg(heights map[uint64]*mt.HeightInfo) *mt.Pong {
	log.Trace()
	var pong mt.Pong
	pong.Height = heights

	return &pong
}

//Transaction package
func NewTxn(txn *ct.Transaction) mt.Message {
	log.Trace()
	var trn mt.Trn
	trn.Txn = txn

	return &trn
}

//version ack package
func NewVerAck() mt.Message {
	log.Trace()
	var verAck mt.VerACK

	return &verAck
}

//Version package
func NewVersion(n p2pnet.P2P, heights map[uint64]*mt.HeightInfo) mt.Message {
	log.Trace()
	var version mt.Version
	version.P = mt.VersionPayload{
		Version:      n.GetVersion(),
		Services:     n.GetServices(),
		SyncPort:     n.GetPort(),
		Nonce:        n.GetID(),
		IsConsensus:  false,
		HttpInfoPort: n.GetHttpInfoPort(),
		TimeStamp:    time.Now().UnixNano(),
		SoftVersion:  config.Version,
		ShardHeights: heights,
	}

	if n.GetRelay() {
		version.P.Relay = 1
	} else {
		version.P.Relay = 0
	}
	if config.DefConfig.P2PNode.HttpInfoPort > 0 {
		version.P.Cap[msgCommon.HTTP_INFO_FLAG] = 0x01
	} else {
		version.P.Cap[msgCommon.HTTP_INFO_FLAG] = 0x00
	}
	return &version
}

//transaction request package
func NewTxnDataReq(shardID common.ShardID, hash common.Uint256) mt.Message {
	log.Trace()
	var dataReq mt.DataReq
	dataReq.DataType = common.TRANSACTION
	dataReq.Hash = hash
	dataReq.ShardID = shardID.ToUint64()

	return &dataReq
}

//block request package
func NewBlkDataReq(shardID common.ShardID, hash common.Uint256) mt.Message {
	log.Trace()
	var dataReq mt.DataReq
	dataReq.DataType = common.BLOCK
	dataReq.Hash = hash
	dataReq.ShardID = shardID.ToUint64()

	return &dataReq
}

//consensus request package
func NewConsensusDataReq(shardID common.ShardID, hash common.Uint256) mt.Message {
	log.Trace()
	var dataReq mt.DataReq
	dataReq.DataType = common.CONSENSUS
	dataReq.Hash = hash
	dataReq.ShardID = shardID.ToUint64()

	return &dataReq
}
