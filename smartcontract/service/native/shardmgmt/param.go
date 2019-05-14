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
	"bytes"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

//
// params for shard creation
// @ParentShardID : local shard ID
// @Creator : account address of shard creator.
// shard creator is also the shard operator after shard activated
//
type CreateShardParam struct {
	ParentShardID common.ShardID
	Creator       common.Address
}

func (this *CreateShardParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ParentShardID); err != nil {
		return fmt.Errorf("serialize: write parent shard id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.Creator); err != nil {
		return fmt.Errorf("serialize: write creator failed, err: %s", err)
	}
	return nil
}

func (this *CreateShardParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ParentShardID, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read parent shard id failed, err: %s", err)
	}
	if this.Creator, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read creator failed, err: %s", err)
	}
	return nil
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
	ShardID           common.ShardID
	NetworkMin        uint32
	StakeAssetAddress common.Address
	GasAssetAddress   common.Address
	GasPrice          uint64
	GasLimit          uint64
	VbftConfigData    []byte
}

func (this *ConfigShardParam) GetConfig() (*config.VBFTConfig, error) {
	cfg := &config.VBFTConfig{}
	err := cfg.Deserialize(bytes.NewReader(this.VbftConfigData))
	return cfg, err
}

func (this *ConfigShardParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardID); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.NetworkMin)); err != nil {
		return fmt.Errorf("serialize: write network min failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.StakeAssetAddress); err != nil {
		return fmt.Errorf("serialize: write stake asset addr failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.GasAssetAddress); err != nil {
		return fmt.Errorf("serialize: write gas asset addr failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.GasPrice); err != nil {
		return fmt.Errorf("serialize: write gas price failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.GasLimit); err != nil {
		return fmt.Errorf("serialize: write gas limit failed, err: %s", err)
	}
	if err := serialization.WriteVarBytes(w, this.VbftConfigData); err != nil {
		return fmt.Errorf("serialize: write cfg data failed, err: %s", err)
	}
	return nil
}

func (this *ConfigShardParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardID, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	networkMin, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read network min failed, err: %s", err)
	}
	this.NetworkMin = uint32(networkMin)
	if this.StakeAssetAddress, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read stake asset addr failed, err: %s", err)
	}
	if this.GasAssetAddress, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read gas asset addr failed, err: %s", err)
	}
	if this.GasPrice, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read gas price failed, err: %s", err)
	}
	if this.GasLimit, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read gas limit failed, err: %s", err)
	}
	if this.VbftConfigData, err = serialization.ReadVarBytes(r); err != nil {
		return fmt.Errorf("deserialize: read config data failed, err: %s", err)
	}
	return nil
}

type ApplyJoinShardParam struct {
	ShardId    common.ShardID
	PeerOwner  common.Address
	PeerPubKey string
}

func (this *ApplyJoinShardParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.PeerOwner); err != nil {
		return fmt.Errorf("serialize: write peer owner failed, err: %s", err)
	}
	if err := serialization.WriteString(w, this.PeerPubKey); err != nil {
		return fmt.Errorf("serialize: write peer pub key failed, err: %s", err)
	}
	return nil
}

func (this *ApplyJoinShardParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.PeerOwner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read peer owner failed, err: %s", err)
	}
	if this.PeerPubKey, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read peer pub key failed, err: %s", err)
	}
	return nil
}

type ApproveJoinShardParam struct {
	ShardId    common.ShardID
	PeerPubKey []string
}

func (this *ApproveJoinShardParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(len(this.PeerPubKey))); err != nil {
		return fmt.Errorf("serialize: write peers num failed, err: %s", err)
	}
	for index, peer := range this.PeerPubKey {
		if err := serialization.WriteString(w, peer); err != nil {
			return fmt.Errorf("serialize: write peer pub key failed, index %d, err: %s", index, err)
		}
	}
	return nil
}

func (this *ApproveJoinShardParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	num, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read peers num failed, err: %s", err)
	}
	this.PeerPubKey = make([]string, num)
	for i := uint64(0); i < num; i++ {
		peer, err := serialization.ReadString(r)
		if err != nil {
			return fmt.Errorf("deserialize: read peer failed, index %d, err: %s", i, err)
		}
		this.PeerPubKey[i] = peer
	}
	return nil
}

//
// param for peer join shard request
// @ShardID : ID of shard which peer node is going to join
// @PeerOwner : wallet address of peer owner (to pay stake token)
// @PeerPubKey : peer public key, to verify message signatures sent from peer, run ontology wallet account
// @StakeAmount : amount of token stake for the peer
//
type JoinShardParam struct {
	ShardID     common.ShardID
	IpAddress   string
	PeerOwner   common.Address
	PeerPubKey  string
	StakeAmount uint64
}

func (this *JoinShardParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardID); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := serialization.WriteString(w, this.IpAddress); err != nil {
		return fmt.Errorf("serialize: write ip failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.PeerOwner); err != nil {
		return fmt.Errorf("serialize: write peer owner failed, err: %s", err)
	}
	if err := serialization.WriteString(w, this.PeerPubKey); err != nil {
		return fmt.Errorf("serialize: write peer pub key failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.StakeAmount); err != nil {
		return fmt.Errorf("serialize: write peer stake amount failed, err: %s", err)
	}
	return nil
}

func (this *JoinShardParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardID, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.IpAddress, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read ip failed, err: %s", err)
	}
	if this.PeerOwner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read peer owner failed, err: %s", err)
	}
	if this.PeerPubKey, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read peer pub key failed, err: %s", err)
	}
	if this.StakeAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read stake amount failed, err: %s", err)
	}
	return nil
}

type ExitShardParam struct {
	ShardId    common.ShardID
	PeerOwner  common.Address
	PeerPubKey string
}

func (this *ExitShardParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.PeerOwner); err != nil {
		return fmt.Errorf("serialize: write peer owner failed, err: %s", err)
	}
	if err := serialization.WriteString(w, this.PeerPubKey); err != nil {
		return fmt.Errorf("serialize: write peer pub key failed, err: %s", err)
	}
	return nil
}

func (this *ExitShardParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.PeerOwner, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read peer owner failed, err: %s", err)
	}
	if this.PeerPubKey, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read peer pub key failed, err: %s", err)
	}
	return nil
}

//
// param of shard-activation request
// The request can only be initiated by operator of the shard
// @ShardID : ID of shard which is to be activated
//
type ActivateShardParam struct {
	ShardID common.ShardID
}

func (this *ActivateShardParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardID); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	return nil
}

func (this *ActivateShardParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardID, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	return nil
}

type CommitDposParam struct {
	ShardID   common.ShardID
	FeeAmount uint64
}

func (this *CommitDposParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardID); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.FeeAmount); err != nil {
		return fmt.Errorf("serialize: write fee amount failed, err: %s", err)
	}
	return nil
}

func (this *CommitDposParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardID, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.FeeAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read fee amount failed, err: %s", err)
	}
	return nil
}

// only can be invoked by ontology program
type XShardHandlingFeeParam struct {
	IsDebt  bool // debt or income
	ShardId common.ShardID
	Fee     uint64
}

func (this *XShardHandlingFeeParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteBool(this.IsDebt)
	sink.WriteShardID(this.ShardId)
	sink.WriteUint64(this.Fee)
}

func (this *XShardHandlingFeeParam) Deserialization(source *common.ZeroCopySource) error {
	var irr, eof bool
	this.IsDebt, irr, eof = source.NextBool()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	shard, err := source.NextShardID()
	if err != nil {
		return err
	}
	this.ShardId = shard
	this.Fee, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

type NotifyRootCommitDPosParam struct {
	ShardId common.ShardID
	Height  uint32
}

func (this *NotifyRootCommitDPosParam) Serialize(w io.Writer) error {
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.Height)); err != nil {
		return fmt.Errorf("serialize: write fee amount failed, err: %s", err)
	}
	return nil
}

func (this *NotifyRootCommitDPosParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if height, err := utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read fee amount failed, err: %s", err)
	} else {
		this.Height = uint32(height)
	}
	return nil
}
