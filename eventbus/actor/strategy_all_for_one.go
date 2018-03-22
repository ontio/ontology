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

import "time"

// NewAllForOneStrategy returns a new SupervisorStrategy which applies the given fault Directive from the decider to the
// failing child and all its children.
//
// This strategy is appropriate when the children have a strong dependency, such that and any single one failing would
// place them all into a potentially invalid state.
func NewAllForOneStrategy(maxNrOfRetries int, withinDuration time.Duration, decider DeciderFunc) SupervisorStrategy {
	return &allForOneStrategy{
		maxNrOfRetries: maxNrOfRetries,
		withinDuration: withinDuration,
		decider:        decider,
	}
}

type allForOneStrategy struct {
	maxNrOfRetries int
	withinDuration time.Duration
	decider        DeciderFunc
}

func (strategy *allForOneStrategy) HandleFailure(supervisor Supervisor, child *PID, rs *RestartStatistics, reason interface{}, message interface{}) {
	directive := strategy.decider(reason)
	switch directive {
	case ResumeDirective:
		//resume the failing child
		logFailure(child, reason, directive)
		supervisor.ResumeChildren(child)
	case RestartDirective:
		children := supervisor.Children()
		//try restart the all the children
		if strategy.shouldStop(rs) {
			logFailure(child, reason, StopDirective)
			supervisor.StopChildren(children...)
		} else {
			logFailure(child, reason, RestartDirective)
			supervisor.RestartChildren(children...)
		}
	case StopDirective:
		children := supervisor.Children()
		//stop all the children, no need to involve the crs
		logFailure(child, reason, directive)
		supervisor.StopChildren(children...)
	case EscalateDirective:
		//send failure to parent
		//supervisor mailbox
		//do not log here, log in the parent handling the error
		supervisor.EscalateFailure(reason, message)
	}
}

func (strategy *allForOneStrategy) shouldStop(rs *RestartStatistics) bool {

	// supervisor says this child may not restart
	if strategy.maxNrOfRetries == 0 {
		return true
	}

	rs.Fail()

	if rs.NumberOfFailures(strategy.withinDuration) > strategy.maxNrOfRetries {
		rs.Reset()
		return true
	}

	return false
}
