package commons

import (
	"fmt"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/crypto"
	"bytes"
)



type VerifyActor struct{
	Count int
}

func (s *VerifyActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize actor here")
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down")
	case *actor.Restarting:
		fmt.Println("Restarting, actor is about restart")

	case *VerifyRequest:
		s.Count++
		//fmt.Println(context.Self().Id, "is verifying...")
		crypto.SetAlg("")
		buf := bytes.NewBuffer(msg.PublicKey)
		pubKey := new(crypto.PubKey)
		err := pubKey.DeSerialize(buf)
		if err != nil {
			fmt.Println("DeSerialize failed.", err)
		}
		err = crypto.Verify(*pubKey,msg.Data,msg.Signature)
		//fmt.Println(context.Self().Id, "done verifying...")
		if err != nil{
			fmt.Println("verify error :", err)
			response:=&VerifyResponse{Seq:msg.Seq,Result:false,ErrorMsg:err.Error()}
			context.Sender().Tell(response)
		}else{
			response:=&VerifyResponse{Seq:msg.Seq,Result:true,ErrorMsg:""}
			context.Sender().Tell(response)
		}
		if s.Count%1 == 0 {
			fmt.Println(s.Count)
		}

	default:
		fmt.Printf("---unknown message%v\n",msg)
	}
}