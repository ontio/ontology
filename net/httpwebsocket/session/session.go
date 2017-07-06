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
	if nCurTime-s.nLastActive > sessionTimeOut { //sec
		return true
	}
	return false
}
