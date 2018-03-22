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
	"github.com/Ontology/eventbus/eventstream"
	"github.com/Ontology/common/log"
	"fmt"
)

//SupervisorEvent is sent on the EventStream when a supervisor have applied a directive to a failing child actor
type SupervisorEvent struct {
	Child     *PID
	Reason    interface{}
	Directive Directive
}

var (
	supervisionSubscriber *eventstream.Subscription
)

func init() {
	supervisionSubscriber = eventstream.Subscribe(func(evt interface{}) {
		if supervisorEvent, ok := evt.(*SupervisorEvent); ok {
			log.Debug("[SUPERVISION]", fmt.Sprintf("actor:%v", supervisorEvent.Child), fmt.Sprintf("directive:%v", supervisorEvent.Directive), fmt.Sprintf("reason:%v", supervisorEvent.Reason))
		}
	})
}
