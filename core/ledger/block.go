package ledger

import (
	tx "GoOnchain/core/transaction"
)

type Block struct {
	Blockheader *Blockheader
	Transcations []*tx.Transaction
}

type Blockheader struct {
	//TODO: define the Blockheader struct(define new uinttype)
	Version uint
	Height uint
	Timestamp uint
	nonce uint64
}