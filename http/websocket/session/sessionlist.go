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
	"sync"
	"github.com/gorilla/websocket"
	"errors"
)

const MAX_SESSION_COUNT = 3000

type SessionList struct {
	sync.RWMutex
	mapOnlineList map[string]*Session //key is SessionId
}

func NewSessionList() *SessionList {
	return &SessionList{
		mapOnlineList: make(map[string]*Session),
	}
}
func (self *SessionList) NewSession(wsConn *websocket.Conn) (session *Session, err error) {
	if self.GetSessionCount() > MAX_SESSION_COUNT {
		return nil, errors.New("over MAX_SESSION_COUNT")
	}
	session, err = newSession(wsConn)
	if err == nil {
		self.addOnlineSession(session)
	}
	return session, err
}
func (self *SessionList) CloseSession(session *Session) {
	if session == nil {
		return
	}
	self.removeSession(session)
	session.close()
}
func (self *SessionList) addOnlineSession(session *Session) {
	if session.GetSessionId() == "" {
		return
	}
	self.Lock()
	defer self.Unlock()
	self.mapOnlineList[session.GetSessionId()] = session
}

func (self *SessionList) removeSession(iSession *Session) (err error) {
	return self.removeSessionById(iSession.GetSessionId())
}

func (self *SessionList) removeSessionById(sSessionId string) (err error) {

	if sSessionId == "" {
		return err
	}
	self.Lock()
	defer self.Unlock()
	delete(self.mapOnlineList, sSessionId)
	return nil
}

func (self *SessionList) GetSessionById(sSessionId string) *Session {
	self.RLock()
	defer self.RUnlock()
	if session, ok := self.mapOnlineList[sSessionId]; ok {
		return session
	}
	return nil

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
