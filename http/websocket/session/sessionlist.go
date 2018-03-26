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
	"github.com/gorilla/websocket"
	"sync"
)

type SessionList struct {
	sync.RWMutex
	mapOnlineList map[string]*Session //key is SessionId
}

func NewSessionList() *SessionList {
	return &SessionList{
		mapOnlineList: make(map[string]*Session),
	}
}
func (this *SessionList) NewSession(wsConn *websocket.Conn) (session *Session, err error) {
	session, err = newSession(wsConn)
	if err == nil {
		this.addOnlineSession(session)
	}
	return session, err
}
func (this *SessionList) CloseSession(session *Session) {
	if session == nil {
		return
	}
	this.removeSession(session)
	session.close()
}
func (this *SessionList) addOnlineSession(session *Session) {
	if session.GetSessionId() == "" {
		return
	}
	this.Lock()
	defer this.Unlock()
	this.mapOnlineList[session.GetSessionId()] = session
}

func (this *SessionList) removeSession(iSession *Session) (err error) {
	return this.removeSessionById(iSession.GetSessionId())
}

func (this *SessionList) removeSessionById(sSessionId string) (err error) {

	if sSessionId == "" {
		return err
	}
	this.Lock()
	defer this.Unlock()
	delete(this.mapOnlineList, sSessionId)
	return nil
}

func (this *SessionList) GetSessionById(sSessionId string) *Session {
	this.RLock()
	defer this.RUnlock()
	if session, ok := this.mapOnlineList[sSessionId]; ok {
		return session
	}
	return nil

}
func (this *SessionList) GetSessionCount() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.mapOnlineList)
}
func (this *SessionList) ForEachSession(visit func(*Session)) {
	this.RLock()
	defer this.RUnlock()
	for _, v := range this.mapOnlineList {
		visit(v)
	}
}
