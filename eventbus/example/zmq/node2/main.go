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
