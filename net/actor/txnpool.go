package actor

import (
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/core/types"
	"github.com/Ontology/eventbus/actor"
	. "github.com/Ontology/txnpool/common"
	"time"
)

var txnPoolPid *actor.PID

func SetTxnPoolPid(txnPid *actor.PID){
	txnPoolPid = txnPid
}

func AddTransaction(transaction *types.Transaction) {
	txnPoolPid.Tell(transaction)
}

func GetTxnPool(byCount bool) []*TXEntry {
	future := txnPoolPid.RequestFuture(&GetTxnPoolReq{ByCount: byCount}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetTxnPoolRsp).TxnPool
}

func GetTransaction(hash common.Uint256) *types.Transaction {
	future := txnPoolPid.RequestFuture(&GetTxnReq{Hash:hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetTxnRsp).Txn
}

func CheckTransaction(hash common.Uint256) bool {
	future := txnPoolPid.RequestFuture(&CheckTxnReq{Hash:hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(CheckTxnRsp).Ok
}

func GetTransactionStatus(hash common.Uint256) []*TXAttr {
	future := txnPoolPid.RequestFuture(&GetTxnStatusReq{Hash:hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetTxnStatusRsp).TxStatus
}

func GetPendingTxn(byCount bool) []*types.Transaction {
	future := txnPoolPid.RequestFuture(&GetPendingTxnReq{ByCount:byCount}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetPendingTxnRsp).Txs
}

func VerifyBlock(height uint32, txs []*types.Transaction) []*VerifyTxResult {
	future := txnPoolPid.RequestFuture(&VerifyBlockReq{Height:height, Txs:txs}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(VerifyBlockRsp).TxnPool
}

func GetTransactionStats(hash common.Uint256) *[]uint64 {
	future := txnPoolPid.RequestFuture(&GetTxnStats{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
	return result.(GetTxnStatsRsp).Count
}

