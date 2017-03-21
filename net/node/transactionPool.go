package node

import (
	"GoOnchain/common"
	"GoOnchain/core/transaction"
	msg "GoOnchain/net/message"
	. "GoOnchain/net/protocol"
	"sync"
)

type TXNPool struct {
	sync.RWMutex
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

func (txnPool *TXNPool) GetTransaction(hash common.Uint256) *transaction.Transaction {
	txnPool.RLock()
	defer txnPool.RUnlock()
	txn := txnPool.list[hash]
	// Fixme need lock
	return txn
}

func (txnPool *TXNPool) AppendTxnPool(txn *transaction.Transaction) bool {
	txnPool.Lock()
	defer txnPool.Unlock()

	// TODO Check the TXN already existed case
	txnPool.list[txn.Hash()] = txn

	return true
}

// Attention: clean the trasaction Pool after the consensus confirmed all of the transcation
func (txnPool *TXNPool) GetTxnPool(cleanPool bool) map[common.Uint256]*transaction.Transaction {
	txnPool.Lock()
	defer txnPool.Unlock()

	list := txnPool.list
	if (cleanPool == true) {
		txnPool.list = make(map[common.Uint256]*transaction.Transaction)
	}
	return list
}

func (txnPool *TXNPool) init() {
	txnPool.list = make(map[common.Uint256]*transaction.Transaction)
}

func (node *node) SynchronizeTxnPool() {
	node.nbrNodes.RLock()
	defer node.nbrNodes.RUnlock()

	for _, n := range node.nbrNodes.List {
		if n.state == ESTABLISH {
			msg.ReqTxnPool(n)
		}
	}
}
