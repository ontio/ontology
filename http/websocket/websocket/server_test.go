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

package websocket

import (
	"sync"
	"testing"
	"time"

	Err "github.com/ontio/ontology/http/base/error"
	"github.com/ontio/ontology/http/base/rest"
)

// blockingConn is a session connection whose write blocks until it is
// released, emulating a slow client whose send buffer is full.
type blockingConn struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func newBlockingConn() *blockingConn {
	return &blockingConn{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (c *blockingConn) WriteMessage(messageType int, data []byte) error {
	c.once.Do(func() { close(c.started) })
	<-c.release
	return nil
}

func (c *blockingConn) SetWriteDeadline(t time.Time) error { return nil }

func (c *blockingConn) Close() error { return nil }

// BroadcastToSubscribers must not hold the global server lock while a slow
// client's write is in flight, otherwise that one client blocks the whole
// broadcast and every control path (subscribe, cleanup, heartbeat) that needs
// the same lock.
func TestBroadcastReleasesLockDuringWrite(t *testing.T) {
	ws := InitWsServer()

	conn := newBlockingConn()
	s, err := ws.SessionList.NewSession(conn)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	ws.SubscribeMap[s.GetSessionId()] = subscribe{SubscribeJsonBlock: true}

	done := make(chan struct{})
	go func() {
		resp := rest.ResponsePack(Err.SUCCESS)
		resp["Action"] = "sendjsonblock"
		ws.BroadcastToSubscribers(nil, WSTOPIC_JSON_BLOCK, resp)
		close(done)
	}()

	select {
	case <-conn.started:
	case <-time.After(2 * time.Second):
		t.Fatal("broadcast never attempted to write")
	}

	// With the write now blocked, the global lock must still be acquirable.
	locked := make(chan struct{})
	go func() {
		ws.Lock()
		ws.Unlock()
		close(locked)
	}()
	select {
	case <-locked:
	case <-time.After(2 * time.Second):
		t.Fatal("global lock was held during a network write")
	}

	close(conn.release)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("broadcast did not finish after the client resumed reading")
	}
}

// Only sessions subscribed to the broadcast topic are written to.
func TestBroadcastOnlyWritesToMatchingSessions(t *testing.T) {
	ws := InitWsServer()

	subscribedConn := &countingConn{}
	subscribed, err := ws.SessionList.NewSession(subscribedConn)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	ws.SubscribeMap[subscribed.GetSessionId()] = subscribe{SubscribeJsonBlock: true}

	otherConn := &countingConn{}
	notSubscribed, err := ws.SessionList.NewSession(otherConn)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	ws.SubscribeMap[notSubscribed.GetSessionId()] = subscribe{SubscribeRawBlock: true}

	resp := rest.ResponsePack(Err.SUCCESS)
	resp["Action"] = "sendjsonblock"
	ws.BroadcastToSubscribers(nil, WSTOPIC_JSON_BLOCK, resp)

	if subscribedConn.writes != 1 {
		t.Fatalf("matching session received %d writes, want 1", subscribedConn.writes)
	}
	if otherConn.writes != 0 {
		t.Fatalf("non-matching session received %d writes", otherConn.writes)
	}
}

type countingConn struct {
	writes int
}

func (c *countingConn) WriteMessage(messageType int, data []byte) error {
	c.writes++
	return nil
}

func (c *countingConn) SetWriteDeadline(t time.Time) error { return nil }

func (c *countingConn) Close() error                       { return nil }
