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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/p2pserver/actor"
	"github.com/ontio/ontology/p2pserver/protocol"
)

type dataReq struct {
	msgHdr
	dataType common.InventoryType
	hash     common.Uint256
}

// Transaction message
type trn struct {
	msgHdr
	txn types.Transaction
}

func (msg trn) Handle(node protocol.Noder) error {
	log.Debug()
	log.Debug("RX Transaction message")
	tx := &msg.txn
	if !node.LocalNode().ExistedID(tx.Hash()) {
		actor.AddTransaction(&msg.txn)
		log.Debug("RX Transaction message hash", msg.txn.Hash())
	}

	return nil
}

func reqTxnData(node protocol.Noder, hash common.Uint256) error {
	var msg dataReq
	msg.dataType = common.TRANSACTION
	buf, _ := msg.Serialization()
	go node.Tx(buf)
	return nil
}

func (msg dataReq) Serialization() ([]byte, error) {
	hdrBuf, err := msg.msgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, msg.dataType)
	if err != nil {
		return nil, err
	}
	msg.hash.Serialize(buf)

	return buf.Bytes(), err
}

func (msg *dataReq) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.msgHdr))
	if err != nil {
		log.Warn("Parse datareq message hdr error")
		return errors.New("Parse datareq message hdr error")
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.dataType))
	if err != nil {
		log.Warn("Parse datareq message dataType error")
		return errors.New("Parse datareq message dataType error")
	}

	err = msg.hash.Deserialize(buf)
	if err != nil {
		log.Warn("Parse datareq message hash error")
		return errors.New("Parse datareq message hash error")
	}
	return nil
}

func NewTxnFromHash(hash common.Uint256) (*types.Transaction, error) {
	txn, err := actor.GetTxnFromLedger(hash)
	if err != nil || txn == nil {
		log.Error("Get transaction with hash error: ", err.Error())
		return nil, err
	}

	return txn, nil
}
func NewTxn(txn *types.Transaction) ([]byte, error) {
	log.Debug()
	var msg trn

	msg.msgHdr.Magic = protocol.NET_MAGIC
	cmd := "tx"
	copy(msg.msgHdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	txn.Serialize(tmpBuffer)
	msg.txn = *txn
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:protocol.CHECKSUM_LEN])
	binary.Read(buf, binary.LittleEndian, &(msg.msgHdr.Checksum))
	msg.msgHdr.Length = uint32(len(b.Bytes()))
	log.Debug("The message payload length is ", msg.msgHdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}

	return m, nil
}

func (msg trn) Serialization() ([]byte, error) {
	hdrBuf, err := msg.msgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	msg.txn.Serialize(buf)

	return buf.Bytes(), err
}

func (msg *trn) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.msgHdr))
	err = msg.txn.Deserialize(buf)
	if err != nil {
		return err
	}

	return nil
}

type txnPool struct {
	msgHdr
}
