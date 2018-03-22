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

package commons

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
		fmt.Println(context.Self().Id," set Privkey")
		s.PrivateKey = msg.PrivKey

	case *SignRequest:
		crypto.SetAlg("")
		fmt.Println(context.Self().Id," is signing")
		signature,err:=crypto.Sign(s.PrivateKey, msg.Data)
		if err!= nil {
			fmt.Println("sign error: ", err)
		}
		response := &SignResponse{Signature:signature,Seq:msg.Seq}
		fmt.Println(context.Self().Id," done signing")
		context.Sender().Request(response,context.Self())

	default:
		fmt.Println("unknown message")
	}
}