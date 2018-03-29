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

package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	"io"
	"github.com/Ontology/net/actor"
	. "github.com/Ontology/net/protocol"
)

var LastInvHash Uint256

type hashReq struct {
	HeaderHashCount uint8
	hashStart       [HASH_LEN]byte
	hashStop        [HASH_LEN]byte
}

type blocksReq struct {
	msgHdr
	p	hashReq
}

type InvPayload struct {
	InvType InventoryType
	Cnt     uint32
	Blk     []byte
}

type Inv struct {
	Hdr msgHdr
	P   InvPayload
}

func NewBlocksReq(n Noder) ([]byte, error) {
	var h blocksReq
	log.Debug("request block hash")
	h.p.HeaderHashCount = 1
	buf, _ := actor.GetCurrentBlockHash()
	copy(h.p.hashStart[:], reverse(buf[:]))

	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(h.p))
	if err != nil {
		log.Error("Binary Write failed at new blocksReq")
		return nil, err
	}

	s := checkSum(p.Bytes())
	h.msgHdr.init("getblocks", s, uint32(len(p.Bytes())))
	m, err := h.Serialization()

	return m, err
}

func (msg blocksReq) Verify(buf []byte) error {
	err := msg.msgHdr.Verify(buf)
	return err
}

func (msg blocksReq) Handle(node Noder) error {
	log.Debug()
	log.Debug("handle blocks request")
	var startHash Uint256
	var stopHash Uint256
	startHash = msg.p.hashStart
	stopHash = msg.p.hashStop

	inv, err := GetInvFromBlockHash(startHash, stopHash)
	if err != nil {
		return err
	}
	buf, err := NewInv(inv)
	if err != nil {
		return err
	}
	go node.Tx(buf)
	return nil
}

func (msg blocksReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *blocksReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &msg)
	return err
}

func (msg Inv) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg Inv) Handle(node Noder) error {
	log.Debug()
	var id Uint256
	str := hex.EncodeToString(msg.P.Blk)
	log.Debug(fmt.Sprintf("The inv type: 0x%x block len: %d, %s\n",
		msg.P.InvType, len(msg.P.Blk), str))

	invType := InventoryType(msg.P.InvType)
	switch invType {
	case TRANSACTION:
		log.Debug("RX TRX message")
		id.Deserialize(bytes.NewReader(msg.P.Blk[:32]))
		if !node.ExistedID(id) {
			reqTxnData(node, id)
		}
	case BLOCK:
		log.Debug("RX block message")
		var i uint32
		count := msg.P.Cnt
		log.Debug("RX inv-block message, hash is ", msg.P.Blk)
		for i = 0; i < count; i++ {
			id.Deserialize(bytes.NewReader(msg.P.Blk[HASH_LEN*i:]))
			isContainBlock, _ := actor.IsContainBlock(id)
			if !isContainBlock && LastInvHash != id {
				LastInvHash = id
				// send the block request
				log.Infof("inv request block hash: %x", id)
				ReqBlkData(node, id)
			}

		}
	case CONSENSUS:
		log.Debug("RX consensus message")
		id.Deserialize(bytes.NewReader(msg.P.Blk[:32]))
		reqConsensusData(node, id)
	default:
		log.Warn("RX unknown inventory message")
	}
	return nil
}

func (msg Inv) Serialization() ([]byte, error) {
	hdrBuf, err := msg.Hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	msg.P.Serialization(buf)

	return buf.Bytes(), err
}

func (msg *Inv) Deserialization(p []byte) error {
	err := msg.Hdr.Deserialization(p)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(p[MSG_HDR_LEN:])
	invType, err := serialization.ReadUint8(buf)
	if err != nil {
		return err
	}
	msg.P.InvType = InventoryType(invType)
	msg.P.Cnt, err = serialization.ReadUint32(buf)
	if err != nil {
		return err
	}

	msg.P.Blk = make([]byte, msg.P.Cnt*HASH_LEN)
	err = binary.Read(buf, binary.LittleEndian, &(msg.P.Blk))

	return err
}

func (msg Inv) invType() InventoryType {
	return msg.P.InvType
}

func GetInvFromBlockHash(starthash Uint256, stophash Uint256) (*InvPayload, error) {
	var count uint32 = 0
	var i uint32
	var empty Uint256
	var startHeight uint32
	var stopHeight uint32
	curHeight, _ := actor.GetCurrentBlockHeight()
	if starthash == empty {
		if stophash == empty {
			if curHeight > MAX_BLK_HDR_CNT {
				count = MAX_BLK_HDR_CNT
			} else {
				count = curHeight
			}
		} else {
			bkStop, err := actor.GetHeaderByHash(stophash)
			if err != nil || bkStop == nil {
				return nil, err
			}
			stopHeight = bkStop.Height
			count = curHeight - stopHeight
			if curHeight > MAX_INV_HDR_CNT {
				count = MAX_INV_HDR_CNT
			}
		}
	} else {
		bkStart, err := actor.GetHeaderByHash(starthash)
		if err != nil || bkStart == nil {
			return nil, err
		}
		startHeight = bkStart.Height
		if stophash != empty {
			bkStop, err := actor.GetHeaderByHash(stophash)
			if err != nil || bkStop == nil {
				return nil, err
			}
			stopHeight = bkStop.Height
			count = startHeight - stopHeight
			if count >= MAX_INV_HDR_CNT {
				count = MAX_INV_HDR_CNT
				stopHeight = startHeight + MAX_INV_HDR_CNT
			}
		} else {
			if startHeight > MAX_INV_HDR_CNT {
				count = MAX_INV_HDR_CNT
			} else {
				count = startHeight
			}
		}
	}
	tmpBuffer := bytes.NewBuffer([]byte{})
	for i = 1; i <= count; i++ {
		//FIXME need add error handle for GetBlockWithHash
		hash, _ := actor.GetBlockHashByHeight(stopHeight + i)
		log.Debug("GetInvFromBlockHash i is ", i, " , hash is ", hash)
		hash.Serialize(tmpBuffer)
	}
	log.Debug("GetInvFromBlockHash hash is ", tmpBuffer.Bytes())
	return NewInvPayload(BLOCK, count, tmpBuffer.Bytes()), nil
}

func NewInvPayload(invType InventoryType, count uint32, msg []byte) *InvPayload {
	return &InvPayload{
		InvType: invType,
		Cnt:     count,
		Blk:     msg,
	}
}

func NewInv(inv *InvPayload) ([]byte, error) {
	var msg Inv

	msg.P.Blk = inv.Blk
	msg.P.InvType = inv.InvType
	msg.P.Cnt = inv.Cnt
	msg.Hdr.Magic = NET_MAGIC
	cmd := "inv"
	copy(msg.Hdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	inv.Serialization(tmpBuffer)

	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg", err.Error())
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:CHECKSUM_LEN])
	binary.Read(buf, binary.LittleEndian, &(msg.Hdr.Checksum))
	msg.Hdr.Length = uint32(len(b.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

func (msg *InvPayload) Serialization(w io.Writer) {
	serialization.WriteUint8(w, uint8(msg.InvType))
	serialization.WriteUint32(w, msg.Cnt)

	binary.Write(w, binary.LittleEndian, msg.Blk)
}
