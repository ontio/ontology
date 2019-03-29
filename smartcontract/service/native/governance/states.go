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

package governance

import (
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
)

type BlackListItem struct {
	PeerPubkey string         //peerPubkey in black list
	Address    common.Address //the owner of this peer
	InitPos    uint64         //initPos of this peer
}

func (this *BlackListItem) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	if err := this.Address.Serialize(w); err != nil {
		return fmt.Errorf("address.Serialize, serialize address error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.InitPos); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize initPos error: %v", err)
	}
	return nil
}

func (this *BlackListItem) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address := new(common.Address)
	err = address.Deserialize(r)
	if err != nil {
		return fmt.Errorf("address.Deserialize, deserialize address error: %v", err)
	}
	initPos, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize initPos error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.Address = *address
	this.InitPos = initPos
	return nil
}

type AuthorizeInfo struct {
	PeerPubkey           string
	Address              common.Address
	ConsensusPos         uint64 //pos deposit in consensus node
	CandidatePos         uint64 //pos deposit in candidate node
	NewPos               uint64 //deposit new pos to consensus or candidate node, it will be calculated in next epoch, you can withdrawal it at any time
	WithdrawConsensusPos uint64 //unAuthorized pos from consensus pos, frozen until next next epoch
	WithdrawCandidatePos uint64 //unAuthorized pos from candidate pos, frozen until next epoch
	WithdrawUnfreezePos  uint64 //unfrozen pos, can withdraw at any time
}

func (this *AuthorizeInfo) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, request peerPubkey error: %v", err)
	}
	if err := this.Address.Serialize(w); err != nil {
		return fmt.Errorf("address.Serialize, serialize address error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.ConsensusPos); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize consensusPos error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.CandidatePos); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize candidatePos error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.NewPos); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize newPos error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.WithdrawConsensusPos); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize withdrawConsensusPos error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.WithdrawCandidatePos); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize withdrawCandidatePos error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.WithdrawUnfreezePos); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize withDrawUnfreezePos error: %v", err)
	}
	return nil
}

func (this *AuthorizeInfo) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address := new(common.Address)
	err = address.Deserialize(r)
	if err != nil {
		return fmt.Errorf("address.Deserialize, deserialize address error: %v", err)
	}
	consensusPos, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize consensusPos error: %v", err)
	}
	candidatePos, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize candidatePos error: %v", err)
	}
	newPos, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize newPos error: %v", err)
	}
	withDrawConsensusPos, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize withDrawConsensusPos error: %v", err)
	}
	withDrawCandidatePos, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize withDrawCandidatePos error: %v", err)
	}
	withDrawUnfreezePos, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize withDrawUnfreezePos error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.Address = *address
	this.ConsensusPos = consensusPos
	this.CandidatePos = candidatePos
	this.NewPos = newPos
	this.WithdrawConsensusPos = withDrawConsensusPos
	this.WithdrawCandidatePos = withDrawCandidatePos
	this.WithdrawUnfreezePos = withDrawUnfreezePos
	return nil
}

type PeerStakeInfo struct {
	Index      uint32
	PeerPubkey string
	Stake      uint64
}

type TotalStake struct {
	//table record each address's total stake in this contract
	Address    common.Address
	Stake      uint64
	TimeOffset uint32
}

func (this *TotalStake) Serialize(w io.Writer) error {
	if err := this.Address.Serialize(w); err != nil {
		return fmt.Errorf("address.Serialize, serialize address error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.Stake); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize stake error: %v", err)
	}
	if err := serialization.WriteUint32(w, this.TimeOffset); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize timeOffset error: %v", err)
	}
	return nil
}

func (this *TotalStake) Deserialize(r io.Reader) error {
	address := new(common.Address)
	err := address.Deserialize(r)
	if err != nil {
		return fmt.Errorf("address.Deserialize, deserialize address error: %v", err)
	}
	stake, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize stake error: %v", err)
	}
	timeOffset, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize timeOffset error: %v", err)
	}
	this.Address = *address
	this.Stake = stake
	this.TimeOffset = timeOffset
	return nil
}

type PenaltyStake struct {
	//table record penalty stake of peer
	PeerPubkey   string //peer pubKey of penalty stake
	InitPos      uint64 //initPos penalty
	AuthorizePos uint64 //authorize pos penalty
	TimeOffset   uint32 //time used for calculate unbound ong
	Amount       uint64 //unbound ong that this penalty unbounded
}

func (this *PenaltyStake) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.InitPos); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize initPos error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.AuthorizePos); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize authorizePos error: %v", err)
	}
	if err := serialization.WriteUint32(w, this.TimeOffset); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize timeOffset error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.Amount); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize amount error: %v", err)
	}
	return nil
}

func (this *PenaltyStake) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	initPos, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize initPos error: %v", err)
	}
	authorizePos, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize authorizePos error: %v", err)
	}
	timeOffset, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize timeOffset error: %v", err)
	}
	amount, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize amount error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.InitPos = initPos
	this.AuthorizePos = authorizePos
	this.TimeOffset = timeOffset
	this.Amount = amount
	return nil
}

type CandidateSplitInfo struct {
	PeerPubkey string
	Address    common.Address
	InitPos    uint64
	Stake      uint64 //total stake, init pos + total pos
	S          uint64 //fee split weight of this peer
}

type PeerAttributes struct {
	PeerPubkey   string
	MaxAuthorize uint64 //max authorzie pos this peer can receive(number of ont), set by peer owner
	T2PeerCost   uint64 //candidate or consensus node doesn't share income percent with authorize users, 100 means node will take all incomes, it will take effect in view T + 2
	T1PeerCost   uint64 //candidate or consensus node doesn't share income percent with authorize users, 100 means node will take all incomes, it will take effect in view T + 1
	TPeerCost    uint64 //candidate or consensus node doesn't share income percent with authorize users, 100 means node will take all incomes, it will take effect in view T
	Field1       []byte //reserved field
	Field2       []byte //reserved field
	Field3       []byte //reserved field
	Field4       []byte //reserved field
}

func (this *PeerAttributes) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteBool, serialize peerPubkey error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.MaxAuthorize); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize maxAuthorize error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.T2PeerCost); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize oldPeerCost error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.T1PeerCost); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize newPeerCost error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.TPeerCost); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize newPeerCost error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Field1); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize field1 error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Field2); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize field2 error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Field3); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize field3 error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Field4); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize field4 error: %v", err)
	}
	return nil
}

func (this *PeerAttributes) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	maxAuthorize, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadBool, deserialize maxAuthorize error: %v", err)
	}
	t2PeerCost, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize t2PeerCost error: %v", err)
	}
	t1PeerCost, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize t1PeerCost error: %v", err)
	}
	tPeerCost, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize tPeerCost error: %v", err)
	}
	field1, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarBytes. deserialize field1 error: %v", err)
	}
	field2, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarBytes. deserialize field2 error: %v", err)
	}
	field3, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarBytes, deserialize field3 error: %v", err)
	}
	field4, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarBytes. deserialize field4 error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.MaxAuthorize = maxAuthorize
	this.T2PeerCost = t2PeerCost
	this.T1PeerCost = t1PeerCost
	this.TPeerCost = tPeerCost
	this.Field1 = field1
	this.Field2 = field2
	this.Field3 = field3
	this.Field4 = field4
	return nil
}

type SplitFeeAddress struct {
	//table record each address's ong motivation
	Address common.Address
	Amount  uint64
}

func (this *SplitFeeAddress) Serialize(w io.Writer) error {
	if err := this.Address.Serialize(w); err != nil {
		return fmt.Errorf("address.Serialize, serialize address error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.Amount); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize amount error: %v", err)
	}
	return nil
}

func (this *SplitFeeAddress) Deserialize(r io.Reader) error {
	address := new(common.Address)
	err := address.Deserialize(r)
	if err != nil {
		return fmt.Errorf("address.Deserialize, deserialize address error: %v", err)
	}
	amount, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize amount error: %v", err)
	}
	this.Address = *address
	this.Amount = amount
	return nil
}
