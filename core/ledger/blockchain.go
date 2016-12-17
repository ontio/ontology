package ledger

import (
	tx "GoOnchain/core/transaction"
)


// Store provides storage for State data
type BlockchainStore interface {
	//TODO: define the state store func
	SaveBlock(*Block) error
}


type Blockchain struct {
	Store BlockchainStore
	TxPool tx.TransactionPool
	Transactions []*tx.Transaction
}

