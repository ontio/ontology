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
	"errors"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
)

type VersionPayload struct {
	Version      uint32
	Services     uint64
	TimeStamp    uint32
	SyncPort     uint16
	HttpInfoPort uint16
	ConsPort     uint16
	Cap          [32]byte
	Nonce        uint64
	// TODO remove tempory to get serilization function passed
	UserAgent   uint8
	StartHeight uint64
	// FIXME check with the specify relay type length
	Relay       uint8
	IsConsensus bool
}
type Version struct {
	Hdr MsgHdr
	P   VersionPayload
	PK  keypair.PublicKey
}

//Check whether header is correct
func (msg Version) Verify(buf []byte) error {
	err := msg.Hdr.Verify(buf)
	return err
}

//Serialize message payload
func (msg Version) Serialization() ([]byte, error) {
	hdrBuf, err := msg.Hdr.Serialization()
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(hdrBuf)
	err = binary.Write(buf, binary.LittleEndian, msg.P)
	if err != nil {
		return nil, err
	}
	keyBuf := keypair.SerializePublicKey(msg.PK)
	err = serialization.WriteVarBytes(buf, keyBuf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

//Deserialize message payload
func (msg *Version) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)

	err := binary.Read(buf, binary.LittleEndian, &(msg.Hdr))
	if err != nil {
		log.Warn("Parse version message hdr error")
		return errors.New("Parse version message hdr error")
	}

	err = binary.Read(buf, binary.LittleEndian, &(msg.P))
	if err != nil {
		log.Warn("Parse version P message error")
		return errors.New("Parse version P message error")
	}

	keyBuf, err := serialization.ReadVarBytes(buf)
	if err != nil {
		return errors.New("Parse pubkey Deserialize failed.")
	}
	pk, err := keypair.DeserializePublicKey(keyBuf)
	if err != nil {
		return errors.New("Parse pubkey Deserialize failed.")
	}
	msg.PK = pk
	return err
}
