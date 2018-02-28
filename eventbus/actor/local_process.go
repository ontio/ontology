package actor

import (
	"sync/atomic"

	"github.com/Ontology/eventbus/mailbox"
)

type localProcess struct {
	mailbox mailbox.Inbound
	dead    int32
}

func (ref *localProcess) SendUserMessage(pid *PID, message interface{}) {
	ref.mailbox.PostUserMessage(message)
}
func (ref *localProcess) SendSystemMessage(pid *PID, message interface{}) {
	ref.mailbox.PostSystemMessage(message)
}

func (ref *localProcess) Stop(pid *PID) {
	atomic.StoreInt32(&ref.dead, 1)
	ref.SendSystemMessage(pid, stopMessage)
}
