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
	"github.com/pborman/uuid"
)

type Session struct {
	sync.Mutex
	mConnection *websocket.Conn
	nLastActive int64
	sessionId  string
}

const SESSION_TIMEOUT int64 = 300

func newSession(wsConn *websocket.Conn) (session *Session, err error) {
	sessionid := uuid.NewUUID().String()
	session = &Session{
		mConnection: wsConn,
		nLastActive: time.Now().Unix(),
		sessionId:  sessionid,
	}
	return session, nil
}

func (self *Session) GetSessionId() string {
	return self.sessionId
}

func (self *Session) close() {
	if self.mConnection != nil {
		self.mConnection.Close()
		self.mConnection = nil
	}
	self.sessionId = ""
}

func (self *Session) UpdateActiveTime() {
	self.Lock()
	defer self.Unlock()
	self.nLastActive = time.Now().Unix()
}

func (self *Session) Send(data []byte) error {
	if self.mConnection == nil {
		return errors.New("WebSocket is null")
	}
	self.Lock()
	defer self.Unlock()
	return self.mConnection.WriteMessage(websocket.TextMessage, data)
}

func (self *Session) SessionTimeoverCheck() bool {

	nCurTime := time.Now().Unix()
	if nCurTime-self.nLastActive > SESSION_TIMEOUT {
		//sec
		return true
	}
	return false
}
