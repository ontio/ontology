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
	"math/big"
	"strconv"

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

type RegisterSyncNodeParam struct {
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
	InitPos    uint64 `json:"initPos"`
}

func (this *RegisterSyncNodeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, request peerPubkey error!")
	}
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, address address error!")
	}
	if err := serialization.WriteUint64(w, this.InitPos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize initPos error!")
	}
	return nil
}

func (this *RegisterSyncNodeParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	this.PeerPubkey = peerPubkey

	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address

	initPos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64, deserialize initPos error!")
	}
	this.InitPos = initPos
	return nil
}

type ApproveSyncNodeParam struct {
	PeerPubkey string `json:"peerPubkey"`
}

func (this *ApproveSyncNodeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
	}
	return nil
}

func (this *ApproveSyncNodeParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type QuitNodeParam struct {
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
}

func (this *QuitNodeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, deserialize peerPubkey error!")
	}
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, deserialize address error!")
	}
	return nil
}

func (this *QuitNodeParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	this.PeerPubkey = peerPubkey

	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address
	return nil
}

type RegisterCandidateParam struct {
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
}

func (this *RegisterCandidateParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, deserialize peerPubkey error!")
	}
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, deserialize address error!")
	}
	return nil
}

func (this *RegisterCandidateParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	this.PeerPubkey = peerPubkey

	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address
	return nil
}

type ApproveCandidateParam struct {
	PeerPubkey string `json:"peerPubkey"`
}

func (this *ApproveCandidateParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
	}
	return nil
}

func (this *ApproveCandidateParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type BlackNodeParam struct {
	PeerPubkey string `json:"peerPubkey"`
}

func (this *BlackNodeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
	}
	return nil
}

func (this *BlackNodeParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type WhiteNodeParam struct {
	PeerPubkey string `json:"peerPubkey"`
}

func (this *WhiteNodeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
	}
	return nil
}

func (this *WhiteNodeParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type PeerPoolList struct {
	Peers []*PeerPool `json:"peers"`
}

type PeerPoolMap struct {
	PeerPoolMap map[string]*PeerPool `json:"peerPoolMap"`
}

func (this *PeerPoolMap) Serialize(w io.Writer) error {
	if err := serialization.WriteVarUint(w, uint64(len(this.PeerPoolMap))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarUint, serialize PeerPoolMap length error!")
	}
	for _, v := range this.PeerPoolMap {
		if err := v.Serialize(w); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialize peerPool error!")
		}
	}
	return nil
}

func (this *PeerPoolMap) Deserialize(r io.Reader) error {
	n, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarUint, deserialize PeerPoolMap length error!")
	}
	for i := 0; uint64(i) < n; i++ {
		peerPool := new(PeerPool)
		if err := peerPool.Deserialize(r); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize peerPool error!")
		}
		this.PeerPoolMap[peerPool.PeerPubkey] = peerPool
	}
	return nil
}

type PeerPool struct {
	Index      uint32 `json:"index"`
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
	Status     Status `json:"status"`
	InitPos    uint64 `json:"initPos"`
	TotalPos   uint64 `json:"totalPos"`
}

func (this *PeerPool) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, this.Index); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize address error!")
	}
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
	}
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize address error!")
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

func (this *PeerPool) Deserialize(r io.Reader) error {
	index, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize index error!")
	}
	this.Index = index

	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	this.PeerPubkey = peerPubkey

	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address

	status := new(Status)
	err = status.Deserialize(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "status.Deserialize. deserialize status error!")
	}
	this.Status = *status

	initPos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64, deserialize initPos error!")
	}
	this.InitPos = initPos

	totalPos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64, deserialize totalPos error!")
	}
	this.TotalPos = totalPos
	return nil
}

type VoteForPeerParam struct {
	Address   string           `json:"address"`
	VoteTable map[string]int64 `json:"voteTable"`
}

func (this *VoteForPeerParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize address error!")
	}
	if err := serialization.WriteVarUint(w, uint64(len(this.VoteTable))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarUint, serialize VoteTable length error!")
	}
	for k, v := range this.VoteTable {
		if err := serialization.WriteString(w, k); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize key error!")
		}
		if err := serialization.WriteString(w, strconv.FormatInt(v, 10)); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize value error!")
		}
	}
	return nil
}

func (this *VoteForPeerParam) Deserialize(r io.Reader) error {
	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address

	n, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarUint, deserialize VoteTable length error!")
	}
	for i := 0; uint64(i) < n; i++ {
		k, err := serialization.ReadString(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize key error!")
		}
		v, err := serialization.ReadString(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize value error!")
		}
		value, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "strconv.ParseInt, deserialize value error!")
		}
		this.VoteTable[k] = value
	}
	return nil
}

type WithDrawParam struct {
	Address       string            `json:"address"`
	WithDrawTable map[string]uint64 `json:"withDrawTable"`
}

func (this *WithDrawParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize address error!")
	}
	if err := serialization.WriteVarUint(w, uint64(len(this.WithDrawTable))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarUint, serialize WithDrawTable length error!")
	}
	for k, v := range this.WithDrawTable {
		if err := serialization.WriteString(w, k); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize key error!")
		}
		if err := serialization.WriteUint64(w, v); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize value error!")
		}
	}
	return nil
}

func (this *WithDrawParam) Deserialize(r io.Reader) error {
	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address

	n, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarUint, deserialize WithDrawTable length error!")
	}
	for i := 0; uint64(i) < n; i++ {
		k, err := serialization.ReadString(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize key error!")
		}
		v, err := serialization.ReadUint64(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64, deserialize value error!")
		}
		this.WithDrawTable[k] = v
	}
	return nil
}

type VoteInfoPool struct {
	PeerPubkey          string `json:"peerPubkey"`
	Address             string `json:"address"`
	ConsensusPos        uint64 `json:"consensusPos"`
	FreezePos           uint64 `json:"freezePos"`
	NewPos              uint64 `json:"newPos"`
	WithDrawPos         uint64 `json:"withDrawPos"`
	WithDrawFreezePos   uint64 `json:"withDrawFreezePos"`
	WithDrawUnfreezePos uint64 `json:"withDrawUnfreezePos"`
}

func (this *VoteInfoPool) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, request peerPubkey error!")
	}
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, address address error!")
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
	if err := serialization.WriteUint64(w, this.WithDrawPos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize withDrawPos error!")
	}
	if err := serialization.WriteUint64(w, this.WithDrawFreezePos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize withDrawFreezePos error!")
	}
	if err := serialization.WriteUint64(w, this.WithDrawUnfreezePos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize withDrawUnfreezePos error!")
	}
	return nil
}

func (this *VoteInfoPool) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	this.PeerPubkey = peerPubkey

	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address

	consensusPos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize consensusPos error!")
	}
	this.ConsensusPos = consensusPos

	freezePos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize freezePos error!")
	}
	this.FreezePos = freezePos

	newPos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize newPos error!")
	}
	this.NewPos = newPos

	withDrawPos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize withDrawPos error!")
	}
	this.WithDrawPos = withDrawPos

	withDrawFreezePos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize withDrawFreezePos error!")
	}
	this.WithDrawFreezePos = withDrawFreezePos

	withDrawUnfreezePos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64. deserialize withDrawUnfreezePos error!")
	}
	this.WithDrawUnfreezePos = withDrawUnfreezePos
	return nil
}

type PeerStakeInfo struct {
	Index      uint32 `json:"index"`
	PeerPubkey string `json:"peerPubkey"`
	Stake      uint64 `json:"stake"`
}

type Configuration struct {
	N                    uint32 `json:"n"`
	C                    uint32 `json:"c"`
	K                    uint32 `json:"k"`
	L                    uint32 `json:"l"`
	BlockMsgDelay        uint32 `json:"block_msg_delay"`
	HashMsgDelay         uint32 `json:"hash_msg_delay"`
	PeerHandshakeTimeout uint32 `json:"peer_handshake_timeout"`
	MaxBlockChangeView   uint32 `json:"max_block_change_view"`
}

func (this *Configuration) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, this.N); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize n error!")
	}
	if err := serialization.WriteUint32(w, this.C); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize c error!")
	}
	if err := serialization.WriteUint32(w, this.K); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize k error!")
	}
	if err := serialization.WriteUint32(w, this.L); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize l error!")
	}
	if err := serialization.WriteUint32(w, this.BlockMsgDelay); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize block_msg_delay error!")
	}
	if err := serialization.WriteUint32(w, this.HashMsgDelay); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize hash_msg_delay error!")
	}
	if err := serialization.WriteUint32(w, this.PeerHandshakeTimeout); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize peer_handshake_timeout error!")
	}
	if err := serialization.WriteUint32(w, this.MaxBlockChangeView); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize max_block_change_view error!")
	}
	return nil
}

func (this *Configuration) Deserialize(r io.Reader) error {
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize n error!")
	}
	this.N = n

	c, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize c error!")
	}
	this.C = c

	k, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize k error!")
	}
	this.K = k

	l, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize l error!")
	}
	this.L = l

	blockMsgDelay, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize blockMsgDelay error!")
	}
	this.BlockMsgDelay = blockMsgDelay

	hashMsgDelay, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize hashMsgDelay error!")
	}
	this.HashMsgDelay = hashMsgDelay

	peerHandshakeTimeout, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize peerHandshakeTimeout error!")
	}
	this.PeerHandshakeTimeout = peerHandshakeTimeout

	maxBlockChangeView, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize maxBlockChangeView error!")
	}
	this.MaxBlockChangeView = maxBlockChangeView
	return nil
}

type VoteCommitDposParam struct {
	Address string `json:"address"`
	Pos     int64  `json:"pos"`
}

func (this *VoteCommitDposParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, address address error!")
	}
	if err := serialization.WriteString(w, strconv.FormatInt(this.Pos, 10)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize pos error!")
	}
	return nil
}

func (this *VoteCommitDposParam) Deserialize(r io.Reader) error {
	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address

	pos, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize pos error!")
	}
	this.Pos, err = strconv.ParseInt(pos, 10, 64)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "strconv.ParseInt, deserialize pos error!")
	}
	return nil
}

type VoteCommitInfoPool struct {
	Address string `json:"address"`
	Pos     uint64 `json:"pos"`
}

func (this *VoteCommitInfoPool) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, address address error!")
	}
	if err := serialization.WriteUint64(w, this.Pos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize pos error!")
	}
	return nil
}

func (this *VoteCommitInfoPool) Deserialize(r io.Reader) error {
	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address

	pos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64, deserialize pos error!")
	}
	this.Pos = pos
	return nil
}

type GovernanceView struct {
	View       *big.Int `json:"view"`
	VoteCommit bool     `json:"voteCommit"`
}

func (this *GovernanceView) Serialize(w io.Writer) error {
	if err := serialization.WriteUint64(w, this.View.Uint64()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize view error!")
	}
	if err := serialization.WriteBool(w, this.VoteCommit); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteBool, serialize voteCommit error!")
	}
	return nil
}

func (this *GovernanceView) Deserialize(r io.Reader) error {
	view, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64, deserialize view error!")
	}
	this.View = new(big.Int).SetUint64(view)

	voteCommit, err := serialization.ReadBool(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadBool, deserialize voteCommit error!")
	}
	this.VoteCommit = voteCommit
	return nil
}
