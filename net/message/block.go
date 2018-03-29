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
	"errors"

	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	"github.com/Ontology/net/actor"
	. "github.com/Ontology/net/protocol"
)

type blockReq struct {
	msgHdr
}

type block struct {
	msgHdr
	blk types.Block
}

func (msg block) Handle(node Noder) error {
	log.Debug("RX block message")
	hash := msg.blk.Hash()
	if con, _ := actor.IsContainBlock(hash); con != true {
		actor.AddBlock(&msg.blk)
		node.RemoveFlightHeight(msg.blk.Header.Height)
	} else {
		log.Debug("Receive duplicated block")
	}
	return nil
}

func (msg dataReq) Handle(node Noder) error {
	log.Debug()
	reqType := common.InventoryType(msg.dataType)
	hash := msg.hash
	switch reqType {
	case common.BLOCK:
		block, err := NewBlockFromHash(hash)
		if err != nil || block == nil {
			log.Debug("Can't get block from hash: ", hash, " ,send not found message")
			b, err := NewNotFound(hash)
			node.Tx(b)
			return err
		}
		log.Debug("block height is ", block.Header.Height, " ,hash is ", hash)
		buf, err := NewBlock(block)
		if err != nil {
			return err
		}
		node.Tx(buf)

	case common.TRANSACTION:
		txn, err := NewTxnFromHash(hash)
		if err != nil {
			return err
		}
		buf, err := NewTxn(txn)
		if err != nil {
			return err
		}
		go node.Tx(buf)
	}
	return nil
}

func NewBlockFromHash(hash common.Uint256) (*types.Block, error) {
	bk, err := actor.GetBlockByHash(hash)
	if err != nil || bk == nil {
		log.Errorf("Get Block error: %s, block hash: %x", err.Error(), hash)
		return nil, err
	}
	return bk, nil
}

func NewBlock(bk *types.Block) ([]byte, error) {
	log.Debug()
	var msg block
	msg.blk = *bk
	msg.msgHdr.Magic = NET_MAGIC
	cmd := "block"
	copy(msg.msgHdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	bk.Serialize(tmpBuffer)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:CHECKSUM_LEN])
	binary.Read(buf, binary.LittleEndian, &(msg.msgHdr.Checksum))
	msg.msgHdr.Length = uint32(len(p.Bytes()))
	log.Debug("The message payload length is ", msg.msgHdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

func ReqBlkData(node Noder, hash common.Uint256) error {
	var msg dataReq
	msg.dataType = common.BLOCK
	msg.hash = hash

	msg.msgHdr.Magic = NET_MAGIC
	copy(msg.msgHdr.CMD[0:7], "getdata")
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(msg.dataType))
	msg.hash.Serialize(p)
	if err != nil {
		log.Error("Binary Write failed at new getdata Msg")
		return err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:CHECKSUM_LEN])
	binary.Read(buf, binary.LittleEndian, &(msg.msgHdr.Checksum))
	msg.msgHdr.Length = uint32(len(p.Bytes()))
	log.Debug("The message payload length is ", msg.msgHdr.Length)

	sendBuf, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return err
	}
	node.Tx(sendBuf)
	return nil
}

func (msg block) Verify(buf []byte) error {
	err := msg.msgHdr.Verify(buf)
	return err
}

func (msg block) Serialization() ([]byte, error) {
	hdrBuf, err := msg.msgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	msg.blk.Serialize(buf)

	return buf.Bytes(), err
}

func (msg *block) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)

	err := binary.Read(buf, binary.LittleEndian, &(msg.msgHdr))
	if err != nil {
		log.Warn("Parse block message hdr error")
		return errors.New("Parse block message hdr error")
	}

	err = msg.blk.Deserialize(buf)
	if err != nil {
		log.Warn("Parse block message error")
		return errors.New("Parse block message error")
	}

	return err
}
