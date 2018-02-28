package common

import (
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	"sync"
)

type TXNAttr struct {
	Height      uint32
	ValidatorID uint8
	Ok          bool
}

type TXNEntry struct {
	Txn   *types.Transaction // transaction which has been verified
	Fee   common.Fixed64     // Total fee per transaction
	Attrs []*TXNAttr         // the result from each validator
}

type TXNPool struct {
	sync.RWMutex
	txnList map[common.Uint256]*TXNEntry
}

func (tp *TXNPool) Init() {
	tp.Lock()
	defer tp.Unlock()
	tp.txnList = make(map[common.Uint256]*TXNEntry)
}

func (tp *TXNPool) AddTxnList(txnEntry *TXNEntry) bool {
	tp.Lock()
	defer tp.Unlock()
	txnHash := txnEntry.Txn.Hash()
	if _, ok := tp.txnList[txnHash]; ok {
		log.Info("Transaction %x already existed in the pool\n", txnHash)
		return false
	}

	tp.txnList[txnHash] = txnEntry
	return true
}

// clean the trasaction Pool with committed transactions.
func (tp *TXNPool) CleanTransactionList(txns []*types.Transaction) error {
	cleaned := 0
	txnsNum := len(txns)
	for _, txn := range txns {
		if txn.TxType == types.BookKeeping {
			txnsNum = txnsNum - 1
			continue
		}
		if tp.delTxnList(txn) {
			cleaned++
		}
	}
	if txnsNum != cleaned {
		log.Info(fmt.Sprintf("The Transactions num Unmatched. Expect %d,got %d .\n",
			txnsNum, cleaned))
	}
	log.Debug(fmt.Sprintf("[cleanTransactionList],transaction %d Requested,%d cleaned, Remains %d in TxPool",
		txnsNum, cleaned, tp.GetTransactionCount()))
	return nil
}

func (tp *TXNPool) CopyTxnList() map[common.Uint256]*TXNEntry {
	tp.RLock()
	defer tp.RUnlock()
	txnMap := make(map[common.Uint256]*TXNEntry, len(tp.txnList))
	for txnId, txnEntry := range tp.txnList {
		txnMap[txnId] = txnEntry
	}
	return txnMap
}

func (tp *TXNPool) delTxnList(txn *types.Transaction) bool {
	tp.Lock()
	defer tp.Unlock()
	txHash := txn.Hash()
	if _, ok := tp.txnList[txHash]; !ok {
		return false
	}
	delete(tp.txnList, txHash)
	return true
}

func (tp *TXNPool) GetTxnPool(byCount bool) []*TXNEntry {
	tp.RLock()
	defer tp.RUnlock()

	count := config.Parameters.MaxTxInBlock
	if count <= 0 {
		byCount = false
	}
	if len(tp.txnList) < count || !byCount {
		count = len(tp.txnList)
	}

	var num int
	txnList := make([]*TXNEntry, count)
	for _, txnEntry := range tp.txnList {
		txnList[num] = txnEntry
		num++
		if num >= count {
			break
		}
	}
	return txnList
}

func (tp *TXNPool) GetTransaction(hash common.Uint256) *types.Transaction {
	tp.RLock()
	defer tp.RUnlock()
	if txn := tp.txnList[hash]; txn == nil {
		return nil
	}
	return tp.txnList[hash].Txn
}

func (tp *TXNPool) GetTxnStatus(hash common.Uint256) *TXNEntry {
	tp.RLock()
	defer tp.RUnlock()
	return tp.txnList[hash]
}

func (tp *TXNPool) GetTransactionCount() int {
	tp.RLock()
	defer tp.RUnlock()
	return len(tp.txnList)
}

func (tp *TXNPool) GetUnverifiedTxs(txs []*types.Transaction) []*types.Transaction {
	tp.RLock()
	defer tp.RUnlock()
	ret := []*types.Transaction{}
	for _, t := range txs {
		if t.TxType == types.BookKeeping {
			continue
		}
		if _, ok := tp.txnList[t.Hash()]; !ok {
			ret = append(ret, t)
		}
	}
	return ret
}
