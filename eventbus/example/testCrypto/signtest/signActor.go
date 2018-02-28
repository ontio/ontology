package signtest

import (
	"github.com/Ontology/eventbus/actor"
	"fmt"
	"github.com/Ontology/crypto"
)

type SignActor struct{
	PrivateKey []byte

}

func (s *SignActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize actor here")
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down")
	case *actor.Restarting:
		fmt.Println("Restarting, actor is about restart")

	case *SetPrivKey:
		//fmt.Println(context.Self().Id," set Privkey")
		s.PrivateKey = msg.PrivKey

	case *SignRequest:
		//fmt.Println(context.Self().Id," is signing")
		signature,_:=crypto.Sign(s.PrivateKey, msg.Data)
		response := &SignResponse{Signature:signature,Seq:msg.Seq}
		//fmt.Println(context.Self().Id," done signing")
		context.Sender().Request(response,context.Self())

	default:
		//fmt.Println("unknown message")
	}
}