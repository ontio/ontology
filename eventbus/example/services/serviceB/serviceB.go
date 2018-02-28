package serviceB

import (
	"fmt"

	"github.com/Ontology/eventbus/actor"
	. "github.com/Ontology/eventbus/example/services/messages"
)

type ServiceB struct {
}

func (this *ServiceB) Receive(context actor.Context) {
	switch msg := context.Message().(type) {

	case *ServiceBRequest:
		fmt.Println("Receive ServiceBRequest:", msg.Message)
		context.Sender().Request(&ServiceBResponse{"response from serviceB"}, context.Self())

	case *ServiceAResponse:
		fmt.Println("Receive ServiceAResonse:", msg.Message)

	default:
		//fmt.Println("unknown message")
	}
}
