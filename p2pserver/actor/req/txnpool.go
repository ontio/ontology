package req

import (
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	"github.com/Ontology/errors"
	"github.com/Ontology/eventbus/actor"
	. "github.com/Ontology/txnpool/common"
	"time"
)

var TxnPoolPid *actor.PID

func SetTxnPoolPid(txnPid *actor.PID) {
	TxnPoolPid = txnPid
}

func AddTransaction(transaction *types.Transaction) {
	TxnPoolPid.Tell(transaction)
}

func GetTxnPool(byCount bool) ([]*TXEntry, error) {
	future := TxnPoolPid.RequestFuture(&GetTxnPoolReq{ByCount: byCount}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetTxnPoolRsp).TxnPool, nil
}

func GetTransaction(hash common.Uint256) (*types.Transaction, error) {
	future := TxnPoolPid.RequestFuture(&GetTxnReq{Hash: hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetTxnRsp).Txn, nil
}

func CheckTransaction(hash common.Uint256) (bool, error) {
	future := TxnPoolPid.RequestFuture(&CheckTxnReq{Hash: hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return false, err
	}
	return result.(CheckTxnRsp).Ok, nil
}

func GetTransactionStatus(hash common.Uint256) ([]*TXAttr, error) {
	future := TxnPoolPid.RequestFuture(&GetTxnStatusReq{Hash: hash}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetTxnStatusRsp).TxStatus, nil
}

func GetPendingTxn(byCount bool) ([]*types.Transaction, error) {
	future := TxnPoolPid.RequestFuture(&GetPendingTxnReq{ByCount: byCount}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetPendingTxnRsp).Txs, nil
}

func VerifyBlock(height uint32, txs []*types.Transaction) ([]*VerifyTxResult, error) {
	future := TxnPoolPid.RequestFuture(&VerifyBlockReq{Height: height, Txs: txs}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(VerifyBlockRsp).TxnPool, nil
}

func GetTransactionStats(hash common.Uint256) (*[]uint64, error) {
	future := TxnPoolPid.RequestFuture(&GetTxnStats{}, 5*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Error(errors.NewErr("ERROR: "), err)
		return nil, err
	}
	return result.(GetTxnStatsRsp).Count, nil
}
