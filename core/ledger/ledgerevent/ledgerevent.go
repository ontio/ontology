package ledgerevent

import "github.com/Ontology/events"

var DefLedgerEvt = NewLedgerEvent()

type LedgerEvent struct {
	evt *events.Event
}

func NewLedgerEvent() *LedgerEvent {
	return &LedgerEvent{
		evt: events.NewEvent(),
	}
}

func (this *LedgerEvent) Notify(eventtype events.EventType, value interface{}) error {
	return this.evt.Notify(eventtype, value)
}

func (this *LedgerEvent) Subscribe(eventtype events.EventType, eventfunc events.EventFunc) events.Subscriber{
	return this.evt.Subscribe(eventtype, eventfunc)
}

func (this *LedgerEvent) UnSubscribe(eventtype events.EventType, subscriber events.Subscriber) {
	this.evt.UnSubscribe(eventtype, subscriber)
}
