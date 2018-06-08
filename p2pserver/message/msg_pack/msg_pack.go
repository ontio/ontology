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

	"github.com/ontio/ontology-crypto/keypair"
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
	var addr mt.Addr
	addr.NodeAddrs = nodeAddrs

	return &addr
}

//Peer address request package
func NewAddrReq() mt.Message {
	var msg mt.AddrReq
	return &msg
}

///block package
func NewBlock(bk *ct.Block) mt.Message {
	log.Debug()
	var blk mt.Block
	blk.Blk = *bk

	return &blk
}

//blk hdr package
func NewHeaders(headers []*ct.Header) mt.Message {
	var blkHdr mt.BlkHeader
	blkHdr.BlkHdr = headers

	return &blkHdr
}

//blk hdr req package
func NewHeadersReq(curHdrHash common.Uint256) mt.Message {
	var h mt.HeadersReq
	h.Len = 1
	buf := curHdrHash
	copy(h.HashEnd[:], buf[:])

	return &h
}

////Consensus info package
func NewConsensus(cp *mt.ConsensusPayload) mt.Message {
	log.Debug()
	var cons mt.Consensus
	cons.Cons = *cp

	return &cons
}

//InvPayload
func NewInvPayload(invType common.InventoryType, count uint32, msg []byte) *mt.InvPayload {
	return &mt.InvPayload{
		InvType: invType,
		Cnt:     count,
		Blk:     msg,
	}
}

//Inv request package
func NewInv(invPayload *mt.InvPayload) mt.Message {
	var inv mt.Inv
	inv.P.Blk = invPayload.Blk
	inv.P.InvType = invPayload.InvType
	inv.P.Cnt = invPayload.Cnt

	return &inv
}

//NotFound package
func NewNotFound(hash common.Uint256) mt.Message {
	log.Debug()
	var notFound mt.NotFound
	notFound.Hash = hash

	return &notFound
}

//ping msg package
func NewPingMsg(height uint64) *mt.Ping {
	log.Debug()
	var ping mt.Ping
	ping.Height = uint64(height)

	return &ping
}

//pong msg package
func NewPongMsg(height uint64) *mt.Pong {
	log.Debug()
	var pong mt.Pong
	pong.Height = uint64(height)

	return &pong
}

//Transaction package
func NewTxn(txn *ct.Transaction) mt.Message {
	log.Debug()
	var trn mt.Trn
	trn.Txn = txn

	return &trn
}

//version ack package
func NewVerAck(isConsensus bool) mt.Message {
	var verAck mt.VerACK
	verAck.IsConsensus = isConsensus

	return &verAck
}

//VersionPayload package
func NewVersionPayload(n p2pnet.P2P, isCons bool, height uint32) mt.VersionPayload {
	vpl := mt.VersionPayload{
		Version:      n.GetVersion(),
		Services:     n.GetServices(),
		SyncPort:     n.GetSyncPort(),
		ConsPort:     n.GetConsPort(),
		Nonce:        n.GetID(),
		IsConsensus:  isCons,
		HttpInfoPort: n.GetHttpInfoPort(),
	}

	vpl.StartHeight = uint64(height)
	if n.GetRelay() {
		vpl.Relay = 1
	} else {
		vpl.Relay = 0
	}
	if config.DefConfig.P2PNode.HttpInfoPort > 0 {
		vpl.Cap[msgCommon.HTTP_INFO_FLAG] = 0x01
	} else {
		vpl.Cap[msgCommon.HTTP_INFO_FLAG] = 0x00
	}

	vpl.UserAgent = 0x00
	vpl.TimeStamp = uint32(time.Now().UTC().UnixNano())

	return vpl
}

//version msg package
func NewVersion(vpl mt.VersionPayload, pk keypair.PublicKey) mt.Message {
	var version mt.Version
	version.P = vpl
	version.PK = pk
	log.Debug("new version msg.pk is ", version.PK)

	return &version
}

//transaction request package
func NewTxnDataReq(hash common.Uint256) mt.Message {
	var dataReq mt.DataReq
	dataReq.DataType = common.TRANSACTION
	dataReq.Hash = hash

	return &dataReq
}

//block request package
func NewBlkDataReq(hash common.Uint256) mt.Message {
	var dataReq mt.DataReq
	dataReq.DataType = common.BLOCK
	dataReq.Hash = hash

	return &dataReq
}

//consensus request package
func NewConsensusDataReq(hash common.Uint256) mt.Message {
	var dataReq mt.DataReq
	dataReq.DataType = common.CONSENSUS
	dataReq.Hash = hash

	return &dataReq
}
