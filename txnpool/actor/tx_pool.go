package actor

import (
	"reflect"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/txnpool/proc"
	ttypes "github.com/ontio/ontology/txnpool/types"
	ctypes "github.com/ontio/ontology/core/types"
	tcomn "github.com/ontio/ontology/txnpool/common"
	"fmt"
)

// TxnPoolActor: Handle the high priority request from Consensus
type TxPoolActor struct {
	txPoolServer *proc.TxPoolServer
	sender       *tcomn.Sender
}

// NewTxPoolActor creates an actor to handle the messages from the consensus
func NewTxPoolActor(svr *proc.TxPoolServer) *TxPoolActor {
	a := &TxPoolActor{}
	a.setServer(svr)
	return a
}

// Receive implements the actor interface
func (self *TxPoolActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("txpool actor started and be ready to receive txPool msg")

	case *actor.Stopping:
		log.Warn("txpool actor stopping")

	case *actor.Restarting:
		log.Warn("txpool actor Restarting")

	case *ttypes.GetTxnPoolReq:
		sender := context.Sender()
		log.Debug("txpool actor Receives getting tx pool req from ", sender)
		res := self.txPoolServer.GetTxEntrysFromPool(msg.ByCount, msg.Height)
		if sender != nil {
			sender.Request(&ttypes.GetTxnPoolRsp{TxnPool: res}, context.Self())
		}

	case *ttypes.GetPendingTxnReq:
		sender := context.Sender()
		log.Debug("txpool actor Receives getting pedning tx req from ", sender)
		res := self.txPoolServer.GetPendingTxs(msg.ByCount)
		if sender != nil {
			sender.Request(&ttypes.GetPendingTxnRsp{Txs: res}, context.Self())
		}

	case *ttypes.VerifyBlockReq:
		sender := context.Sender()
		log.Debug("txpool actor Receives verifying block req from ", sender)
		if msg == nil || len(msg.Txs) == 0 {
			return
		}
		self.txPoolServer.AddVerifyBlock(msg.Height, msg.Txs, sender)

	case *message.SaveBlockCompleteMsg:
		sender := context.Sender()
		log.Debug("txpool actor Receives block complete event from ", sender)
		if msg.Block != nil {
			self.txPoolServer.RemoveTransactionsFromPool(msg.Block.Transactions)
		}
		//below is tx status
	case *ttypes.AppendTxReq:
		sender := msg.Sender
		log.Debug("txpool-tx actor Receives tx from ", sender.Sender())
		self.handleTransaction(sender, context.Self(), msg.Tx)

	case *ttypes.GetTxFromPoolReq:
		sender := context.Sender()
		log.Debug("txpool-tx actor Receives getting tx req from ", sender)
		res := self.txPoolServer.GetTxFromPool(msg.Hash)
		if sender != nil {
			sender.Request(&ttypes.GetTxFromPoolRsp{Txn: res},
				context.Self())
		}

	case *ttypes.GetTxVerifyResultStaticsReq:
		sender := context.Sender()
		log.Debug("txpool-tx actor Receives getting tx stats from ", sender)
		res := self.txPoolServer.GetVerifyResultStatistics()
		if sender != nil {
			sender.Request(&ttypes.GetTxVerifyResultStaticsRsp{Count: res},
				context.Self())
		}

	case *ttypes.IsTxInPoolReq:
		sender := context.Sender()
		log.Debug("txpool-tx actor Receives checking tx req from ", sender)
		res := self.txPoolServer.IsContainTx(msg.Hash)
		if sender != nil {
			sender.Request(&ttypes.IsTxInPoolRsp{Ok: res},
				context.Self())
		}

	case *ttypes.GetTxVerifyResultReq:
		sender := context.Sender()
		log.Debug("txpool-tx actor Receives getting tx status req from ", sender)
		res := self.txPoolServer.GetTxVerifyStatus(msg.Hash)
		if sender != nil {
			if res == nil {
				sender.Request(&ttypes.GetTxVerifyResultRsp{Hash: msg.Hash,
					VerifyResults: nil}, context.Self())
			} else {
				sender.Request(&ttypes.GetTxVerifyResultRsp{Hash: res.Hash,
					VerifyResults: res.VerifyResults}, context.Self())
			}
		}

	default:
		log.Debug("txpool actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

// handleTransaction handles a transaction from network and http
func (self *TxPoolActor) handleTransaction(sender ttypes.SenderType, pid *actor.PID,
	txn *ctypes.Transaction) {
	self.txPoolServer.Increase(ttypes.Received)

	if self.txPoolServer.GetTxFromPool(txn.Hash()) != nil {
		log.Debug(fmt.Sprintf("Transaction %x already in the txn pool",
			txn.Hash()))
		self.txPoolServer.Increase(ttypes.Duplicate)
	} else if self.txPoolServer.GetTxCountFromPool() >= ttypes.MAX_CAPACITY {
		log.Warn("Transaction pool is full", txn.Hash())
		self.txPoolServer.Increase(ttypes.Failure)
	} else {
		<-self.txPoolServer.Slots
		self.txPoolServer.PutTransaction(txn, sender)
	}
}

func (self *TxPoolActor) setServer(svr *proc.TxPoolServer) {
	self.txPoolServer = svr
}
