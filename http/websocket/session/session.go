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
	"time"

	"github.com/gorilla/websocket"
	cfg "github.com/ontio/ontology/common/config"
	"github.com/pborman/uuid"
)

type TxHashInfo struct {
	TxHash    string //transaction hash
	StartTime int64  //recv tranasction time
}

// Conn is the subset of a websocket connection that a session needs. It is an
// interface so the bounded-write behaviour of Send can be exercised by tests.
type Conn interface {
	WriteMessage(messageType int, data []byte) error
	SetWriteDeadline(t time.Time) error
	Close() error
}

//websocket session
type Session struct {
	sync.Mutex
	mConnection Conn
	nLastActive int64        //last active time
	sessionId   string       //session id
	TxHashArr   []TxHashInfo //transaction hashes in this session
}

const SESSION_TIMEOUT int64 = 300

// SESSION_WRITE_TIMEOUT bounds a single websocket write. Without a deadline a
// client that stops reading would block the writer forever once its send
// buffer fills up, which in turn blocks the global broadcast path that drives
// every other subscriber. The value is intentionally generous compared to
// normal message sizes but small enough to keep a stalled peer from wedging
// the push pipeline.
const SESSION_WRITE_TIMEOUT = 10 * time.Second

//create new session
func newSession(wsConn Conn) *Session {
	sessionid := uuid.NewUUID().String()
	session := &Session{
		mConnection: wsConn,
		nLastActive: time.Now().Unix(),
		sessionId:   sessionid,
		TxHashArr:   []TxHashInfo{},
	}
	return session
}

func (self *Session) GetSessionId() string {
	return self.sessionId
}

//close session
func (self *Session) Close() {
	self.Lock()
	defer self.Unlock()
	if self.mConnection != nil {
		self.mConnection.Close()
		self.mConnection = nil
	}
	// The session id is intentionally left untouched. It is the key used to
	// clean up the subscription and transaction-hash maps, and closing a
	// session can happen before that cleanup runs (for example when the
	// timeout ticker closes it first). Clearing the id here would make the
	// later cleanup delete the empty-string key and leak the real entries.
}

//update active time
func (self *Session) UpdateActiveTime() {
	self.Lock()
	defer self.Unlock()
	self.nLastActive = time.Now().Unix()
}

//send to client
func (self *Session) Send(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	self.Lock()
	defer self.Unlock()
	if self.mConnection == nil {
		return errors.New("WebSocket is null")
	}

	// Bound the network write so a slow or stalled client cannot block the
	// caller indefinitely.
	self.mConnection.SetWriteDeadline(time.Now().Add(SESSION_WRITE_TIMEOUT))
	if err := self.mConnection.WriteMessage(websocket.TextMessage, data); err != nil {
		// A write that failed - in particular one that exceeded the deadline -
		// leaves the websocket connection in an undefined state. Tear it down
		// so the session is no longer a target for later broadcasts and will
		// be reclaimed by the session timeout cleanup.
		self.mConnection.Close()
		self.mConnection = nil
		return err
	}
	return nil
}

//sesseion time over check
func (self *Session) SessionTimeoverCheck() bool {
	self.Lock()
	defer self.Unlock()
	nCurTime := time.Now().Unix()
	if nCurTime-self.nLastActive > SESSION_TIMEOUT {
		//sec
		return true
	}
	return false
}

//remove timew over transaction hashes
func (self *Session) RemoveTimeoverTxHashes() (remove []TxHashInfo) {
	self.Lock()
	defer self.Unlock()
	index := len(self.TxHashArr)
	now := time.Now().Unix()
	for k, v := range self.TxHashArr {
		if (now - v.StartTime) < int64(cfg.DEFAULT_GEN_BLOCK_TIME*10) {
			index = k
			break
		}
	}
	remove = self.TxHashArr[0:index]
	self.TxHashArr = self.TxHashArr[index:]
	return remove
}

//append transaction hash
func (self *Session) AppendTxHash(txhash string) {
	self.Lock()
	defer self.Unlock()
	self.TxHashArr = append(self.TxHashArr, TxHashInfo{txhash, time.Now().Unix()})
}
