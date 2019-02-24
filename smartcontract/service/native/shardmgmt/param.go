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

package shardmgmt

import (
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

//
// CommonParam wrapper for all smart contract interfaces
//
type CommonParam struct {
	Input []byte
}

func (this *CommonParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.Input); err != nil {
		return fmt.Errorf("CommonParam serialize write failed: %s", err)
	}
	return nil
}

func (this *CommonParam) Deserialize(r io.Reader) error {
	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("CommonParam deserialize read failed: %s", err)
	}
	this.Input = buf
	return nil
}

//
// params for shard creation
// @ParentShardID : local shard ID
// @Creator : account address of shard creator.
// shard creator is also the shard operator after shard activated
//
type CreateShardParam struct {
	ParentShardID uint64         `json:"parent_shard_id"`
	Creator       common.Address `json:"creator"`
}

func (this *CreateShardParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *CreateShardParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

//
// params for shard configuration
// @ShardID : ID of shard which is to be configured
// @NetworkMin : min node count of shard network
// @StakeAssetAddress : contract address of token. shard is based on PoS. (ONT address)
// @GasAssetAddress : contract address of gas token. (ONG address)
// @...
//

type ConfigShardParam struct {
	ShardID           uint64             `json:"shard_id"`
	NetworkMin        uint32             `json:"network_min"`
	StakeAssetAddress common.Address     `json:"stake_asset_address"`
	GasAssetAddress   common.Address     `json:"gas_asset_address"`
	VbftConfigData    *config.VBFTConfig `json:"vbft_config_data"`
}

func (this *ConfigShardParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ConfigShardParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

//
// param for peer join shard request
// @ShardID : ID of shard which peer node is going to join
// @PeerOwner : wallet address of peer owner (to pay stake token)
// @PeerAddress : wallet address for peer fee split (to get paid gas)
// @PeerPubKey : peer public key, to verify message signatures sent from peer
// @StakeAmount : amount of token stake for the peer
//
type JoinShardParam struct {
	ShardID     uint64         `json:"shard_id"`
	PeerOwner   common.Address `json:"peer_owner"`
	PeerAddress string         `json:"peer_address"`
	PeerPubKey  string         `json:"peer_pub_key"`
	StakeAmount uint64         `json:"stake_amount"`
}

func (this *JoinShardParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *JoinShardParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

//
// param of shard-activation request
// The request can only be initiated by operator of the shard
// @ShardID : ID of shard which is to be activated
//
type ActivateShardParam struct {
	ShardID uint64 `json:"shard_id"`
}

func (this *ActivateShardParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ActivateShardParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}
