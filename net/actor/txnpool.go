package actor

import (
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/core/types"
	"github.com/Ontology/eventbus/actor"
	. "github.com/Ontology/txnpool/common"
	tp "github.com/Ontology/txnpool/proc"
	"time"
)

var TxnPid *actor.PID

func AddTransaction(transaction *types.Transaction) {
	TxnPid.Tell(transaction)
}

func GetTxnPool(byCount bool) []*TXEntry {
	future := TxnPid.RequestFuture(&GetTxnPoolReq{ByCount: byCount}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetTxnPoolRsp).TxnPool
}

func GetTransaction(hash common.Uint256) *types.Transaction {
	future := TxnPid.RequestFuture(&GetTxnReq{Hash:hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetTxnRsp).Txn
}

func CheckTransaction(hash common.Uint256) bool {
	future := TxnPid.RequestFuture(&CheckTxnReq{Hash:hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(CheckTxnRsp).Ok
}

func GetTransactionStatus(hash common.Uint256) []*TXAttr {
	future := TxnPid.RequestFuture(&GetTxnStatusReq{Hash:hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetTxnStatusRsp).TxStatus
}

func GetPendingTxn(byCount bool) []*types.Transaction {
	future := TxnPid.RequestFuture(&GetPendingTxnReq{ByCount:byCount}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetPendingTxnRsp).Txs
}

func VerifyBlock(height uint32, txs []*types.Transaction) []*VerifyTxResult {
	future := TxnPid.RequestFuture(&VerifyBlockReq{Height:height, Txs:txs}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(VerifyBlockRsp).TxnPool
}

func GetTransactionStats(hash common.Uint256) *[]uint64 {
	future := TxnPid.RequestFuture(&GetTxnStats{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetTxnStatsRsp).Count
}

