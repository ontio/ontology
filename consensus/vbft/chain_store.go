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
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
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

func (self *ChainStore) GetVbftConfigInfo() (*vconfig.Configuration, error) {
	storageKey := &states.StorageKey{
		CodeHash: genesis.GovernanceContractAddress,
		Key:      append([]byte(gov.VBFT_CONFIG)),
	}
	vbft, err := ledger.DefLedger.GetStorageItem(storageKey.CodeHash, storageKey.Key)
	if err != nil {
		return nil, err
	}
	config := &govcon.Configuration{}
	if err := json.Unmarshal(vbft, config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %s", err)
	}
	chainconfig := &vconfig.Configuration{
		View:                 uint32(1),
		N:                    config.N,
		C:                    config.C,
		K:                    config.K,
		L:                    config.L,
		BlockMsgDelay:        config.BlockMsgDelay,
		HashMsgDelay:         config.HashMsgDelay,
		PeerHandshakeTimeout: config.PeerHandshakeTimeout,
		MaxBlockChangeView:   config.MaxBlockChangeView,
	}
	return chainconfig, nil
}

func (self *ChainStore) GetPeersConfig() ([]*vconfig.PeerStakeInfo, error) {
	storageKey := &states.StorageKey{
		CodeHash: genesis.GovernanceContractAddress,
		Key:      append([]byte(gov.PEER_POOL)),
	}
	peers, err := ledger.DefLedger.FindStorageItem(storageKey.CodeHash, storageKey.Key)
	if err != nil {
		return nil, err
	}
	var peerstakes []*vconfig.PeerStakeInfo
	for _, peer := range peers {
		peersconfig := &govcon.PeerPool{}
		if err := json.Unmarshal(peer, peersconfig); err != nil {
			return nil, fmt.Errorf("unmarshal peersconfig: %s", err)
		}

		config := &vconfig.PeerStakeInfo{
			Index:  uint32(peersconfig.Index.Uint64()),
			NodeID: peersconfig.PeerPubkey,
			Stake:  (peersconfig.InitPos.Uint64() + peersconfig.TotalPos.Uint64()),
		}
		peerstakes = append(peerstakes, config)
	}
	return peerstakes, nil
}

func (self *ChainStore) GetForceUpdate() (bool, error) {
	storageKey := &states.StorageKey{
		CodeHash: genesis.GovernanceContractAddress,
		Key:      append([]byte(gov.GOVERNANCE_VIEW)),
	}
	force, err := ledger.DefLedger.GetStorageItem(storageKey.CodeHash, storageKey.Key)
	if err != nil {
		return false, err
	}
	config := &govcon.GovernanceView{}
	if err := json.Unmarshal(force, config); err != nil {
		return false, fmt.Errorf("unmarshal config: %s", err)
	}
	return config.VoteCommit, nil
}
