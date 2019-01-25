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
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

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

type ConfigShardParam struct {
	ShardID           uint64         `json:"shard_id"`
	NetworkMin        uint32         `json:"network_min"`
	StakeAssetAddress common.Address `json:"stake_asset_address"`
	GasAssetAddress   common.Address `json:"gas_asset_address"`
	ConfigTestData    []byte         `json:"config_test_data"`
}

func (this *ConfigShardParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ConfigShardParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

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

type ActivateShardParam struct {
	ShardID uint64 `json:"shard_id"`
}

func (this *ActivateShardParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ActivateShardParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}
