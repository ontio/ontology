package main

import (
	"runtime"

	"fmt"

	"time"

	"github.com/Ontology/common/log"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/example/zmq/messages"
	"github.com/Ontology/eventbus/mailbox"
	"github.com/Ontology/eventbus/zmqremote"
)

func main() {
	log.Init()
	log.Debug("test")
	runtime.GOMAXPROCS(runtime.NumCPU() * 1)
	runtime.GC()

	zmqremote.Start("127.0.0.1:8080")

	var sender *actor.PID
	props := actor.
		FromFunc(
			func(context actor.Context) {
				switch msg := context.Message().(type) {
				case *messages.StartRemote:
					//fmt.Println("Done server!")
					fmt.Println("Starting")
					sender = msg.Sender
					context.Respond(&messages.Start{})
				case *messages.Ping:
					//fmt.Println("ping")
					sender.Tell(&messages.Pong{})
				}
			}).
		WithMailbox(mailbox.Bounded(1000000))

	pid, _ := actor.SpawnNamed(props, "remote")
	fmt.Println(pid)

	for {
		time.Sleep(1 * time.Second)
	}
}
