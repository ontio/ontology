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
	"time"
)



func main()  {

	runtime.GOMAXPROCS(runtime.NumCPU() * 1)
	runtime.GC()

	log.Init()
	remote.Start("172.26.127.138:9081")
	vfprops := actor.FromProducer(func() actor.Actor { return &commons.VerifyActor{} })
	actor.SpawnNamed(vfprops, "verify3")

	for{
		time.Sleep(1 * time.Second)
	}
}