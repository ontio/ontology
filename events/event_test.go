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

package events

import (
	"fmt"
	"testing"
)

func TestNewEvent(t *testing.T) {
	event := NewEvent()

	var subscriber1 EventFunc = func(v interface{}) {
		fmt.Println("subscriber1 event func.")
	}

	var subscriber2 EventFunc = func(v interface{}) {
		fmt.Println("subscriber2 event func.")
	}

	fmt.Println("Subscribe...")
	sub1 := event.Subscribe(EventReplyTx, subscriber1)
	event.Subscribe(EventSaveBlock, subscriber2)

	fmt.Println("Notify...")
	event.Notify(EventReplyTx, nil)

	fmt.Println("Notify All...")
	event.NotifyAll()

	event.UnSubscribe(EventReplyTx, sub1)
	fmt.Println("Notify All after unsubscribe sub1...")
	event.NotifyAll()

}
