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
	"encoding/hex"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/cmd/utils"
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

func (this *ChainManager) getChildShards() map[uint64]*ShardInfo {

	shards := make(map[uint64]*ShardInfo)

	for _, shardInfo := range this.shards {
		if shardInfo.ConnType == CONN_TYPE_CHILD {
			shards[shardInfo.ShardID] = shardInfo
		}
	}

	return shards
}

func (self *ChainManager) initShardInfo(shardID uint64, shard *shardstates.ShardState) (*ShardInfo, error) {
	if shardID != shard.ShardID {
		return nil, fmt.Errorf("unmatched shard ID with shardstate")
	}

	peerPK := hex.EncodeToString(keypair.SerializePublicKey(self.account.PublicKey))
	info := &ShardInfo{}
	if i, present := self.shards[shard.ShardID]; present {
		info = i
	}
	info.ShardID = shard.ShardID
	info.ParentShardID = shard.ParentShardID

	if _, present := shard.Peers[peerPK]; present {
		// peer is in the shard
		// build shard config
		if self.shardID == shard.ShardID {
			// self shards
			info.ConnType = CONN_TYPE_SELF
		} else if self.parentShardID == shard.ShardID {
			// parent shard
			info.ConnType = CONN_TYPE_PARENT
		} else if self.shardID == shard.ParentShardID {
			// child shard
			info.ConnType = CONN_TYPE_CHILD
		}
	} else {
		if self.shardID == shard.ParentShardID {
			// child shards
			info.ConnType = CONN_TYPE_CHILD
		} else if self.parentShardID == shard.ParentShardID {
			// sib shards
			info.ConnType = CONN_TYPE_SIB
		}
	}

	if info.ConnType != CONN_TYPE_UNKNOWN {
		self.shards[shard.ShardID] = info
	}
	return info, nil
}

func (self *ChainManager)buildShardCommandArgs(cmdArgs map[string]string, shardInfo *ShardInfo) ([]string, error) {
	args := make([]string, 0)
	shardArgs := make(map[string]string)
	for _, flag := range utils.CmdFlagsForSharding {
		shardArgs[flag.GetName()] = ""
	}
	shardArgs[utils.ShardIDFlag.GetName()] = fmt.Sprintf("%d", shardInfo.ShardID)
	shardArgs[utils.ShardPortFlag.GetName()] = fmt.Sprintf("%d",  uint(uint64(self.shardPort) + shardInfo.ShardID - self.shardID))
	shardArgs[utils.ParentShardIDFlag.GetName()] = fmt.Sprintf("%d", shardInfo.ParentShardID)
	shardArgs[utils.ParentShardPortFlag.GetName()] = fmt.Sprintf("%d", self.shardPort)

	// copy all args to new shard command, except sharding related flags
	for n, v := range cmdArgs {
		if n == utils.ConsensusPortFlag.GetName() {
			continue
		}

		if shardCfg, present := shardArgs[n]; !present {
			args = append(args, "--" + n+"="+v)
		} else if len(shardCfg) > 0 {
			args = append(args, "--" + n+"="+shardCfg)
		}
	}

	return args, nil
}
