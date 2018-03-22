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

package store

import (
	states "github.com/Ontology/core/states"
	."github.com/Ontology/common"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	"github.com/Ontology/core/payload"
)
// ILedgerStore provides func with store package.
type ILedgerStore interface {
	InitLedgerStoreWithGenesisBlock(genesisblock *types.Block, defaultBookKeeper []*crypto.PubKey) error
	Close() error
	AddHeaders(headers []*types.Header) error
	AddBlock(block *types.Block) error
	GetCurrentBlockHash() Uint256
	GetCurrentBlockHeight() uint32
	GetCurrentHeaderHeight() uint32
	GetCurrentHeaderHash() Uint256
	GetBlockHash(height uint32) Uint256
	GetHeaderByHash(blockHash Uint256) (*types.Header, error)
	GetHeaderByHeight(height uint32) (*types.Header, error)
	GetBlockByHash(blockHash Uint256) (*types.Block, error)
	GetBlockByHeight(height uint32) (*types.Block, error)
	GetTransaction(txHash Uint256) (*types.Transaction, uint32, error)
	IsContainBlock(blockHash Uint256) (bool, error)
	IsContainTransaction(txHash Uint256) (bool, error)
	GetBlockRootWithNewTxRoot(txRoot Uint256) Uint256
	GetContractState(contractHash Address) (*payload.DeployCode, error)
	GetBookKeeperState() (*states.BookKeeperState, error)
	GetStorageItem(key *states.StorageKey) (*states.StorageItem, error)
	PreExecuteContract(tx *types.Transaction) ([]interface{}, error)
}


