package proc

import (
	"fmt"
	"github.com/Ontology/common/log"
	tx "github.com/Ontology/core/types"
	"github.com/Ontology/eventbus/actor"
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

func (ta *TxActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("Server started and be ready to receive tx msg")
	case *actor.Stopping:
		log.Info("Server stopping")
	case *actor.Restarting:
		log.Info("Server Restarting")
	case *tx.Transaction:
		log.Info("Server Receives txn message")
		ta.server.increaseStats(tc.RcvStats)
		if txn := ta.server.getTransaction(msg.Hash()); txn != nil {
			log.Info(fmt.Sprintf("Transaction %x already in the txn pool",
				msg.Hash()))
			ta.server.increaseStats(tc.DuplicateStats)
		} else {
			ta.server.assginTXN2Worker(msg)
		}
	case *tc.GetTxnReq:
		res := ta.server.getTransaction(msg.Hash)
		context.Sender().Request(&tc.GetTxnRsp{Txn: res}, context.Self())
	case *tc.GetTxnStats:
		res := ta.server.getStats()
		context.Sender().Request(&tc.GetTxnStatsRsp{Count: res}, context.Self())
	case *tc.CheckTxnReq:
		res := ta.server.CheckTx(msg.Hash)
		context.Sender().Request(&tc.CheckTxnRsp{Ok: res}, context.Self())
	case *tc.GetTxnStatusReq:
		res := ta.server.GetTxStatusReq(msg.Hash)
		context.Sender().Request(res, context.Self())
	default:
		log.Info("Unknown msg type", msg)
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
		log.Info("Server started and be ready to receive txPool msg")
	case *actor.Stopping:
		log.Info("Server stopping")
	case *actor.Restarting:
		log.Info("Server Restarting")
	case *tc.GetTxnPoolReq:
		res := tpa.server.GetTxPool(msg.ByCount)
		context.Sender().Request(&tc.GetTxnPoolRsp{TxnPool: res}, context.Self())
	case *tc.GetPendingTxnReq:
		res := tpa.server.GetPendingTxs(msg.ByCount)
		context.Sender().Request(&tc.GetPendingTxnRsp{Txs: res}, context.Self())
	default:
		log.Info("Unknown msg type", msg)
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
		fmt.Println("llls")
		v := Validator{
			Pid:       msg.Sender,
			CheckType: msg.Type,
		}
		vpa.server.registerValidator(msg.Id, v)
	case *types.UnRegisterValidator:
		log.Infof("validator %v disconnected", msg.Id)
		pid := vpa.server.GetValidatorPID(msg.Id)
		if pid != nil {
			vpa.server.unRegisterValidator(msg.Id)
			pid.Tell(&types.UnRegisterAck{Id: msg.Id})
		}

	case *types.CheckResponse:
		log.Info("Server Receives verify rsp message")
		vpa.server.assignRsp2Worker(msg)
	default:
		log.Info("Unknown msg type", msg)
	}
}

func (vpa *VerifyRspActor) setServer(s *TXPoolServer) {
	vpa.server = s
}
