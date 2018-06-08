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
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/p2pserver/common"
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
	P  VersionPayload
	PK keypair.PublicKey
}

//Serialize message payload
func (this Version) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(this.P))
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. payload:%v", this.P))
	}
	serialization.WriteVarBytes(p, keypair.SerializePublicKey(this.PK))
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNetPackFail, fmt.Sprintf("write error. publickey:%v", this.PK))
	}

	return p.Bytes(), nil
}

func (this *Version) CmdType() string {
	return common.VERSION_TYPE
}

//Deserialize message payload
func (this *Version) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)

	err := binary.Read(buf, binary.LittleEndian, &(this.P))
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("read payload error. buf:%v", buf))
	}

	keyBuf, err := serialization.ReadVarBytes(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("read public key buffer error. buf:%v", buf))
	}
	pk, err := keypair.DeserializePublicKey(keyBuf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, fmt.Sprintf("deserialize public key error. keyBuf:%v", keyBuf))
	}
	this.PK = pk
	return nil
}
