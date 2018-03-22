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

//Directive is an enum for supervision actions
type Directive int

// Directive determines how a supervisor should handle a faulting actor
const (
	// ResumeDirective instructs the supervisor to resume the actor and continue processing messages
	ResumeDirective Directive = iota

	// RestartDirective instructs the supervisor to discard the actor, replacing it with a new instance,
	// before processing additional messages
	RestartDirective

	// StopDirective instructs the supervisor to stop the actor
	StopDirective

	// EscalateDirective instructs the supervisor to escalate handling of the failure to the actor's parent supervisor
	EscalateDirective
)
