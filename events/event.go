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

package events

import (
	"errors"
	"sync"
)

type Event struct {
	m           sync.RWMutex
	subscribers map[EventType]map[Subscriber]EventFunc
}

func NewEvent() *Event {
	return &Event{
		subscribers: make(map[EventType]map[Subscriber]EventFunc),
	}
}

//  adds a new subscriber to Event.
func (e *Event) Subscribe(eventtype EventType, eventfunc EventFunc) Subscriber {
	e.m.Lock()
	defer e.m.Unlock()

	sub := make(chan interface{})
	_, ok := e.subscribers[eventtype]
	if !ok {
		e.subscribers[eventtype] = make(map[Subscriber]EventFunc)
	}
	e.subscribers[eventtype][sub] = eventfunc

	return sub
}

// UnSubscribe removes the specified subscriber
func (e *Event) UnSubscribe(eventtype EventType, subscriber Subscriber) (err error) {
	e.m.Lock()
	defer e.m.Unlock()

	subEvent, ok := e.subscribers[eventtype]
	if !ok {
		err = errors.New("No event type.")
		return
	}

	delete(subEvent, subscriber)
	close(subscriber)

	return
}

//Notify subscribers that Subscribe specified event
func (e *Event) Notify(eventtype EventType, value interface{}) (err error) {
	e.m.RLock()
	defer e.m.RUnlock()

	subs, ok := e.subscribers[eventtype]
	if !ok {
		err = errors.New("No event type.")
		return
	}

	for _, event := range subs {
		go e.NotifySubscriber(event, value)
	}
	return
}

func (e *Event) NotifySubscriber(eventfunc EventFunc, value interface{}) {
	if eventfunc == nil {
		return
	}

	//invode subscriber event func
	eventfunc(value)

}

//Notify all event subscribers
func (e *Event) NotifyAll() (errs []error) {
	e.m.RLock()
	defer e.m.RUnlock()

	for eventtype, _ := range e.subscribers {
		if err := e.Notify(eventtype, nil); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
