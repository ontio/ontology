package ledger

import (
	tx "GoOnchain/core/transaction"
)

// TranscationPool provides storage for transactions in the pending
// transaction pool.
type TranscationPool interface {

	//  add a transaction to the pool.
	Add(*tx.Transaction) error

	//returns all transactions that were in the pool.
	Dump() ([]*tx.Transaction, error)
}