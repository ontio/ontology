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
