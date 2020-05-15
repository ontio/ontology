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
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/event"
	cstates "github.com/ontio/ontology/smartcontract/states"
)

type ExecuteResult struct {
	WriteSet        *overlaydb.MemDB
	Hash            common.Uint256
	MerkleRoot      common.Uint256
	CrossStates     []common.Uint256
	CrossStatesRoot common.Uint256
	Notify          []*event.ExecuteNotify
}

// LedgerStore provides func with store package.
type LedgerStore interface {
	InitLedgerStoreWithGenesisBlock(genesisblock *types.Block, defaultBookkeeper []keypair.PublicKey) error
	Close() error
	AddHeaders(headers []*types.Header) error
	AddBlock(block *types.Block, ccMsg *types.CrossChainMsg, stateMerkleRoot common.Uint256) error
	ExecuteBlock(b *types.Block) (ExecuteResult, error)                                       // called by consensus
	SubmitBlock(b *types.Block, crossChainMsg *types.CrossChainMsg, exec ExecuteResult) error // called by consensus
	GetStateMerkleRoot(height uint32) (result common.Uint256, err error)
	GetCurrentBlockHash() common.Uint256
	GetCurrentBlockHeight() uint32
	GetCurrentHeaderHeight() uint32
	GetCurrentHeaderHash() common.Uint256
	GetBlockHash(height uint32) common.Uint256
	GetHeaderByHash(blockHash common.Uint256) (*types.Header, error)
	GetRawHeaderByHash(blockHash common.Uint256) (*types.RawHeader, error)
	GetHeaderByHeight(height uint32) (*types.Header, error)
	GetBlockByHash(blockHash common.Uint256) (*types.Block, error)
	GetBlockByHeight(height uint32) (*types.Block, error)
	GetTransaction(txHash common.Uint256) (*types.Transaction, uint32, error)
	IsContainBlock(blockHash common.Uint256) (bool, error)
	IsContainTransaction(txHash common.Uint256) (bool, error)
	GetBlockRootWithNewTxRoots(startHeight uint32, txRoots []common.Uint256) common.Uint256
	GetMerkleProof(m, n uint32) ([]common.Uint256, error)
	GetContractState(contractHash common.Address) (*payload.DeployCode, error)
	GetBookkeeperState() (*states.BookkeeperState, error)
	GetStorageItem(key *states.StorageKey) (*states.StorageItem, error)
	PreExecuteContract(tx *types.Transaction) (*cstates.PreExecResult, error)
	PreExecuteContractBatch(txes []*types.Transaction, atomic bool) ([]*cstates.PreExecResult, uint32, error)
	GetEventNotifyByTx(tx common.Uint256) (*event.ExecuteNotify, error)
	GetEventNotifyByBlock(height uint32) ([]*event.ExecuteNotify, error)

	//cross chain states root
	GetCrossStatesRoot(height uint32) (common.Uint256, error)
	GetCrossChainMsg(height uint32) (*types.CrossChainMsg, error)
	GetCrossStatesProof(height uint32, key []byte) ([]byte, error)
	EnableBlockPrune(numBeforeCurr uint32)
}
