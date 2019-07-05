/*
 * Copyright (C) 2018 The ontology Authors
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

package TestCommon

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/chainmgr"
)

var TestConfigs map[common.ShardID]*config.OntologyConfig

func init() {
	TestConfigs = make(map[common.ShardID]*config.OntologyConfig)
}

func GetConfig(t *testing.T, shardID common.ShardID) *config.OntologyConfig {
	if TestConfigs[shardID] == nil {
		InitConfig(t, shardID)
	}
	return TestConfigs[shardID]
}

func InitConfig(t *testing.T, shardID common.ShardID) {
	shardName := chainmgr.GetShardName(shardID)
	CreateAccount(t, shardName+"_adminOntID")
	CreateAccount(t, shardName+"_peerOwner0")
	CreateAccount(t, shardName+"_peerOwner1")
	CreateAccount(t, shardName+"_peerOwner2")
	CreateAccount(t, shardName+"_peerOwner3")
	CreateAccount(t, shardName+"_peerOwner4")
	CreateAccount(t, shardName+"_peerOwner5")
	CreateAccount(t, shardName+"_peerOwner6")
	CreateAccount(t, shardName+"_user1") // shard_0_user1

	cfg := config.NewOntologyConfig()
	acc := GetAccount(shardName + "_adminOntID")
	cfg.Genesis.VBFT.AdminOntID = fmt.Sprintf("did:ont:%s", acc.Address.ToBase58())
	for i := 0; i < 7; i++ {
		peer := GetAccount(GetOwnerName(shardID, i))
		cfg.Genesis.VBFT.Peers[i] = &config.VBFTPeerStakeInfo{
			Index:      uint32(i + 1),
			Address:    peer.Address.ToBase58(),
			PeerPubkey: hex.EncodeToString(keypair.SerializePublicKey(peer.PublicKey)),
			InitPos:    100000,
		}
	}
	cfg.Genesis.VBFT.BlockMsgDelay = 5000
	cfg.Genesis.VBFT.HashMsgDelay = 5000

	TestConfigs[shardID] = cfg
}

func InitSoloConfig(t *testing.T, shardID common.ShardID) {
	shardName := chainmgr.GetShardName(shardID)
	CreateAccount(t, shardName+"_adminOntID")
	CreateAccount(t, shardName+"_peerOwner0")

	cfg := config.NewOntologyConfig()
	cfg.Genesis.ConsensusType = config.CONSENSUS_TYPE_SOLO
	cfg.Genesis.SOLO.GenBlockTime = 1

	curPk := hex.EncodeToString(keypair.SerializePublicKey(GetAccount(shardName + "_peerOwner0").PublicKey))
	cfg.Genesis.SOLO.Bookkeepers = []string{curPk}
	TestConfigs[shardID] = cfg
}
