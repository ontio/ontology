package main

import (
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/example/testCrypto/signtest"
	"time"
	"runtime"
)



func main()  {
	//runtime.GOMAXPROCS(runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU() * 1)
	runtime.GC()
	props := actor.FromProducer(func() actor.Actor { return &signtest.BusynessActor{Datas:make(map[string][]byte)} })

	bActor:=actor.Spawn(props)

	//var wg sync.WaitGroup
	//
	//wg.Add(1)
	//start := time.Now()
	bActor.Tell(&signtest.RunMsg{})
	//wg.Wait()
	//elapsed := time.Since(start)
	//fmt.Printf("Elapsed %s\n", elapsed)

	for{
		time.Sleep(1 * time.Second)
	}
}