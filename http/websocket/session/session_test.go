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

package session

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeConn is a stand-in websocket connection that records the write deadline
// it was given and can be made to fail a write.
type fakeConn struct {
	mu       sync.Mutex
	deadline time.Time
	wrote    bool
	closed   bool
	writeErr error
}

func (c *fakeConn) WriteMessage(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.wrote = true
	return c.writeErr
}

func (c *fakeConn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deadline = t
	return nil
}

func (c *fakeConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

// A write that fails (for example because it exceeded the write deadline) must
// tear the connection down so the stalled session stops being a broadcast
// target and later sends fail fast instead of blocking.
func TestSendClosesConnectionOnWriteError(t *testing.T) {
	fc := &fakeConn{writeErr: errors.New("i/o timeout")}
	s := &Session{mConnection: fc}

	start := time.Now()
	if err := s.Send([]byte("hello")); err == nil {
		t.Fatal("expected Send to return the write error")
	}

	fc.mu.Lock()
	deadline, closed := fc.deadline, fc.closed
	fc.mu.Unlock()

	if deadline.IsZero() {
		t.Fatal("Send did not set a write deadline")
	}
	if !deadline.After(start) {
		t.Fatalf("write deadline %v is not in the future", deadline)
	}
	if budget := deadline.Sub(start); budget > SESSION_WRITE_TIMEOUT+time.Second {
		t.Fatalf("write deadline too far out: %v", budget)
	}
	if !closed {
		t.Fatal("connection was not closed after a failed write")
	}

	// Once the connection is torn down, further sends must not touch it.
	if err := s.Send([]byte("again")); err == nil {
		t.Fatal("expected Send to fail after the connection was closed")
	}
	if !fc.closed {
		t.Fatal("connection state changed unexpectedly")
	}
}

func TestSendSkipsEmptyData(t *testing.T) {
	fc := &fakeConn{}
	s := &Session{mConnection: fc}

	if err := s.Send(nil); err != nil {
		t.Fatalf("Send(nil) returned %v", err)
	}
	fc.mu.Lock()
	wrote := fc.wrote
	fc.mu.Unlock()
	if wrote {
		t.Fatal("Send(nil) unexpectedly wrote to the connection")
	}
}
