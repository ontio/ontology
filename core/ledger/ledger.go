package ledger

import (
	tx "GoOnchain/core/transaction"
	"GoOnchain/crypto"
	. "GoOnchain/common"
)

var DefaultLedger *Ledger

// Ledger - the struct for onchainDNA ledger
type Ledger struct {
	Blockchain *Blockchain
	State      *State
}

func (l *Ledger) IsDoubleSpend(Tx *tx.Transaction) error {
	//TODO: implement ledger IsDoubleSpend

	return nil
}

func GetMinerAddress(miners []*crypto.PubKey) Uint160 {
	//TODO: GetMinerAddress()
	return Uint160{}
}
