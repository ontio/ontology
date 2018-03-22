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

type deadLetterProcess struct{}

var (
	deadLetter           Process = &deadLetterProcess{}
	deadLetterSubscriber *eventstream.Subscription
)

func init() {
	deadLetterSubscriber = eventstream.Subscribe(func(msg interface{}) {
		if deadLetter, ok := msg.(*DeadLetterEvent); ok {
			log.Debug("[DeadLetter]:", fmt.Sprintf("%v",deadLetter))
		}
	})

	//this subscriber may not be deactivated.
	//it ensures that Watch commands that reach a stopped actor gets a Terminated message back.
	//This can happen if one actor tries to Watch a PID, while another thread sends a Stop message.
	eventstream.Subscribe(func(msg interface{}) {
		if deadLetter, ok := msg.(*DeadLetterEvent); ok {
			if m, ok := deadLetter.Message.(*Watch); ok {
				//we know that this is a local actor since we get it on our own event stream, thus the address is not terminated
				m.Watcher.sendSystemMessage(&Terminated{AddressTerminated: false, Who: deadLetter.PID})
			}
		}
	})
}

// A DeadLetterEvent is published via event.Publish when a message is sent to a nonexistent PID
type DeadLetterEvent struct {
	PID     *PID        // The invalid process, to which the message was sent
	Message interface{} // The message that could not be delivered
	Sender  *PID        // the process that sent the Message
}

func (*deadLetterProcess) SendUserMessage(pid *PID, message interface{}) {
	_, msg, sender := UnwrapEnvelope(message)
	eventstream.Publish(&DeadLetterEvent{
		PID:     pid,
		Message: msg,
		Sender:  sender,
	})
}

func (*deadLetterProcess) SendSystemMessage(pid *PID, message interface{}) {
	eventstream.Publish(&DeadLetterEvent{
		PID:     pid,
		Message: message,
	})
}

func (ref *deadLetterProcess) Stop(pid *PID) {
	ref.SendSystemMessage(pid, stopMessage)
}
