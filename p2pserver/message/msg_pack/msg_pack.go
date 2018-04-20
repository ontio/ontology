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
	"github.com/ontio/ontology/p2pserver/actor/req"
	msgCommon "github.com/ontio/ontology/p2pserver/common"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	p2pnet "github.com/ontio/ontology/p2pserver/net/protocol"
)

//Peer address package
func NewAddrs(nodeAddrs []msgCommon.PeerAddr, count uint64) ([]byte, error) {
	var addr mt.Addr
	addr.NodeAddrs = nodeAddrs
	addr.NodeCnt = count

	m, err := addr.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

//Peer address request package
func NewAddrReq() ([]byte, error) {
	var msg mt.AddrReq

	buf, err := msg.Serialization()
	if err != nil {
		return nil, err
	}
	return buf, err
}

///block package
func NewBlock(bk *ct.Block) ([]byte, error) {
	log.Debug()
	var blk mt.Block
	blk.Blk = *bk

	m, err := blk.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

//blk hdr package
func NewHeaders(headers []ct.Header, count uint32) ([]byte, error) {
	var blkHdr mt.BlkHeader
	blkHdr.Cnt = count
	blkHdr.BlkHdr = headers

	m, err := blkHdr.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

//blk hdr req package
func NewHeadersReq(curHdrHash common.Uint256) ([]byte, error) {
	var h mt.HeadersReq
	h.P.Len = 1
	buf := curHdrHash
	copy(h.P.HashEnd[:], buf[:])

	m, err := h.Serialization()
	return m, err
}

////Consensus info package
func NewConsensus(cp *mt.ConsensusPayload) ([]byte, error) {
	log.Debug()
	var cons mt.Consensus
	cons.Cons = *cp
	m, err := cons.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
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
func NewInv(invPayload *mt.InvPayload) ([]byte, error) {
	var inv mt.Inv
	inv.P.Blk = invPayload.Blk
	inv.P.InvType = invPayload.InvType
	inv.P.Cnt = invPayload.Cnt

	m, err := inv.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

//NotFound package
func NewNotFound(hash common.Uint256) ([]byte, error) {
	log.Debug()
	var notFound mt.NotFound
	notFound.Hash = hash

	m, err := notFound.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

//ping msg package
func NewPingMsg(height uint64) ([]byte, error) {
	log.Debug()
	var ping mt.Ping
	ping.Height = uint64(height)

	m, err := ping.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

//pong msg package
func NewPongMsg(height uint64) ([]byte, error) {
	log.Debug()
	var pong mt.Pong
	pong.Height = uint64(height)

	m, err := pong.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

//Transaction package
func NewTxn(txn *ct.Transaction) ([]byte, error) {
	log.Debug()
	var trn mt.Trn
	trn.Txn = *txn

	m, err := trn.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

//version ack package
func NewVerAck(isConsensus bool) ([]byte, error) {
	var verAck mt.VerACK
	verAck.IsConsensus = isConsensus

	m, err := verAck.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

//VersionPayload package
func NewVersionPayload(n p2pnet.P2P, isCons bool) mt.VersionPayload {
	vpl := mt.VersionPayload{
		Version:      n.GetVersion(),
		Services:     n.GetServices(),
		SyncPort:     n.GetSyncPort(),
		ConsPort:     n.GetConsPort(),
		Nonce:        n.GetID(),
		IsConsensus:  isCons,
		HttpInfoPort: n.GetHttpInfoPort(),
	}

	height, _ := req.GetCurrentBlockHeight()
	vpl.StartHeight = uint64(height)
	if n.GetRelay() {
		vpl.Relay = 1
	} else {
		vpl.Relay = 0
	}
	if config.Parameters.HttpInfoPort > 0 {
		vpl.Cap[msgCommon.HTTP_INFO_FLAG] = 0x01
	} else {
		vpl.Cap[msgCommon.HTTP_INFO_FLAG] = 0x00
	}

	vpl.UserAgent = 0x00
	vpl.TimeStamp = uint32(time.Now().UTC().UnixNano())

	return vpl
}

//version msg package
func NewVersion(vpl mt.VersionPayload, pk keypair.PublicKey) ([]byte, error) {
	log.Debug()
	var version mt.Version
	version.P = vpl
	version.PK = pk
	log.Debug("new version msg.pk is ", version.PK)

	m, err := version.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

//transaction request package
func NewTxnDataReq(hash common.Uint256) ([]byte, error) {
	var dataReq mt.DataReq
	dataReq.DataType = common.TRANSACTION
	dataReq.Hash = hash

	buf, _ := dataReq.Serialization()
	return buf, nil
}

//block request package
func NewBlkDataReq(hash common.Uint256) ([]byte, error) {
	var dataReq mt.DataReq
	dataReq.DataType = common.BLOCK
	dataReq.Hash = hash

	sendBuf, err := dataReq.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return sendBuf, nil
}

//consensus request package
func NewConsensusDataReq(hash common.Uint256) ([]byte, error) {
	var dataReq mt.DataReq
	dataReq.DataType = common.CONSENSUS
	dataReq.Hash = hash
	buf, _ := dataReq.Serialization()
	return buf, nil
}
