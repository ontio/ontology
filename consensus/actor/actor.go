package actor

import (
	"errors"
	"time"

	"github.com/Ontology/core/types"
	ontErrors "github.com/Ontology/errors"
	"github.com/Ontology/eventbus/actor"
	txpool "github.com/Ontology/txnpool/common"
)

type TxPoolActor struct {
	Pool *actor.PID
}

func (self *TxPoolActor) GetTxnPool(byCount bool, height uint32) []*txpool.TXEntry {
	poolmsg := &txpool.GetTxnPoolReq{ByCount: byCount}
	future := self.Pool.RequestFuture(poolmsg, time.Second*10)
	entry, err := future.Result()
	if err != nil {
		return nil
	}

	txs := entry.(*txpool.GetTxnPoolRsp).TxnPool
	return txs
}

func (self *TxPoolActor) VerifyBlock(txs []*types.Transaction, height uint32) error {
	poolmsg := &txpool.VerifyBlockReq{Txs: txs, Height: height}
	future := self.Pool.RequestFuture(poolmsg, time.Second*10)
	entry, err := future.Result()
	if err != nil {
		return err
	}

	txentry := entry.(*txpool.VerifyBlockRsp).TxnPool
	for _, entry := range txentry {
		if entry.ErrCode != ontErrors.ErrNoError {
			return errors.New(entry.ErrCode.Error())
		}
	}

	return nil
}

type P2PActor struct {
	P2P *actor.PID
}

type LedgerActor struct {
	Ledger *actor.PID
}
