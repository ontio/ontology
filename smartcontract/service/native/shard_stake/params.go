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

package shard_stake

import (
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type InitShardParam struct {
	ShardId        common.ShardID
	StakeAssetAddr common.Address
	MinStake       uint64
}

func (this *InitShardParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.StakeAssetAddr); err != nil {
		return fmt.Errorf("serialize: write stake asset addr failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.MinStake); err != nil {
		return fmt.Errorf("serialize: write min stake failed, err: %s", err)
	}
	return nil
}

func (this *InitShardParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.StakeAssetAddr, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read stake asset addr failed, err: %s", err)
	}
	if this.MinStake, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read min stake failed, err: %s", err)
	}
	return nil
}

type PeerAmount struct {
	PeerPubKey string
	Amount     uint64
}

func (this *PeerAmount) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubKey); err != nil {
		return fmt.Errorf("serialize: write peer pub key failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.Amount); err != nil {
		return fmt.Errorf("serialize: write amount failed, err: %s", err)
	}
	return nil
}

func (this *PeerAmount) Deserialize(r io.Reader) error {
	var err error = nil
	if this.PeerPubKey, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read peer pub key failed, err: %s", err)
	}
	if this.Amount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read amount failed, err: %s", err)
	}
	return nil
}

type PeerStakeParam struct {
	ShardId   common.ShardID
	PeerOwner common.Address
	Value     *PeerAmount
}

func (this *PeerStakeParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.PeerOwner); err != nil {
		return fmt.Errorf("serialize: write peer owner failed, err: %s", err)
	}
	if err := this.Value.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write peer amount failed, err: %s", err)
	}
	return nil
}

func (this *PeerStakeParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.PeerOwner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read peer owner failed, err: %s", err)
	}
	this.Value = &PeerAmount{}
	if err := this.Value.Deserialize(r); err != nil {
		return fmt.Errorf("serialize: read peer amount failed, err: %s", err)
	}
	return nil
}

type UnfreezeFromShardParam struct {
	ShardId common.ShardID
	User    common.Address
	Value   []*PeerAmount
}

func (this *UnfreezeFromShardParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write user addr failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(len(this.Value))); err != nil {
		return fmt.Errorf("serialize: write value len failed, err: %s", err)
	}
	for index, value := range this.Value {
		if err := value.Serialize(w); err != nil {
			return fmt.Errorf("serialize: write value failed, index %d, err: %s", index, err)
		}
	}
	return nil
}

func (this *UnfreezeFromShardParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user addr failed, err: %s", err)
	}
	num, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read value len failed, err: %s", err)
	}
	this.Value = make([]*PeerAmount, num)
	for i := uint64(0); i < num; i++ {
		value := &PeerAmount{}
		if err := value.Deserialize(r); err != nil {
			return fmt.Errorf("deserialize: read value failed, index %d, err: %s", i, err)
		}
		this.Value[i] = value
	}
	return nil
}

type WithdrawStakeAssetParam struct {
	ShardId common.ShardID
	User    common.Address
}

func (this *WithdrawStakeAssetParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write address failed, err: %s", err)
	}
	return nil
}

func (this *WithdrawStakeAssetParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read address failed, err: %s", err)
	}
	return nil
}

type WithdrawFeeParam struct {
	ShardId common.ShardID
	User    common.Address
}

func (this *WithdrawFeeParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write address failed, err: %s", err)
	}
	return nil
}

func (this *WithdrawFeeParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read address failed, err: %s", err)
	}
	return nil
}

// only can be invoked by shard call while shard commit dpos, so use self-define zero copy serialize
type CommitDposParam struct {
	ShardId   common.ShardID
	Height    uint32
	Hash      common.Uint256
	FeeAmount uint64
	Debt      map[common.ShardID]map[View]uint64 // should pay handling fee to other shard
	Income    map[common.ShardID]map[View]uint64 // should receive handling fee from other shard
}

func (this *CommitDposParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteShardID(this.ShardId)
	sink.WriteUint32(this.Height)
	sink.WriteHash(this.Hash)
	sink.WriteUint64(this.FeeAmount)
	sink.WriteUint64(uint64(len(this.Debt)))
	for shard, viewInfo := range this.Debt {
		sink.WriteShardID(shard)
		sink.WriteUint64(uint64(len(viewInfo)))
		for view, fee := range viewInfo {
			sink.WriteUint32(uint32(view))
			sink.WriteUint64(fee)
		}
	}
	sink.WriteUint64(uint64(len(this.Income)))
	for shard, viewInfo := range this.Income {
		sink.WriteShardID(shard)
		sink.WriteUint64(uint64(len(viewInfo)))
		for view, fee := range viewInfo {
			sink.WriteUint32(uint32(view))
			sink.WriteUint64(fee)
		}
	}
}

func (this *CommitDposParam) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	var err error = nil
	this.ShardId, err = source.NextShardID()
	if err != nil {
		return fmt.Errorf("deseialization: read shard id failed, err: %s", err)
	}
	this.Height, eof = source.NextUint32()
	this.Hash, eof = source.NextHash()
	this.FeeAmount, eof = source.NextUint64()
	debtNum, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("deseialization: %s", io.ErrUnexpectedEOF)
	}
	this.Debt = make(map[common.ShardID]map[View]uint64)
	for i := uint64(0); i < debtNum; i++ {
		shard, err := source.NextShardID()
		if err != nil {
			return fmt.Errorf("deseialization: read debt shard id failed, index %d, err: %s", i, err)
		}
		num, eof := source.NextUint64()
		if eof {
			return io.ErrUnexpectedEOF
		}
		viewInfo := make(map[View]uint64)
		for i := uint64(0); i < num; i++ {
			view, eof := source.NextUint32()
			fee, eof := source.NextUint64()
			if eof {
				return fmt.Errorf("deseialization: read debt view fee info failed, shard %d, index %d, err: %s",
					shard.ToUint64(), i, err)
			}
			viewInfo[View(view)] = fee
		}
		this.Debt[shard] = viewInfo
	}
	incomeNum, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("deseialization: %s", io.ErrUnexpectedEOF)
	}
	this.Income = make(map[common.ShardID]map[View]uint64)
	for i := uint64(0); i < incomeNum; i++ {
		shard, err := source.NextShardID()
		if err != nil {
			return fmt.Errorf("deseialization: read income shard id failed, index %d, err: %s", i, err)
		}
		num, eof := source.NextUint64()
		if eof {
			return io.ErrUnexpectedEOF
		}
		viewInfo := make(map[View]uint64)
		for i := uint64(0); i < num; i++ {
			view, eof := source.NextUint32()
			fee, eof := source.NextUint64()
			if eof {
				return fmt.Errorf("deseialization: read incom view fee info failed, shard %d, index %d, err: %s",
					shard.ToUint64(), i, err)
			}
			viewInfo[View(view)] = fee
		}
		this.Income[shard] = viewInfo
	}
	return nil
}

type SetMinStakeParam struct {
	ShardId common.ShardID
	Amount  uint64
}

func (this *SetMinStakeParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.Amount); err != nil {
		return fmt.Errorf("serialize: write amount failed, err: %s", err)
	}
	return nil
}

func (this *SetMinStakeParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.Amount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read amount failed, err: %s", err)
	}
	return nil
}

type UserStakeParam struct {
	ShardId common.ShardID
	User    common.Address
	Value   []*PeerAmount
}

func (this *UserStakeParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write user addr failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(len(this.Value))); err != nil {
		return fmt.Errorf("serialize: write value len failed, err: %s", err)
	}
	for index, value := range this.Value {
		if err := value.Serialize(w); err != nil {
			return fmt.Errorf("serialize: write value failed, index %d, err: %s", index, err)
		}
	}
	return nil
}

func (this *UserStakeParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user addr failed, err: %s", err)
	}
	num, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read value len failed, err: %s", err)
	}
	this.Value = make([]*PeerAmount, num)
	for i := uint64(0); i < num; i++ {
		value := &PeerAmount{}
		if err := value.Deserialize(r); err != nil {
			return fmt.Errorf("deserialize: read value failed, index %d, err: %s", i, err)
		}
		this.Value[i] = value
	}
	return nil
}

type ChangeMaxAuthorizationParam struct {
	ShardId common.ShardID
	User    common.Address
	Value   *PeerAmount
}

func (this *ChangeMaxAuthorizationParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write addr failed, err: %s", err)
	}
	if err := this.Value.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write value failed, err: %s", err)
	}
	return nil
}

func (this *ChangeMaxAuthorizationParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user addr failed, err: %s", err)
	}
	this.Value = &PeerAmount{}
	if err = this.Value.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read value failed, err: %s", err)
	}
	return nil
}

type ChangeProportionParam struct {
	ShardId common.ShardID
	User    common.Address
	Value   *PeerAmount
}

func (this *ChangeProportionParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write addr failed, err: %s", err)
	}
	if err := this.Value.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write value failed, err: %s", err)
	}
	return nil
}

func (this *ChangeProportionParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user addr failed, err: %s", err)
	}
	this.Value = &PeerAmount{}
	if err = this.Value.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read value failed, err: %s", err)
	}
	return nil
}

type DeletePeerParam struct {
	ShardId common.ShardID
	Peers   []string
}

func (this *DeletePeerParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(len(this.Peers))); err != nil {
		return fmt.Errorf("serialize: write peer num failed, err: %s", err)
	}
	for index, peer := range this.Peers {
		if err := serialization.WriteString(w, peer); err != nil {
			return fmt.Errorf("serialize: write pub key failed, index %d, err: %s", index, err)
		}
	}
	return nil
}

func (this *DeletePeerParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	peersNum, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read peers num failed, err: %s", err)
	}
	peers := make([]string, 0)
	for i := uint64(0); i < peersNum; i++ {
		peer, err := serialization.ReadString(r)
		if err != nil {
			return fmt.Errorf("deserialize: deserialize pub key failed, index %d, err: %s", i, err)
		}
		peers = append(peers, peer)
	}
	this.Peers = peers
	return nil
}

type PeerExitParam struct {
	ShardId common.ShardID
	Peer    string
}

func (this *PeerExitParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := serialization.WriteString(w, this.Peer); err != nil {
		return fmt.Errorf("serialize: write pub key failed, err: %s", err)
	}
	return nil
}

func (this *PeerExitParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}

	if this.Peer, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read peer failed, err: %s", err)
	}
	return nil
}

type GetPeerInfoParam struct {
	ShardId common.ShardID
	View    uint64
}

func (this *GetPeerInfoParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.View); err != nil {
		return fmt.Errorf("serialize: write view failed, err: %s", err)
	}
	return nil
}

func (this *GetPeerInfoParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.View, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read view failed, err: %s", err)
	}
	return nil
}

type GetUserStakeInfoParam struct {
	ShardId common.ShardID
	View    uint64
	User    common.Address
}

func (this *GetUserStakeInfoParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.View); err != nil {
		return fmt.Errorf("serialize: write view failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write addr failed, err: %s", err)
	}
	return nil
}

func (this *GetUserStakeInfoParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.View, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read view failed, err: %s", err)
	}
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read addr failed, err: %s", err)
	}
	return nil
}

type GetXShardFeeInfoParam struct {
	ShardId common.ShardID
	View    uint64
}

func (this *GetXShardFeeInfoParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.View); err != nil {
		return fmt.Errorf("serialize: write view failed, err: %s", err)
	}
	return nil
}

func (this *GetXShardFeeInfoParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.View, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read view failed, err: %s", err)
	}
	return nil
}
