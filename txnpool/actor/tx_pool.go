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

	case *ttypes.GetTxnPoolReq:
		sender := context.Sender()

		log.Debug("txpool actor Receives getting tx pool req from ", sender)

		res := self.server.GetTxEntrysFromPool(msg.ByCount, msg.Height)
		if sender != nil {
			sender.Request(&ttypes.GetTxnPoolRsp{TxnPool: res}, context.Self())
		}

	case *ttypes.GetPendingTxnReq:
		sender := context.Sender()

		log.Debug("txpool actor Receives getting pedning tx req from ", sender)

		res := self.server.GetPendingTxs(msg.ByCount)
		if sender != nil {
			sender.Request(&ttypes.GetPendingTxnRsp{Txs: res}, context.Self())
		}

	case *ttypes.VerifyBlockReq:
		sender := context.Sender()

		log.Debug("txpool actor Receives verifying block req from ", sender)

		self.server.HandleVerifyBlockReq(msg, sender)

	case *message.SaveBlockCompleteMsg:
		sender := context.Sender()

		log.Debug("txpool actor Receives block complete event from ", sender)

		if msg.Block != nil {
			self.server.RemoveTransactionsFromPool(msg.Block.Transactions)
		}

	default:
		log.Debug("txpool actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *TxPoolActor) setServer(svr *proc.TXPoolServer) {
	self.server = svr
}

