package txnpool

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/eventhub"
	tc "github.com/Ontology/txnpool/common"
	tp "github.com/Ontology/txnpool/proc"
	"github.com/Ontology/validator/db"
	"github.com/Ontology/validator/stateful"
	"github.com/Ontology/validator/stateless"
	"sync"
	"testing"
	"time"
)

var (
	txn   *types.Transaction
	topic string
)

func init() {
	crypto.SetAlg("")
	log.Init(log.Path, log.Stdout)
	topic = "TXN"

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

func Benchmark_RCV(b *testing.B) {
	var s *tp.TXNPoolServer
	var wg sync.WaitGroup

	eh := eventhub.GlobalEventHub

	// Start validator routine
	ca := stateless.NewVerifier(tc.SignatureV)
	caProps := actor.FromProducer(func() actor.Actor {
		return ca
	})
	caPid := actor.Spawn(caProps)
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
	statefulPid := actor.Spawn(statefulProps)
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

	eventPid := actor.Spawn(eventProps)
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

	txnPoolPid := actor.Spawn(txnPoolProps)
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

	txnPid := actor.Spawn(txnProps)
	if txnPid == nil {
		fmt.Println("Fail to start txn actor")
		return
	}
	s.RegisterActor(tc.TxActor, txnPid)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			var j int
			defer wg.Done()
			for {
				j++
				txnPid.Tell(txn)

				if j >= 10000 {
					return
				}
			}
		}()
	}

	wg.Wait()
	time.Sleep(1 * time.Second)
	txnPoolPid.Tell(&tp.GetTxnPoolReq{ByCount: true})
	txnPoolPid.Tell(&tp.GetPendingTxnReq{ByCount: true})
	time.Sleep(2 * time.Second)

	s.Stop()
	eh.Unsubscribe(topic, caPid)
	caPid.Stop()
	eh.Unsubscribe(topic, statefulPid)
	statefulPid.Stop()
}
