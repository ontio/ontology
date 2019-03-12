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
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/core/types"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

const (
	SHARD_CREATE_FEE = 100 * 1000000000 // 100 ong
)

const (
	SHARD_STATE_CREATED    = iota
	SHARD_STATE_CONFIGURED  // all parameter configured
	SHARD_STATE_ACTIVE      // started
	SHARD_STATE_STOPPING    // started
	SHARD_STATE_ARCHIVED
)

type NodeType uint

const (
	CONDIDATE_NODE = iota
	CONSENSUS_NODE
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
	Index            uint32         `json:"index"`
	PeerOwner        common.Address `json:"peer_owner"`
	PeerPubKey       string         `json:"peer_pub_key"`
	StakeAmount      uint64         `json:"stake_amount"`
	UserStakeAmount  uint64         `json:"user_stake_amount"`
	NodeType         NodeType       `json:"node_type"`
	MaxAuthorization uint64         `json:"max_authorization"`
	Proportion       uint64         `json:"proportion"`
}

type ShardState struct {
	ShardID             types.ShardID  `json:"shard_id"`
	Creator             common.Address `json:"creator"`
	State               uint32         `json:"state"`
	GenesisParentHeight uint32         `json:"genesis_parent_height"`
	Config              *ShardConfig   `json:"config"`

	Peers map[keypair.PublicKey]*PeerShardStakeInfo `json:"peers"`
}

func (this *ShardState) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardState) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

func (this *ShardState) GetPeerStakeInfo(pubKey string) (*PeerShardStakeInfo, keypair.PublicKey, error) {
	pubKeyData, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, nil, fmt.Errorf("GetPeerStakeInfo: decode param pub key failed, err: %s", err)
	}
	paramPubkey, err := keypair.DeserializePublicKey(pubKeyData)
	if err != nil {
		return nil, nil, fmt.Errorf("GetPeerStakeInfo: deserialize param pub key failed, err: %s", err)
	}
	shardPeerStakeInfo, ok := this.Peers[paramPubkey]
	if !ok {
		return nil, nil, fmt.Errorf("GetPeerStakeInfo: peer %s not exist", pubKey)
	}
	return shardPeerStakeInfo, paramPubkey, nil
}

type View uint64 // shard consensus epoch index
