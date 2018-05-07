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

package types

import (
	"bytes"
	"encoding/binary"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/log"
)

type Consensus struct {
	MsgHdr
	Cons ConsensusPayload
}

type PeerStateUpdate struct {
	PeerPubKey keypair.PublicKey
	Connected  bool
}

//Serialize message payload
func (this *Consensus) Serialization() ([]byte, error) {

	p := bytes.NewBuffer([]byte{})
	this.Cons.Serialize(p)
	checkSumBuf := CheckSum(p.Bytes())
	this.MsgHdr.Init("consensus", checkSumBuf, uint32(len(p.Bytes())))
	log.Debug("NewConsensus The message payload length is ", this.MsgHdr.Length)

	hdrBuf, err := this.MsgHdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	data := append(buf.Bytes(), p.Bytes()...)
	return data, nil
}

//Deserialize message payload
func (this *Consensus) Deserialization(p []byte) error {
	log.Debug()
	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(this.MsgHdr))
	err = this.Cons.Deserialize(buf)
	return err
}
