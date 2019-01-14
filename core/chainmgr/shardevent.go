
package chainmgr

import (
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/events"
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

