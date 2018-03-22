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

package zmqremote

import (
	"github.com/Ontology/eventbus/actor"
)

type process struct {
	pid *actor.PID
}

func newProcess(pid *actor.PID) actor.Process {
	return &process{
		pid: pid,
	}
}

func (ref *process) SendUserMessage(pid *actor.PID, message interface{}) {
	header, msg, sender := actor.UnwrapEnvelope(message)
	SendMessage(pid, header, msg, sender, -1)
}

func SendMessage(pid *actor.PID, header actor.ReadonlyMessageHeader, message interface{}, sender *actor.PID, serializerID int32) {
	rd := &remoteDeliver{
		header:       header,
		message:      message,
		sender:       sender,
		target:       pid,
		serializerID: serializerID,
	}

	endpointManager.remoteDeliver(rd)
}

func (ref *process) SendSystemMessage(pid *actor.PID, message interface{}) {

	//intercept any Watch messages and direct them to the endpoint manager
	switch msg := message.(type) {
	case *actor.Watch:
		rw := &remoteWatch{
			Watcher: msg.Watcher,
			Watchee: pid,
		}
		endpointManager.remoteWatch(rw)
	case *actor.Unwatch:
		ruw := &remoteUnwatch{
			Watcher: msg.Watcher,
			Watchee: pid,
		}
		endpointManager.remoteUnwatch(ruw)
	default:
		SendMessage(pid, nil, message, nil, -1)
	}
}

func (ref *process) Stop(pid *actor.PID) {
	ref.SendSystemMessage(pid, stopMessage)
}
