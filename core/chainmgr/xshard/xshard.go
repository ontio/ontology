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

package xshard

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/ledger"
	sComm "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shard_stake"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func GetShardTxsByParentHeight(start, end uint32) map[common.ShardID][]*types.Transaction {
	return nil
}

func GetShardView(lgr *ledger.Ledger, shardID common.ShardID) (*utils.ChangeView, error) {
	if lgr == nil {
		return nil, fmt.Errorf("get shard view,ledger is nil")
	}
	shardIDBytes := utils.GetUint64Bytes(shardID.ToUint64())
	viewKey := shard_stake.GenShardViewKey(shardIDBytes)
	viewBytes, err := lgr.GetStorageItem(utils.ShardStakeAddress, viewKey)
	if err == sComm.ErrNotFound {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("GetShardView: get current view: %s", err)
	}
	changeView := &utils.ChangeView{}
	if err := changeView.Deserialize(bytes.NewBuffer(viewBytes)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize changeView error: %v", err)
	}
	return changeView, nil
}

func GetShardState(lgr *ledger.Ledger, shardID common.ShardID) (*shardstates.ShardState, error) {
	if lgr == nil {
		return nil, fmt.Errorf("get shard state,ledger is nil ")
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

func GetShardPeerStakeInfo(lgr *ledger.Ledger, shardID common.ShardID, shardView uint32) (map[string]*shard_stake.PeerViewInfo, error) {
	if lgr == nil {
		return nil, fmt.Errorf("get shard peer stake info,ledger is nil")
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
