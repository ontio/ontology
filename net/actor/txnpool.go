package actor

import (
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/core/types"
	"github.com/Ontology/eventbus/actor"
	"sync"
	"time"
)

var TxnPid *actor.PID

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

type GetTxnStatsReq struct{}
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
	Txn   *types.Transaction
	Attrs []*TXNAttr
}

type GetTxnPoolRsp struct {
	TxnPool []*TXNEntry
}

type CleanTxnPoolReq struct {
	TxnPool []*types.Transaction
}

type GetTxnReq struct {
	hash common.Uint256
}
type GetTxnRsp struct {
	txn *types.Transaction
}

func AddTransaction(transaction *types.Transaction) {
	TxnPid.Tell(transaction)
}

func CleanTxnPool(TxnPool []*types.Transaction) {
	TxnPid.Tell(&CleanTxnPoolReq{TxnPool})
}

func GetTxnPool(byCount bool) []*TXNEntry {
	future := TxnPid.RequestFuture(&GetTxnPoolReq{ByCount: byCount}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetTxnPoolRsp).TxnPool
}

func GetTransaction(hash common.Uint256) *types.Transaction {
	future := TxnPid.RequestFuture(&GetTxnReq{hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetTxnRsp).txn
}

func GetTxnStats(hash common.Uint256) *[]uint64 {
	future := TxnPid.RequestFuture(&GetTxnStatsReq{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetTxnStatsRsp).count
}
