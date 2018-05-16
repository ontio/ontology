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

package vbft

import "testing"

func constructEventTimer() *EventTimer {
	server := constructServer()
	return NewEventTimer(server)
}

func TestStartTimer(t *testing.T) {
	eventtimer := constructEventTimer()
	err := eventtimer.StartTimer(1, 10)
	t.Logf("TestStartTimer: %v", err)
}

func TestCancelTimer(t *testing.T) {
	eventtimer := constructEventTimer()
	err := eventtimer.StartTimer(1, 10)
	t.Logf("TestStartTimer: %v", err)
	err = eventtimer.CancelTimer(1)
	t.Logf("TestCancelTimer: %v", err)
}

func TestStartEventTimer(t *testing.T) {
	eventtimer := constructEventTimer()
	err := eventtimer.startEventTimer(EventProposeBlockTimeout, 1)
	t.Logf("TestStartEventTimer: %v", err)
}

func TestCancelEventTimer(t *testing.T) {
	eventtimer := constructEventTimer()
	err := eventtimer.startEventTimer(EventProposeBlockTimeout, 1)
	t.Logf("startEventTimer: %v", err)
	err = eventtimer.cancelEventTimer(EventProposeBlockTimeout, 1)
	t.Logf("cancelEventTimer: %v", err)
}
