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

type RegisterCandidateParam struct {
	PeerPubkey string
	Address    common.Address
	InitPos    uint32
}

func (this *RegisterCandidateParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, request peerPubkey error!")
	}
	if err := this.Address.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, address address error!")
	}
	if err := serialization.WriteUint32(w, this.InitPos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint64, serialize initPos error!")
	}
	return nil
}

func (this *RegisterCandidateParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	address := new(common.Address)
	err = address.Deserialize(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Deserialize, deserialize address error!")
	}
	initPos, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint64, deserialize initPos error!")
	}
	this.PeerPubkey = peerPubkey
	this.Address = *address
	this.InitPos = initPos
	return nil
}

type QuitNodeParam struct {
	PeerPubkey string
	Address    common.Address
}

func (this *QuitNodeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, deserialize peerPubkey error!")
	}
	if err := this.Address.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Deserialize, deserialize address error!")
	}
	return nil
}

func (this *QuitNodeParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	address := new(common.Address)
	err = address.Deserialize(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Deserialize, deserialize address error!")
	}
	this.PeerPubkey = peerPubkey
	this.Address = *address
	return nil
}

type ApproveCandidateParam struct {
	PeerPubkey string
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

type RejectCandidateParam struct {
	PeerPubkey string
}

func (this *RejectCandidateParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
	}
	return nil
}

func (this *RejectCandidateParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type BlackNodeParam struct {
	PeerPubkey string
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
	PeerPubkey string
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

type VoteForPeerParam struct {
	Address        common.Address
	PeerPubkeyList []string
	PosList        []uint32
}

func (this *VoteForPeerParam) Serialize(w io.Writer) error {
	if len(this.PeerPubkeyList) > 1024 {
		return errors.NewErr("length of input list > 1024!")
	}
	if len(this.PeerPubkeyList) != len(this.PosList) {
		return errors.NewErr("length of PeerPubkeyList != length of PosList!")
	}
	if err := this.Address.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Serialize, serialize address error!")
	}
	if err := serialization.WriteUint32(w, uint32(len(this.PeerPubkeyList))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize peerPubkeyList length error!")
	}
	for _, v := range this.PeerPubkeyList {
		if err := serialization.WriteString(w, v); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
		}
	}
	if err := serialization.WriteUint32(w, uint32(len(this.PosList))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize posList length error!")
	}
	for _, v := range this.PosList {
		if err := serialization.WriteUint32(w, v); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize pos error!")
		}
	}
	return nil
}

func (this *VoteForPeerParam) Deserialize(r io.Reader) error {
	address := new(common.Address)
	err := address.Deserialize(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Deserialize, deserialize address error!")
	}
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize peerPubkeyList length error!")
	}
	if n > 1024 {
		return errors.NewErr("length of input list > 1024!")
	}
	peerPubkeyList := make([]string, 0)
	for i := 0; uint32(i) < n; i++ {
		k, err := serialization.ReadString(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
		}
		peerPubkeyList = append(peerPubkeyList, k)
	}
	m, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize posList length error!")
	}
	posList := make([]uint32, 0)
	for i := 0; uint32(i) < m; i++ {
		k, err := serialization.ReadUint32(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize pos error!")
		}
		posList = append(posList, k)
	}
	if m != n {
		return errors.NewErr("length of PeerPubkeyList != length of PosList!")
	}
	this.Address = *address
	this.PeerPubkeyList = peerPubkeyList
	this.PosList = posList
	return nil
}

type WithdrawParam struct {
	Address        common.Address
	PeerPubkeyList []string
	WithdrawList   []uint32
}

func (this *WithdrawParam) Serialize(w io.Writer) error {
	if len(this.PeerPubkeyList) > 1024 {
		return errors.NewErr("length of input list > 1024!")
	}
	if len(this.PeerPubkeyList) != len(this.WithdrawList) {
		return errors.NewErr("length of PeerPubkeyList != length of WithdrawList, contract params error!")
	}
	if err := this.Address.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Serialize, serialize address error!")
	}
	if err := serialization.WriteUint32(w, uint32(len(this.PeerPubkeyList))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize peerPubkeyList length error!")
	}
	for _, v := range this.PeerPubkeyList {
		if err := serialization.WriteString(w, v); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize peerPubkey error!")
		}
	}
	if err := serialization.WriteUint32(w, uint32(len(this.WithdrawList))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize withdrawList length error!")
	}
	for _, v := range this.WithdrawList {
		if err := serialization.WriteUint32(w, v); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize withdraw error!")
		}
	}
	return nil
}

func (this *WithdrawParam) Deserialize(r io.Reader) error {
	address := new(common.Address)
	err := address.Deserialize(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Deserialize, deserialize address error!")
	}
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize peerPubkeyList length error!")
	}
	if n > 1024 {
		return errors.NewErr("length of input list > 1024!")
	}
	peerPubkeyList := make([]string, 0)
	for i := 0; uint32(i) < n; i++ {
		k, err := serialization.ReadString(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize peerPubkey error!")
		}
		peerPubkeyList = append(peerPubkeyList, k)
	}
	m, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize withdrawList length error!")
	}
	withdrawList := make([]uint32, 0)
	for i := 0; uint32(i) < m; i++ {
		k, err := serialization.ReadUint32(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize withdraw error!")
		}
		withdrawList = append(withdrawList, k)
	}
	if m != n {
		return errors.NewErr("length of PeerPubkeyList != length of WithdrawList, contract params error!")
	}
	this.Address = *address
	this.PeerPubkeyList = peerPubkeyList
	this.WithdrawList = withdrawList
	return nil
}

type Configuration struct {
	N                    uint32
	C                    uint32
	K                    uint32
	L                    uint32
	BlockMsgDelay        uint32
	HashMsgDelay         uint32
	PeerHandshakeTimeout uint32
	MaxBlockChangeView   uint32
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
	c, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize c error!")
	}
	k, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize k error!")
	}
	l, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize l error!")
	}
	blockMsgDelay, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize blockMsgDelay error!")
	}
	hashMsgDelay, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize hashMsgDelay error!")
	}
	peerHandshakeTimeout, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize peerHandshakeTimeout error!")
	}
	maxBlockChangeView, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize maxBlockChangeView error!")
	}
	this.N = n
	this.C = c
	this.K = k
	this.L = l
	this.BlockMsgDelay = blockMsgDelay
	this.HashMsgDelay = hashMsgDelay
	this.PeerHandshakeTimeout = peerHandshakeTimeout
	this.MaxBlockChangeView = maxBlockChangeView
	return nil
}

type GlobalParam struct {
	CandidateFee uint32
	MinInitStake uint32
	CandidateNum uint32
	A            uint32
	B            uint32
	Yita         uint32
}

func (this *GlobalParam) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, this.CandidateFee); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize candidateFee error!")
	}
	if err := serialization.WriteUint32(w, this.MinInitStake); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize minInitStake error!")
	}
	if err := serialization.WriteUint32(w, this.CandidateNum); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize candidateNum error!")
	}
	if err := serialization.WriteUint32(w, this.A); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize a error!")
	}
	if err := serialization.WriteUint32(w, this.B); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize b error!")
	}
	if err := serialization.WriteUint32(w, this.Yita); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize yita error!")
	}
	return nil
}

func (this *GlobalParam) Deserialize(r io.Reader) error {
	candidateFee, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize candidateFee error!")
	}
	minInitStake, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize minInitStake error!")
	}
	candidateNum, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize candidateNum error!")
	}
	a, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize a error!")
	}
	b, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize b error!")
	}
	yita, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize yita error!")
	}
	this.CandidateFee = candidateFee
	this.MinInitStake = minInitStake
	this.CandidateNum = candidateNum
	this.A = a
	this.B = b
	this.Yita = yita
	return nil
}
