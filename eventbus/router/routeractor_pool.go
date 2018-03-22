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

package router

import (
	"sync"
	"time"

	"github.com/Ontology/eventbus/actor"
)

type poolRouterActor struct {
	props  *actor.Props
	config RouterConfig
	state  Interface
	wg     *sync.WaitGroup
}

func (a *poolRouterActor) Receive(context actor.Context) {
	switch m := context.Message().(type) {
	case *actor.Started:
		a.config.OnStarted(context, a.props, a.state)
		a.wg.Done()

	case *AddRoutee:
		r := a.state.GetRoutees()
		if r.Contains(m.PID) {
			return
		}
		context.Watch(m.PID)
		r.Add(m.PID)
		a.state.SetRoutees(r)

	case *RemoveRoutee:
		r := a.state.GetRoutees()
		if !r.Contains(m.PID) {
			return
		}

		context.Unwatch(m.PID)
		r.Remove(m.PID)
		a.state.SetRoutees(r)
		// sleep for 1ms before sending the poison pill
		// This is to give some time to the routee actor receive all
		// the messages. Specially due to the synchronization conditions in
		// consistent hash router, where a copy of hmc can be obtained before
		// the update and cause messages routed to a dead routee if there is no
		// delay. This is a best effort approach and 1ms seems to be acceptable
		// in terms of both delay it cause to the router actor and the time it
		// provides for the routee to receive messages before it dies.
		time.Sleep(time.Millisecond * 1)
		m.PID.Tell(&actor.PoisonPill{})

	case *BroadcastMessage:
		msg := m.Message
		sender := context.Sender()
		a.state.GetRoutees().ForEach(func(i int, pid actor.PID) {
			pid.Request(msg, sender)
		})

	case *GetRoutees:
		r := a.state.GetRoutees()
		routees := make([]*actor.PID, r.Len())
		r.ForEach(func(i int, pid actor.PID) {
			routees[i] = &pid
		})

		context.Respond(&Routees{routees})
	}
}
