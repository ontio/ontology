package actor

import (
	"fmt"
	"reflect"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology-eventbus/actor"
	ctypes "github.com/ontio/ontology/core/types"
	ttypes "github.com/ontio/ontology/txnpool/types"
	"github.com/ontio/ontology/txnpool/proc"
)


// TxnActor: Handle the low priority msg from P2P and API
type TxActor struct {
	txPoolServer *proc.TxPoolServer
}

// NewTxActor creates an actor to handle the transaction-based messages from
// network and http
func NewTxActor(s *proc.TxPoolServer) *TxActor {
	a := &TxActor{}
	a.setServer(s)
	return a
}

// handleTransaction handles a transaction from network and http
func (self *TxActor) handleTransaction(sender ttypes.SenderType, pid *actor.PID,
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
		self.txPoolServer.AssignTxToWorker(txn, sender)
	}
}

// Receive implements the actor interface
func (self *TxActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("txpool-tx actor started and be ready to receive tx msg")

	case *actor.Stopping:
		log.Warn("txpool-tx actor stopping")

	case *actor.Restarting:
		log.Warn("txpool-tx actor Restarting")

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
		log.Warn("txpool-tx actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *TxActor) setServer(s *proc.TxPoolServer) {
	self.txPoolServer = s
}

