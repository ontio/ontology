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
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/example/testRemoteCrypto/commons"
	"runtime"
	"github.com/Ontology/eventbus/remote"
	"github.com/Ontology/common/log"
	"github.com/Ontology/eventbus/eventhub"
	"fmt"
	"time"
	"sync"
)



func main()  {

	runtime.GOMAXPROCS(runtime.NumCPU() * 1)
	runtime.GC()

	var wg sync.WaitGroup
	log.Init()
	remote.Start("172.26.127.139:9080")
	props := actor.FromProducer(func() actor.Actor { return &commons.BusynessActor{Datas:make(map[string][]byte), WgStop: &wg} })

	bActor, _ := actor.SpawnNamed(props, "busi")

	signActor := actor.NewPID("172.26.127.133:9080", "sign")
	vfActor1 := actor.NewPID("172.26.127.133:9081", "verify1")
	vfActor2 := actor.NewPID("172.26.127.136:9081", "verify2")
	vfActor3 := actor.NewPID("172.26.127.138:9081", "verify3")

	eventhub.GlobalEventHub.Subscribe(commons.SetTOPIC, signActor)
	eventhub.GlobalEventHub.Subscribe(commons.SigTOPIC, signActor)
	eventhub.GlobalEventHub.Subscribe(commons.VerifyTOPIC,vfActor1)
	eventhub.GlobalEventHub.Subscribe(commons.VerifyTOPIC,vfActor2)
	eventhub.GlobalEventHub.Subscribe(commons.VerifyTOPIC,vfActor3)

	wg.Add(1)
	start := time.Now()

	bActor.Tell(&commons.RunMsg{})
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Elapsed %s\n", elapsed)
	x := int(float32(commons.Loop2) / (float32(elapsed) / float32(time.Second)))
	fmt.Printf("Msg per sec %v", x)

	for {
		time.Sleep(1*time.Second)
	}
}