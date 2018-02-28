package main

import (
	"fmt"
	"github.com/Ontology/common/log"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/eventhub"
	tc "github.com/Ontology/txnpool/common"
	tp "github.com/Ontology/txnpool/proc"
	"github.com/Ontology/validator/db"
	"github.com/Ontology/validator/stateful"
	"github.com/Ontology/validator/stateless"
	"github.com/Ontology/eventbus/remote"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/payload"
    "time"
    "github.com/Ontology/common"
    "bytes"
	"encoding/hex"
)

var (
	txn   *types.Transaction
)


func init() {
	log.Init(log.Path, log.Stdout)

	bookKeepingPayload := &payload.BookKeeping{
		Nonce: uint64(time.Now().UnixNano()),
	}

	txn = &types.Transaction{
		Version:    0,
		Attributes: []*types.TxAttribute{},
		TxType:     types.BookKeeping,
		Payload:    bookKeepingPayload,
	}

	tempStr := "3369930accc1ddd067245e8edadcd9bea207ba5e1753ac18a51df77a343bfe92"
	hex, _ := hex.DecodeString(tempStr)
	var hash common.Uint256
	hash.Deserialize(bytes.NewReader(hex))
	txn.SetHash(hash)
}

func main() {
	remote.Start("192.168.153.130:50010")

	var s *tp.TXNPoolServer
	var stopCh chan bool

	stopCh = make(chan bool)

	eh := eventhub.GlobalEventHub

	// Start validator routine
	ca := stateless.NewVerifier(tc.SignatureV)
	caProps := actor.FromProducer(func() actor.Actor {
		return ca
	})
	caPid, _ := actor.SpawnNamed(caProps, "SignatureV")
	ca.SetPID(caPid)
	eh.Subscribe(tc.TOPIC, caPid)

	store, err := db.NewStore("temp.db")
	if err != nil {
		return
	}

	dbVerifier := stateful.NewDBVerifier(tc.StatefulV, store)

	statefulProps := actor.FromProducer(func() actor.Actor {
		return dbVerifier
	})
	statefulPid, _ := actor.SpawnNamed(statefulProps, "StatefulV")
	dbVerifier.SetPID(statefulPid)
	eh.Subscribe(tc.TOPIC, statefulPid)

	// Start txnpool server to receive msgs from p2p, consensus and valdiators
	s = tp.NewTxnPoolServer(tc.MAXWORKERNUM)
	s.SetEventHub(eh)

	// Initialize an actor to handle the rsp msg from valdiators
	eventActor := tp.NewVerifyRspActor(s)
	eventProps := actor.FromProducer(func() actor.Actor {
		return eventActor
	})

	eventPid, _ := actor.SpawnNamed(eventProps, "RspEvent")
	if eventPid == nil {
		fmt.Println("Fail to start verify rsp actor")
		return
	}
	s.RegisterActor(tc.VerifyRspActor, eventPid)

	// Initialize an actor to handle the msg from consensus
	tpa := tp.NewTxnPoolActor(s)
	txnPoolProps := actor.FromProducer(func() actor.Actor {
		return tpa
	})

	txnPoolPid, _ := actor.SpawnNamed(txnPoolProps, "TxPool")
	if txnPoolPid == nil {
		fmt.Println("Fail to start txnpool actor")
		return
	}
	s.RegisterActor(tc.TxPoolActor, txnPoolPid)

	// Initialize an actor to handle the txn msg from p2p and api
	ta := tp.NewTxnActor(s)
	txnProps := actor.FromProducer(func() actor.Actor {
		return ta
	})
    
	txnPid, _ := actor.SpawnNamed(txnProps, "Txn")
	if txnPid == nil {
		fmt.Println("Fail to start txn actor")
		return
	}
	s.RegisterActor(tc.TxActor, txnPid)
	<-stopCh
}
