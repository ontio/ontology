package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Ontology/eventbus/actor"
	. "github.com/Ontology/eventbus/example/services/messages"
	"github.com/Ontology/eventbus/example/services/serviceA"
	"github.com/Ontology/eventbus/example/services/serviceB"
)

func main() {
	sva := actor.FromProducer(func() actor.Actor { return &serviceA.ServiceA{} })
	svb := actor.FromProducer(func() actor.Actor { return &serviceB.ServiceB{} })

	pipA, _ := actor.SpawnNamed(sva, "serviceA")
	pipB, _ := actor.SpawnNamed(svb, "serviceB")

	pipA.Request(&ServiceARequest{"TESTA"}, pipB)

	pipB.Request(&ServiceBRequest{"TESTB"}, pipA)
	time.Sleep(2 * time.Second)

	f := pipA.RequestFuture(1, 50*time.Microsecond)
	result, err := f.Result()
	if err != nil {
		fmt.Println("errors:", err.Error())
	}
	fmt.Println("get sync call result :" + strconv.Itoa(result.(int)))

}
