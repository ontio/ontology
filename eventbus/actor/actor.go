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

package actor

// The Producer type is a function that creates a new actor
type Producer func() Actor

// Actor is the interface that defines the Receive method.
//
// Receive is sent messages to be processed from the mailbox associated with the instance of the actor
type Actor interface {
	Receive(c Context)
}

// The ActorFunc type is an adapter to allow the use of ordinary functions as actors to process messages
type ActorFunc func(c Context)

// Receive calls f(c)
func (f ActorFunc) Receive(c Context) {
	f(c)
}

type SenderFunc func(c Context, target *PID, envelope *MessageEnvelope)

//FromProducer creates a props with the given actor producer assigned
func FromProducer(actorProducer Producer) *Props {
	return &Props{actorProducer: actorProducer}
}

//FromFunc creates a props with the given receive func assigned as the actor producer
func FromFunc(f ActorFunc) *Props {
	return FromProducer(func() Actor { return f })
}

func FromSpawnFunc(spawn SpawnFunc) *Props {
	return &Props{spawner: spawn}
}

//Deprecated: FromInstance is deprecated
//Please use FromProducer(func() actor.Actor {...}) instead
func FromInstance(template Actor) *Props {
	return &Props{actorProducer: makeProducerFromInstance(template)}
}

//Deprecated: makeProducerFromInstance is deprecated.
func makeProducerFromInstance(a Actor) Producer {
	return func() Actor {
		return a
	}
}
