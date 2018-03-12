package actor

import (
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	"github.com/Ontology/eventbus/actor"
	. "github.com/Ontology/txnpool/common"
	"github.com/Ontology/errors"
	"time"
)

var txnPoolPid *actor.PID

func SetTxnPoolPid(txnPid *actor.PID){
	txnPoolPid = txnPid
}

func AddTransaction(transaction *types.Transaction) {
	txnPoolPid.Tell(transaction)
}

func GetTxnPool(byCount bool) ([]*TXEntry, error) {
	future := txnPoolPid.RequestFuture(&GetTxnPoolReq{ByCount: byCount}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetTxnPoolRsp).TxnPool, nil
}

func GetTransaction(hash common.Uint256) (*types.Transaction, error) {
	future := txnPoolPid.RequestFuture(&GetTxnReq{Hash:hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetTxnRsp).Txn, nil
}

func CheckTransaction(hash common.Uint256) (bool, error) {
	future := txnPoolPid.RequestFuture(&CheckTxnReq{Hash:hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return false, err
	}
	return result.(CheckTxnRsp).Ok, nil
}

func GetTransactionStatus(hash common.Uint256) ([]*TXAttr, error) {
	future := txnPoolPid.RequestFuture(&GetTxnStatusReq{Hash:hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetTxnStatusRsp).TxStatus, nil
}

func GetPendingTxn(byCount bool) ([]*types.Transaction, error) {
	future := txnPoolPid.RequestFuture(&GetPendingTxnReq{ByCount:byCount}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetPendingTxnRsp).Txs, nil
}

func VerifyBlock(height uint32, txs []*types.Transaction) ([]*VerifyTxResult, error) {
	future := txnPoolPid.RequestFuture(&VerifyBlockReq{Height:height, Txs:txs}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(VerifyBlockRsp).TxnPool, nil
}

func GetTransactionStats(hash common.Uint256) (*[]uint64, error) {
	future := txnPoolPid.RequestFuture(&GetTxnStats{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetTxnStatsRsp).Count, nil
}

