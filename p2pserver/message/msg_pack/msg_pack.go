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
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
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

	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, addr.NodeCnt)
	if err != nil {
		log.Error("Binary Write failed at new Msg: ", err.Error())
		return nil, err
	}

	err = binary.Write(p, binary.LittleEndian, addr.NodeAddrs)
	if err != nil {
		log.Error("Binary Write failed at new Msg: ", err.Error())
		return nil, err
	}

	checkSumBuf := mt.CheckSum(p.Bytes())
	addr.Hdr.Init("addr", checkSumBuf, uint32(len(p.Bytes())))
	log.Debug("The message payload length is ", addr.Hdr.Length)

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
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.Hdr.Init("getaddr", sum, 0)

	buf, err := msg.Serialization()
	if err != nil {
		return nil, err
	}

	str := hex.EncodeToString(buf)
	log.Debug("The message get addr length is: ", len(buf), " ", str)
	return buf, err
}

///block package
func NewBlock(bk *ct.Block) ([]byte, error) {
	log.Debug()
	var blk mt.Block
	blk.Blk = *bk

	tmpBuffer := bytes.NewBuffer([]byte{})
	bk.Serialize(tmpBuffer)

	checkSumBuf := mt.CheckSum(tmpBuffer.Bytes())
	blk.Init("block", checkSumBuf, uint32(len(tmpBuffer.Bytes())))
	log.Debug("The message payload length is ", blk.Length)

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

	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint32(tmpBuffer, blkHdr.Cnt)
	for _, header := range headers {
		header.Serialize(tmpBuffer)
	}

	checkSumBuf := mt.CheckSum(tmpBuffer.Bytes())
	blkHdr.Hdr.Init("headers", checkSumBuf, uint32(len(tmpBuffer.Bytes())))
	log.Debug("The message payload length is ", blkHdr.Hdr.Length)

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

	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(h.P))
	if err != nil {
		log.Error("Binary Write failed at new headersReq")
		return nil, err
	}

	s := mt.CheckSum(p.Bytes())
	h.Hdr.Init("getheaders", s, uint32(len(p.Bytes())))

	m, err := h.Serialization()
	return m, err
}

//blk request package
func NewBlocksReq(curBlkHash common.Uint256) ([]byte, error) {
	log.Debug("request block hash")
	var h mt.BlocksReq
	h.P.HeaderHashCount = 1
	buf := curBlkHash
	copy(h.P.HashStart[:], mt.Reverse(buf[:]))

	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(h.P))
	if err != nil {
		log.Error("Binary Write failed at new blocksReq")
		return nil, err
	}

	s := mt.CheckSum(p.Bytes())
	h.MsgHdr.Init("getblocks", s, uint32(len(p.Bytes())))
	m, err := h.Serialization()
	return m, err
}

////Consensus info package
func NewConsensus(cp *mt.ConsensusPayload) ([]byte, error) {
	log.Debug()
	var cons mt.Consensus

	tmpBuffer := bytes.NewBuffer([]byte{})
	cp.Serialize(tmpBuffer)
	cons.Cons = *cp

	checkSumBuf := mt.CheckSum(tmpBuffer.Bytes())
	cons.Init("consensus", checkSumBuf, uint32(len(tmpBuffer.Bytes())))
	log.Debug("NewConsensus The message payload length is ", cons.Length)

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

	tmpBuffer := bytes.NewBuffer([]byte{})
	invPayload.Serialization(tmpBuffer)

	checkSumBuf := mt.CheckSum(tmpBuffer.Bytes())
	inv.Hdr.Init("inv", checkSumBuf, uint32(len(tmpBuffer.Bytes())))
	log.Debug("NewInv The message payload length is ", inv.Hdr.Length)

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

	tmpBuffer := bytes.NewBuffer([]byte{})
	notFound.Hash.Serialize(tmpBuffer)

	checkSumBuf := mt.CheckSum(tmpBuffer.Bytes())
	notFound.Init("notfound", checkSumBuf, uint32(len(tmpBuffer.Bytes())))
	log.Debug("The message payload length is ", notFound.MsgHdr.Length)

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

	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(tmpBuffer, ping.Height)

	checkSumBuf := mt.CheckSum(tmpBuffer.Bytes())
	ping.Hdr.Init("ping", checkSumBuf, uint32(len(tmpBuffer.Bytes())))
	log.Debug("NewPingMsg The message payload length is ", ping.Hdr.Length)

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

	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(tmpBuffer, pong.Height)

	checkSumBuf := mt.CheckSum(tmpBuffer.Bytes())
	pong.Init("pong", checkSumBuf, uint32(len(tmpBuffer.Bytes())))
	log.Debug("NewPongMsg The message payload length is ", pong.Length)

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

	tmpBuffer := bytes.NewBuffer([]byte{})
	txn.Serialize(tmpBuffer)
	trn.Txn = *txn

	checkSumBuf := mt.CheckSum(tmpBuffer.Bytes())
	trn.Init("tx", checkSumBuf, uint32(len(tmpBuffer.Bytes())))
	log.Debug("NewTxn The message payload length is ", trn.Length)

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

	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteBool(tmpBuffer, verAck.IsConsensus)

	checkSumBuf := mt.CheckSum(tmpBuffer.Bytes())
	verAck.Init("verack", checkSumBuf, uint32(len(tmpBuffer.Bytes())))
	log.Debug("NewVerAck The message payload length is ", verAck.Length)

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

	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(version.P))
	serialization.WriteVarBytes(p, keypair.SerializePublicKey(version.PK))
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}

	checkSumBuf := mt.CheckSum(p.Bytes())
	version.Hdr.Init("version", checkSumBuf, uint32(len(p.Bytes())))
	log.Debug("NewVersion The message payload length is ", version.Hdr.Length)

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

	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(dataReq.DataType))
	dataReq.Hash.Serialize(p)
	if err != nil {
		log.Error("Binary Write failed at new getdata Msg")
		return nil, err
	}

	checkSumBuf := mt.CheckSum(p.Bytes())
	dataReq.Init("getdata", checkSumBuf, uint32(len(p.Bytes())))
	log.Debug("NewBlkDataReq The message payload length is ", dataReq.Length)

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
