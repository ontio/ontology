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

package eventhub

import (
	"math/rand"

	"github.com/Ontology/common/log"
	"github.com/Ontology/eventbus/actor"
	"github.com/orcaman/concurrent-map"
)

type PublishPolicy int

type RoundRobinState struct {
	state map[string]int
}

const (
	PUBLISH_POLICY_ALL = iota
	PUBLISH_POLICY_ROUNDROBIN
	PUBLISH_POLICY_RANDOM
)

type EventHub struct {
	//sync.RWMutex
	Subscribers cmap.ConcurrentMap
	RoundRobinState
}

type Event struct {
	Publisher *actor.PID
	Topic     string
	Message   interface{}
	Policy    PublishPolicy
}

var GlobalEventHub = &EventHub{Subscribers: cmap.New(), RoundRobinState: RoundRobinState{make(map[string]int)}}

func (this *EventHub) Publish(event *Event) {
	//go func() {
	actors, ok := this.Subscribers.Get(event.Topic)
	if !ok {
		log.Info("no subscribers yet!")
		return
	}
	subscribers := actors.([]*actor.PID)
	this.sendEventByPolicy(subscribers, event, this.RoundRobinState)
	//}()
}

func (this *EventHub) Subscribe(topic string, subscriber *actor.PID) {
	subscribers, _ := this.Subscribers.Get(topic)

	//defer this.RWMutex.Unlock()
	//this.RWMutex.Lock()
	if subscribers == nil {
		this.Subscribers.Set(topic, []*actor.PID{subscriber})
	} else {
		this.Subscribers.Set(topic, append(subscribers.([]*actor.PID), subscriber))
	}

}

func (this *EventHub) Unsubscribe(topic string, subscriber *actor.PID) {

	tmpslice, ok := this.Subscribers.Get(topic)
	if !ok {
		log.Debug("No subscriber on topic:%s yet.\n", topic)
		return
	}
	//defer this.RWMutex.Unlock()
	//this.RWMutex.Lock()
	subscribers := tmpslice.([]*actor.PID)
	for i, s := range subscribers {
		if s == subscriber {
			this.Subscribers.Set(topic, append(subscribers[0:i], subscribers[i+1:]...))
			return
		}
	}

}

func (this *EventHub) sendEventByPolicy(subscribers []*actor.PID, event *Event, state RoundRobinState) {
	switch event.Policy {
	case PUBLISH_POLICY_ALL:
		for _, subscriber := range subscribers {
			subscriber.Request(event.Message, event.Publisher)
		}
	case PUBLISH_POLICY_RANDOM:
		length := len(subscribers)
		if length == 0 {
			log.Info("no subscribers yet!")
			return
		}
		var i int
		i = rand.Intn(length)
		subscribers[i].Request(event.Message, event.Publisher)
	case PUBLISH_POLICY_ROUNDROBIN:
		latestIdx := state.state[event.Topic]
		i := latestIdx + 1
		if i < 0 {
			latestIdx = 0
			i = 0
		}
		state.state[event.Topic] = i
		mod := len(subscribers)
		subscribers[i%mod].Request(event.Message, event.Publisher)
	}
}

func (this *EventHub) RemovePID(pid actor.PID) {
	if this.Subscribers.Count() == 0 {
		return
	}
	keys := this.Subscribers.Keys()
	for index, _ := range keys {
		this.Unsubscribe(keys[index], &pid)
	}
}
