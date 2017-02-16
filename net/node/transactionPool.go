package node

import (
	"GoOnchain/common"
	"GoOnchain/core/transaction"
	"sync"
)

type TXNPool struct {
	lock sync.RWMutex
	list map[common.Uint256]*transaction.Transaction
}

// Reference process
// func (net neter) AddTransaction(Transaction tx) bool {
// 	if (Blockchain.Default == null) {
// 		return false
// 	}

// 	//TODO Lock the memory pool
// 	if (MemoryPool.ContainsKey(tx.Hash)) {
// 		return false
// 	}
// 	if (Blockchain.Default.ContainsTransaction(tx.Hash)) {
// 		return false
// 	}
// 	if (!tx.Verify(MemoryPool.Values)) {
// 		return false
// 	}
// 	AddingTransactionEventArgs args = new AddingTransactionEventArgs(tx);
// 	AddingTransaction.Invoke(this, args);
// 	if (!args.Cancel) MemoryPool.Add(tx.Hash, tx);
// 	return !args.Cancel;

// }

func (txnPool TXNPool) GetTransaction(hash common.Uint256) *transaction.Transaction {
	// Fixme need lock
	return txnPool.list[hash]
}

func (txnPool *TXNPool) AppendTxnPool(txn *transaction.Transaction) bool {
	// TODO add the lock
	txnPool.list[txn.Hash()] = txn
	// TODO Check the TXN already existed case
	return true
}

func (txnPool TXNPool) GetTxnPool() map[common.Uint256]*transaction.Transaction {
	return txnPool.list
}

func (txnPool *TXNPool) init() {
	txnPool.list = make(map[common.Uint256]*transaction.Transaction)
}
