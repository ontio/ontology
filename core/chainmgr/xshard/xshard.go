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
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shard_stake"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func GetShardTxsByParentHeight(start, end uint32) map[types.ShardID][]*types.Transaction {
	return nil
}

func GetShardView(lgr *ledger.Ledger, shardID types.ShardID) (*utils.ChangeView, error) {
	return nil, nil
}

func GetShardState(lgr *ledger.Ledger, shardID types.ShardID) (*shardstates.ShardState, error) {
	return nil, nil
}

func GetShardPeerStakeInfo(lgr *ledger.Ledger, shardID types.ShardID, shardView uint32) (map[string]*shard_stake.PeerViewInfo, error) {
	return nil, nil
}
