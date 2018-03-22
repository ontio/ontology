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
	"sync"

	"github.com/Ontology/common/log"
	"fmt"
)

type guardiansValue struct {
	sync.RWMutex
	guardians map[SupervisorStrategy]*guardianProcess
}

var guardians = &guardiansValue{guardians: make(map[SupervisorStrategy]*guardianProcess)}

func (gs *guardiansValue) getGuardianPid(s SupervisorStrategy) *PID {
	gs.Lock()
	defer gs.Unlock()
	if g, ok := gs.guardians[s]; ok {
		return g.pid
	}
	g := gs.newGuardian(s)
	gs.guardians[s] = g
	//gs.guardians.Store(s, g)
	return g.pid
}

// newGuardian creates and returns a new actor.guardianProcess with a timeout of duration d
func (gs *guardiansValue) newGuardian(s SupervisorStrategy) *guardianProcess {
	ref := &guardianProcess{strategy: s}
	id := ProcessRegistry.NextId()

	pid, ok := ProcessRegistry.Add(ref, "guardian"+id)
	if !ok {
		log.Error("failed to register guardian process", fmt.Sprintf("pid:%v", pid))
	}

	ref.pid = pid
	return ref
}

type guardianProcess struct {
	pid      *PID
	strategy SupervisorStrategy
}

func (g *guardianProcess) SendUserMessage(pid *PID, message interface{}) {
	panic(errors.New("Guardian actor cannot receive any user messages"))
}

func (g *guardianProcess) SendSystemMessage(pid *PID, message interface{}) {
	if msg, ok := message.(*Failure); ok {
		g.strategy.HandleFailure(g, msg.Who, msg.RestartStats, msg.Reason, msg.Message)
	}
}

func (g *guardianProcess) Stop(pid *PID) {
	//Ignore
}

func (g *guardianProcess) Children() []*PID {
	panic(errors.New("Guardian does not hold its children PIDs"))
}

func (*guardianProcess) EscalateFailure(reason interface{}, message interface{}) {
	panic(errors.New("Guardian cannot escalate failure"))
}

func (*guardianProcess) RestartChildren(pids ...*PID) {
	for _, pid := range pids {
		pid.sendSystemMessage(restartMessage)
	}
}

func (*guardianProcess) StopChildren(pids ...*PID) {
	for _, pid := range pids {
		pid.sendSystemMessage(stopMessage)
	}
}

func (*guardianProcess) ResumeChildren(pids ...*PID) {
	for _, pid := range pids {
		pid.sendSystemMessage(resumeMailboxMessage)
	}
}
