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
	server *proc.TXPoolServer
}

// NewTxActor creates an actor to handle the transaction-based messages from
// network and http
func NewTxActor(s *proc.TXPoolServer) *TxActor {
	a := &TxActor{}
	a.setServer(s)
	return a
}

// handleTransaction handles a transaction from network and http
func (self *TxActor) handleTransaction(sender ttypes.SenderType, pid *actor.PID,
	txn *ctypes.Transaction) {
	self.server.IncreaseStats(ttypes.RcvStats)

	if self.server.GetTransactionFromPool(txn.Hash()) != nil {
		log.Debug(fmt.Sprintf("Transaction %x already in the txn pool",
			txn.Hash()))

		self.server.IncreaseStats(ttypes.DuplicateStats)
	} else if self.server.GetTransactionCountFromPool() >= ttypes.MAX_CAPACITY {
		log.Warn("Transaction pool is full", txn.Hash())

		self.server.IncreaseStats(ttypes.FailureStats)
	} else {
		<-self.server.Slots
		self.server.AssignTxToWorker(txn, sender)
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

	case *ttypes.GetTxnReq:
		sender := context.Sender()

		log.Debug("txpool-tx actor Receives getting tx req from ", sender)

		res := self.server.GetTransactionFromPool(msg.Hash)
		if sender != nil {
			sender.Request(&ttypes.GetTxnRsp{Txn: res},
				context.Self())
		}

	case *ttypes.GetTxnStats:
		sender := context.Sender()

		log.Debug("txpool-tx actor Receives getting tx stats from ", sender)

		res := self.server.GetStats()
		if sender != nil {
			sender.Request(&ttypes.GetTxnStatsRsp{Count: res},
				context.Self())
		}

	case *ttypes.CheckTxnReq:
		sender := context.Sender()

		log.Debug("txpool-tx actor Receives checking tx req from ", sender)

		res := self.server.IsContainTx(msg.Hash)
		if sender != nil {
			sender.Request(&ttypes.CheckTxnRsp{Ok: res},
				context.Self())
		}

	case *ttypes.GetTxnStatusReq:
		sender := context.Sender()

		log.Debug("txpool-tx actor Receives getting tx status req from ", sender)

		res := self.server.GetTxStatus(msg.Hash)
		if sender != nil {
			if res == nil {
				sender.Request(&ttypes.GetTxnStatusRsp{Hash: msg.Hash,
					VerifyResults: nil}, context.Self())
			} else {
				sender.Request(&ttypes.GetTxnStatusRsp{Hash: res.Hash,
					VerifyResults: res.Attrs}, context.Self())
			}
		}

	default:
		log.Warn("txpool-tx actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *TxActor) setServer(s *proc.TXPoolServer) {
	self.server = s
}

