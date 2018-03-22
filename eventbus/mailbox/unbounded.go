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
	"github.com/Ontology/eventbus/internal/queue/goring"
	"github.com/Ontology/eventbus/internal/queue/mpsc"
)

type unboundedMailboxQueue struct {
	userMailbox *goring.Queue
}

func (q *unboundedMailboxQueue) Push(m interface{}) {
	q.userMailbox.Push(m)
}

func (q *unboundedMailboxQueue) Pop() interface{} {
	m, o := q.userMailbox.Pop()
	if o {
		return m
	}
	return nil
}

// Unbounded returns a producer which creates an unbounded mailbox
func Unbounded(mailboxStats ...Statistics) Producer {
	return func(invoker MessageInvoker, dispatcher Dispatcher) Inbound {
		q := &unboundedMailboxQueue{
			userMailbox: goring.New(10),
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
