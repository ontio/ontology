package main

import (
	"github.com/Ontology/eventbus/actor"
	"runtime"
	"github.com/Ontology/eventbus/example/testRemoteCrypto/commons"
	"github.com/Ontology/eventbus/remote"
	"github.com/Ontology/common/log"
	"time"
)

func main()  {

	runtime.GOMAXPROCS(runtime.NumCPU() * 1)
	runtime.GC()

	log.Init()
	remote.Start("172.26.127.133:9080")
	signprops := actor.FromProducer(func() actor.Actor { return &commons.SignActor{} })
	actor.SpawnNamed(signprops, "sign")


	for{
		time.Sleep(1 * time.Second)
	}
}