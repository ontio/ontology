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

import "github.com/Ontology/eventbus/eventstream"

type optionFn func()

// WithDeadLetterSubscriber option replaces the default DeadLetterEvent event subscriber with fn.
//
// fn will only receive *DeadLetterEvent messages
//
// Specifying nil will clear the existing.
func WithDeadLetterSubscriber(fn func(evt interface{})) optionFn {
	return func() {
		if deadLetterSubscriber != nil {
			eventstream.Unsubscribe(deadLetterSubscriber)
		}
		if fn != nil {
			deadLetterSubscriber = eventstream.Subscribe(fn).
				WithPredicate(func(m interface{}) bool {
					_, ok := m.(*DeadLetterEvent)
					return ok
				})
		}
	}
}

// WithSupervisorSubscriber option replaces the default SupervisorEvent event subscriber with fn.
//
// fn will only receive *SupervisorEvent messages
//
// Specifying nil will clear the existing.
func WithSupervisorSubscriber(fn func(evt interface{})) optionFn {
	return func() {
		if supervisionSubscriber != nil {
			eventstream.Unsubscribe(supervisionSubscriber)
		}
		if fn != nil {
			supervisionSubscriber = eventstream.Subscribe(fn).
				WithPredicate(func(m interface{}) bool {
					_, ok := m.(*SupervisorEvent)
					return ok
				})
		}
	}
}

// SetOptions is used to configure the actor system
func SetOptions(opts ...optionFn) {
	for _, opt := range opts {
		opt()
	}
}
