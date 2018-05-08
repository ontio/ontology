package test

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
	"time"
)

func TestWs(t *testing.T) {
	for i := 0; i < 5; i++ {
		go work(t, i)
	}
	time.Sleep(time.Second * 5)
}

func work(t *testing.T, id int) {
	u := url.URL{Scheme: "ws", Host: "localhost:20335", Path: ""}
	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		assert.Error(t, err)
	}
	defer ws.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				assert.Error(t, err)
				return
			}
			assert.Contains(t, string(message), "SUCCESS")
		}
	}()

	ticker := time.NewTicker(time.Nanosecond)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			err := ws.WriteMessage(websocket.TextMessage, []byte("{\"Action\":\"heartbeat\",\"SubscribeBlockTxHashs\":true}"))
			if err != nil {
				assert.Error(t, err)
				return
			}
		}
	}
}
