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
	shardstates "github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

func (self *ChainManager) initShardInfo(shard *shardstates.ShardState) *ShardInfo {
	info := &ShardInfo{}
	if i, present := self.shards[shard.ShardID]; present {
		info = i
	}
	info.ShardID = shard.ShardID

	seedList := make([]string, 0)
	for _, p := range shard.Peers {
		seedList = append(seedList, p.IpAddress)
	}
	info.SeedList = seedList
	self.shards[shard.ShardID] = info
	return info
}
