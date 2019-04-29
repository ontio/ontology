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

package chainmgr

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/types"
	shardstates "github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

//
// buildShardConfig: generate OntologyConfig for shard
//
func (self *ChainManager) buildShardConfig(shardID types.ShardID, shardState *shardstates.ShardState) (*config.OntologyConfig, error) {
	if cfg := self.GetShardConfig(shardID); cfg != nil {
		return cfg, nil
	}

	if shardState.State != shardstates.SHARD_STATE_ACTIVE {
		return nil, fmt.Errorf("Shard not active: %d", shardState.State)
	}

	// copy current config
	buf := new(bytes.Buffer)
	if err := config.DefConfig.Serialize(buf); err != nil {
		return nil, fmt.Errorf("serialize parent config: %s", err)
	}
	shardConfig := &config.OntologyConfig{}
	if err := shardConfig.Deserialize(buf); err != nil {
		return nil, fmt.Errorf("init child config: %s", err)
	}

	if shardConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_SOLO {
		// add seeds and bookkeepers to config
		bookkeepers := make([]string, 0)
		for peerPK, _ := range shardState.Peers {
			bookkeepers = append(bookkeepers, peerPK)
		}
		shardConfig.Genesis.SOLO.Bookkeepers = bookkeepers
	} else if shardConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_VBFT {
		peers := make([]*config.VBFTPeerStakeInfo, 0)
		shardView, err := GetShardView(self.mainLedger, shardState.ShardID)
		if err != nil {
			return nil, fmt.Errorf("buildShardConfig GetShardView: failed, err: %s", err)
		}
		peerStakeInfo, err := GetShardPeerStakeInfo(self.mainLedger, shardState.ShardID, shardView.View)
		if err != nil {
			return nil, fmt.Errorf("buildShardConfig GetShardPeerStakeInfo: failed, err: %s", err)
		}
		seedList := make([]string, 0)
		for peerPK, info := range shardState.Peers {
			seedList = append(seedList, info.IpAddress)
			vbftpeerstakeinfo := &config.VBFTPeerStakeInfo{
				Index:      info.Index,
				PeerPubkey: peerPK,
				Address:    info.PeerOwner.ToBase58(),
			}
			stakeInfo, ok := peerStakeInfo[peerPK]
			if !ok {
				return nil, fmt.Errorf("buildShardConfig: peer %s has not stake info", vbftpeerstakeinfo.PeerPubkey)
			}
			vbftpeerstakeinfo.InitPos = stakeInfo.UserStakeAmount + stakeInfo.InitPos
			peers = append(peers, vbftpeerstakeinfo)
		}
		sort.SliceStable(peers, func(i, j int) bool {
			if peers[i].Index > peers[j].Index {
				return true
			}
			return false
		})
		shardConfig.Genesis.SeedList = seedList
		shardConfig.Genesis.VBFT.N = shardState.Config.VbftCfg.N
		shardConfig.Genesis.VBFT.C = shardState.Config.VbftCfg.C
		shardConfig.Genesis.VBFT.K = shardState.Config.VbftCfg.K
		shardConfig.Genesis.VBFT.L = shardState.Config.VbftCfg.L
		shardConfig.Genesis.VBFT.BlockMsgDelay = shardState.Config.VbftCfg.BlockMsgDelay
		shardConfig.Genesis.VBFT.HashMsgDelay = shardState.Config.VbftCfg.HashMsgDelay
		shardConfig.Genesis.VBFT.PeerHandshakeTimeout = shardState.Config.VbftCfg.PeerHandshakeTimeout
		shardConfig.Genesis.VBFT.MaxBlockChangeView = shardState.Config.VbftCfg.MaxBlockChangeView
		shardConfig.Genesis.VBFT.MinInitStake = shardState.Config.VbftCfg.MinInitStake
		shardConfig.Genesis.VBFT.AdminOntID = shardState.Config.VbftCfg.AdminOntID
		shardConfig.Genesis.VBFT.VrfValue = shardState.Config.VbftCfg.VrfValue
		shardConfig.Genesis.VBFT.VrfProof = shardState.Config.VbftCfg.VrfProof
		shardConfig.Genesis.VBFT.Peers = peers
	} else {
		return nil, fmt.Errorf("only solo suppported")
	}
	// TODO: init config for shard $shardID, including genesis config, data dir, net port, etc
	shardName := GetShardName(shardID)
	shardConfig.P2PNode.NetworkName = shardName

	// init child shard config
	shardConfig.Shard = &config.ShardConfig{
		ShardID:             shardID,
		GenesisParentHeight: shardState.GenesisParentHeight,
	}

	shardConfig.Rpc = config.DefConfig.Rpc
	shardConfig.Restful = config.DefConfig.Restful

	shardConfig.Common.GasPrice = shardState.Config.GasPrice
	shardConfig.Common.GasLimit = shardState.Config.GasLimit
	return shardConfig, nil
}
