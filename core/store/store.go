package store

import (
	states "github.com/Ontology/core/states"
	."github.com/Ontology/common"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
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
	GetCurrentStateRoot() (Uint256, error)
	GetBlockRootWithNewTxRoot(txRoot Uint256) Uint256
	GetAssetState(assetId Uint256) (*states.AssetState, error)
	GetContractState(contractHash Uint160) (*states.ContractState, error)
	GetAccountState(programHash Uint160) (*states.AccountState, error)
	GetBookKeeperState() (*states.BookKeeperState, error)
	GetStorageItem(key *states.StorageKey) (*states.StorageItem, error)
	PreExecuteContract(tx *types.Transaction) ([]interface{}, error)
}

