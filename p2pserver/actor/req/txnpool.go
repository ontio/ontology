package req

import (
	"time"

	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	"github.com/Ontology/errors"
	"github.com/Ontology/eventbus/actor"
	txnpool "github.com/Ontology/txnpool/common"
)

var TxnPoolPid *actor.PID

func SetTxnPoolPid(txnPid *actor.PID) {
	TxnPoolPid = txnPid
}

func AddTransaction(transaction *types.Transaction) {
	TxnPoolPid.Tell(transaction)
}

func GetTxnPool(byCount bool) ([]*txnpool.TXEntry, error) {
	future := TxnPoolPid.RequestFuture(&txnpool.GetTxnPoolReq{ByCount: byCount}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(txnpool.GetTxnPoolRsp).TxnPool, nil
}

func GetTransaction(hash common.Uint256) (*types.Transaction, error) {
	future := TxnPoolPid.RequestFuture(&txnpool.GetTxnReq{Hash: hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(txnpool.GetTxnRsp).Txn, nil
}

func CheckTransaction(hash common.Uint256) (bool, error) {
	future := TxnPoolPid.RequestFuture(&txnpool.CheckTxnReq{Hash: hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return false, err
	}
	return result.(txnpool.CheckTxnRsp).Ok, nil
}

func GetTransactionStatus(hash common.Uint256) ([]*txnpool.TXAttr, error) {
	future := TxnPoolPid.RequestFuture(&txnpool.GetTxnStatusReq{Hash: hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(txnpool.GetTxnStatusRsp).TxStatus, nil
}

func GetPendingTxn(byCount bool) ([]*types.Transaction, error) {
	future := TxnPoolPid.RequestFuture(&txnpool.GetPendingTxnReq{ByCount: byCount}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(txnpool.GetPendingTxnRsp).Txs, nil
}

func VerifyBlock(height uint32, txs []*types.Transaction) ([]*txnpool.VerifyTxResult, error) {
	future := TxnPoolPid.RequestFuture(&txnpool.VerifyBlockReq{Height: height, Txs: txs}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(txnpool.VerifyBlockRsp).TxnPool, nil
}

func GetTransactionStats(hash common.Uint256) (*[]uint64, error) {
	future := TxnPoolPid.RequestFuture(&txnpool.GetTxnStats{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(txnpool.GetTxnStatsRsp).Count, nil
}
