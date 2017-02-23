package ledger

import (
	. "GoOnchain/common"
	. "GoOnchain/core/asset"
	tx "GoOnchain/core/transaction"
)

// ILedgerStore provides func with store package.
type ILedgerStore interface {
	//TODO: define the state store func
	SaveBlock(b *Block) error
	GetBlock(hash Uint256) (*Block, error)
	GetBlockHash(height uint32) (Uint256, error)
	InitLedgerStore(ledger *Ledger) error

	SaveHeader(header *Header) error
	GetHeader(hash Uint256) (*Header, error)

	GetTransaction(hash Uint256) (*tx.Transaction,error)

	SaveAsset(asset *Asset) error
	GetAsset(hash Uint256) (*Asset, error)

	GetCurrentBlockHash() Uint256
}
