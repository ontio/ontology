package main

import (
	"fmt"
	"runtime"

	"github.com/Ontology/common/log"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/crypto"
	"github.com/Ontology/eventbus/example/ontCrypto/remotePerformance/messages"
	"github.com/Ontology/eventbus/remote"
	"time"
)

func main() {
	log.Init()
	runtime.GOMAXPROCS(runtime.NumCPU() * 1)
	runtime.GC()

	remote.Start("127.0.0.1:8080")

	crypto.SetAlg("")
	var sender *actor.PID
	var pubKey crypto.PubKey
	props := actor.
		FromFunc(
			func(context actor.Context) {
				switch msg := context.Message().(type) {
				case *messages.StartRemote:
					fmt.Println("Starting")
					sender = msg.Sender
					fmt.Println("Starting")
					sk, pk, err := crypto.GenKeyPair()
					fmt.Println(sk)
					pubKey = pk
					if err != nil {
						fmt.Println(err)
					}
					context.Respond(&messages.Start{PriKey: sk})
				case *messages.Ping:
					err := crypto.Verify(pubKey, msg.Data, msg.Signature)
					if err == nil {
						sender.Tell(&messages.Pong{IfOK: "yes"})
					} else {
						sender.Tell(&messages.Pong{IfOK: "no"})
					}
				}
			})

	actor.SpawnNamed(props, "remote")

	for {
		time.Sleep(1*time.Second)
	}
}
