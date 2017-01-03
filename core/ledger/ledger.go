package ledger


import (
	tx "GoOnchain/core/transaction"
)
// Ledger - the struct for onchainDNA ledger
type Ledger struct {
	Blockchain *Blockchain
	State      *State
}

func (l *Ledger) IsDoubleSpend(Tx *tx.Transaction) error {
	//TODO: implement ledger IsDoubleSpend

	return  nil
}

