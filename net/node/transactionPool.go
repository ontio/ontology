package node

import (
	"fmt"
	"sort"
	"sync"

	"github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/types"
	va "github.com/Ontology/core/validation"
	ontError "github.com/Ontology/errors"
)

// Genesis transaction will not be added to pool, so use height == 0 to indicate transaction not packed in block
type PoolTransaction struct {
	tx *types.Transaction
	//height uint32
	fee common.Fixed64
}

type TXNPool struct {
	sync.RWMutex
	txnList map[common.Uint256]*PoolTransaction // transaction which have been verifyed will put into this map
}

func (this *TXNPool) init() {
	this.txnList = make(map[common.Uint256]*PoolTransaction)
}

//append transaction to txnpool when check ok.
//1.check transaction. 2.check with ledger(db) 3.check with pool
func (this *TXNPool) AppendTxnPool(txn *types.Transaction) ontError.ErrCode {
	//verify transaction with Concurrency
	if errCode := va.VerifyTransaction(txn); errCode != ontError.ErrNoError {
		log.Infof("Transaction verification failed %x\n", txn.Hash())
		return errCode
	}
	if errCode := va.VerifyTransactionWithLedger(txn, ledger.DefaultLedger); errCode != ontError.ErrNoError {
		log.Infof("Transaction verification with ledger failed %x\n", txn.Hash())
		return errCode
	}

	//verify transaction by pool with lock
	this.RLock()
	defer this.RUnlock()
	if _, ok := this.txnList[txn.Hash()]; ok {
		return ontError.ErrDuplicatedTx
	}

	this.txnList[txn.Hash()] = &PoolTransaction{
		tx:  txn,
		fee: txn.GetTotalFee(),
	}

	return ontError.ErrNoError
}

//get the transaction in txnpool
func (this *TXNPool) GetTxnPool(byCount bool) (map[common.Uint256]*types.Transaction, common.Fixed64) {
	this.Lock()
	defer this.Unlock()

	orderByFee := make([]*PoolTransaction, 0, len(this.txnList))
	for _, ptx := range this.txnList {
		orderByFee = append(orderByFee, ptx)
	}
	sort.Sort(OrderByNetWorkFee(orderByFee))

	maxcount := config.Parameters.MaxTxInBlock
	if maxcount <= 0 || byCount == false || maxcount > len(orderByFee) {
		maxcount = len(orderByFee)
	}

	txnMap := make(map[common.Uint256]*types.Transaction, maxcount)
	var networkFeeSum common.Fixed64
	for i := 0; i < maxcount; i++ {
		ptx := orderByFee[i]
		networkFeeSum += common.Fixed64(ptx.fee)
		txnMap[ptx.tx.Hash()] = ptx.tx
	}

	return txnMap, networkFeeSum
}

//get the transaction in txnpool
func (this *TXNPool) GetTxnPoolTxlist() []common.Uint256 {
	this.RLock()
	list := make([]common.Uint256, 0, len(this.txnList))
	for k := range this.txnList {
		list = append(list, k)
	}
	this.RUnlock()
	return list
}

//clean the trasaction Pool with committed block.
func (this *TXNPool) CleanTransactions(txns []*types.Transaction) error {
	this.Lock()
	defer this.Unlock()

	txnsNum := len(txns)
	poolLen := len(this.txnList)
	for _, txn := range txns {
		if txn.TxType == types.BookKeeping {
			txnsNum = txnsNum - 1
			continue
		}

		delete(this.txnList, txn.Hash())
	}

	cleaned := poolLen - len(this.txnList)
	if txnsNum != cleaned {
		log.Info(fmt.Sprintf("The Transactions num Unmatched. Expect %d, got %d .\n", txnsNum, cleaned))
	}
	log.Debug(fmt.Sprintf("[cleanTransactionList],transaction %d Requested, %d cleaned, Remains %d in TxPool", txnsNum, cleaned, this.GetTransactionCount()))

	return nil
}

//get the transaction by hash
func (this *TXNPool) GetTransaction(hash common.Uint256) *types.Transaction {
	this.RLock()
	defer this.RUnlock()
	return this.txnList[hash].tx
}

// clean the trasaction Pool with committed transactions.
func (this *TXNPool) GetTransactionCount() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.txnList)
}

type OrderByNetWorkFee []*PoolTransaction

func (n OrderByNetWorkFee) Len() int { return len(n) }

func (n OrderByNetWorkFee) Swap(i, j int) { n[i], n[j] = n[j], n[i] }

func (n OrderByNetWorkFee) Less(i, j int) bool { return n[j].fee < n[i].fee }
