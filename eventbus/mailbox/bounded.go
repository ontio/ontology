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

package mailbox

import (
	"github.com/Ontology/eventbus/internal/queue/mpsc"
	rbqueue "github.com/Workiva/go-datastructures/queue"
)

type boundedMailboxQueue struct {
	userMailbox *rbqueue.RingBuffer
	dropping    bool
}

func (q *boundedMailboxQueue) Push(m interface{}) {
	if q.dropping {
		if q.userMailbox.Len() > 0 && q.userMailbox.Cap()-1 == q.userMailbox.Len() {
			q.userMailbox.Get()
		}
	}
	q.userMailbox.Put(m)
}

func (q *boundedMailboxQueue) Pop() interface{} {
	if q.userMailbox.Len() > 0 {
		m, _ := q.userMailbox.Get()
		return m
	}
	return nil
}

// Bounded returns a producer which creates an bounded mailbox of the specified size
func Bounded(size int, mailboxStats ...Statistics) Producer {
	return bounded(size, false, mailboxStats...)
}

// Bounded dropping returns a producer which creates an bounded mailbox of the specified size that drops front element on push
func BoundedDropping(size int, mailboxStats ...Statistics) Producer {
	return bounded(size, true, mailboxStats...)
}

func bounded(size int, dropping bool, mailboxStats ...Statistics) Producer {
	return func(invoker MessageInvoker, dispatcher Dispatcher) Inbound {
		q := &boundedMailboxQueue{
			userMailbox: rbqueue.NewRingBuffer(uint64(size)),
			dropping:    dropping,
		}
		return &defaultMailbox{
			systemMailbox: mpsc.New(),
			userMailbox:   q,
			invoker:       invoker,
			mailboxStats:  mailboxStats,
			dispatcher:    dispatcher,
		}
	}
}
