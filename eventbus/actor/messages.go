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

//An AutoReceiveMessage is a special kind of user message that will be handled in some way automatially by the actor
type AutoReceiveMessage interface {
	AutoReceiveMessage()
}

//NotInfluenceReceiveTimeout messages will not reset the ReceiveTimeout timer of an actor that receives the message
type NotInfluenceReceiveTimeout interface {
	NotInfluenceReceiveTimeout()
}

// A SystemMessage message is reserved for specific lifecycle messages used by the actor system
type SystemMessage interface {
	SystemMessage()
}

// A ReceiveTimeout message is sent to an actor after the Context.ReceiveTimeout duration has expired
type ReceiveTimeout struct{}

// A Restarting message is sent to an actor when the actor is being restarted by the system due to a failure
type Restarting struct{}

// A Stopping message is sent to an actor prior to the actor being stopped
type Stopping struct{}

// A Stopped message is sent to the actor once it has been stopped. A stopped actor will receive no further messages
type Stopped struct{}

// A Started message is sent to an actor once it has been started and ready to begin receiving messages.
type Started struct{}

// Restart is message sent by the actor system to control the lifecycle of an actor
type Restart struct{}

type Failure struct {
	Who          *PID
	Reason       interface{}
	RestartStats *RestartStatistics
	Message      interface{}
}

type continuation struct {
	message interface{}
	f       func()
}

func (*Restarting) AutoReceiveMessage() {}
func (*Stopping) AutoReceiveMessage()   {}
func (*Stopped) AutoReceiveMessage()    {}
func (*PoisonPill) AutoReceiveMessage() {}

func (*Started) SystemMessage()      {}
func (*Stop) SystemMessage()         {}
func (*Watch) SystemMessage()        {}
func (*Unwatch) SystemMessage()      {}
func (*Terminated) SystemMessage()   {}
func (*Failure) SystemMessage()      {}
func (*Restart) SystemMessage()      {}
func (*continuation) SystemMessage() {}

var (
	restartingMessage     interface{} = &Restarting{}
	stoppingMessage       interface{} = &Stopping{}
	stoppedMessage        interface{} = &Stopped{}
	poisonPillMessage     interface{} = &PoisonPill{}
	receiveTimeoutMessage interface{} = &ReceiveTimeout{}
)

var (
	restartMessage        interface{} = &Restart{}
	startedMessage        interface{} = &Started{}
	stopMessage           interface{} = &Stop{}
	resumeMailboxMessage  interface{} = &mailbox.ResumeMailbox{}
	suspendMailboxMessage interface{} = &mailbox.SuspendMailbox{}
)
