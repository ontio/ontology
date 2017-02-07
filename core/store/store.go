package store

import(
	. "GoOnchain/core/ledger"
	. "GoOnchain/core/store/LevelDBStore"
)

func NewLedgerStore() ILedgerStore {
	// TODO: read config file decide which db to use.
	ldbs,_ := NewLevelDBStore("E:\\AntSharesCore\\Chain")

	return ldbs
}


