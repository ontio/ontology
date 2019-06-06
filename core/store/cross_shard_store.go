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

package store

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
)

//CrossShardStore provides func with cross shard msg store package
type CrossShardStore interface {
	Close() error
	SaveCrossShardMsgByHash(msgHash common.Uint256, crossShardMsg *types.CrossShardMsg) error
	GetCrossShardMsgByHash(msgHash common.Uint256) (*types.CrossShardMsg, error)
	SaveAllShardIDs(shardIDs []common.ShardID) error
	GetAllShardIDs() ([]common.ShardID, error)
	SaveCrossShardHash(shardID common.ShardID, msgHash common.Uint256) error
	GetCrossShardHash(shardID common.ShardID) (common.Uint256, error)
	AddShardConsensusConfig(shardID common.ShardID, height uint32, value []byte) error
	GetShardConsensusConfig(shardID common.ShardID, height uint32) ([]byte, error)
	AddShardConsensusHeight(shardID common.ShardID, value []uint32) error
	GetShardConsensusHeight(shardID common.ShardID) ([]uint32, error)
}
