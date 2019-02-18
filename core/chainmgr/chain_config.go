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

	self.setShardConfig(shardID, shardConfig)

	return shardConfig, nil
}

func GetShardName(shardID uint64) string {
	return fmt.Sprintf("shard_%d", shardID)
}
