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
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

type PID struct {
	Address string
	Id      string
	p       *Process
}

/*
func (m *PID) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	str := fmt.Sprintf("{\"Address\":\"%v\", \"Id\":\"%v\"}", m.Address, m.Id)
	return []byte(str), nil
}*/

func (pid *PID) ref() Process {
	p := (*Process)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&pid.p))))
	if p != nil {
		if l, ok := (*p).(*localProcess); ok && atomic.LoadInt32(&l.dead) == 1 {
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&pid.p)), nil)
		} else {
			return *p
		}
	}

	ref, exists := ProcessRegistry.Get(pid)
	if exists {
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&pid.p)), unsafe.Pointer(&ref))
	}

	return ref
}

// Tell sends a messages asynchronously to the PID
func (pid *PID) Tell(message interface{}) {
	pid.ref().SendUserMessage(pid, message)
}

// Request sends a messages asynchronously to the PID. The actor may send a response back via respondTo, which is
// available to the receiving actor via Context.Sender
func (pid *PID) Request(message interface{}, respondTo *PID) {
	env := &MessageEnvelope{
		Message: message,
		Header:  nil,
		Sender:  respondTo,
	}
	pid.ref().SendUserMessage(pid, env)
}

// RequestFuture sends a message to a given PID and returns a Future
func (pid *PID) RequestFuture(message interface{}, timeout time.Duration) *Future {
	future := NewFuture(timeout)
	env := &MessageEnvelope{
		Message: message,
		Header:  nil,
		Sender:  future.PID(),
	}
	pid.ref().SendUserMessage(pid, env)
	return future
}

func (pid *PID) sendSystemMessage(message interface{}) {
	pid.ref().SendSystemMessage(pid, message)
}

func (pid *PID) StopFuture() *Future {
	future := NewFuture(10 * time.Second)

	pid.sendSystemMessage(&Watch{Watcher: future.pid})
	pid.Stop()

	return future
}

func (pid *PID) GracefulStop() {
	pid.StopFuture().Wait()
}

//Stop the given PID
func (pid *PID) Stop() {
	pid.ref().Stop(pid)
}

func pidFromKey(key string, p *PID) {
	i := strings.IndexByte(key, '#')
	if i == -1 {
		p.Address = ProcessRegistry.Address
		p.Id = key
	} else {
		p.Address = key[:i]
		p.Id = key[i+1:]
	}
}

func (pid *PID) key() string {
	if pid.Address == ProcessRegistry.Address {
		return pid.Id
	}
	return pid.Address + "#" + pid.Id
}

func (pid *PID) String() string {
	if pid == nil {
		return "nil"
	}
	return pid.Address + "/" + pid.Id
}

//NewPID returns a new instance of the PID struct
func NewPID(address, id string) *PID {
	return &PID{
		Address: address,
		Id:      id,
	}
}

//NewLocalPID returns a new instance of the PID struct with the address preset
func NewLocalPID(id string) *PID {
	return &PID{
		Address: ProcessRegistry.Address,
		Id:      id,
	}
}
