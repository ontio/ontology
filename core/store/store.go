package store

import(
	. "github.com/DNAProject/DNA/core/ledger"
	. "github.com/DNAProject/DNA/core/store/LevelDBStore"
)

func NewLedgerStore() ILedgerStore {
	// TODO: read config file decide which db to use.
	ldbs,_ := NewLevelDBStore("Chain")

	return ldbs
}


