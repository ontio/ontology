package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/Ontology/eventbus/actor"
)

type ping struct{ val int }
type pingActor struct{}

var start, end int64

func (state *pingActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize actor here")
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down")
	case *actor.Restarting:
		fmt.Println("Restarting, actor is about restart")
	case *ping:
		val := msg.val
		if val < 10000000 {
			context.Sender().Request(&ping{val: val + 1}, context.Self())
		} else {
			end = time.Now().UnixNano()
			fmt.Printf("%s end %d\n", context.Self().Id, end)
		}
	}
}
func main() {
	fmt.Printf("test pingpang")
	runtime.GOMAXPROCS(runtime.NumCPU())
	props := actor.FromProducer(func() actor.Actor { return &pingActor{} })
	actora := actor.Spawn(props)
	actorb := actor.Spawn(props)
	start = time.Now().UnixNano()
	fmt.Printf("begin time %d\n", start)
	actora.Request(&ping{val: 1}, actorb)

	time.Sleep(10 * time.Second)
	fmt.Println((end - start) / 1000000)
	actora.Stop()
	actorb.Stop()
}
