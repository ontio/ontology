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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	gov "github.com/ontio/ontology/smartcontract/service/native/governance"
	"github.com/ontio/ontology/smartcontract/service/native/shard_stake"
	shardstates "github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func getRootCurrentViewPeerItem(native *native.NativeService, pubKey string) (*utils.PeerPoolItem, error) {
	peerPoolMap, err := getRootCurrentViewPeerMap(native)
	if err != nil {
		return nil, fmt.Errorf("getRootCurrentViewPeerItem: failed, err: %s", err)
	}
	item, ok := peerPoolMap.PeerPoolMap[pubKey]
	if !ok {
		return nil, fmt.Errorf("getRootCurrentViewPeerItem: peer not exist")
	}
	return item, nil
}

func getRootCurrentViewPeerMap(native *native.NativeService) (*utils.PeerPoolMap, error) {
	//get current view
	view, err := utils.GetView(native, utils.GovernanceContractAddress, []byte(gov.GOVERNANCE_VIEW))
	if err != nil {
		return nil, fmt.Errorf("getRootCurrentViewPeerMap: get view error: %s", err)
	}
	//get peerPoolMap
	peerPoolMap, err := utils.GetPeerPoolMap(native, utils.GovernanceContractAddress, view, gov.PEER_POOL)
	if err != nil {
		return nil, fmt.Errorf("getRootCurrentViewPeerMap: get peerPoolMap error: %s", err)
	}
	return peerPoolMap, nil
}

func initStakeContractShard(native *native.NativeService, id common.ShardID, minStake uint64, stakeAsset common.Address) error {
	param := &shard_stake.InitShardParam{
		ShardId:        id,
		MinStake:       minStake,
		StakeAssetAddr: stakeAsset,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		return fmt.Errorf("initStakeContractShard: failed, err: %s", err)
	}
	if _, err := native.NativeCall(utils.ShardStakeAddress, shard_stake.INIT_SHARD, bf.Bytes()); err != nil {
		return fmt.Errorf("initStakeContractShard: failed, err: %s", err)
	}
	return nil
}

func peerInitStake(native *native.NativeService, param *JoinShardParam) error {
	callParam := &shard_stake.PeerStakeParam{
		ShardId:   param.ShardID,
		PeerOwner: param.PeerOwner,
		Value:     &shard_stake.PeerAmount{PeerPubKey: param.PeerPubKey, Amount: param.StakeAmount},
	}
	bf := new(bytes.Buffer)
	if err := callParam.Serialize(bf); err != nil {
		return fmt.Errorf("peerInitStake: failed, err: %s", err)
	}
	if _, err := native.NativeCall(utils.ShardStakeAddress, shard_stake.PEER_STAKE, bf.Bytes()); err != nil {
		return fmt.Errorf("peerInitStake: failed, err: %s", err)
	}
	return nil
}

func peerExit(native *native.NativeService, shardId common.ShardID, peer string) error {
	param := &shard_stake.PeerExitParam{
		ShardId: shardId,
		Peer:    peer,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		return fmt.Errorf("peerExit: failed, err: %s", err)
	}
	if _, err := native.NativeCall(utils.ShardStakeAddress, shard_stake.PEER_EXIT, bf.Bytes()); err != nil {
		return fmt.Errorf("peerExit: failed, err: %s", err)
	}
	return nil
}

func deletePeer(native *native.NativeService, shardId common.ShardID, peers []string) error {
	param := &shard_stake.DeletePeerParam{
		ShardId: shardId,
		Peers:   peers,
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		return fmt.Errorf("deletePeer: failed, err: %s", err)
	}
	if _, err := native.NativeCall(utils.ShardStakeAddress, shard_stake.DELETE_PEER, bf.Bytes()); err != nil {
		return fmt.Errorf("deletePeer: failed, err: %s", err)
	}
	return nil
}

func updateNewCfg(native *native.NativeService, shard *shardstates.ShardState, newCfg *utils.Configuration) {
	shard.Config.VbftCfg.N = newCfg.N
	shard.Config.VbftCfg.C = newCfg.C
	shard.Config.VbftCfg.K = newCfg.K
	shard.Config.VbftCfg.L = newCfg.L
	shard.Config.VbftCfg.BlockMsgDelay = newCfg.BlockMsgDelay
	shard.Config.VbftCfg.HashMsgDelay = newCfg.HashMsgDelay
	shard.Config.VbftCfg.PeerHandshakeTimeout = newCfg.PeerHandshakeTimeout
	shard.Config.VbftCfg.MaxBlockChangeView = newCfg.MaxBlockChangeView
	setShardState(native, utils.ShardMgmtContractAddress, shard)
}

func preCommitDpos(native *native.NativeService, shardId common.ShardID) error {
	bf := new(bytes.Buffer)
	if err := utils.SerializeShardId(bf, shardId); err != nil {
		return fmt.Errorf("preCommitDpos: serialize shardId failed, err: %s", err)
	}
	if _, err := native.NativeCall(utils.ShardStakeAddress, shard_stake.PRE_COMMIT_DPOS, bf.Bytes()); err != nil {
		return fmt.Errorf("preCommitDpos: failed, err: %s", err)
	}
	return nil
}
