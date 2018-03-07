package txnpool

import (
	"github.com/Ontology/common/log"
	"github.com/Ontology/eventbus/actor"
	tc "github.com/Ontology/txnpool/common"
	tp "github.com/Ontology/txnpool/proc"
)

func startActor(obj interface{}) *actor.PID {
	props := actor.FromProducer(func() actor.Actor {
		return obj.(actor.Actor)
	})

	pid := actor.Spawn(props)
	if pid == nil {
		log.Error("Fail to start actor")
		return nil
	}
	return pid
}

func StartTxnPoolServer() *tp.TXPoolServer {
	var s *tp.TXPoolServer

	/* Start txnpool server to receive msgs from p2p,
	 * consensus and valdiators
	 */
	s = tp.NewTxPoolServer(tc.MAXWORKERNUM)

	// Initialize an actor to handle the msgs from valdiators
	rspActor := tp.NewVerifyRspActor(s)
	rspPid := startActor(rspActor)
	if rspPid == nil {
		log.Error("Fail to start verify rsp actor")
		return nil
	}
	s.RegisterActor(tc.VerifyRspActor, rspPid)

	// Initialize an actor to handle the msgs from consensus
	tpa := tp.NewTxPoolActor(s)
	txPoolPid := startActor(tpa)
	if txPoolPid == nil {
		log.Error("Fail to start txnpool actor")
		return nil
	}
	s.RegisterActor(tc.TxPoolActor, txPoolPid)

	// Initialize an actor to handle the msgs from p2p and api
	ta := tp.NewTxActor(s)
	txPid := startActor(ta)
	if txPid == nil {
		log.Error("Fail to start txn actor")
		return nil
	}
	s.RegisterActor(tc.TxActor, txPid)
	return s
}
