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

package increment

import (
	"fmt"
	"sync"

	ethcomm "github.com/ethereum/go-ethereum/common"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
)

// IncrementValidator do increment check of transaction
type IncrementValidator struct {
	mutex      sync.Mutex
	blocks     []map[common.Uint256]bool
	baseHeight uint32
	maxBlocks  int
	nonces     []map[common.Address]uint64
}

func NewIncrementValidator(maxBlocks int) *IncrementValidator {
	if maxBlocks <= 0 {
		maxBlocks = 20
	}
	return &IncrementValidator{
		maxBlocks: maxBlocks,
	}
}

func (self *IncrementValidator) Clean() {
	self.mutex.Lock()
	self.blocks = nil
	self.baseHeight = 0
	self.nonces = nil
	self.mutex.Unlock()
}

// BlockRange returns the block range [start, end) this validator can check
func (self *IncrementValidator) BlockRange() (start uint32, end uint32) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.blockRange()
}

func (self *IncrementValidator) blockRange() (start uint32, end uint32) {
	return self.baseHeight, self.baseHeight + uint32(len(self.blocks))
}

// AddBlock add a new block to this validator
func (self *IncrementValidator) AddBlock(block *types.Block) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if len(self.blocks) == 0 {
		self.baseHeight = block.Header.Height
	}

	if self.baseHeight+uint32(len(self.blocks)) != block.Header.Height {
		start, end := self.blockRange()
		log.Errorf("discontinue block is not allowed: [start, end)=[%d, %d), block height= %d",
			start, end, block.Header.Height)
		return
	}

	if len(self.blocks) >= self.maxBlocks {
		self.blocks = self.blocks[1:]
		self.baseHeight += 1

		self.nonces = self.nonces[1:]
	}
	txHashes := make(map[common.Uint256]bool)
	nonceMap := make(map[common.Address]uint64)
	for _, tx := range block.Transactions {
		txhash := tx.Hash()
		txHashes[txhash] = true
		if tx.IsEipTx() {
			nonceMap[tx.Payer] = uint64(tx.Nonce) + 1
		}
	}
	self.nonces = append(self.nonces, nonceMap)
	self.blocks = append(self.blocks, txHashes)
}

// Verfiy does increment check start at startHeight
func (self *IncrementValidator) Verify(tx *types.Transaction, startHeight uint32, nonceCtx map[common.Address]uint64) error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if startHeight < self.baseHeight {
		return fmt.Errorf("can not do increment validation: startHeight %v < self.baseHeight %v", startHeight, self.baseHeight)
	}
	for i := int(startHeight - self.baseHeight); i < len(self.blocks); i++ {
		if _, ok := self.blocks[i][tx.Hash()]; ok {
			return fmt.Errorf("tx duplicated")
		}
	}
	//check nonce
	if tx.IsEipTx() {
		//1st tx for account
		if nonceCtx[tx.Payer] == 0 {
			//get the nonce from cache
			for i := 0; i < len(self.blocks); i++ {
				latestNonce := self.nonces[i][tx.Payer]
				if latestNonce != 0 {
					nonceCtx[tx.Payer] = latestNonce
				}
			}
			//still empty, load from ledger store
			if nonceCtx[tx.Payer] == 0 {
				acct, err := ledger.DefLedger.GetEthAccount(ethcomm.BytesToAddress(tx.Payer[:]))
				if err != nil {
					return err
				}
				nonceCtx[tx.Payer] = acct.Nonce
			}
		}

		if uint64(tx.Nonce) != nonceCtx[tx.Payer] {
			log.Infof("wrong nonce for %s, exptected: %d, got: %d", tx.Payer.ToBase58(),
				nonceCtx[tx.Payer], tx.Nonce)
			return fmt.Errorf("wrong nonce for %s, exptected: %d, got: %d", tx.Payer.ToBase58(),
				nonceCtx[tx.Payer], tx.Nonce)
		}

		nonceCtx[tx.Payer] = uint64(tx.Nonce) + 1
	}

	return nil
}
