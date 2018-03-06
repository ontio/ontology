package actor

import (
	"github.com/Ontology/transaction"
	"github.com/Ontology/common"
	"github.com/ONTID/eventbus/actor"
	"time"
	"fmt"
	"sync"
)

var TxnPid actor.PID

type TxnStatsType uint8
const (
	_ TxnStatsType = iota
	RcvStats
	SuccessStats
	FailureStats
	DuplicateStats
	SigErrStats
	StateErrStats
	MAXSTATS
)

type txnStats struct {
	sync.RWMutex
	count []uint64
}

type GetTxnStatsReq struct {}
type GetTxnStatsRsp struct {
	count *[]uint64
}

type GetTxnPoolReq struct {
	ByCount bool
}

type TXNAttr struct {
	Height      uint32
	ValidatorID uint8
	Ok          bool
}

type TXNEntry struct {
	Txn   *transaction.Transaction
	Attrs []*TXNAttr
}

type GetTxnPoolRsp struct {
	TxnPool []*TXNEntry
}

type CleanTxnPoolReq struct {
	TxnPool []*transaction.Transaction
}

type GetTxnReq struct {
	hash common.Uint256
}
type GetTxnRsp struct {
	txn *transaction.Transaction
}

func AddTransaction(transaction *transaction.Transaction){
	TxnPid.Tell(transaction)
}

func CleanTxnPool(TxnPool []*transaction.Transaction){
	TxnPid.Tell(&CleanTxnPoolReq{TxnPool})
}

func GetTxnPool(byCount bool)([]*TXNEntry){
	future := TxnPid.RequestFuture(&GetTxnPoolReq{ByCount: byCount}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetTxnPoolRsp).TxnPool
}

func GetTransaction(hash common.Uint256)(*transaction.Transaction){
	future := TxnPid.RequestFuture(&GetTxnReq{hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetTxnRsp).txn
}

func GetTxnStats(hash common.Uint256)(*[]uint64){
	future := TxnPid.RequestFuture(&GetTxnStatsReq{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetTxnStatsRsp).count
}
