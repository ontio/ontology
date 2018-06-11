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
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	pc	"github.com/ontio/ontology/p2pserver/common"
)

type EmergencyActionResponse struct {
	PubKey   keypair.PublicKey
	SigOnBlk []byte
	RspSig   []byte
	hash     *common.Uint256
}

func (this *EmergencyActionResponse) Serialize(w io.Writer) error {
	var err error
	err = serialization.WriteVarBytes(w, keypair.SerializePublicKey(this.PubKey))
	if err != nil {
		return fmt.Errorf("failed to serialize responser public key %v. PubKey %v",
			err, this.PubKey)
	}
	err = serialization.WriteVarBytes(w, this.SigOnBlk)
	if err != nil {
		return fmt.Errorf("failed to serialize responser sig on block %v. signature %v",
			err, this.RspSig)
	}
	err = serialization.WriteVarBytes(w, this.RspSig)
	if err != nil {
		return fmt.Errorf("failed to serialize response sig %v. signature %v", err, this.RspSig)
	}
	return nil
}

func (this *EmergencyActionResponse) Deserialize(r io.Reader) error {
	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("failed to read buf for responser public key %v", err)
	}
	this.PubKey, err = keypair.DeserializePublicKey(buf)
	if err != nil {
		return fmt.Errorf("failed to deserialize responser public key %v", err)
	}

	sigOnBlk, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("failed to read responser signature on block %v", err)
	}
	this.SigOnBlk = sigOnBlk

	rspSig, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("failed to read responser signature on block %v", err)
	}
	this.RspSig = rspSig
	return nil
}

func (this *EmergencyActionResponse) Hash() common.Uint256 {
	if this.hash != nil {
		return *this.hash
	}

	buf := new(bytes.Buffer)
	serialization.WriteVarBytes(buf, keypair.SerializePublicKey(this.PubKey))
	serialization.WriteVarBytes(buf, this.SigOnBlk)

	temp := sha256.Sum256(buf.Bytes())
	hash := common.Uint256(sha256.Sum256(temp[:]))
	this.hash = &hash
	return *this.hash
}

type EmergencyRspMsg struct {
	Payload EmergencyActionResponse
}

func (this *EmergencyRspMsg) CmdType() string {
	return pc.EMERGENCY_RSP_TYPE
}

// Serialization the EmergencyRspMsg for network
func (this *EmergencyRspMsg) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	err := this.Payload.Serialize(p)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize EmergencyRspMsg %v", err)
	}

	return p.Bytes(), nil
}

// Deserialization the EmergencyReqMsg from network
func (this *EmergencyRspMsg) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := this.Payload.Deserialize(buf)
	if err != nil {
		return err
	}
	return nil
}
