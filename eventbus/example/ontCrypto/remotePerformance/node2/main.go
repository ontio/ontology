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
