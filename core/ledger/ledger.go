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

package ledger

import (
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/store/ledgerstore"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/event"
)

var DefLedger *Ledger

type Ledger struct {
	ldgStore store.LedgerStore
}

func NewLedger() (*Ledger, error) {
	ldgStore, err := ledgerstore.NewLedgerStore()
	if err != nil {
		return nil, fmt.Errorf("NewLedgerStore error %s", err)
	}
	return &Ledger{
		ldgStore: ldgStore,
	}, nil
}

func (this *Ledger) GetStore() store.LedgerStore {
	return this.ldgStore
}

func (this *Ledger) Init(defaultBookkeeper []keypair.PublicKey) error {
	genesisBlock, err := genesis.GenesisBlockInit(defaultBookkeeper)
	if err != nil {
		return fmt.Errorf("genesisBlock error %s", err)
	}
	err = this.ldgStore.InitLedgerStoreWithGenesisBlock(genesisBlock, defaultBookkeeper)
	if err != nil {
		return fmt.Errorf("InitLedgerStoreWithGenesisBlock error %s", err)
	}
	return nil
}

func (this *Ledger) AddHeaders(headers []*types.Header) error {
	return this.ldgStore.AddHeaders(headers)
}

func (this *Ledger) AddBlock(block *types.Block) error {
	return this.ldgStore.AddBlock(block)
}

func (this *Ledger) GetBlockRootWithNewTxRoot(txRoot common.Uint256) common.Uint256 {
	return this.ldgStore.GetBlockRootWithNewTxRoot(txRoot)
}

func (this *Ledger) GetBlockByHeight(height uint32) (*types.Block, error) {
	return this.ldgStore.GetBlockByHeight(height)
}

func (this *Ledger) GetBlockByHash(blockHash common.Uint256) (*types.Block, error) {
	return this.ldgStore.GetBlockByHash(blockHash)
}

func (this *Ledger) GetHeaderByHeight(height uint32) (*types.Header, error) {
	return this.ldgStore.GetHeaderByHeight(height)
}

func (this *Ledger) GetHeaderByHash(blockHash common.Uint256) (*types.Header, error) {
	return this.ldgStore.GetHeaderByHash(blockHash)
}

func (this *Ledger) GetBlockHash(height uint32) common.Uint256 {
	return this.ldgStore.GetBlockHash(height)
}

func (this *Ledger) GetTransaction(txHash common.Uint256) (*types.Transaction, error) {
	tx, _, err := this.ldgStore.GetTransaction(txHash)
	return tx, err
}

func (this *Ledger) GetTransactionWithHeight(txHash common.Uint256) (*types.Transaction, uint32, error) {
	return this.ldgStore.GetTransaction(txHash)
}

func (this *Ledger) GetCurrentBlockHeight() uint32 {
	return this.ldgStore.GetCurrentBlockHeight()
}

func (this *Ledger) GetCurrentBlockHash() common.Uint256 {
	return this.ldgStore.GetCurrentBlockHash()
}

func (this *Ledger) GetCurrentHeaderHeight() uint32 {
	return this.ldgStore.GetCurrentHeaderHeight()
}

func (this *Ledger) GetCurrentHeaderHash() common.Uint256 {
	return this.ldgStore.GetCurrentHeaderHash()
}

func (this *Ledger) IsContainTransaction(txHash common.Uint256) (bool, error) {
	return this.ldgStore.IsContainTransaction(txHash)
}

func (this *Ledger) IsContainBlock(blockHash common.Uint256) (bool, error) {
	return this.ldgStore.IsContainBlock(blockHash)
}

func (this *Ledger) GetCurrentStateRoot() (common.Uint256, error) {
	return common.Uint256{}, nil
}

func (this *Ledger) GetBookkeeperState() (*states.BookkeeperState, error) {
	return this.ldgStore.GetBookkeeperState()
}

func (this *Ledger) GetStorageItem(codeHash common.Address, key []byte) ([]byte, error) {
	storageKey := &states.StorageKey{
		CodeHash: codeHash,
		Key:      key,
	}
	storageItem, err := this.ldgStore.GetStorageItem(storageKey)
	if err != nil {
		return nil, fmt.Errorf("GetStorageItem error %s", err)
	}
	if storageItem == nil {
		return nil, nil
	}
	return storageItem.Value, nil
}

func (this *Ledger) GetContractState(contractHash common.Address) (*payload.DeployCode, error) {
	return this.ldgStore.GetContractState(contractHash)
}

func (this *Ledger) PreExecuteContract(tx *types.Transaction) ([]interface{}, error) {
	return this.ldgStore.PreExecuteContract(tx)
}

func (this *Ledger) GetEventNotifyByTx(tx common.Uint256) ([]*event.NotifyEventInfo, error) {
	return this.ldgStore.GetEventNotifyByTx(tx)
}

func (this *Ledger) GetEventNotifyByBlock(height uint32) ([]common.Uint256, error) {
	return this.ldgStore.GetEventNotifyByBlock(height)
}
