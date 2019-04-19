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
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/ledger"
	sComm "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/smartcontract/service/native/shard_stake"
	shardsysmsg "github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
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
		ShardID:       shardID,
		Config:        cfg,
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
	if ver != shardmgmt.VERSION_CONTRACT_SHARD_MGMT {
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

func GetShardState(lgr *ledger.Ledger, shardID common.ShardID) (*shardstates.ShardState, error) {
	if lgr == nil {
		return nil, fmt.Errorf("get shard state, nil ledger")
	}

	shardIDBytes := utils.GetUint64Bytes(shardID.ToUint64())
	key := append([]byte(shardmgmt.KEY_SHARD_STATE), shardIDBytes...)
	data, err := lgr.GetStorageItem(utils.ShardMgmtContractAddress, key)
	if err == sComm.ErrNotFound {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("get shardmgmt shard state: %s", err)
	}

	shardState := &shardstates.ShardState{}
	if err := shardState.Deserialization(common.NewZeroCopySource(data)); err != nil {
		return nil, fmt.Errorf("des shardmgmt shard state: %s", err)
	}

	return shardState, nil
}

func GetShardView(lgr *ledger.Ledger, shardID common.ShardID) (*utils.ChangeView, error) {
	shardIDBytes := utils.GetUint64Bytes(shardID.ToUint64())
	viewKey := shard_stake.GenShardViewKey(shardIDBytes)
	viewBytes, err := lgr.GetStorageItem(utils.ShardStakeAddress, viewKey)
	if err == sComm.ErrNotFound {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("GetShardPeerStakeInfo: get current view: %s", err)
	}
	changeView := &utils.ChangeView{}
	if err := changeView.Deserialize(bytes.NewBuffer(viewBytes)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize changeView error: %v", err)
	}
	return changeView, nil
}

func GetShardPeerStakeInfo(lgr *ledger.Ledger, shardID types.ShardID, shardView uint32) (map[string]*shard_stake.PeerViewInfo, error) {
	if lgr == nil {
		return nil, fmt.Errorf("GetShardPeerStakeInfo: nil ledger")
	}
	shardIDBytes := utils.GetUint64Bytes(shardID.ToUint64())

	viewBytes := utils.GetUint32Bytes(shardView)
	viewInfoKey := shard_stake.GenShardViewInfoKey(shardIDBytes, viewBytes)
	infoData, err := lgr.GetStorageItem(utils.ShardStakeAddress, viewInfoKey)
	if err == sComm.ErrNotFound {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("GetShardPeerStakeInfo: get current view info: %s", err)
	}
	info := shard_stake.ViewInfo{}
	if err := info.Deserialization(common.NewZeroCopySource(infoData)); err != nil {
		return nil, fmt.Errorf("GetShardPeerStakeInfo: deserialize view info: %s", err)
	}
	return info.Peers, nil
}

func GetRequestedRemoteShards(lgr *ledger.Ledger, blockNum uint32) ([]common.ShardID, error) {
	if lgr == nil {
		return nil, fmt.Errorf("uninitialized chain mgr")
	}
	blockNumBytes := utils.GetUint32Bytes(blockNum)
	key := utils.ConcatKey(utils.ShardSysMsgContractAddress, []byte(shardsysmsg.KEY_SHARDS_IN_BLOCK), blockNumBytes)
	toShardsBytes, err := xshard_state.GetKVStorageItem(key)
	if err == xshard_state.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get remote msg toShards in blk %d: %s", blockNum, err)
	}

	req := &shardsysmsg.ToShardsInBlock{}
	if err := req.Deserialization(common.NewZeroCopySource(toShardsBytes)); err != nil {
		return nil, fmt.Errorf("deserialize toShards: %s", err)
	}
	return req.Shards, nil
}

func GetRequestsToRemoteShard(lgr *ledger.Ledger, blockHeight uint32, toShard common.ShardID) ([][]byte, error) {
	if lgr == nil {
		return nil, fmt.Errorf("nil ledger")
	}

	blockNumBytes := utils.GetUint32Bytes(blockHeight)
	shardIDBytes := utils.GetUint64Bytes(toShard.ToUint64())
	key := utils.ConcatKey(utils.ShardSysMsgContractAddress, []byte(shardsysmsg.KEY_REQS_IN_BLOCK), blockNumBytes, shardIDBytes)
	reqBytes, err := xshard_state.GetKVStorageItem(key)
	if err == xshard_state.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get remote msg to shard %d in blk %d: %s", toShard, blockHeight, err)
	}

	req := &shardsysmsg.ReqsInBlock{}
	if err := req.Deserialization(common.NewZeroCopySource(reqBytes)); err != nil {
		return nil, fmt.Errorf("deserialize remote msg to shard %d: %s", toShard, err)
	}
	return req.Reqs, nil
}
