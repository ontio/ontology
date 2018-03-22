/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"github.com/Ontology/common/log"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/example/zmq/messages"

	"sync"

	"runtime"
	"time"

	"bytes"
	"fmt"

	"github.com/Ontology/eventbus/mailbox"
	"github.com/Ontology/eventbus/zmqremote"
)

type localActor struct {
	count        int
	wgStop       *sync.WaitGroup
	messageCount int
}

func (state *localActor) Receive(context actor.Context) {
	switch context.Message().(type) {
	case *messages.Pong:
		//fmt.Println("pong")
		state.count++
		if state.count%50000 == 0 {
			fmt.Println(state.count)
		}
		if state.count == state.messageCount {
			state.wgStop.Done()
		}
		//case *messages.Start:
		//	fmt.Println("Done!")
		//	state.wgStop.Done()
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
	log.Debug("test")
	runtime.GOMAXPROCS(runtime.NumCPU() * 1)
	runtime.GC()

	var wg sync.WaitGroup

	messageCount := 50000

	zmqremote.Start("127.0.0.1:8081")

	props := actor.
		FromProducer(newLocalActor(&wg, messageCount)).
		WithMailbox(mailbox.Bounded(1000000))

	pid := actor.Spawn(props)
	fmt.Println(pid)

	remotePid := actor.NewPID("127.0.0.1:8080", "remote")
	remotePid.
		RequestFuture(&messages.StartRemote{
			Sender: pid,
		}, 5*time.Second).
		Wait()

	wg.Add(1)

	start := time.Now()
	fmt.Println("Starting to send")

	bb := bytes.NewBuffer([]byte(""))
	for i := 0; i < 2000; i++ {
		bb.WriteString("1234567890")
	}
	message := &messages.Ping{Data: bb.Bytes()}
	for i := 0; i < messageCount; i++ {
		remotePid.Tell(message)
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Elapsed %s", elapsed)

	x := int(float32(messageCount*2) / (float32(elapsed) / float32(time.Second)))
	fmt.Printf("Msg per sec %v", x)
}
