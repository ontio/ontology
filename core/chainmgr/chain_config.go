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

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

//
// buildShardConfig: generate OntologyConfig for shard
//
func (self *ChainManager) buildShardConfig(shardID uint64, shardState *shardstates.ShardState) (*config.OntologyConfig, error) {
	if cfg := self.GetShardConfig(shardID); cfg != nil {
		return cfg, nil
	}

	if shardState.State != shardstates.SHARD_STATE_ACTIVE {
		return nil, fmt.Errorf("Shard not active: %d", shardState.State)
	}

	// TODO: check if shardID is in children
	childShards := self.getChildShards()
	if _, present := childShards[shardID]; !present {
		return nil, fmt.Errorf("ShardID:%d not in children", shardID)
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

	// FIXME: only solo supported
	if shardConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_SOLO {
		// add seeds and bookkeepers to config
		seedlist := make([]string, 0)
		bookkeepers := make([]string, 0)
		for peerPK, info := range shardState.Peers {
			seedlist = append(seedlist, info.PeerAddress)
			bookkeepers = append(bookkeepers, peerPK)
		}
		shardConfig.Genesis.SOLO.Bookkeepers = bookkeepers
		shardConfig.Genesis.SeedList = seedlist
	} else if shardConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_VBFT {
		seedlist := make([]string, 0)
		peers := make([]*config.VBFTPeerStakeInfo, 0)
		var index uint32
		index = 1
		for peerPK, info := range shardState.Peers {
			seedlist = append(seedlist, info.PeerAddress)
			vbftpeerstakeinfo := &config.VBFTPeerStakeInfo{
				Index:      index,
				PeerPubkey: peerPK,
				Address:    info.PeerOwner.ToBase58(),
				InitPos:    info.StakeAmount,
			}
			peers = append(peers, vbftpeerstakeinfo)
			index++
		}
		shardConfig.Genesis.SeedList = seedlist
		shardConfig.Genesis.VBFT.N = shardState.Config.VbftConfigData.N
		shardConfig.Genesis.VBFT.C = shardState.Config.VbftConfigData.C
		shardConfig.Genesis.VBFT.K = shardState.Config.VbftConfigData.K
		shardConfig.Genesis.VBFT.L = shardState.Config.VbftConfigData.L
		shardConfig.Genesis.VBFT.BlockMsgDelay = shardState.Config.VbftConfigData.BlockMsgDelay
		shardConfig.Genesis.VBFT.HashMsgDelay = shardState.Config.VbftConfigData.HashMsgDelay
		shardConfig.Genesis.VBFT.PeerHandshakeTimeout = shardState.Config.VbftConfigData.PeerHandshakeTimeout
		shardConfig.Genesis.VBFT.MaxBlockChangeView = shardState.Config.VbftConfigData.MaxBlockChangeView
		shardConfig.Genesis.VBFT.MinInitStake = shardState.Config.VbftConfigData.MinInitStake
		shardConfig.Genesis.VBFT.AdminOntID = shardState.Config.VbftConfigData.AdminOntID
		shardConfig.Genesis.VBFT.VrfValue = shardState.Config.VbftConfigData.VrfValue
		shardConfig.Genesis.VBFT.VrfProof = shardState.Config.VbftConfigData.VrfProof
		shardConfig.Genesis.VBFT.Peers = shardState.Config.VbftConfigData.Peers
	} else {
		return nil, fmt.Errorf("only solo suppported")
	}
	// TODO: init config for shard $shardID, including genesis config, data dir, net port, etc
	shardName := GetShardName(shardID)
	shardConfig.P2PNode.NodePort = 10000 + uint(shardID)
	shardConfig.P2PNode.NetworkName = shardName

	// init child shard config
	shardConfig.Shard = &config.ShardConfig{
		ShardID:              shardID,
		ParentShardID:        self.shardID,
		GenesisParentHeight:  shardState.GenesisParentHeight,
		ShardPort:            uint(uint64(self.shardPort) + shardID - self.shardID),
		ParentShardIPAddress: config.DEFAULT_PARENTSHARD_IPADDR,
		ParentShardPort:      self.shardPort,
	}

	shardConfig.Rpc = config.DefConfig.Rpc
	shardConfig.Restful = config.DefConfig.Restful

	return shardConfig, nil
}
