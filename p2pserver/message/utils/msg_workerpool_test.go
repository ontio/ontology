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

package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/protocol"
)

type testMsgPayload struct{}

func (this *testMsgPayload) Serialization(sink *common.ZeroCopySink) error {
	return nil
}

func (this *testMsgPayload) Deserialization(source *common.ZeroCopySource) error {
	return nil
}

func (this *testMsgPayload) CmdType() string {
	return "TestMsg"
}

func msgHandler(data *types.MsgPayload, p2p p2p.P2P, pid *actor.PID, args ...interface{}) {
	log.Info("msg handler")
}

func TestMsgWPStartStop(t *testing.T) {
	msgWP := &msgWorkerPool{
		maxWorkerCount: 5,
	}
	assert.NotNil(t, msgWP)

	msgWP.init()
	msgWP.start()
	msgWP.stop()
}

func TestMsgWPBWithinMaxWorkerCount(t *testing.T) {
	msgWP := &msgWorkerPool{
		maxWorkerCount: 5,
	}
	assert.NotNil(t, msgWP)

	msgWP.init()
	msgWP.start()

	for i := 0; i < 3; i++ {
		mJobItem := &msgJobItem{
			msgPayload: &types.MsgPayload{Payload: &testMsgPayload{}},
			msgHandler: testHandler,
		}
		bRecvSucess := msgWP.receiveMsg(mJobItem)
		assert.Equal(t, true, bRecvSucess)
	}

	msgWP.stop()
}

func TestMsgWPBBeyondMaxWorkerCount(t *testing.T) {
	msgWP := &msgWorkerPool{
		maxWorkerCount: 5,
	}
	assert.NotNil(t, msgWP)

	msgWP.init()
	msgWP.start()

	for i := 0; i < 8; i++ {
		mJobItem := &msgJobItem{
			msgPayload: &types.MsgPayload{Payload: &testMsgPayload{}},
			msgHandler: msgHandler,
		}
		bRecvSucess := msgWP.receiveMsg(mJobItem)
		assert.Equal(t, true, bRecvSucess)

		time.Sleep(time.Millisecond * 100)
	}

	msgWP.stop()
}

func TestMsgWPBAutoClean(t *testing.T) {
	msgWP := &msgWorkerPool{
		maxWorkerCount: 5,
	}
	assert.NotNil(t, msgWP)

	msgWP.init()
	msgWP.start()

	for i := 0; i < 3; i++ {
		mJobItem := &msgJobItem{
			msgPayload: &types.MsgPayload{Payload: &testMsgPayload{}},
			msgHandler: testHandler,
		}
		bRecvSucess := msgWP.receiveMsg(mJobItem)
		assert.Equal(t, true, bRecvSucess)
	}

	time.Sleep(time.Second * 15)
	assert.Equal(t, uint(0), msgWP.curWorkerCount)
	assert.Equal(t, 0, len(*msgWP.waitingWokers["TestMsg"]))

	msgWP.stop()
}
