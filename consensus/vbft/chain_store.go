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
	"fmt"
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/states"
	gov "github.com/ontio/ontology/smartcontract/service/native/governance"
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

func (self *ChainStore) GetVbftConfigInfo() (*config.VBFTConfig, error) {
	storageKey := &states.StorageKey{
		CodeHash: genesis.GovernanceContractAddress,
		Key:      append([]byte(gov.VBFT_CONFIG)),
	}
	data, err := ledger.DefLedger.GetStorageItem(storageKey.CodeHash, storageKey.Key)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(data)
	n, err := serialization.ReadUint32(buffer)
	if err != nil {
		return nil, err
	}
	c, err := serialization.ReadUint32(buffer)
	if err != nil {
		return nil, err
	}
	k, err := serialization.ReadUint32(buffer)
	if err != nil {
		return nil, err
	}
	l, err := serialization.ReadUint32(buffer)
	if err != nil {
		return nil, err
	}
	blockmsgdelay, err := serialization.ReadUint32(buffer)
	if err != nil {
		return nil, err
	}
	hashmsgdelay, err := serialization.ReadUint32(buffer)
	if err != nil {
		return nil, err
	}
	peerhandshaketimeout, err := serialization.ReadUint32(buffer)
	if err != nil {
		return nil, err
	}
	maxblockchangeview, err := serialization.ReadUint32(buffer)
	if err != nil {
		return nil, err
	}
	chainconfig := &config.VBFTConfig{
		N:                    n,
		C:                    c,
		K:                    k,
		L:                    l,
		BlockMsgDelay:        blockmsgdelay,
		HashMsgDelay:         hashmsgdelay,
		PeerHandshakeTimeout: peerhandshaketimeout,
		MaxBlockChangeView:   maxblockchangeview,
	}
	return chainconfig, nil
}

func (self *ChainStore) GetPeersConfig() ([]*config.VBFTPeerStakeInfo, error) {
	goveranceview, err := self.GetGovernanceView()
	if err != nil {
		return nil, err
	}
	storageKey := &states.StorageKey{
		CodeHash: genesis.GovernanceContractAddress,
		Key:      append([]byte(gov.PEER_POOL), goveranceview.View.Bytes()...),
	}
	data, err := ledger.DefLedger.GetStorageItem(storageKey.CodeHash, storageKey.Key)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(data)
	len, err := serialization.ReadVarUint(buffer, 0)
	if err != nil {
		return nil, err
	}
	var peerstakes []*config.VBFTPeerStakeInfo
	for i := 0; i < int(len); i++ {
		index, err := serialization.ReadUint32(buffer)
		if err != nil {
			return nil, err
		}
		peerpubkey, err := serialization.ReadString(buffer)
		if err != nil {
			return nil, err
		}
		_, err = serialization.ReadString(buffer)
		if err != nil {
			return nil, err
		}
		_, err = serialization.ReadUint8(buffer)
		if err != nil {
			return nil, err
		}
		initpos, err := serialization.ReadUint64(buffer)
		if err != nil {
			return nil, err
		}
		totalpos, err := serialization.ReadUint64(buffer)
		if err != nil {
			return nil, err
		}
		config := &config.VBFTPeerStakeInfo{
			Index:      index,
			PeerPubkey: peerpubkey,
			InitPos:    initpos + totalpos,
		}
		peerstakes = append(peerstakes, config)
	}
	return peerstakes, nil
}

func (self *ChainStore) isForceUpdate() (bool, error) {
	goveranceview, err := self.GetGovernanceView()
	if err != nil {
		return false, err
	}
	return goveranceview.VoteCommit, nil
}

func (self *ChainStore) GetGovernanceView() (*gov.GovernanceView, error) {
	storageKey := &states.StorageKey{
		CodeHash: genesis.GovernanceContractAddress,
		Key:      append([]byte(gov.GOVERNANCE_VIEW)),
	}
	data, err := ledger.DefLedger.GetStorageItem(storageKey.CodeHash, storageKey.Key)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(data)
	view, err := serialization.ReadUint64(buffer)
	if err != nil {
		return nil, err
	}
	votecommit, err := serialization.ReadBool(buffer)
	if err != nil {
		return nil, err
	}
	config := &gov.GovernanceView{
		View:       new(big.Int).SetUint64(view),
		VoteCommit: votecommit,
	}
	return config, nil
}
