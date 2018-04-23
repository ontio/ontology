package actor

import (
	"reflect"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/txnpool/proc"
	ttypes "github.com/ontio/ontology/txnpool/types"
)

// TxnPoolActor: Handle the high priority request from Consensus
type TxPoolActor struct {
	txPoolServer *proc.TxPoolServer
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
		self.txPoolServer.AddVerifyBlock(msg.Height,msg.Txs, sender)

	case *message.SaveBlockCompleteMsg:
		sender := context.Sender()

		log.Debug("txpool actor Receives block complete event from ", sender)

		if msg.Block != nil {
			self.txPoolServer.RemoveTransactionsFromPool(msg.Block.Transactions)
		}

	default:
		log.Debug("txpool actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *TxPoolActor) setServer(svr *proc.TxPoolServer) {
	self.txPoolServer = svr
}

