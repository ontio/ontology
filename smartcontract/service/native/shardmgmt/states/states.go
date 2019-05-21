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
	"math/big"
	"sort"
	"strings"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shard_stake"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type ShardMgmtFeeType uint8

const (
	TYPE_CREATE_SHARD_FEE ShardMgmtFeeType = iota
	TYPE_JOIN_SHARD_FEE
)

const (
	DEFAULT_CREATE_SHARD_FEE = 100 * 1000000000 // 100 ong
	DEFAULT_JOIN_SHARD_FEE   = 100 * 1000000000 // 100 ong
)

const (
	SHARD_STATE_CREATED    = iota
	SHARD_STATE_CONFIGURED // all parameter configured
	SHARD_PEER_JOIND       // has some peer joined
	SHARD_STATE_ACTIVE     // started
	SHARD_STATE_STOPPING   // started
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

func (this *ShardConfig) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.GasPrice)
	sink.WriteUint64(this.GasLimit)
	sink.WriteUint32(this.NetworkSize)
	sink.WriteAddress(this.StakeAssetAddress)
	sink.WriteAddress(this.GasAssetAddress)
	this.VbftCfg.Serialization(sink)
}

func (this *ShardConfig) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.GasPrice, eof = source.NextUint64()
	this.GasLimit, eof = source.NextUint64()
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
	ShardID             common.ShardID
	Creator             common.Address
	State               uint32
	GenesisParentHeight uint32
	Config              *ShardConfig

	Peers map[string]*PeerShardStakeInfo
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
		return peers[i].PeerPubKey > peers[j].PeerPubKey
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

func (this *ShardState) UpdateDposInfo(native *native.NativeService) error {
	currentView, err := shard_stake.GetShardCurrentViewIndex(native, this.ShardID)
	if err != nil {
		return fmt.Errorf("updateDposInfo: failed, err: %s", err)
	}
	currentViewInfo, err := shard_stake.GetShardViewInfo(native, this.ShardID, currentView)
	if err != nil {
		return fmt.Errorf("updateDposInfo: failed, err: %s", err)
	}
	peerStakeInfo := make([]*shard_stake.PeerViewInfo, 0)
	for _, peer := range currentViewInfo.Peers {
		peerStakeInfo = append(peerStakeInfo, peer)
	}
	sort.SliceStable(peerStakeInfo, func(i, j int) bool {
		stakeI := peerStakeInfo[i].InitPos + peerStakeInfo[i].UserStakeAmount
		stakeJ := peerStakeInfo[j].InitPos + peerStakeInfo[j].UserStakeAmount
		if stakeI == stakeJ {
			return peerStakeInfo[i].PeerPubKey > peerStakeInfo[j].PeerPubKey
		} else {
			return stakeI > stakeJ
		}
	})
	consensusCount := uint32(0)
	for _, peer := range peerStakeInfo {
		peerState, ok := this.Peers[peer.PeerPubKey]
		if !ok {
			return fmt.Errorf("updateDposInfo: stake peer %s isn't exist in shard state", peer.PeerPubKey)
		}
		if peerState.NodeType == CONSENSUS_NODE || peerState.NodeType == CONDIDATE_NODE {
			if consensusCount < this.Config.VbftCfg.K {
				peerState.NodeType = CONSENSUS_NODE
				consensusCount++
			} else {
				peerState.NodeType = CONDIDATE_NODE
			}
		}
	}
	return nil
}

type ShardCommitDposInfo struct {
	TransferId          *big.Int                   `json:"transfer_id"`
	Height              uint32                     `json:"height"`
	Hash                common.Uint256             `json:"hash"`
	FeeAmount           uint64                     `json:"fee_amount"`
	XShardHandleFeeInfo *shard_stake.XShardFeeInfo `json:"xshard_handle_fee_info"`
}

func (this *ShardCommitDposInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(common.BigIntToNeoBytes(this.TransferId))
	sink.WriteUint32(this.Height)
	sink.WriteHash(this.Hash)
	sink.WriteUint64(this.FeeAmount)
	this.XShardHandleFeeInfo.Serialization(sink)
}

func (this *ShardCommitDposInfo) Deserialization(source *common.ZeroCopySource) error {
	id, _, irr, eof := source.NextVarBytes()
	if irr {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.TransferId = common.BigIntFromNeoBytes(id)
	this.Height, eof = source.NextUint32()
	this.Hash, eof = source.NextHash()
	this.FeeAmount, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.XShardHandleFeeInfo = &shard_stake.XShardFeeInfo{}
	if err := this.XShardHandleFeeInfo.Deserialization(source); err != nil {
		return err
	}
	return nil
}
