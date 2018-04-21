package actor

import (
	"reflect"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/txnpool/proc"
	tcomn "github.com/ontio/ontology/txnpool/common"
)

// TxnPoolActor: Handle the high priority request from Consensus
type TxPoolActor struct {
	server *proc.TXPoolServer
}

// NewTxPoolActor creates an actor to handle the messages from the consensus
func NewTxPoolActor(svr *proc.TXPoolServer) *TxPoolActor {
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

	case *tcomn.GetTxnPoolReq:
		sender := context.Sender()

		log.Debug("txpool actor Receives getting tx pool req from ", sender)

		res := self.server.GetTxPool(msg.ByCount, msg.Height)
		if sender != nil {
			sender.Request(&tcomn.GetTxnPoolRsp{TxnPool: res}, context.Self())
		}

	case *tcomn.GetPendingTxnReq:
		sender := context.Sender()

		log.Debug("txpool actor Receives getting pedning tx req from ", sender)

		res := self.server.GetPendingTxs(msg.ByCount)
		if sender != nil {
			sender.Request(&tcomn.GetPendingTxnRsp{Txs: res}, context.Self())
		}

	case *tcomn.VerifyBlockReq:
		sender := context.Sender()

		log.Debug("txpool actor Receives verifying block req from ", sender)

		self.server.VerifyBlock(msg, sender)

	case *message.SaveBlockCompleteMsg:
		sender := context.Sender()

		log.Debug("txpool actor Receives block complete event from ", sender)

		if msg.Block != nil {
			self.server.CleanTransactionList(msg.Block.Transactions)
		}

	default:
		log.Debug("txpool actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *TxPoolActor) setServer(svr *proc.TXPoolServer) {
	self.server = svr
}

