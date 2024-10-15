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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
)

type PendingBlock struct {
	block        *VbftBlock
	execResult   *store.ExecuteResult
	hasSubmitted bool
}

type ChainStore struct {
	db              *ledger.Ledger
	ChainedBlockNum uint32
	pendingBlocks   map[uint32]*PendingBlock
}

func OpenBlockStore(db *ledger.Ledger) (*ChainStore, error) {
	chainstore := &ChainStore{
		db:              db,
		ChainedBlockNum: db.GetCurrentBlockHeight(),
		pendingBlocks:   make(map[uint32]*PendingBlock),
	}
	merkleRoot, err := db.GetStateMerkleRoot(chainstore.ChainedBlockNum)
	if err != nil {
		log.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", chainstore.ChainedBlockNum, err)
		return nil, fmt.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", chainstore.ChainedBlockNum, err)
	}
	crossStatesRoot, err := db.GetCrossStatesRoot(chainstore.ChainedBlockNum)
	if err != nil {
		log.Errorf("GetCrossStatesRoot blockNum:%d, error :%s", chainstore.ChainedBlockNum, err)
		return nil, fmt.Errorf("GetCrossStatesRoot blockNum:%d, error :%s", chainstore.ChainedBlockNum, err)
	}
	writeSet := overlaydb.NewMemDB(1, 1)
	block, err := chainstore.GetBlock(chainstore.ChainedBlockNum)
	if err != nil {
		return nil, err
	}
	log.Debugf("chainstore openblockstore pendingBlocks height:%d,", chainstore.ChainedBlockNum)

	chainstore.pendingBlocks[chainstore.ChainedBlockNum] = &PendingBlock{block: block, execResult: &store.ExecuteResult{WriteSet: writeSet, MerkleRoot: merkleRoot, CrossStatesRoot: crossStatesRoot}, hasSubmitted: true}
	return chainstore, nil
}

func (self *ChainStore) GetExecMerkleRoot(blkNum uint32) (common.Uint256, error) {
	if blk, present := self.pendingBlocks[blkNum]; blk != nil && present {
		return blk.execResult.MerkleRoot, nil
	}
	merkleRoot, err := self.db.GetStateMerkleRoot(blkNum)
	if err != nil {
		log.Infof("GetStateMerkleRoot blockNum:%d, error :%s", blkNum, err)
		return common.Uint256{}, fmt.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", blkNum, err)
	} else {
		return merkleRoot, nil
	}
}

func (self *ChainStore) GetCrossStatesRoot(blkNum uint32) (common.Uint256, error) {
	if blk, present := self.pendingBlocks[blkNum]; blk != nil && present {
		return blk.execResult.CrossStatesRoot, nil
	}
	statesRoot, err := self.db.GetCrossStatesRoot(blkNum)
	if err != nil {
		log.Infof("GetCrossStatesRoot blockNum:%d, error :%s", blkNum, err)
		return common.UINT256_EMPTY, fmt.Errorf("GetCrossStatesRoot blockNum:%d, error :%s", blkNum, err)
	} else {
		return statesRoot, nil
	}
}

func (self *ChainStore) ReloadFromLedger() {
	height := self.db.GetCurrentBlockHeight()
	if height > self.ChainedBlockNum {
		self.ChainedBlockNum = height
		self.pendingBlocks = make(map[uint32]*PendingBlock)
		log.Debug("chainstore ReloadFromLedger pendingBlocks")
	}
}

func (self *ChainStore) AddBlock(block *VbftBlock) (result *store.ExecuteResult, err error) {
	blkNum := block.getBlockNum()
	if blkNum != self.ChainedBlockNum+1 {
		log.Warnf("chain store adding chained block(%d, %d)", blkNum, self.ChainedBlockNum)
		return nil, fmt.Errorf("chain store adding chained block(%d, %d)", blkNum, self.ChainedBlockNum)
	}

	err = self.SubmitBlock(blkNum - 1)
	if err != nil {
		log.Errorf("chainstore blkNum:%d, SubmitBlock: %s", blkNum-1, err)
		return nil, err
	}
	execResult, err := self.db.ExecuteBlock(block.Block)
	if err != nil {
		log.Errorf("chainstore AddBlock GetBlockExecResult: %s", err)
		return nil, fmt.Errorf("chainstore AddBlock GetBlockExecResult: %s", err)
	}
	h := block.Block.Hash()
	log.Debugf("execResult:%+v, AddBlock execResult height:%d, hash: %s \n", execResult, block.Block.Header.Height, h.ToHexString())
	log.Debugf("chainstore addblock pendingBlocks height:%d", blkNum)

	self.pendingBlocks[blkNum] = &PendingBlock{block: block, execResult: &execResult, hasSubmitted: false}

	self.ChainedBlockNum = blkNum
	return &execResult, nil
}

func (self *ChainStore) SubmitBlock(blkNum uint32) error {
	if submitBlk := self.pendingBlocks[blkNum]; submitBlk != nil && !submitBlk.hasSubmitted {
		err := self.db.SubmitBlock(submitBlk.block.Block, submitBlk.block.CrossChainMsg, *submitBlk.execResult)
		if err != nil {
			return fmt.Errorf("ledger add submitBlk (%d, %d, %d) failed: %s", blkNum, self.ChainedBlockNum, self.db.GetCurrentBlockHeight(), err)
		}
		delete(self.pendingBlocks, blkNum-1)
		submitBlk.hasSubmitted = true
	}
	return nil
}

func (self *ChainStore) GetBlock(blockNum uint32) (*VbftBlock, error) {
	if blk, present := self.pendingBlocks[blockNum]; present {
		return blk.block, nil
	}
	block, err := self.db.GetBlockByHeight(blockNum)
	if err != nil {
		return nil, err
	}
	prevMerkleRoot := common.Uint256{}
	var crossChainMsg *types.CrossChainMsg
	if blockNum > 1 {
		prevMerkleRoot, err = self.db.GetStateMerkleRoot(blockNum - 1)
		if err != nil {
			log.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", blockNum, err)
			return nil, fmt.Errorf("GetStateMerkleRoot blockNum:%d, error :%s", blockNum, err)
		}
		crossChainMsg, err = self.db.GetCrossChainMsg(blockNum - 1)
		if err != nil {
			log.Errorf("GetCrossChainMsg blockNum:%d, error :%s", blockNum, err)
			return nil, fmt.Errorf("v blockNum:%d, error :%s", blockNum, err)
		}
	}
	return initVbftBlock(block, crossChainMsg, prevMerkleRoot)
}
