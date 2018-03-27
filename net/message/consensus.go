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
	"io"

	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/signature"
	"github.com/Ontology/net/actor"
	. "github.com/Ontology/net/protocol"
	"github.com/ontio/ontology-crypto/keypair"
)

type ConsensusPayload struct {
	Version         uint32
	PrevHash        common.Uint256
	Height          uint32
	BookkeeperIndex uint16
	Timestamp       uint32
	Data            []byte

	Owner     keypair.PublicKey
	Signature []byte

	hash common.Uint256
}

type consensus struct {
	msgHdr
	cons ConsensusPayload
}

func (cp *ConsensusPayload) Hash() common.Uint256 {
	return common.Uint256{}
}

func (cp *ConsensusPayload) Verify() error {
	buf := new(bytes.Buffer)
	cp.SerializeUnsigned(buf)
	err := signature.Verify(cp.Owner, buf.Bytes(), cp.Signature)
	if err != nil {
		err = errors.New("consensus failed: signature verification failed")
	}

	return err
}

func (cp *ConsensusPayload) ToArray() []byte {
	b := new(bytes.Buffer)
	cp.Serialize(b)
	return b.Bytes()
}

func (cp *ConsensusPayload) InvertoryType() common.InventoryType {
	return common.CONSENSUS
}

func (cp *ConsensusPayload) GetMessage() []byte {
	//TODO: GetMessage
	return []byte{}
}

func (msg consensus) Handle(node Noder) error {
	log.Debug()
	if actor.ConsensusPid != nil {
		actor.ConsensusPid.Tell(&msg.cons)
	}
	return nil
}

func reqConsensusData(node Noder, hash common.Uint256) error {
	var msg dataReq
	msg.dataType = common.CONSENSUS
	// TODO handle the hash array case
	msg.hash = hash

	buf, _ := msg.Serialization()
	go node.Tx(buf)

	return nil
}
func (cp *ConsensusPayload) Type() common.InventoryType {

	//TODO:Temporary add for Interface signature.SignableData use.
	return common.CONSENSUS
}

func (cp *ConsensusPayload) SerializeUnsigned(w io.Writer) error {
	serialization.WriteUint32(w, cp.Version)
	cp.PrevHash.Serialize(w)
	serialization.WriteUint32(w, cp.Height)
	serialization.WriteUint16(w, cp.BookkeeperIndex)
	serialization.WriteUint32(w, cp.Timestamp)
	serialization.WriteVarBytes(w, cp.Data)
	return nil

}

func (cp *ConsensusPayload) Serialize(w io.Writer) error {
	err := cp.SerializeUnsigned(w)
	if err != nil {
		return err
	}
	buf := keypair.SerializePublicKey(cp.Owner)
	err = serialization.WriteVarBytes(w, buf)
	if err != nil {
		return err
	}

	err = serialization.WriteVarBytes(w, cp.Signature)
	if err != nil {
		return err
	}

	return err
}

func (msg *consensus) Serialization() ([]byte, error) {
	hdrBuf, err := msg.msgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = msg.cons.Serialize(buf)

	return buf.Bytes(), err
}

func (cp *ConsensusPayload) DeserializeUnsigned(r io.Reader) error {
	var err error
	cp.Version, err = serialization.ReadUint32(r)
	if err != nil {
		log.Warn("consensus item Version Deserialize failed.")
		return errors.New("consensus item Version Deserialize failed.")
	}

	preBlock := new(common.Uint256)
	err = preBlock.Deserialize(r)
	if err != nil {
		log.Warn("consensus item preHash Deserialize failed.")
		return errors.New("consensus item preHash Deserialize failed.")
	}
	cp.PrevHash = *preBlock

	cp.Height, err = serialization.ReadUint32(r)
	if err != nil {
		log.Warn("consensus item Height Deserialize failed.")
		return errors.New("consensus item Height Deserialize failed.")
	}

	cp.BookkeeperIndex, err = serialization.ReadUint16(r)
	if err != nil {
		log.Warn("consensus item BookkeeperIndex Deserialize failed.")
		return errors.New("consensus item BookkeeperIndex Deserialize failed.")
	}

	cp.Timestamp, err = serialization.ReadUint32(r)
	if err != nil {
		log.Warn("consensus item Timestamp Deserialize failed.")
		return errors.New("consensus item Timestamp Deserialize failed.")
	}

	cp.Data, err = serialization.ReadVarBytes(r)
	if err != nil {
		log.Warn("consensus item Data Deserialize failed.")
		return errors.New("consensus item Data Deserialize failed.")
	}

	return nil
}

func (cp *ConsensusPayload) Deserialize(r io.Reader) error {
	err := cp.DeserializeUnsigned(r)

	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		log.Warn("consensus item Owner deserialize failed, " + err.Error())
		return errors.New("consensus item Owner deserialize failed.")
	}
	cp.Owner, err = keypair.DeserializePublicKey(buf)
	if err != nil {
		log.Warn("consensus item Owner deserialize failed, " + err.Error())
		return errors.New("consensus item Owner deserialize failed.")
	}

	cp.Signature, err = serialization.ReadVarBytes(r)

	return err
}

func (msg *consensus) Deserialization(p []byte) error {
	log.Debug()
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.msgHdr))
	err = msg.cons.Deserialize(buf)
	return err
}

func NewConsensus(cp *ConsensusPayload) ([]byte, error) {
	log.Debug()
	var msg consensus
	msg.msgHdr.Magic = NET_MAGIC
	cmd := "consensus"
	copy(msg.msgHdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	cp.Serialize(tmpBuffer)
	msg.cons = *cp
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		log.Error("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.msgHdr.Checksum))
	msg.msgHdr.Length = uint32(len(b.Bytes()))
	log.Debug("NewConsensus The message payload length is ", msg.msgHdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		log.Error("Error Convert net message ", err.Error())
		return nil, err
	}
	return m, nil
}
