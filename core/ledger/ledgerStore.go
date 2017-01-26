package ledger

import (
	. "GoOnchain/common"
)

// ILedgerStore provides func with store package.
type ILedgerStore interface {
	//TODO: define the state store func
	SaveBlock(b *Block) error
	GetBlock(hash []byte) (*Block, error)
	GetBlockHash(height uint32) Uint256
	InitLedgerStore(ledger *Ledger) error
	GetHeader(hash Uint256) (*Header, error)
}
