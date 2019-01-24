/*
 * Copyright (C) 2019 The ontology Authors
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

package chainmgr

import (
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/events/message"
)

type ShardEventActor struct {
	ShardSystemEventHandler func(v interface{})
}

func (actor *ShardEventActor) Receive(c actor.Context) {
	switch msg := c.Message().(type) {
	case *message.ShardSystemEventMsg:
		actor.ShardSystemEventHandler(msg)
	default:
	}
}

func subscribeShardSystemEvent(handler func(v interface{})) {
	var props = actor.FromProducer(func() actor.Actor {
		return &ShardEventActor{ShardSystemEventHandler: handler}
	})

	pid := actor.Spawn(props)
	sub := events.NewActorSubscriber(pid)
	sub.Subscribe(message.TOPIC_SHARD_SYSTEM_EVENT)
}

func (self *ChainManager) handleShardSystemEvent(msg *message.ShardSystemEventMsg) {
	if msg == nil {
		return
	}
	switch msg.Event.EventType {
	default:
	}
}
