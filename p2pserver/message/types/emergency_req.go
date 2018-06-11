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
	"github.com/ontio/ontology/core/types"
	pc "github.com/ontio/ontology/p2pserver/common"
)

type EmergencyReason uint8

const (
	FalseConsensus EmergencyReason = iota
	StoppedConsensus
)

type EmergencyEvidence uint8

const (
	ConsensusMessage EmergencyEvidence = iota
	ConflictBlock
)

type EmergencyActionRequest struct {
	Reason           EmergencyReason
	Evidence         EmergencyEvidence
	ProposalBlkNum   uint32
	ProposalBlk      *types.Block
	ProposerPK       keypair.PublicKey
	ProposerSigOnBlk []byte
	AdminSigs        []*types.Sig
	ReqPK            keypair.PublicKey
	ReqSig           []byte
	hash             *common.Uint256
}

// Serialize the EmergencyActionRequest
func (this *EmergencyActionRequest) Serialize(w io.Writer) error {
	var err error
	_, err = w.Write([]byte{byte(this.Reason), byte(this.Evidence)})
	if err != nil {
		return fmt.Errorf("failed to serialze %v. Reason %d, Evidence %d",
			err, this.Reason, this.Evidence)
	}

	err = serialization.WriteUint32(w, this.ProposalBlkNum)
	if err != nil {
		return fmt.Errorf("failed to serialize %v. proposal block num %d",
			err, this.ProposalBlkNum)
	}

	this.ProposalBlk.Serialize(w)

	err = serialization.WriteVarBytes(w, keypair.SerializePublicKey(this.ProposerPK))
	if err != nil {
		return fmt.Errorf("failed to serialize proposer public key %v. ProperPK %v",
			err, this.ProposerPK)
	}

	err = serialization.WriteVarBytes(w, this.ProposerSigOnBlk)
	if err != nil {
		return fmt.Errorf("failed to serialize proposer sig on block %v. ProposerSigOnBlk %v",
			err, this.ProposerSigOnBlk)
	}

	err = serialization.WriteVarUint(w, uint64(len(this.AdminSigs)))
	if err != nil {
		return fmt.Errorf("failed to serialize the length of admin sigs %v. len %d",
			err, len(this.AdminSigs))
	}

	for _, sig := range this.AdminSigs {
		err = sig.Serialize(w)
		if err != nil {
			return fmt.Errorf("failed to serialize signature %v. sig %v", err, sig)
		}
	}

	err = serialization.WriteVarBytes(w, keypair.SerializePublicKey(this.ReqPK))
	if err != nil {
		return fmt.Errorf("failed to serialize request public key %v. ReqPK %v",
			err, this.ReqPK)
	}

	err = serialization.WriteVarBytes(w, this.ReqSig)
	if err != nil {
		return fmt.Errorf("failed to serialize request sig %v. signature %v",
			err, this.ReqSig)
	}
	return nil
}

// Deserialize the EmergencyActionRequest
func (this *EmergencyActionRequest) Deserialize(r io.Reader) error {
	var tempByte [2]byte
	var err error
	r.Read(tempByte[:])
	this.Reason = EmergencyReason(tempByte[0])
	this.Evidence = EmergencyEvidence(tempByte[1])

	this.ProposalBlkNum, err = serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("failed to read proposal block num %v", err)
	}

	if this.ProposalBlk == nil {
		this.ProposalBlk = new(types.Block)
	}
	this.ProposalBlk.Deserialize(r)

	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("failed to read buf for proposer public key %v", err)
	}
	this.ProposerPK, err = keypair.DeserializePublicKey(buf)
	if err != nil {
		return fmt.Errorf("failed to deserialize proposer public key %v", err)
	}

	sigOnBlk, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("failed to read signature on block %v", err)
	}
	this.ProposerSigOnBlk = sigOnBlk

	num, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return fmt.Errorf("failed to read the length of admin sigs %v", err)
	}

	this.AdminSigs = make([]*types.Sig, 0, num)
	for i := 0; i < int(num); i++ {
		sig := new(types.Sig)
		err := sig.Deserialize(r)
		if err != nil {
			return fmt.Errorf("failed to deserialize admin signature %v", err)
		}
		this.AdminSigs = append(this.AdminSigs, sig)
	}

	buf, err = serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("failed to read buf for request public key %v", err)
	}
	this.ReqPK, err = keypair.DeserializePublicKey(buf)
	if err != nil {
		return fmt.Errorf("failed to deserialize request public key %v", err)
	}

	reqSig, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("failed to read request signature %v", err)
	}
	this.ReqSig = reqSig
	return nil
}

// Hash returns EmergencyActionRequest hash
func (this *EmergencyActionRequest) Hash() common.Uint256 {
	if this.hash != nil {
		return *this.hash
	}

	buf := new(bytes.Buffer)
	buf.Write([]byte{byte(this.Reason), byte(this.Evidence)})
	serialization.WriteUint32(buf, this.ProposalBlkNum)
	this.ProposalBlk.Serialize(buf)
	serialization.WriteVarBytes(buf, keypair.SerializePublicKey(this.ProposerPK))
	serialization.WriteVarBytes(buf, this.ProposerSigOnBlk)

	for _, sig := range this.AdminSigs {
		sig.Serialize(buf)
	}

	temp := sha256.Sum256(buf.Bytes())
	hash := common.Uint256(sha256.Sum256(temp[:]))
	this.hash = &hash
	return *this.hash
}

type EmergencyReqMsg struct {
	Payload EmergencyActionRequest
}

func (this *EmergencyReqMsg) CmdType() string {
	return pc.EMERGENCY_REQ_TYPE
}

// Serialization the EmergencyReqMsg for network
func (this *EmergencyReqMsg) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	err := this.Payload.Serialize(p)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize emergencyReqMsg %v", err)
	}

	return p.Bytes(), nil
}

// Deserialization the EmergencyReqMsg from network
func (this *EmergencyReqMsg) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	err := this.Payload.Deserialize(buf)
	if err != nil {
		return err
	}
	return nil
}
