package serviceA

import (
	"fmt"

	"github.com/Ontology/eventbus/actor"
	. "github.com/Ontology/eventbus/example/services/messages"
)

type ServiceA struct {
}

func (this *ServiceA) Receive(context actor.Context) {
	switch msg := context.Message().(type) {

	case *ServiceARequest:
		fmt.Println("Receive ServiceARequest:", msg.Message)
		context.Sender().Tell(&ServiceAResponse{"I got your message"})

	case *ServiceBResponse:
		fmt.Println("Receive ServiceBResponse:", msg.Message)

	case int:
		context.Sender().Tell(msg + 1)

	default:
		fmt.Printf("unknown message:%v\n", msg)
	}
}
