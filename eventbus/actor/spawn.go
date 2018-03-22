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

import (
	"errors"
)

// ErrNameExists is the error used when an existing name is used for spawning an actor.
var ErrNameExists = errors.New("spawn: name exists")

type SpawnFunc func(id string, props *Props, parent *PID) (*PID, error)

// DefaultSpawner conforms to Spawner and is used to spawn a local actor
var DefaultSpawner SpawnFunc = spawn

// Spawn starts a new actor based on props and named with a unique id
func Spawn(props *Props) *PID {
	pid, _ := SpawnNamed(props, ProcessRegistry.NextId())
	return pid
}

// SpawnPrefix starts a new actor based on props and named using a prefix followed by a unique id
func SpawnPrefix(props *Props, prefix string) (*PID, error) {
	return SpawnNamed(props, prefix+ProcessRegistry.NextId())
}

// SpawnNamed starts a new actor based on props and named using the specified name
//
// If name exists, error will be ErrNameExists
func SpawnNamed(props *Props, name string) (*PID, error) {
	var parent *PID
	if props.guardianStrategy != nil {
		parent = guardians.getGuardianPid(props.guardianStrategy)
	}
	return props.spawn(name, parent)
}

func spawn(id string, props *Props, parent *PID) (*PID, error) {
	lp := &localProcess{}
	pid, absent := ProcessRegistry.Add(lp, id)
	if !absent {
		return pid, ErrNameExists
	}

	cell := newLocalContext(props.actorProducer, props.getSupervisor(), props.inboundMiddleware, props.outboundMiddleware, parent)
	mb := props.produceMailbox(cell, props.getDispatcher())
	lp.mailbox = mb
	var ref Process = lp
	pid.p = &ref
	cell.self = pid
	mb.Start()
	mb.PostSystemMessage(startedMessage)

	return pid, nil
}
