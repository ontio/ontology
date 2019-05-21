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

package vbft

import (
	"fmt"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	actorTypes "github.com/ontio/ontology/consensus/actor"
	shardmsg "github.com/ontio/ontology/core/chainmgr/message"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/store"
	com "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/xshard_types"
	eventmsg "github.com/ontio/ontology/events/message"
	p2pmsg "github.com/ontio/ontology/p2pserver/message/types"
)

type PendingBlock struct {
	block      *Block
	execResult *store.ExecuteResult
}
type ChainStore struct {
	db              *ledger.Ledger
	chainedBlockNum uint32
	pendingBlocks   map[uint32]*PendingBlock
	pid             *actor.PID
	needSubmitBlock bool
	shardID         common.ShardID
	p2p             *actorTypes.P2PActor
}

func OpenBlockStore(db *ledger.Ledger, serverPid *actor.PID, p2p *actorTypes.P2PActor, shardID common.ShardID) (*ChainStore, error) {
	chainstore := &ChainStore{
		db:              db,
		chainedBlockNum: db.GetCurrentBlockHeight(),
		pendingBlocks:   make(map[uint32]*PendingBlock),
		pid:             serverPid,
		needSubmitBlock: false,
		shardID:         shardID,
		p2p:             p2p,
	}
	merkleRoot, err := db.GetStateMerkleRoot(chainstore.chainedBlockNum)
	if err != nil {
		log.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", chainstore.chainedBlockNum, err)
		return nil, fmt.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", chainstore.chainedBlockNum, err)
	}
	writeSet := overlaydb.NewMemDB(1, 1)
	block, err := chainstore.GetBlock(chainstore.chainedBlockNum)
	if err != nil {
		return nil, err
	}
	chainstore.pendingBlocks[chainstore.chainedBlockNum] = &PendingBlock{block: block, execResult: &store.ExecuteResult{WriteSet: writeSet, MerkleRoot: merkleRoot}}
	return chainstore, nil
}

func (self *ChainStore) close() {
	if self.db != nil {
		self.db.Close()
	}
}

func (self *ChainStore) GetChainedBlockNum() uint32 {
	return self.chainedBlockNum
}

func (self *ChainStore) GetExecMerkleRoot(blkNum uint32) (common.Uint256, error) {
	if blk, present := self.pendingBlocks[blkNum]; blk != nil && present {
		return blk.execResult.MerkleRoot, nil
	}
	merkleRoot, err := self.db.GetStateMerkleRoot(blkNum)
	if err != nil {
		log.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", blkNum, err)
		return common.Uint256{}, fmt.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", blkNum, err)
	} else {
		return merkleRoot, nil
	}

}

func (self *ChainStore) GetExecShardNotify(blkNum uint32) []xshard_types.CommonShardMsg {
	if blk, present := self.pendingBlocks[blkNum]; blk != nil && present {
		return blk.execResult.ShardNotify
	} else {
		return nil
	}
}

func (self *ChainStore) GetExecWriteSet(blkNum uint32) *overlaydb.MemDB {
	if blk, present := self.pendingBlocks[blkNum]; blk != nil && present {
		return blk.execResult.WriteSet
	}
	return nil
}

func (self *ChainStore) ReloadFromLedger() {
	height := self.db.GetCurrentBlockHeight()
	if height > self.chainedBlockNum {
		// update chainstore height
		self.chainedBlockNum = height
		// remove persisted pending blocks
		newPending := make(map[uint32]*PendingBlock)
		for blkNum, blk := range self.pendingBlocks {
			if blkNum > height {
				newPending[blkNum] = blk
			}
		}
		// update pending blocks
		self.pendingBlocks = newPending
	}
}

func (self *ChainStore) AddBlock(block *Block) error {
	if block == nil {
		return fmt.Errorf("try add nil block")
	}

	if block.getBlockNum() <= self.GetChainedBlockNum() {
		log.Warnf("chain store adding chained block(%d, %d)", block.getBlockNum(), self.GetChainedBlockNum())
		return nil
	}

	if block.Block.Header == nil {
		panic("nil block header")
	}
	blkNum := self.GetChainedBlockNum() + 1
	var err error
	if self.needSubmitBlock {
		if submitBlk, present := self.pendingBlocks[blkNum-1]; submitBlk != nil && present {
			err := self.db.SubmitBlock(submitBlk.block.Block, *submitBlk.execResult)
			//todo broadcast p2p block
			if submitBlk.block.CrossMsg != nil {
				self.broadCrossMsg(submitBlk.block.CrossMsg, blkNum-1)
			}
			if err != nil && blkNum > self.GetChainedBlockNum() {
				return fmt.Errorf("ledger add submitBlk (%d, %d) failed: %s", blkNum, self.GetChainedBlockNum(), err)
			}
			if _, present := self.pendingBlocks[blkNum-2]; present {
				delete(self.pendingBlocks, blkNum-2)
			}
		} else {
			return nil
		}
	}
	execResult, err := self.db.ExecuteBlock(block.Block)
	if err != nil {
		log.Errorf("chainstore AddBlock GetBlockExecResult: %s", err)
		return fmt.Errorf("chainstore AddBlock GetBlockExecResult: %s", err)
	}
	self.pendingBlocks[blkNum] = &PendingBlock{block: block, execResult: &execResult}
	self.needSubmitBlock = true
	self.pid.Tell(
		&eventmsg.BlockConsensusComplete{
			Block: block.Block,
		})
	self.chainedBlockNum = blkNum
	blkNum++
	return nil
}

func (self *ChainStore) GetBlock(blockNum uint32) (*Block, error) {
	if blk, present := self.pendingBlocks[blockNum]; present {
		return blk.block, nil
	}
	block, err := self.db.GetBlockByHeight(uint32(blockNum))
	if err != nil {
		return nil, err
	}
	prevMerkleRoot := common.Uint256{}
	if blockNum > 1 {
		prevMerkleRoot, err = self.db.GetStateMerkleRoot(blockNum - 1)
		if err != nil {
			log.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", blockNum, err)
			return nil, fmt.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", blockNum, err)
		}
	}
	return initVbftBlock(block, prevMerkleRoot)
}

func (self *ChainStore) broadCrossMsg(crossShardMsgs *CrossShardMsgs, height uint32) {
	var hashes []common.Uint256
	for _, crossShard := range crossShardMsgs.CrossMsgs {
		hashes = append(hashes, crossShard.MsgHash)
	}
	shardMsgMap := make(map[common.ShardID][]xshard_types.CommonShardMsg)
	msgs := self.GetExecShardNotify(height - 1)
	for _, msg := range msgs {
		shardMsgMap[msg.GetTargetShardID()] = append(shardMsgMap[msg.GetTargetShardID()], msg)
	}
	if len(hashes) == 0 {
		return
	}
	msgRoot := common.ComputeMerkleRoot(hashes)
	for _, crossMsg := range crossShardMsgs.CrossMsgs {
		shardMsg, present := shardMsgMap[crossMsg.ShardID]
		if !present {
			log.Errorf("broadcast cross msg not found :%v", crossMsg.ShardID)
			continue
		}
		crossShardMsg := &shardmsg.CrossShardMsg{
			FromShardID:       self.shardID,
			MsgHeight:         crossShardMsgs.Height,
			SignMsgHeight:     height,
			CrossShardMsgRoot: msgRoot,
			ShardMsg:          shardMsg,
			ShardMsgHashs:     crossShardMsgs.CrossMsgs,
		}
		preMsgHash, err := self.db.GetShardMsgHash(crossMsg.ShardID)
		if err != nil {
			if err != com.ErrNotFound {
				log.Errorf("chainstore getshardmsghash err:%s", err)
			}
		} else {
			crossShardMsg.PreCrossShardMsgHash = preMsgHash
		}
		sink := common.ZeroCopySink{}
		crossShardMsg.Serialization(&sink)
		msg := &p2pmsg.CrossShardPayload{
			Version: common.VERSION_SUPPORT_SHARD,
			ShardID: crossMsg.ShardID,
			Data:    sink.Bytes(),
		}
		self.p2p.Broadcast(msg)
	}
}
