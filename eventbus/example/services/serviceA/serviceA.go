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
