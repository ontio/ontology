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
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type Status uint8

func (this *Status) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint8(uint8(*this))
}

func (this *Status) Deserialization(source *common.ZeroCopySource) error {
	status, eof := source.NextUint8()
	if eof {
		return fmt.Errorf("serialization.ReadUint8, deserialize status error: %v", io.ErrUnexpectedEOF)
	}
	*this = Status(status)
	return nil
}

type BlackListItem struct {
	PeerPubkey string         //peerPubkey in black list
	Address    common.Address //the owner of this peer
	InitPos    uint64         //initPos of this peer
}

func (this *BlackListItem) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	this.Address.Serialization(sink)
	sink.WriteUint64(this.InitPos)
}

func (this *BlackListItem) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address := new(common.Address)
	err = address.Deserialization(source)
	if err != nil {
		return fmt.Errorf("address.Deserialize, deserialize address error: %v", err)
	}
	initPos, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("serialization.ReadUint64, deserialize initPos error: %v", io.ErrUnexpectedEOF)
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

func (this *PeerPoolMap) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint32(uint32(len(this.PeerPoolMap)))

	var peerPoolItemList []*PeerPoolItem
	for _, v := range this.PeerPoolMap {
		peerPoolItemList = append(peerPoolItemList, v)
	}
	sort.SliceStable(peerPoolItemList, func(i, j int) bool {
		return peerPoolItemList[i].PeerPubkey > peerPoolItemList[j].PeerPubkey
	})
	for _, v := range peerPoolItemList {
		v.Serialization(sink)
	}
	return nil
}

func (this *PeerPoolMap) Deserialization(source *common.ZeroCopySource) error {

	n, err := utils.DecodeUint32(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize PeerPoolMap length error: %v", err)
	}
	peerPoolMap := make(map[string]*PeerPoolItem)
	for i := 0; uint32(i) < n; i++ {
		peerPoolItem := new(PeerPoolItem)
		if err := peerPoolItem.Deserialization(source); err != nil {
			return fmt.Errorf("deserialize peerPool error: %v", err)
		}
		peerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem
	}
	this.PeerPoolMap = peerPoolMap
	return nil
}

type PeerPoolListForVm struct {
	PeerPoolList []*PeerPoolItemForVm
}

func (this *PeerPoolListForVm) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(uint32(len(this.PeerPoolList)))
	for _, v := range this.PeerPoolList {
		v.Serialization(sink)
	}
}

type PeerPoolItem struct {
	Index      uint32         //peer index
	PeerPubkey string         //peer pubkey
	Address    common.Address //peer owner
	Status     Status         //peer status
	InitPos    uint64         //peer initPos
	TotalPos   uint64         //total authorize pos this peer received
}

func (this *PeerPoolItem) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Index)
	sink.WriteString(this.PeerPubkey)
	this.Address.Serialization(sink)
	this.Status.Serialization(sink)
	sink.WriteUint64(this.InitPos)
	sink.WriteUint64(this.TotalPos)
}

func (this *PeerPoolItem) Deserialization(source *common.ZeroCopySource) error {
	index, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("serialization.ReadUint32, deserialize index error: %v", io.ErrUnexpectedEOF)
	}
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address := new(common.Address)
	err = address.Deserialization(source)
	if err != nil {
		return fmt.Errorf("address.Deserialize, deserialize address error: %v", err)
	}
	status := new(Status)
	err = status.Deserialization(source)
	if err != nil {
		return fmt.Errorf("status.Deserialize. deserialize status error: %v", err)
	}
	initPos, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("serialization.ReadUint64, deserialize initPos error: %v", io.ErrUnexpectedEOF)
	}
	totalPos, err := utils.DecodeUint64(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize totalPos error: %v", err)
	}
	this.Index = index
	this.PeerPubkey = peerPubkey
	this.Address = *address
	this.Status = *status
	this.InitPos = initPos
	this.TotalPos = totalPos
	return nil
}

// readable for neovm, only used for GetPeerPool, GetPeerInfo and GetPeerPoolByAddress
type PeerPoolItemForVm struct {
	Index       uint32         //peer index
	PeerAddress common.Address //peer address
	Address     common.Address //peer owner
	Status      Status         //peer status
	InitPos     uint64         //peer initPos
	TotalPos    uint64         //total authorize pos this peer received
}

func (this *PeerPoolItemForVm) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Index)
	this.PeerAddress.Serialization(sink)
	this.Address.Serialization(sink)
	this.Status.Serialization(sink)
	sink.WriteUint64(this.InitPos)
	sink.WriteUint64(this.TotalPos)
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

func (this *AuthorizeInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	this.Address.Serialization(sink)
	sink.WriteUint64(this.ConsensusPos)
	sink.WriteUint64(this.CandidatePos)
	sink.WriteUint64(this.NewPos)
	sink.WriteUint64(this.WithdrawConsensusPos)
	sink.WriteUint64(this.WithdrawCandidatePos)
	sink.WriteUint64(this.WithdrawUnfreezePos)
}

func (this *AuthorizeInfo) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address := new(common.Address)
	err = address.Deserialization(source)
	if err != nil {
		return fmt.Errorf("address.Deserialize, deserialize address error: %v", err)
	}
	consensusPos, err := utils.DecodeUint64(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize consensusPos error: %v", err)
	}
	candidatePos, err := utils.DecodeUint64(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize candidatePos error: %v", err)
	}
	newPos, err := utils.DecodeUint64(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize newPos error: %v", err)
	}
	withDrawConsensusPos, err := utils.DecodeUint64(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize withDrawConsensusPos error: %v", err)
	}
	withDrawCandidatePos, err := utils.DecodeUint64(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize withDrawCandidatePos error: %v", err)
	}
	withDrawUnfreezePos, err := utils.DecodeUint64(source)
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

type GovernanceView struct {
	View   uint32
	Height uint32
	TxHash common.Uint256
}

func (this *GovernanceView) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, this.View); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize view error: %v", err)
	}
	if err := serialization.WriteUint32(w, this.Height); err != nil {
		return fmt.Errorf("serialization.WriteBool, serialize height error: %v", err)
	}
	if err := this.TxHash.Serialize(w); err != nil {
		return fmt.Errorf("txHash.Serialize, serialize txHash error: %v", err)
	}
	return nil
}

func (this *GovernanceView) Deserialize(r io.Reader) error {
	view, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize view error: %v", err)
	}
	height, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize height error: %v", err)
	}
	txHash := new(common.Uint256)
	if err := txHash.Deserialize(r); err != nil {
		return fmt.Errorf("txHash.Deserialize, deserialize txHash error: %v", err)
	}
	this.View = view
	this.Height = height
	this.TxHash = *txHash
	return nil
}

type TotalStake struct { //table record each address's total stake in this contract
	Address    common.Address
	Stake      uint64
	TimeOffset uint32
}

func (this *TotalStake) Serialization(sink *common.ZeroCopySink) {
	this.Address.Serialization(sink)
	sink.WriteUint64(this.Stake)
	sink.WriteUint32(this.TimeOffset)
}

func (this *TotalStake) Deserialization(source *common.ZeroCopySource) error {
	address := new(common.Address)
	err := address.Deserialization(source)
	if err != nil {
		return fmt.Errorf("address.Deserialize, deserialize address error: %v", err)
	}
	stake, err := utils.DecodeUint64(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize stake error: %v", err)
	}
	timeOffset, err := utils.DecodeUint32(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize timeOffset error: %v", err)
	}
	this.Address = *address
	this.Stake = stake
	this.TimeOffset = timeOffset
	return nil
}

type PenaltyStake struct { //table record penalty stake of peer
	PeerPubkey   string //peer pubKey of penalty stake
	InitPos      uint64 //initPos penalty
	AuthorizePos uint64 //authorize pos penalty
	TimeOffset   uint32 //time used for calculate unbound ong
	Amount       uint64 //unbound ong that this penalty unbounded
}

func (this *PenaltyStake) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	sink.WriteUint64(this.InitPos)
	sink.WriteUint64(this.AuthorizePos)
	sink.WriteUint32(this.TimeOffset)
	sink.WriteUint64(this.Amount)
}

func (this *PenaltyStake) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	initPos, err := utils.DecodeUint64(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize initPos error: %v", err)
	}
	authorizePos, err := utils.DecodeUint64(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64. deserialize authorizePos error: %v", err)
	}
	timeOffset, err := utils.DecodeUint32(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize timeOffset error: %v", err)
	}
	amount, err := utils.DecodeUint64(source)
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
	T2PeerCost   uint64 //candidate or consensus node doesn't share initpos income percent with authorize users, 100 means node will take all incomes, it will take effect in view T + 2
	T1PeerCost   uint64 //candidate or consensus node doesn't share initpos income percent with authorize users, 100 means node will take all incomes, it will take effect in view T + 1
	TPeerCost    uint64 //candidate or consensus node doesn't share initpos income percent with authorize users, 100 means node will take all incomes, it will take effect in view T
	T2StakeCost  uint64 //candidate or consensus node doesn't share stake income percent with authorize users, it will take effect in view T + 2, 101 means 0, 0 means null
	T1StakeCost  uint64 //candidate or consensus node doesn't share stake income percent with authorize users, it will take effect in view T + 1, 101 means 0, 0 means null
	TStakeCost   uint64 //candidate or consensus node doesn't share stake income percent with authorize users, it will take effect in view T, 101 means 0, 0 means null
	Field4       []byte //reserved field
}

func (this *PeerAttributes) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	sink.WriteUint64(this.MaxAuthorize)
	sink.WriteUint64(this.T2PeerCost)
	sink.WriteUint64(this.T1PeerCost)
	sink.WriteUint64(this.TPeerCost)
	utils.EncodeVarUint(sink, this.T2StakeCost)
	utils.EncodeVarUint(sink, this.T1StakeCost)
	utils.EncodeVarUint(sink, this.TStakeCost)
	sink.WriteVarBytes(this.Field4)
}

func (this *PeerAttributes) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeString, deserialize peerPubkey error: %v", err)
	}
	maxAuthorize, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("source.NextUint64, deserialize maxAuthorize error: %v", io.ErrUnexpectedEOF)
	}
	t2PeerCost, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("source.NextUint64, deserialize t2PeerCost error: %v", io.ErrUnexpectedEOF)
	}
	t1PeerCost, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("source.NextUint64, deserialize t1PeerCost error: %v", io.ErrUnexpectedEOF)
	}
	tPeerCost, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("source.NextUint64, deserialize tPeerCost error: %v", io.ErrUnexpectedEOF)
	}
	t2StakeCost, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize t2StakeCost error: %v", err)
	}
	t1StakeCost, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize t1StakeCost error: %v", err)
	}
	tStakeCost, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize tStakeCost error: %v", err)
	}
	field4, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize field4 error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.MaxAuthorize = maxAuthorize
	this.T2PeerCost = t2PeerCost
	this.T1PeerCost = t1PeerCost
	this.TPeerCost = tPeerCost
	this.T2StakeCost = t2StakeCost
	this.T1StakeCost = t1StakeCost
	this.TStakeCost = tStakeCost
	this.Field4 = field4
	return nil
}

type SplitFeeAddress struct { //table record each address's ong motivation
	Address common.Address
	Amount  uint64
}

func (this *SplitFeeAddress) Serialization(sink *common.ZeroCopySink) {
	this.Address.Serialization(sink)
	sink.WriteUint64(this.Amount)
}

func (this *SplitFeeAddress) Deserialization(source *common.ZeroCopySource) error {
	address := new(common.Address)
	err := address.Deserialization(source)
	if err != nil {
		return fmt.Errorf("address.Deserialize, deserialize address error: %v", err)
	}
	amount, err := utils.DecodeUint64(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize amount error: %v", err)
	}
	this.Address = *address
	this.Amount = amount
	return nil
}
