package proc

import (
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	tx "github.com/Ontology/core/types"
	"github.com/Ontology/eventbus/actor"
	tc "github.com/Ontology/txnpool/common"
)

func NewTxnActor(s *TXNPoolServer) *TxnActor {
	a := &TxnActor{}
	a.setServer(s)
	return a
}

func NewTxnPoolActor(s *TXNPoolServer) *TxnPoolActor {
	a := &TxnPoolActor{}
	a.setServer(s)
	return a
}

func NewVerifyRspActor(s *TXNPoolServer) *VerifyRspActor {
	a := &VerifyRspActor{}
	a.setServer(s)
	return a
}

// TxnActor: Handle the low priority msg from P2P and API
type GetTxnReq struct {
	Hash common.Uint256
}

type GetTxnRsp struct {
	Txn *tx.Transaction
}

type CheckTxnReq struct {
	Hash common.Uint256
}

type CheckTxnRsp struct {
	Ok bool
}

type GetTxnStatusReq struct {
	Hash common.Uint256
}

type GetTxnStatusRsp struct {
	TxnStatus *tc.TXNEntry
}

type GetTxnStats struct {
}

type GetTxnStatsRsp struct {
	count *[]uint64
}

type TxnActor struct {
	server *TXNPoolServer
}

func (ta *TxnActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("Server started and be ready to receive txn")
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
	case *GetTxnReq:
		res := ta.server.getTransaction(msg.Hash)
		context.Sender().Request(&GetTxnRsp{Txn: res}, context.Self())
	case *GetTxnStats:
		res := ta.server.getStats()
		context.Sender().Request(&GetTxnStatsRsp{count: res}, context.Self())
	case *CheckTxnReq:
		res := ta.server.CheckTxn(msg.Hash)
		context.Sender().Request(&CheckTxnRsp{Ok: res}, context.Self())
	case *GetTxnStatusReq:
		res := ta.server.GetTxnStatusReq(msg.Hash)
		context.Sender().Request(&GetTxnStatusRsp{TxnStatus: res}, context.Self())
	default:
		log.Info("Unknown msg type", msg)
	}
}

func (ta *TxnActor) setServer(s *TXNPoolServer) {
	ta.server = s
}

// TxnPoolActor: Handle the high priority request from Consensus
type GetTxnPoolReq struct {
	ByCount bool
}

type GetTxnPoolRsp struct {
	TxnPool []*tc.TXNEntry
}

type GetPendingTxnReq struct {
	ByCount bool
}

type GetPendingTxnRsp struct {
	Txs []*tx.Transaction
}

type GetUnverifiedTxsReq struct {
	Txs []*tx.Transaction
}

type GetUnverifiedTxsRsp struct {
	Txs []*tx.Transaction
}

type CleanTxnPoolReq struct {
	TxnPool []*tx.Transaction
}

type TxnPoolActor struct {
	server *TXNPoolServer
}

func (tpa *TxnPoolActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("Server started and be ready to receive txn")
	case *actor.Stopping:
		log.Info("Server stopping")
	case *actor.Restarting:
		log.Info("Server Restarting")
	case *GetTxnPoolReq:
		res := tpa.server.GetTxnPool(msg.ByCount)
		context.Sender().Request(&GetTxnPoolRsp{TxnPool: res}, context.Self())
	case *CleanTxnPoolReq:
		tpa.server.CleanTransactionList(msg.TxnPool)
	case *GetPendingTxnReq:
		res := tpa.server.GetPendingTxs(msg.ByCount)
		context.Sender().Request(&GetPendingTxnRsp{Txs: res}, context.Self())
	case *GetUnverifiedTxsReq:
		res := tpa.server.GetUnverifiedTxs(msg.Txs)
		context.Sender().Request(&GetUnverifiedTxsRsp{Txs: res}, context.Self())
	default:
		log.Info("Unknown msg type", msg)
	}
}

func (tpa *TxnPoolActor) setServer(s *TXNPoolServer) {
	tpa.server = s
}

// VerifyRspActor: Handle the response from the validators
type VerifyRspActor struct {
	server *TXNPoolServer
}

func (vpa *VerifyRspActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("Server started and be ready to receive txn")
	case *actor.Stopping:
		log.Info("Server stopping")
	case *actor.Restarting:
		log.Info("Server Restarting")
	case *tc.VerifyRsp:
		log.Info("Server Receives verify rsp message")
		vpa.server.assignRsp2Worker(msg)
	default:
		log.Info("Unknown msg type", msg)
	}
}

func (vpa *VerifyRspActor) setServer(s *TXNPoolServer) {
	vpa.server = s
}
