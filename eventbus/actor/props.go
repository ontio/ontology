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

import "github.com/Ontology/eventbus/mailbox"

type InboundMiddleware func(next ActorFunc) ActorFunc
type OutboundMiddleware func(next SenderFunc) SenderFunc

// Props represents configuration to define how an actor should be created
type Props struct {
	actorProducer       Producer
	mailboxProducer     mailbox.Producer
	guardianStrategy    SupervisorStrategy
	supervisionStrategy SupervisorStrategy
	inboundMiddleware   []InboundMiddleware
	outboundMiddleware  []OutboundMiddleware
	dispatcher          mailbox.Dispatcher
	spawner             SpawnFunc
}

func (props *Props) getDispatcher() mailbox.Dispatcher {
	if props.dispatcher == nil {
		return defaultDispatcher
	}
	return props.dispatcher
}

func (props *Props) getSupervisor() SupervisorStrategy {
	if props.supervisionStrategy == nil {
		return defaultSupervisionStrategy
	}
	return props.supervisionStrategy
}

func (props *Props) produceMailbox(invoker mailbox.MessageInvoker, dispatcher mailbox.Dispatcher) mailbox.Inbound {
	if props.mailboxProducer == nil {
		return defaultMailboxProducer(invoker, dispatcher)
	}
	return props.mailboxProducer(invoker, dispatcher)
}

func (props *Props) spawn(id string, parent *PID) (*PID, error) {
	if props.spawner != nil {
		return props.spawner(id, props, parent)
	}
	return DefaultSpawner(id, props, parent)
}

// Assign one or more middlewares to the props
func (props *Props) WithMiddleware(middleware ...InboundMiddleware) *Props {
	props.inboundMiddleware = append(props.inboundMiddleware, middleware...)
	return props
}

func (props *Props) WithOutboundMiddleware(middleware ...OutboundMiddleware) *Props {
	props.outboundMiddleware = append(props.outboundMiddleware, middleware...)
	return props
}

//WithMailbox assigns the desired mailbox producer to the props
func (props *Props) WithMailbox(mailbox mailbox.Producer) *Props {
	props.mailboxProducer = mailbox
	return props
}

//WithGuardian assigns a guardian strategy to the props
func (props *Props) WithGuardian(guardian SupervisorStrategy) *Props {
	props.guardianStrategy = guardian
	return props
}

//WithSupervisor assigns a supervision strategy to the props
func (props *Props) WithSupervisor(supervisor SupervisorStrategy) *Props {
	props.supervisionStrategy = supervisor
	return props
}

//WithDispatcher assigns a dispatcher to the props
func (props *Props) WithDispatcher(dispatcher mailbox.Dispatcher) *Props {
	props.dispatcher = dispatcher
	return props
}

//WithSpawnFunc assigns a custom spawn func to the props, this is mainly for internal usage
func (props *Props) WithSpawnFunc(spawn SpawnFunc) *Props {
	props.spawner = spawn
	return props
}

//WithFunc assigns a receive func to the props
func (props *Props) WithFunc(f ActorFunc) *Props {
	props.actorProducer = func() Actor { return f }
	return props
}

//WithProducer assigns a actor producer to the props
func (props *Props) WithProducer(p Producer) *Props {
	props.actorProducer = p
	return props
}

//Deprecated: WithInstance is deprecated.
func (props *Props) WithInstance(a Actor) *Props {
	props.actorProducer = makeProducerFromInstance(a)
	return props
}
