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

package node

import (
	"fmt"

	"github.com/Ontology/events"
)

type eventQueue struct {
	Consensus  *events.Event
	Block      *events.Event
	Disconnect *events.Event
}

func (eq *eventQueue) init() {
	eq.Consensus = events.NewEvent()
	eq.Block = events.NewEvent()
	eq.Disconnect = events.NewEvent()
}

func (eq *eventQueue) GetEvent(eventName string) *events.Event {
	switch eventName {
	case "consensus":
		return eq.Consensus
	case "block":
		return eq.Block
	case "disconnect":
		return eq.Disconnect
	default:
		fmt.Printf("Unknow event registe")
		return nil
	}
}
