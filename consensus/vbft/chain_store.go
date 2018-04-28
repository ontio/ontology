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
	"github.com/ontio/ontology/consensus/actor"
	"github.com/ontio/ontology/core/ledger"
)

type ChainStore struct {
	db              *actor.LedgerActor
	chainedBlockNum uint64
	pendingBlocks   map[uint64]*Block
}

func OpenBlockStore(lgr *actor.LedgerActor) (*ChainStore, error) {
	return &ChainStore{
		db:              lgr,
		chainedBlockNum: uint64(ledger.DefLedger.GetCurrentBlockHeight()),
		pendingBlocks:   make(map[uint64]*Block),
	}, nil
}

func (self *ChainStore) close() {
	// TODO: any action on ledger actor??
}

func (self *ChainStore) GetChainedBlockNum() uint64 {
	return self.chainedBlockNum
}

func (self *ChainStore) AddBlock(block *Block, blockHash common.Uint256) error {
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
	self.pendingBlocks[block.getBlockNum()] = block

	blkNum := self.GetChainedBlockNum() + 1
	for {
		if blk, present := self.pendingBlocks[blkNum]; blk != nil && present {
			log.Infof("ledger adding chained block (%d, %d)", blkNum, self.GetChainedBlockNum())

			err := ledger.DefLedger.AddBlock(blk.Block)
			if err != nil && blkNum > self.GetChainedBlockNum() {
				return fmt.Errorf("ledger add blk (%d, %d) failed: %s", blkNum, self.GetChainedBlockNum(), err)
			}

			self.chainedBlockNum = blkNum
			if blkNum != uint64(ledger.DefLedger.GetCurrentBlockHeight()) {
				log.Errorf("!!! chain store added chained block (%d, %d): %s",
					blkNum, ledger.DefLedger.GetCurrentBlockHeight(), err)
			}

			delete(self.pendingBlocks, blkNum)
			blkNum++
		} else {
			break
		}
	}

	return nil
}

//
// SetBlock is used when recovering from fork-chain
//
func (self *ChainStore) SetBlock(block *Block, blockHash common.Uint256) error {

	err := ledger.DefLedger.AddBlock(block.Block)
	self.chainedBlockNum = uint64(ledger.DefLedger.GetCurrentBlockHeight())
	if err != nil {
		return fmt.Errorf("ledger failed to add block: %s", err)
	}

	return nil
}

func (self *ChainStore) GetBlock(blockNum uint64) (*Block, error) {

	if blk, present := self.pendingBlocks[blockNum]; present {
		return blk, nil
	}

	block, err := ledger.DefLedger.GetBlockByHeight(uint32(blockNum))
	if err != nil {
		return nil, err
	}

	return initVbftBlock(block)
}
