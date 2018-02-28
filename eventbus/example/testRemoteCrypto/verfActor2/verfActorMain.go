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
	remote.Start("172.26.127.136:9081")
	vfprops := actor.FromProducer(func() actor.Actor { return &commons.VerifyActor{} })
	actor.SpawnNamed(vfprops, "verify2")


	for{
		time.Sleep(1 * time.Second)
	}
}