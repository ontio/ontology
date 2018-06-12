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
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
)

type Status int

func (this *Status) Serialize(w io.Writer) error {
	if err := serialization.WriteUint8(w, uint8(*this)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint8, serialize status error!")
	}
	return nil
}

func (this *Status) Deserialize(r io.Reader) error {
	status, err := serialization.ReadUint8(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint8, deserialize status error!")
	}
	*this = Status(status)
	return nil
}

type BlackListItem struct {
	PeerPubkey string
	Address    common.Address
	InitPos    uint64
}

func (this *BlackListItem) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
	}
	if err := this.Address.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Serialize, serialize address error!")
	}
	if err := serialization.WriteUint64(w, this.InitPos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize initPos error!")
	}
	return nil
}

func (this *BlackListItem) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	address := new(common.Address)
	err = address.Deserialize(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Deserialize, deserialize address error!")
	}
	initPos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64, deserialize initPos error!")
	}
	this.PeerPubkey = peerPubkey
	this.Address = *address
	this.InitPos = initPos
	return nil
}

type PeerPoolList struct {
	Peers []*PeerPoolItem
}

type PeerPoolMap struct {
	PeerPoolMap map[string]*PeerPoolItem
}

func (this *PeerPoolMap) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(this.PeerPoolMap))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize PeerPoolMap length error!")
	}
	for _, v := range this.PeerPoolMap {
		if err := v.Serialize(w); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialize peerPool error!")
		}
	}
	return nil
}

func (this *PeerPoolMap) Deserialize(r io.Reader) error {
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize PeerPoolMap length error!")
	}
	peerPoolMap := make(map[string]*PeerPoolItem)
	for i := 0; uint32(i) < n; i++ {
		peerPoolItem := new(PeerPoolItem)
		if err := peerPoolItem.Deserialize(r); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize peerPool error!")
		}
		peerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem
	}
	this.PeerPoolMap = peerPoolMap
	return nil
}

type PeerPoolItem struct {
	Index      uint32
	PeerPubkey string
	Address    common.Address
	Status     Status
	InitPos    uint64
	TotalPos   uint64
}

func (this *PeerPoolItem) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, this.Index); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize address error!")
	}
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
	}
	if err := this.Address.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Serialize, serialize address error!")
	}
	if err := this.Status.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "this.Status.Serialize, serialize Status error!")
	}
	if err := serialization.WriteUint64(w, this.InitPos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize initPos error!")
	}
	if err := serialization.WriteUint64(w, this.TotalPos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize totalPos error!")
	}
	return nil
}

func (this *PeerPoolItem) Deserialize(r io.Reader) error {
	index, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize index error!")
	}
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	address := new(common.Address)
	err = address.Deserialize(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Deserialize, deserialize address error!")
	}
	status := new(Status)
	err = status.Deserialize(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "status.Deserialize. deserialize status error!")
	}
	initPos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64, deserialize initPos error!")
	}
	totalPos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64, deserialize totalPos error!")
	}
	this.Index = index
	this.PeerPubkey = peerPubkey
	this.Address = *address
	this.Status = *status
	this.InitPos = initPos
	this.TotalPos = totalPos
	return nil
}

type VoteInfo struct {
	PeerPubkey          string
	Address             common.Address
	ConsensusPos        uint64
	FreezePos           uint64
	NewPos              uint64
	WithdrawPos         uint64
	WithdrawFreezePos   uint64
	WithdrawUnfreezePos uint64
}

func (this *VoteInfo) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, request peerPubkey error!")
	}
	if err := this.Address.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Serialize, serialize address error!")
	}
	if err := serialization.WriteUint64(w, this.ConsensusPos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize consensusPos error!")
	}
	if err := serialization.WriteUint64(w, this.FreezePos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize freezePos error!")
	}
	if err := serialization.WriteUint64(w, this.NewPos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize newPos error!")
	}
	if err := serialization.WriteUint64(w, this.WithdrawPos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize withDrawPos error!")
	}
	if err := serialization.WriteUint64(w, this.WithdrawFreezePos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize withDrawFreezePos error!")
	}
	if err := serialization.WriteUint64(w, this.WithdrawUnfreezePos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize withDrawUnfreezePos error!")
	}
	return nil
}

func (this *VoteInfo) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	address := new(common.Address)
	err = address.Deserialize(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Deserialize, deserialize address error!")
	}
	consensusPos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize consensusPos error!")
	}
	freezePos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize freezePos error!")
	}
	newPos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize newPos error!")
	}
	withDrawPos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize withDrawPos error!")
	}
	withDrawFreezePos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize withDrawFreezePos error!")
	}
	withDrawUnfreezePos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize withDrawUnfreezePos error!")
	}
	this.PeerPubkey = peerPubkey
	this.Address = *address
	this.ConsensusPos = consensusPos
	this.FreezePos = freezePos
	this.NewPos = newPos
	this.WithdrawPos = withDrawPos
	this.WithdrawFreezePos = withDrawFreezePos
	this.WithdrawUnfreezePos = withDrawUnfreezePos
	return nil
}

type PeerStakeInfo struct {
	Index      uint32
	PeerPubkey string
	Stake      uint64
}

type GovernanceView struct {
	View   uint32
	Height uint32
	TxHash common.Uint256
}

func (this *GovernanceView) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, this.View); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize view error!")
	}
	if err := serialization.WriteUint32(w, this.Height); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteBool, serialize height error!")
	}
	if err := this.TxHash.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "txHash.Serialize, serialize txHash error!")
	}
	return nil
}

func (this *GovernanceView) Deserialize(r io.Reader) error {
	view, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize view error!")
	}
	height, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize height error!")
	}
	txHash := new(common.Uint256)
	if err := txHash.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "txHash.Deserialize, deserialize txHash error!")
	}
	this.View = view
	this.Height = height
	this.TxHash = *txHash
	return nil
}

type TotalStake struct {
	Address    common.Address
	Stake      uint64
	TimeOffset uint32
}

func (this *TotalStake) Serialize(w io.Writer) error {
	if err := this.Address.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Serialize, serialize address error!")
	}
	if err := serialization.WriteUint64(w, this.Stake); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize stake error!")
	}
	if err := serialization.WriteUint32(w, this.TimeOffset); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize timeOffset error!")
	}
	return nil
}

func (this *TotalStake) Deserialize(r io.Reader) error {
	address := new(common.Address)
	err := address.Deserialize(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Deserialize, deserialize address error!")
	}
	stake, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64, deserialize stake error!")
	}
	timeOffset, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64, deserialize timeOffset error!")
	}
	this.Address = *address
	this.Stake = stake
	this.TimeOffset = timeOffset
	return nil
}

type PenaltyStake struct {
	PeerPubkey string
	InitPos    uint64
	VotePos    uint64
	TimeOffset uint32
	Amount     uint64
}

func (this *PenaltyStake) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, request peerPubkey error!")
	}
	if err := serialization.WriteUint64(w, this.InitPos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize initPos error!")
	}
	if err := serialization.WriteUint64(w, this.VotePos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize votePos error!")
	}
	if err := serialization.WriteUint32(w, this.TimeOffset); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize timeOffset error!")
	}
	if err := serialization.WriteUint64(w, this.Amount); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize amount error!")
	}
	return nil
}

func (this *PenaltyStake) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	initPos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize initPos error!")
	}
	votePos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize votePos error!")
	}
	timeOffset, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64, deserialize timeOffset error!")
	}
	amount, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize amount error!")
	}
	this.PeerPubkey = peerPubkey
	this.InitPos = initPos
	this.VotePos = votePos
	this.TimeOffset = timeOffset
	this.Amount = amount
	return nil
}

type CandidateSplitInfo struct {
	PeerPubkey string
	Address    common.Address
	InitPos    uint64
	Stake      uint64
	S          uint64
}

type SyncNodeSplitInfo struct {
	PeerPubkey string
	Address    common.Address
	InitPos    uint64
	S          uint64
}
