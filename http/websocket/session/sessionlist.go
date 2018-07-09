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

// Package session privides websocket server manager sessionlist
package session

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

const MAX_SESSION_COUNT = 3000

type SessionList struct {
	sync.RWMutex
	mapOnlineList map[string]*Session //key is SessionId
}

// new websocket session list
func NewSessionList() *SessionList {
	return &SessionList{
		mapOnlineList: make(map[string]*Session),
	}
}
func (self *SessionList) NewSession(wsConn *websocket.Conn) (session *Session, err error) {
	if self.GetSessionCount() > MAX_SESSION_COUNT {
		return nil, errors.New("over MAX_SESSION_COUNT")
	}
	session = newSession(wsConn)

	self.Lock()
	self.mapOnlineList[session.GetSessionId()] = session
	self.Unlock()

	return session, nil
}
func (self *SessionList) CloseSession(session *Session) {
	if session == nil {
		return
	}
	self.removeSession(session)
	session.Close()
}

func (self *SessionList) removeSession(session *Session) {
	self.Lock()
	defer self.Unlock()
	delete(self.mapOnlineList, session.GetSessionId())
}

func (self *SessionList) GetSessionById(sessionId string) *Session {
	self.RLock()
	defer self.RUnlock()
	return self.mapOnlineList[sessionId]

}

func (self *SessionList) GetSessionCount() int {
	self.RLock()
	defer self.RUnlock()
	return len(self.mapOnlineList)
}

func (self *SessionList) ForEachSession(visit func(*Session)) {
	self.RLock()
	defer self.RUnlock()
	for _, v := range self.mapOnlineList {
		visit(v)
	}
}
