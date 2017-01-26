package ledger

import (
	. "GoOnchain/common"
	tx "GoOnchain/core/transaction"
	"GoOnchain/crypto"
	. "GoOnchain/errors"
	"errors"
)

var DefaultLedger *Ledger

// Ledger - the struct for onchainDNA ledger
type Ledger struct {
	Blockchain *Blockchain
	State      *State
	Store      ILedgerStore
}

func (l *Ledger) IsDoubleSpend(Tx *tx.Transaction) error {
	//TODO: implement ledger IsDoubleSpend

	return nil
}

func GetDefaultLedger() (*Ledger, error) {
	if DefaultLedger == nil {
		return nil, NewDetailErr(errors.New("DefaultLedger GetDefaultLedger failed,DefaultLedger not Exist."), ErrNoCode, "")
	}
	return DefaultLedger, nil
}

func GetMinerAddress(miners []*crypto.PubKey) Uint160 {
	//TODO: GetMinerAddress()
	return Uint160{}
}
