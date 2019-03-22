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
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"io"
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

const (
	SHARD_CREATE_FEE  = 100 * 1000000000 // 100 ong
	DEFAULT_MIN_STAKE = 100000
)

const (
	SHARD_STATE_CREATED    = iota
	SHARD_STATE_CONFIGURED // all parameter configured
	SHARD_STATE_ACTIVE     // started
	SHARD_STATE_STOPPING   // started
	SHARD_STATE_ARCHIVED
)

type NodeType uint

const (
	CONDIDATE_NODE = iota
	CONSENSUS_NODE
	QUIT_CONSENSUS_NODE
	QUITING_CONSENSUS_NODE
)

type ShardMgmtGlobalState struct {
	NextSubShardIndex uint16 `json:"next_sub_shard_index"`
}

// FIXME: replace all json marshal

func (this *ShardMgmtGlobalState) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardMgmtGlobalState) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type ShardConfig struct {
	NetworkSize       uint32             `json:"network_size"`
	StakeAssetAddress common.Address     `json:"stake_asset_address"`
	GasAssetAddress   common.Address     `json:"gas_asset_address"`
	VbftConfigData    *config.VBFTConfig `json:"vbft_config_data"`
}

func (this *ShardConfig) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardConfig) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type PeerShardStakeInfo struct {
	Index      uint32         `json:"index"`
	PeerOwner  common.Address `json:"peer_owner"`
	PeerPubKey string         `json:"peer_pub_key"`
	NodeType   NodeType       `json:"node_type"`
}

func (this *PeerShardStakeInfo) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.Index)); err != nil {
		return fmt.Errorf("serialize: write index failed, err: %s", err)
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
		return peers[i].Index < peers[j].Index
	})
	for _, peer := range peers {
		if err := peer.Serialize(w); err != nil {
			return fmt.Errorf("serialzie: write peer failed, index %d, err: %s", peer.Index, err)
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
