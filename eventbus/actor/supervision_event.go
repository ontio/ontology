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
