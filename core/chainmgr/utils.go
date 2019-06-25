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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/ledger"
	sComm "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	shardstates "github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func (self *ChainManager) GetShardConfig(shardID common.ShardID) *config.OntologyConfig {
	self.lock.RLock()
	defer self.lock.RUnlock()
	if s := self.shards[shardID]; s != nil {
		return s.Config
	}
	return nil
}

func (self *ChainManager) setShardConfig(shardID common.ShardID, cfg *config.OntologyConfig) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	if info := self.shards[shardID]; info != nil {
		info.Config = cfg
		return nil
	}

	self.shards[shardID] = &ShardInfo{
		ShardID: shardID,
		Config:  cfg,
	}
	return nil
}

func (self *ChainManager) updateShardConfig(shardID common.ShardID, shardcfg *shardstates.ShardConfig) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	if info, present := self.shards[shardID]; present {
		info.Config.Genesis.VBFT = shardcfg.VbftCfg
		info.Config.Common.GasPrice = shardcfg.GasPrice
		info.Config.Common.GasLimit = shardcfg.GasLimit
	} else {
		cfg := &config.OntologyConfig{
			Genesis: &config.GenesisConfig{
				VBFT: shardcfg.VbftCfg,
			},
			Common: &config.CommonConfig{
				GasLimit: shardcfg.GasLimit,
				GasPrice: shardcfg.GasPrice,
			},
		}
		self.shards[shardID] = &ShardInfo{
			ShardID: shardID,
			Config:  cfg,
		}
	}
	return nil
}

func GetShardMgmtGlobalState(lgr *ledger.Ledger) (*shardstates.ShardMgmtGlobalState, error) {
	if lgr == nil {
		return nil, fmt.Errorf("get shard global state, nil ledger")
	}

	data, err := lgr.GetStorageItem(utils.ShardMgmtContractAddress, []byte(shardmgmt.KEY_VERSION))
	if err == sComm.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get shardmgmt version: %s", err)
	}
	if len(data) == 0 {
		return nil, nil
	}

	ver, err := serialization.ReadUint32(bytes.NewBuffer(data))
	if ver != utils.VERSION_CONTRACT_SHARD_MGMT {
		return nil, fmt.Errorf("uncompatible version: %d", ver)
	}

	data, err = lgr.GetStorageItem(utils.ShardMgmtContractAddress, []byte(shardmgmt.KEY_GLOBAL_STATE))
	if err == sComm.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get shardmgmt global state: %s", err)
	}
	if len(data) == 0 {
		return nil, nil
	}

	globalState := &shardstates.ShardMgmtGlobalState{}
	if err := globalState.Deserialization(common.NewZeroCopySource(data)); err != nil {
		return nil, fmt.Errorf("des shardmgmt global state: %s", err)
	}

	return globalState, nil
}
