package main

import (
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/remote"

	"sync"

	"runtime"
	"time"

	"bytes"
	"fmt"

	"github.com/Ontology/common/log"
	"github.com/Ontology/crypto"
	"github.com/Ontology/eventbus/example/ontCrypto/remotePerformance/messages"
)

type localActor struct {
	count        int
	wgStop       *sync.WaitGroup
	messageCount int
}

func (state *localActor) Receive(context actor.Context) {
	switch context.Message().(type) {
	case *messages.Pong:
		state.count++
		//fmt.Println("Pong")
		if state.count%50000 == 0 {
			fmt.Println(state.count)
		}
		if state.count == state.messageCount {
			state.wgStop.Done()
		}
		//case *messages.Pong:
		//	if msg.IfOK == "ok" {
		//		state.wgStop.Done()
		//	} else {
		//		state.wgStop.Done()
		//	}
		//}
	}
}

func newLocalActor(stop *sync.WaitGroup, messageCount int) actor.Producer {
	return func() actor.Actor {
		return &localActor{
			wgStop:       stop,
			messageCount: messageCount,
		}
	}
}

func main() {
	log.Init()
	runtime.GOMAXPROCS(runtime.NumCPU() * 1)
	runtime.GC()

	var wg sync.WaitGroup
	crypto.SetAlg("")

	messageCount := 500

	//remote.DefaultSerializerID = 1
	remote.Start("127.0.0.1:8081")

	props := actor.
		FromProducer(newLocalActor(&wg, messageCount))

	pid := actor.Spawn(props)

	remotePid := actor.NewPID("127.0.0.1:8080", "remote")
	sk, _ := remotePid.
		RequestFuture(&messages.StartRemote{
			Sender: pid,
		}, 5*time.Second).
		Result()
	fmt.Println(sk)
	sk1 := sk.(*messages.Start).PriKey
	wg.Add(1)

	start := time.Now()
	fmt.Println("Starting to send")

	bb := bytes.NewBuffer([]byte("s"))

	for i := 0; i < 200000; i++ {
		bb.WriteString("1234567890")
	}
	data := bb.Bytes()

	signature, err := crypto.Sign(sk1, data)
	fmt.Println(len(signature))
	fmt.Println(len(data))
	if err != nil {
		fmt.Println(err)
	}
	message := &messages.Ping{Signature: signature, Data: data}
	for i := 0; i < messageCount; i++ {
		remotePid.Tell(message)
		//time.Sleep(5000 * time.Millisecond)
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Elapsed %s", elapsed)

	x := int(float32(messageCount*2) / (float32(elapsed) / float32(time.Second)))
	fmt.Printf("Msg per sec %v", x)
}
