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
	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

func (this *ChainManager) addShardBlockInfo(blkInfo *message.ShardBlockInfo) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if err := this.blockPool.AddBlock(blkInfo); err != nil {
		return err
	}

	return nil
}

func (this *ChainManager) getShardBlockInfo(shardID uint64, height uint64) *message.ShardBlockInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()

	return this.blockPool.GetBlock(shardID, height)
}

func (this *ChainManager) addShardEvent(evt *shardstates.ShardEventState) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if err := this.blockPool.AddEvent(this.shardID, evt); err != nil {
		return err
	}
	return nil
}

func (this *ChainManager) updateShardBlockInfo(shardID uint64, height uint64, blk *types.Block, shardTxs map[uint64]*message.ShardBlockTx) {
	this.lock.Lock()
	defer this.lock.Unlock()

	blkInfo := this.blockPool.GetBlock(shardID, height)
	if blkInfo == nil {
		return
	}

	blkInfo.Header = &message.ShardBlockHeader{Header: blk.Header}
	blkInfo.ShardTxs = shardTxs
}
