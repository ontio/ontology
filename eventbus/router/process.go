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
	"sync/atomic"

	"github.com/Ontology/eventbus/actor"
)

// process serves as a proxy to the router implementation and forwards messages directly to the routee. This
// optimization avoids serializing router messages through an actor
type process struct {
	router   *actor.PID
	state    Interface
	mu       sync.Mutex
	watchers actor.PIDSet
	stopping int32
}

func (ref *process) SendUserMessage(pid *actor.PID, message interface{}) {
	_, msg, _ := actor.UnwrapEnvelope(message)
	if _, ok := msg.(ManagementMessage); !ok {
		ref.state.RouteMessage(message)
	} else {
		r, _ := actor.ProcessRegistry.Get(ref.router)
		// Always send the original message to the router actor,
		// since if the message is enveloped, the sender need to get a response.
		r.SendUserMessage(pid, message)
	}
}

func (ref *process) SendSystemMessage(pid *actor.PID, message interface{}) {
	switch msg := message.(type) {
	case *actor.Watch:
		if atomic.LoadInt32(&ref.stopping) == 1 {
			if r, ok := actor.ProcessRegistry.Get(msg.Watcher); ok {
				r.SendSystemMessage(msg.Watcher, &actor.Terminated{Who: pid})
			}
			return
		}
		ref.mu.Lock()
		ref.watchers.Add(msg.Watcher)
		ref.mu.Unlock()

	case *actor.Unwatch:
		ref.mu.Lock()
		ref.watchers.Remove(msg.Watcher)
		ref.mu.Unlock()

	case *actor.Stop:
		term := &actor.Terminated{Who: pid}
		ref.mu.Lock()
		ref.watchers.ForEach(func(_ int, other actor.PID) {
			if r, ok := actor.ProcessRegistry.Get(&other); ok {
				r.SendSystemMessage(&other, term)
			}
		})
		ref.mu.Unlock()

	default:
		r, _ := actor.ProcessRegistry.Get(ref.router)
		r.SendSystemMessage(pid, message)

	}
}

func (ref *process) Stop(pid *actor.PID) {
	if atomic.SwapInt32(&ref.stopping, 1) == 1 {
		return
	}

	ref.router.StopFuture().Wait()
	actor.ProcessRegistry.Remove(pid)
	ref.SendSystemMessage(pid, &actor.Stop{})
}
