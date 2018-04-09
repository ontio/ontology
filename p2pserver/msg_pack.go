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

package p2pserver

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"time"

	"github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	"github.com/Ontology/p2pserver/actor/req"
	msgCommon "github.com/Ontology/p2pserver/common"
	msg "github.com/Ontology/p2pserver/message"
	"github.com/Ontology/p2pserver/peer"
)

func NewAddrs(nodeAddrs []msgCommon.PeerAddr, count uint64) ([]byte, error) {
	var addr msg.Addr
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

	checkSumBuf := msg.CheckSum(p.Bytes())
	addr.Hdr.Init("addr", checkSumBuf, uint32(len(p.Bytes())))
	log.Debug("The message payload length is ", addr.Hdr.Length)

	m, err := addr.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewAddrReq() ([]byte, error) {
	var msg msg.AddrReq
	// Fixme the check is the []byte{0} instead of 0
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

func NewBlock(bk *types.Block) ([]byte, error) {
	log.Debug()
	var blk msg.Block
	blk.Blk = *bk

	tmpBuffer := bytes.NewBuffer([]byte{})
	bk.Serialize(tmpBuffer)

	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}

	checkSumBuf := msg.CheckSum(p.Bytes())
	blk.Init("block", checkSumBuf, uint32(len(p.Bytes())))
	log.Debug("The message payload length is ", blk.Length)

	m, err := blk.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewHeaders(headers []types.Header, count uint32) ([]byte, error) {
	var blkHdr msg.BlkHeader
	blkHdr.Cnt = count
	blkHdr.BlkHdr = headers

	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint32(tmpBuffer, blkHdr.Cnt)
	for _, header := range headers {
		header.Serialize(tmpBuffer)
	}
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}

	checkSumBuf := msg.CheckSum(b.Bytes())
	blkHdr.Hdr.Init("headers", checkSumBuf, uint32(len(b.Bytes())))
	log.Debug("The message payload length is ", blkHdr.Hdr.Length)

	m, err := blkHdr.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewHeadersReq(curHdrHash common.Uint256) ([]byte, error) {
	var h msg.HeadersReq
	h.P.Len = 1
	buf := curHdrHash
	copy(h.P.HashEnd[:], buf[:])

	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(h.P))
	if err != nil {
		log.Error("Binary Write failed at new headersReq")
		return nil, err
	}

	s := msg.CheckSum(p.Bytes())
	h.Hdr.Init("getheaders", s, uint32(len(p.Bytes())))

	m, err := h.Serialization()
	return m, err
}

func NewBlocksReq(curBlkHash common.Uint256) ([]byte, error) {
	log.Debug("request block hash")
	var h msg.BlocksReq
	// Fixme correct with the exactly request length
	h.P.HeaderHashCount = 1
	//Fixme! Should get the remote Node height.
	buf := curBlkHash
	copy(h.P.HashStart[:], msg.Reverse(buf[:]))

	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(h.P))
	if err != nil {
		log.Error("Binary Write failed at new blocksReq")
		return nil, err
	}

	s := msg.CheckSum(p.Bytes())
	h.MsgHdr.Init("getblocks", s, uint32(len(p.Bytes())))
	m, err := h.Serialization()
	return m, err
}

func NewConsensus(cp *msg.ConsensusPayload) ([]byte, error) {
	log.Debug()
	var cons msg.Consensus

	tmpBuffer := bytes.NewBuffer([]byte{})
	cp.Serialize(tmpBuffer)
	cons.Cons = *cp
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}

	checkSumBuf := msg.CheckSum(b.Bytes())
	cons.Init("consensus", checkSumBuf, uint32(len(b.Bytes())))
	log.Debug("NewConsensus The message payload length is ", cons.Length)

	m, err := cons.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewInv(invPayload *msg.InvPayload) ([]byte, error) {
	var inv msg.Inv
	inv.P.Blk = invPayload.Blk
	inv.P.InvType = invPayload.InvType
	inv.P.Cnt = invPayload.Cnt

	tmpBuffer := bytes.NewBuffer([]byte{})
	invPayload.Serialization(tmpBuffer)

	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg", err.Error())
		return nil, err
	}

	checkSumBuf := msg.CheckSum(b.Bytes())
	inv.Hdr.Init("inv", checkSumBuf, uint32(len(b.Bytes())))
	log.Debug("NewInv The message payload length is ", inv.Hdr.Length)

	m, err := inv.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewNotFound(hash common.Uint256) ([]byte, error) {
	log.Debug()
	var notFound msg.NotFound
	notFound.Hash = hash

	tmpBuffer := bytes.NewBuffer([]byte{})
	notFound.Hash.Serialize(tmpBuffer)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new notfound Msg")
		return nil, err
	}

	checkSumBuf := msg.CheckSum(p.Bytes())
	notFound.Init("notfound", checkSumBuf, uint32(len(p.Bytes())))
	log.Debug("The message payload length is ", notFound.MsgHdr.Length)

	m, err := notFound.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

func NewPingMsg(height uint64) ([]byte, error) {
	log.Debug()
	var ping msg.Ping
	ping.Height = uint64(height)

	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(tmpBuffer, ping.Height)
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}

	checkSumBuf := msg.CheckSum(b.Bytes())
	ping.Hdr.Init("ping", checkSumBuf, uint32(len(b.Bytes())))
	log.Debug("NewPingMsg The message payload length is ", ping.Hdr.Length)

	m, err := ping.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewPongMsg(height uint64) ([]byte, error) {
	log.Debug()
	var pong msg.Pong
	pong.Height = uint64(height)

	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteUint64(tmpBuffer, pong.Height)
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}

	checkSumBuf := msg.CheckSum(b.Bytes())
	pong.Init("pong", checkSumBuf, uint32(len(b.Bytes())))
	log.Debug("NewPongMsg The message payload length is ", pong.Length)

	m, err := pong.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewTxn(txn *types.Transaction) ([]byte, error) {
	log.Debug()
	var trn msg.Trn

	tmpBuffer := bytes.NewBuffer([]byte{})
	txn.Serialize(tmpBuffer)
	trn.Txn = *txn
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}

	checkSumBuf := msg.CheckSum(b.Bytes())
	trn.Init("tx", checkSumBuf, uint32(len(b.Bytes())))
	log.Debug("NewTxn The message payload length is ", trn.Length)

	m, err := trn.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

func NewVerAck(isConsensus bool) ([]byte, error) {
	var verAck msg.VerACK
	verAck.IsConsensus = isConsensus

	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteBool(tmpBuffer, verAck.IsConsensus)
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}

	checkSumBuf := msg.CheckSum(b.Bytes())
	verAck.Init("verack", checkSumBuf, uint32(len(b.Bytes())))
	log.Debug("NewVerAck The message payload length is ", verAck.Length)

	m, err := verAck.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}

func NewVersionPayload(p *peer.Peer, isCons bool) msg.VersionPayload {
	vpl := msg.VersionPayload{
		Version:      p.GetVersion(),
		Services:     p.GetServices(),
		SyncPort:     p.GetSyncPort(),
		ConsPort:     p.GetConsPort(),
		Nonce:        p.GetID(),
		IsConsensus:  isCons,
		HttpInfoPort: p.GetHttpInfoPort(),
	}

	height, _ := req.GetCurrentBlockHeight()
	vpl.StartHeight = uint64(height)
	if p.GetRelay() {
		vpl.Relay = 1
	} else {
		vpl.Relay = 0
	}
	if config.Parameters.HttpInfoStart {
		vpl.Cap[msg.HTTP_INFO_FLAG] = 0x01
	} else {
		vpl.Cap[msg.HTTP_INFO_FLAG] = 0x00
	}

	vpl.UserAgent = 0x00
	vpl.TimeStamp = uint32(time.Now().UTC().UnixNano())

	return vpl
}

func NewVersion(vpl msg.VersionPayload, pk *crypto.PubKey) ([]byte, error) {
	log.Debug()
	var version msg.Version
	version.P = vpl
	version.PK = pk
	log.Debug("new version msg.pk is ", version.PK)

	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(version.P))
	version.PK.Serialize(p)
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}

	checkSumBuf := msg.CheckSum(p.Bytes())
	version.Hdr.Init("version", checkSumBuf, uint32(len(p.Bytes())))
	log.Debug("NewVersion The message payload length is ", version.Hdr.Length)

	m, err := version.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

func NewTxnDataReq(hash common.Uint256) ([]byte, error) {
	var dataReq msg.DataReq
	dataReq.DataType = common.TRANSACTION
	// TODO handle the hash array case
	dataReq.Hash = hash

	buf, _ := dataReq.Serialization()
	return buf, nil
}

func NewBlkDataReq(hash common.Uint256) ([]byte, error) {
	var dataReq msg.DataReq
	dataReq.DataType = common.BLOCK
	dataReq.Hash = hash

	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(dataReq.DataType))
	dataReq.Hash.Serialize(p)
	if err != nil {
		log.Error("Binary Write failed at new getdata Msg")
		return nil, err
	}

	checkSumBuf := msg.CheckSum(p.Bytes())
	dataReq.Init("getdata", checkSumBuf, uint32(len(p.Bytes())))
	log.Debug("NewBlkDataReq The message payload length is ", dataReq.Length)

	sendBuf, err := dataReq.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return sendBuf, nil
}

func NewConsensusDataReq(hash common.Uint256) ([]byte, error) {
	var dataReq msg.DataReq
	dataReq.DataType = common.CONSENSUS
	// TODO handle the hash array case
	dataReq.Hash = hash
	buf, _ := dataReq.Serialization()
	return buf, nil
}
