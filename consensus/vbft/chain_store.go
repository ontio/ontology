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
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/states"
	gov "github.com/ontio/ontology/smartcontract/service/native"
	govcon "github.com/ontio/ontology/smartcontract/service/native/states"
)

type ChainStore struct {
	db              *ledger.Ledger
	chainedBlockNum uint64
	pendingBlocks   map[uint64]*Block
}

func OpenBlockStore(db *ledger.Ledger) (*ChainStore, error) {
	return &ChainStore{
		db:              db,
		chainedBlockNum: uint64(db.GetCurrentBlockHeight()),
		pendingBlocks:   make(map[uint64]*Block),
	}, nil
}

func (self *ChainStore) close() {
	// TODO: any action on ledger actor??
}

func (self *ChainStore) GetChainedBlockNum() uint64 {
	return self.chainedBlockNum
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
	self.pendingBlocks[block.getBlockNum()] = block

	blkNum := self.GetChainedBlockNum() + 1
	for {
		if blk, present := self.pendingBlocks[blkNum]; blk != nil && present {
			log.Infof("ledger adding chained block (%d, %d)", blkNum, self.GetChainedBlockNum())

			err := self.db.AddBlock(blk.Block)
			if err != nil && blkNum > self.GetChainedBlockNum() {
				return fmt.Errorf("ledger add blk (%d, %d) failed: %s", blkNum, self.GetChainedBlockNum(), err)
			}

			self.chainedBlockNum = blkNum
			if blkNum != uint64(self.db.GetCurrentBlockHeight()) {
				log.Errorf("!!! chain store added chained block (%d, %d): %s",
					blkNum, self.db.GetCurrentBlockHeight(), err)
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

	err := self.db.AddBlock(block.Block)
	self.chainedBlockNum = uint64(self.db.GetCurrentBlockHeight())
	if err != nil {
		return fmt.Errorf("ledger failed to add block: %s", err)
	}

	return nil
}

func (self *ChainStore) GetBlock(blockNum uint64) (*Block, error) {

	if blk, present := self.pendingBlocks[blockNum]; present {
		return blk, nil
	}

	block, err := self.db.GetBlockByHeight(uint32(blockNum))
	if err != nil {
		return nil, err
	}

	return initVbftBlock(block)
}

func (self *ChainStore) GetVbftConfigInfo() (*govcon.Configuration, error) {
	storageKey := &states.StorageKey{
		CodeHash: genesis.GovernanceContractAddress,
		Key:      append([]byte(gov.VBFT_CONFIG)),
	}
	vbft, err := ledger.DefLedger.GetStorageItem(storageKey.CodeHash, storageKey.Key)
	if err != nil {
		return nil, err
	}
	chainconfig := &govcon.Configuration{}
	if err := json.Unmarshal(vbft, chainconfig); err != nil {
		return nil, fmt.Errorf("unmarshal chainconfig: %s", err)
	}
	return chainconfig, nil
}

func (self *ChainStore) GetForceUpdate() (bool, error) {
	storageKey := &states.StorageKey{
		CodeHash: genesis.GovernanceContractAddress,
		Key:      append([]byte(gov.FORCE_COMMIT)),
	}
	isforce, err := ledger.DefLedger.GetStorageItem(storageKey.CodeHash, storageKey.Key)
	if err != nil {
		return false, err
	}
	if bytes.Compare(isforce, []byte{1}) == 0 {
		return true, nil
	}
	return false, nil
}

func (self *ChainStore) GetNewDopsInfo() ([]uint64, error) {
	// calculate peer ranks
	config, err := self.GetVbftConfigInfo()
	if err != nil {
		return nil, fmt.Errorf("GetVbftConfigInfo err:%s", err)
	}
	scale := config.L/config.K - 1
	if scale <= 0 {
		return nil, fmt.Errorf(" L is equal or less than K!")
	}
	peers := config.Peers
	peerRanks := make([]uint64, 0)
	var sum uint64
	for i := 0; i < int(config.K); i++ {
		sum += peers[i].Stake
	}
	for i := 0; i < int(config.K); i++ {
		if peers[i].Stake == 0 {
			return nil, fmt.Errorf(fmt.Sprintf("peers rank %d, has zero stake!", i))
		}
		s := uint64(math.Ceil(float64(peers[i].Stake) * float64(scale) * float64(config.K) / float64(sum)))
		peerRanks = append(peerRanks, s)
	}
	// calculate dpos table
	dposTable := make([]uint64, 0)
	for i := 0; i < int(config.K); i++ {
		for j := uint64(0); j < peerRanks[i]; j++ {
			dposTable = append(dposTable, peers[i].Index)
		}
	}
	// shuffle
	for i := len(dposTable) - 1; i > 0; i-- {
		h, err := gov.Shufflehash(common.Uint256{}, 1, peers[dposTable[i]].PeerPubkey, i)
		//	h, err := gov.Shufflehash(native.Tx.Hash(), native.Height, peers[dposTable[i]].PeerPubkey, i)
		if err != nil {
			return nil, fmt.Errorf("[commitDpos] Failed to calculate hash value!")
		}
		j := h % uint64(i)
		dposTable[i], dposTable[j] = dposTable[j], dposTable[i]
	}
	log.Debugf("DPOS table is:", dposTable)
	return dposTable, nil
}
