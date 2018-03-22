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

package signtest

import (
	"fmt"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/crypto"
)



type VerifyActor struct{

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
		//fmt.Println(context.Self().Id, "is verifying...")
		err := crypto.Verify(msg.PublicKey,msg.Data,msg.Signature)
		//fmt.Println(context.Self().Id, "done verifying...")
		if err != nil{
			response:=&VerifyResponse{Seq:msg.Seq,Result:false,ErrorMsg:err.Error()}
			context.Sender().Tell(response)
		}else{
			response:=&VerifyResponse{Seq:msg.Seq,Result:true,ErrorMsg:""}
			context.Sender().Tell(response)
		}

	default:
		fmt.Printf("---unknown message%v\n",msg)
	}
}