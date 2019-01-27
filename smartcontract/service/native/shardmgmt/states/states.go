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
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

const (
	SHARD_STATE_CREATED    = iota
	SHARD_STATE_CONFIGURED // all parameter configured
	SHARD_STATE_ACTIVE     // started
	SHARD_STATE_STOPPING   // started
	SHARD_STATE_ARCHIVED
)

type ShardMgmtGlobalState struct {
	NextShardID uint64 `json:"next_shard_id"`
}

func (this *ShardMgmtGlobalState) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardMgmtGlobalState) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type ShardConfig struct {
	NetworkSize       uint32         `json:"network_size"`
	StakeAssetAddress common.Address `json:"stake_asset_address"`
	GasAssetAddress   common.Address `json:"gas_asset_address"`
	TestData          []byte         `json:"test_data"`
}

func (this *ShardConfig) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardConfig) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type PeerShardStakeInfo struct {
	PeerOwner   common.Address `json:"peer_owner"`
	PeerAddress string         `json:"peer_address"`
	StakeAmount uint64         `json:"stake_amount"`
}

type ShardState struct {
	ShardID             uint64                         `json:"shard_id"`
	ParentShardID       uint64                         `json:"parent_shard_id"`
	Creator             common.Address                 `json:"creator"`
	State               uint32                         `json:"state"`
	GenesisParentHeight uint64                         `json:"genesis_parent_height"`
	Config              *ShardConfig                   `json:"config"`
	Peers               map[string]*PeerShardStakeInfo `json:"peers"`
}

func (this *ShardState) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardState) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}
