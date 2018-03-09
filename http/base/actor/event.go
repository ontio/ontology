package actor

import (
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/events/message"
	"github.com/Ontology/events"
)

var defLedgerEvtPid *actor.PID
var smartcodeEvtPid *actor.PID

type EventActor struct{
	blockPersistCompleted func(v interface{})
	smartCodeEvt func(v interface{})
}

func (t *EventActor) Receive(c actor.Context) {
	switch msg := c.Message().(type) {
	case *message.SaveBlockCompleteMsg:
		t.blockPersistCompleted(msg.Block)
	case *message.SmartCodeEventMsg:
		t.smartCodeEvt(msg.Event)
	default:
		//fmt.Println(msg)
	}
}

//ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventBlockPersistCompleted, SendBlock2WSclient)
//ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventSmartCode, PushSmartCodeEvent)
func SubscribeEvent(topic string,handler func(v interface{})) {
	var props = actor.FromProducer(func() actor.Actor {
		if topic == message.TopicSaveBlockComplete{
			return &EventActor{blockPersistCompleted:handler}
		}else if topic == message.TopicSmartCodeEvent{
			return &EventActor{smartCodeEvt:handler}
		}else{
			return &EventActor{}
		}
	})
	var pid = actor.Spawn(props)
	var sub = events.NewActorSubscriber(pid)
	sub.Subscribe(topic)
}
