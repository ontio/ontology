package proc

import (
	"fmt"
	"reflect"

	"github.com/Ontology/common/log"
	tx "github.com/Ontology/core/types"
	"github.com/Ontology/errors"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/events/message"
	tc "github.com/Ontology/txnpool/common"
	"github.com/Ontology/validator/types"
)

func NewTxActor(s *TXPoolServer) *TxActor {
	a := &TxActor{}
	a.setServer(s)
	return a
}

func NewTxPoolActor(s *TXPoolServer) *TxPoolActor {
	a := &TxPoolActor{}
	a.setServer(s)
	return a
}

func NewVerifyRspActor(s *TXPoolServer) *VerifyRspActor {
	a := &VerifyRspActor{}
	a.setServer(s)
	return a
}

// TxnActor: Handle the low priority msg from P2P and API
type TxActor struct {
	server *TXPoolServer
}

// Handle a new transaction
func (ta *TxActor) handleTransaction(sender, self *actor.PID,
	txn *tx.Transaction) {
	ta.server.increaseStats(tc.RcvStats)

	if ta.server.getTransaction(txn.Hash()) != nil {
		log.Info(fmt.Sprintf("Transaction %x already in the txn pool",
			txn.Hash()))

		ta.server.increaseStats(tc.DuplicateStats)

		if sender == nil {
			return
		}
		rsp := &tc.TxRsp{
			Hash:    txn.Hash(),
			ErrCode: errors.ErrNoError,
		}
		sender.Request(rsp, self)
	} else if ta.server.getTransactionCount() >= tc.MAXCAPACITY {
		log.Info("Transaction pool is full", txn.Hash())

		ta.server.increaseStats(tc.FailureStats)

		if sender == nil {
			return
		}

		rsp := &tc.TxRsp{
			Hash:    txn.Hash(),
			ErrCode: errors.ErrUnknown,
		}
		sender.Request(rsp, self)
	} else {
		ta.server.assginTXN2Worker(txn, sender)
	}
}

func (ta *TxActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("Server started and be ready to receive tx msg")

	case *actor.Stopping:
		log.Info("Server stopping")

	case *actor.Restarting:
		log.Info("Server Restarting")

	case *tx.Transaction:
		sender := context.Sender()

		log.Info("Server Receives tx from ", sender)

		ta.handleTransaction(sender, context.Self(), msg)

	case *tc.GetTxnReq:
		sender := context.Sender()

		log.Info("Server Receives getting tx req from ", sender)

		res := ta.server.getTransaction(msg.Hash)
		if sender != nil {
			sender.Request(&tc.GetTxnRsp{Txn: res},
				context.Self())
		}

	case *tc.GetTxnStats:
		sender := context.Sender()

		log.Info("Server Receives getting tx stats from ", sender)

		res := ta.server.getStats()
		if sender != nil {
			sender.Request(&tc.GetTxnStatsRsp{Count: res},
				context.Self())
		}

	case *tc.CheckTxnReq:
		sender := context.Sender()

		log.Info("Server Receives checking tx req from ", sender)

		res := ta.server.checkTx(msg.Hash)
		if sender != nil {
			sender.Request(&tc.CheckTxnRsp{Ok: res},
				context.Self())
		}

	case *tc.GetTxnStatusReq:
		sender := context.Sender()

		log.Info("Server Receives getting tx status req from ", sender)

		res := ta.server.getTxStatusReq(msg.Hash)
		if sender != nil {
			sender.Request(&tc.GetTxnStatusRsp{Hash: res.Hash,
				TxStatus: res.Attrs}, context.Self())
		}

	default:
		log.Info("txpool-tx actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (ta *TxActor) setServer(s *TXPoolServer) {
	ta.server = s
}

// TxnPoolActor: Handle the high priority request from Consensus
type TxPoolActor struct {
	server *TXPoolServer
}

func (tpa *TxPoolActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("txpool actor started and be ready to receive txPool msg")

	case *actor.Stopping:
		log.Info("txpool actor stopping")

	case *actor.Restarting:
		log.Info("txpool actor Restarting")

	case *tc.GetTxnPoolReq:
		sender := context.Sender()

		log.Info("txpool actor Receives getting tx pool req from ", sender)

		res := tpa.server.getTxPool(msg.ByCount)
		if sender != nil {
			sender.Request(&tc.GetTxnPoolRsp{TxnPool: res}, context.Self())
		}

	case *tc.GetPendingTxnReq:
		sender := context.Sender()

		log.Info("txpool actor Receives getting pedning tx req from ", sender)

		res := tpa.server.getPendingTxs(msg.ByCount)
		if sender != nil {
			sender.Request(&tc.GetPendingTxnRsp{Txs: res}, context.Self())
		}

	case *tc.VerifyBlockReq:
		sender := context.Sender()

		log.Info("txpool actor Receives verifying block req from ", sender)

		tpa.server.verifyBlock(msg, sender)

	case *message.SaveBlockCompleteMsg:
		sender := context.Sender()

		log.Info("txpool actor Receives block complete event from ", sender)

		if msg.Block != nil {
			tpa.server.cleanTransactionList(msg.Block.Transactions)
		}

	default:
		log.Info("txpool actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (tpa *TxPoolActor) setServer(s *TXPoolServer) {
	tpa.server = s
}

// VerifyRspActor: Handle the response from the validators
type VerifyRspActor struct {
	server *TXPoolServer
}

func (vpa *VerifyRspActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("Server started and be ready to receive validator's msg")

	case *actor.Stopping:
		log.Info("Server stopping")

	case *actor.Restarting:
		log.Info("Server Restarting")

	case *types.RegisterValidator:
		log.Infof("validator %v connected", msg.Sender)
		vpa.server.registerValidator(msg)

	case *types.UnRegisterValidator:
		log.Infof("validator %d:%v disconnected", msg.Type, msg.Id)

		vpa.server.unRegisterValidator(msg.Type, msg.Id)

	case *types.CheckResponse:
		log.Info("Server Receives verify rsp message")

		vpa.server.assignRsp2Worker(msg)

	default:
		log.Info("txpool-verify actor:Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (vpa *VerifyRspActor) setServer(s *TXPoolServer) {
	vpa.server = s
}
