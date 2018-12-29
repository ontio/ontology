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

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/events/message"
)

type PendingBlock struct {
	block      *Block
	execResult *store.ExecuteResult
}

type ChainStore struct {
	db              *ledger.Ledger
	chainedBlockNum uint32
	pendingBlocks   map[uint32]*PendingBlock
	server          *Server
	needSubmitBlock bool
}

func OpenBlockStore(db *ledger.Ledger, server *Server) (*ChainStore, error) {
	return &ChainStore{
		db:              db,
		chainedBlockNum: db.GetCurrentBlockHeight(),
		pendingBlocks:   make(map[uint32]*PendingBlock),
		server:          server,
		needSubmitBlock: false,
	}, nil
}

func (self *ChainStore) close() {
	// TODO: any action on ledger actor??
}

func (self *ChainStore) GetChainedBlockNum() uint32 {
	return self.chainedBlockNum
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
	for {
		var err error
		if self.needSubmitBlock {
			if submitBlk, present := self.pendingBlocks[blkNum-1]; submitBlk != nil && present {
				err := self.db.SubmitBlock(submitBlk.block.Block, *submitBlk.execResult)
				if err != nil && blkNum > self.GetChainedBlockNum() {
					return fmt.Errorf("ledger add submitBlk (%d, %d) failed: %s", blkNum, self.GetChainedBlockNum(), err)
				}
				if _, present := self.pendingBlocks[blkNum-2]; present {
					delete(self.pendingBlocks, blkNum-2)
				}
			} else {
				break
			}
		}
		execResult, err := self.db.ExecuteBlock(block.Block)
		if err != nil {
			log.Errorf("chainstore AddBlock GetBlockExecResult: %s", err)
			return fmt.Errorf("chainstore AddBlock GetBlockExecResult: %s", err)
		}
		self.pendingBlocks[blkNum] = &PendingBlock{block: block, execResult: &execResult}
		self.needSubmitBlock = true
		self.server.pid.Tell(
			&message.BlockConsensusComplete{
				Block: block.Block,
			})
		self.chainedBlockNum = blkNum
		blkNum++
		break
	}

	return nil
}

//
// SetBlock is used when recovering from fork-chain
//
func (self *ChainStore) SetBlock(blkNum uint32, blk *PendingBlock) {
	self.pendingBlocks[blkNum] = blk
}

func (self *ChainStore) GetBlock(blockNum uint32) (*Block, error) {

	if blk, present := self.pendingBlocks[blockNum]; present {
		return blk.block, nil
	}

	block, err := self.db.GetBlockByHeight(uint32(blockNum))
	if err != nil {
		return nil, err
	}

	return initVbftBlock(block)
}
