/*
 * Copyright (C) 2019 The ontology Authors
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

package shardstates

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	SHARD_CREATE_FEE  = 100 * 1000000000 // 100 ong
	DEFAULT_MIN_STAKE = 100000
)

const (
	SHARD_STATE_CREATED    = iota
	SHARD_STATE_CONFIGURED  // all parameter configured
	SHARD_STATE_ACTIVE      // started
	SHARD_STATE_STOPPING    // started
	SHARD_STATE_ARCHIVED
)

type NodeType uint8

const (
	CONDIDATE_NODE NodeType = iota
	CONSENSUS_NODE
	QUIT_CONSENSUS_NODE
	QUITING_CONSENSUS_NODE
)

type ShardMgmtGlobalState struct {
	NextSubShardIndex uint16
}

func (this *ShardMgmtGlobalState) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.NextSubShardIndex)); err != nil {
		return fmt.Errorf("serialize: write NextSubShardIndex failed, err: %s", err)
	}
	return nil
}

func (this *ShardMgmtGlobalState) Deserialize(r io.Reader) error {
	index, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read NextSubShardIndex failed, err: %s", err)
	}
	this.NextSubShardIndex = uint16(index)
	return nil
}

func (this *ShardMgmtGlobalState) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint16(this.NextSubShardIndex)
}

func (this *ShardMgmtGlobalState) Deserialization(source *common.ZeroCopySource) error {
	id, eof := source.NextUint16()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.NextSubShardIndex = id
	return nil
}

type ShardConfig struct {
	GasPrice          uint64
	GasLimit          uint64
	NetworkSize       uint32
	StakeAssetAddress common.Address
	GasAssetAddress   common.Address
	VbftCfg           *config.VBFTConfig
}

func (this *ShardConfig) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.NetworkSize)); err != nil {
		return fmt.Errorf("serialize: write net size failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.StakeAssetAddress); err != nil {
		return fmt.Errorf("serialize: write stake asset addr failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.GasAssetAddress); err != nil {
		return fmt.Errorf("serialize: write gas asset addr failed, err: %s", err)
	}
	if err := this.VbftCfg.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write config failed, err: %s", err)
	}
	return nil
}

func (this *ShardConfig) Deserialize(r io.Reader) error {
	netSize, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read net size failed, err: %s", err)
	}
	this.NetworkSize = uint32(netSize)
	this.StakeAssetAddress, err = utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("deserialize: read stake asset addr failed, err: %s", err)
	}
	this.GasAssetAddress, err = utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("deserialize: read gas asset addr failed, err: %s", err)
	}
	this.VbftCfg = &config.VBFTConfig{}
	if err := this.VbftCfg.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read config failed, err: %s", err)
	}
	return nil
}

func (this *ShardConfig) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.NetworkSize)
	sink.WriteAddress(this.StakeAssetAddress)
	sink.WriteAddress(this.GasAssetAddress)
	this.VbftCfg.Serialization(sink)
}

func (this *ShardConfig) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.NetworkSize, eof = source.NextUint32()
	this.StakeAssetAddress, eof = source.NextAddress()
	this.GasAssetAddress, eof = source.NextAddress()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.VbftCfg = &config.VBFTConfig{}
	return this.VbftCfg.Deserialization(source)
}

type PeerShardStakeInfo struct {
	Index      uint32
	IpAddress  string
	PeerOwner  common.Address
	PeerPubKey string
	NodeType   NodeType
}

func (this *PeerShardStakeInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.Index)); err != nil {
		return fmt.Errorf("serialize: write index failed, err: %s", err)
	}
	if err := serialization.WriteString(w, this.IpAddress); err != nil {
		return fmt.Errorf("serialize: write ip address failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.PeerOwner); err != nil {
		return fmt.Errorf("serialize: write peer owner failed, err: %s", err)
	}
	if err := serialization.WriteString(w, this.PeerPubKey); err != nil {
		return fmt.Errorf("serialize: write peer pub key failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.NodeType)); err != nil {
		return fmt.Errorf("serialize: write node type failed, err: %s", err)
	}
	return nil
}

func (this *PeerShardStakeInfo) Deserialize(r io.Reader) error {
	index, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read index failed, err: %s", err)
	}
	this.Index = uint32(index)
	ipAddr, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("deserialize: read ip addr failed, err: %s", err)
	}
	this.IpAddress = ipAddr
	owner, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("deserialize: read peer owner failed, err: %s", err)
	}
	this.PeerOwner = owner
	if this.PeerPubKey, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read peer pub key failed, err: %s", err)
	}
	nodeType, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read node type failed, err: %s", err)
	}
	this.NodeType = NodeType(nodeType)
	return nil
}

func (this *PeerShardStakeInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Index)
	sink.WriteString(this.IpAddress)
	sink.WriteAddress(this.PeerOwner)
	sink.WriteString(this.PeerPubKey)
	sink.WriteUint8(uint8(this.NodeType))
}

func (this *PeerShardStakeInfo) Deserialization(source *common.ZeroCopySource) error {
	var irregular, eof bool
	this.Index, eof = source.NextUint32()
	this.IpAddress, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}
	this.PeerOwner, eof = source.NextAddress()
	this.PeerPubKey, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}
	nodeType, eof := source.NextUint8()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.NodeType = NodeType(nodeType)
	return nil
}

type ShardState struct {
	ShardID             types.ShardID
	Creator             common.Address
	State               uint32
	GenesisParentHeight uint32
	Config              *ShardConfig

	Peers map[string]*PeerShardStakeInfo
}

func (this *ShardState) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.ShardID.ToUint64()); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.Creator); err != nil {
		return fmt.Errorf("serialize: write creator failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.State)); err != nil {
		return fmt.Errorf("serialize: write state failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.GenesisParentHeight)); err != nil {
		return fmt.Errorf("serialize: write genesis parent height failed, err: %s", err)
	}
	if err := this.Config.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write config failed, err: %s", err)
	}
	peersNum := uint64(len(this.Peers))
	if err := utils.WriteVarUint(w, peersNum); err != nil {
		return fmt.Errorf("serialize: write peers num failed, err: %s", err)
	}
	peers := make([]*PeerShardStakeInfo, 0)
	for _, peer := range this.Peers {
		peers = append(peers, peer)
	}
	sort.SliceStable(peers, func(i, j int) bool {
		return peers[i].PeerPubKey < peers[j].PeerPubKey
	})
	for _, peer := range peers {
		if err := peer.Serialize(w); err != nil {
			return fmt.Errorf("serialize: write peer failed, index %d, err: %s", peer.Index, err)
		}
	}
	return nil
}

func (this *ShardState) Deserialize(r io.Reader) error {
	id, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	shardId, err := types.NewShardID(id)
	if err != nil {
		return fmt.Errorf("deserialize: generate shard id failed, err: %s", err)
	}
	this.ShardID = shardId
	if this.Creator, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read creator failed, err: %s", err)
	}
	state, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read state failed, err: %s", err)
	}
	this.State = uint32(state)
	height, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read genesis parent height failed, err: %s", err)
	}
	this.GenesisParentHeight = uint32(height)
	this.Config = &ShardConfig{}
	if err := this.Config.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read shard config failed, err: %s", err)
	}
	peersNum, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read peers num failed, err: %s", err)
	}
	peers := make(map[string]*PeerShardStakeInfo)
	for i := uint64(0); i < peersNum; i++ {
		peer := &PeerShardStakeInfo{}
		if err := peer.Deserialize(r); err != nil {
			return fmt.Errorf("deserialize: read peer failed, index %d, err: %s", i, err)
		}
		peers[peer.PeerPubKey] = peer
	}
	this.Peers = peers
	return nil
}

func (this *ShardState) Serialization(sink *common.ZeroCopySink) {
	utils.SerializationShardId(sink, this.ShardID)
	sink.WriteAddress(this.Creator)
	sink.WriteUint32(this.State)
	sink.WriteUint32(this.GenesisParentHeight)
	this.Config.Serialization(sink)
	sink.WriteUint64(uint64(len(this.Peers)))
	peers := make([]*PeerShardStakeInfo, 0)
	for _, peer := range this.Peers {
		peers = append(peers, peer)
	}
	sort.SliceStable(peers, func(i, j int) bool {
		return peers[i].PeerPubKey < peers[j].PeerPubKey
	})
	for _, peer := range peers {
		peer.Serialization(sink)
	}
}

func (this *ShardState) Deserialization(source *common.ZeroCopySource) error {
	shardId, err := utils.DeserializationShardId(source)
	if err != nil {
		return fmt.Errorf("dese shardId: %s", err)
	}
	var eof bool
	this.ShardID = shardId
	this.Creator, eof = source.NextAddress()
	this.State, eof = source.NextUint32()
	this.GenesisParentHeight, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Config = &ShardConfig{}
	err = this.Config.Deserialization(source)
	if err != nil {
		return fmt.Errorf("dese config: %s", err)
	}
	peersNum, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Peers = make(map[string]*PeerShardStakeInfo)
	for i := uint64(0); i < peersNum; i++ {
		peer := &PeerShardStakeInfo{}
		if err := peer.Deserialization(source); err != nil {
			return fmt.Errorf("read peer, index %d, err: %s", i, err)
		}
		this.Peers[strings.ToLower(peer.PeerPubKey)] = peer
	}
	return nil
}
