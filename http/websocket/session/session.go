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
	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
	"sync"
	"time"
)

type Session struct {
	sync.Mutex
	mConnection *websocket.Conn
	nLastActive int64
	sSessionId  string
}

const sessionTimeOut int64 = 120

func (s *Session) GetSessionId() string {
	return s.sSessionId
}

func newSession(wsConn *websocket.Conn) (session *Session, err error) {
	sSessionId := uuid.NewUUID().String()
	session = &Session{
		mConnection: wsConn,
		nLastActive: time.Now().Unix(),
		sSessionId:  sSessionId,
	}
	return session, nil
}

func (s *Session) close() {
	if s.mConnection != nil {
		s.mConnection.Close()
		s.mConnection = nil
	}
	s.sSessionId = ""
}

func (s *Session) UpdateActiveTime() {
	s.Lock()
	defer s.Unlock()
	s.nLastActive = time.Now().Unix()
}

func (s *Session) Send(data []byte) error {
	if s.mConnection == nil {
		return errors.New("WebSocket is null")
	}
	//https://godoc.org/github.com/gorilla/websocket
	s.Lock()
	defer s.Unlock()
	return s.mConnection.WriteMessage(websocket.TextMessage, data)
}

func (s *Session) SessionTimeoverCheck() bool {

	nCurTime := time.Now().Unix()
	if nCurTime - s.nLastActive > sessionTimeOut {
		//sec
		return true
	}
	return false
}
