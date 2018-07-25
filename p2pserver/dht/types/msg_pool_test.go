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

package types

import (
	"encoding/binary"
	"testing"

	"github.com/ontio/ontology/common/log"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.Init(log.PATH, log.Stdout)
}

func TestMsgPool(t *testing.T) {
	messagePool := NewRequestPool(func(requestId RequestId) {
		log.Infof("timeout: request id %s", requestId)
	})
	assert.NotNil(t, messagePool)

	node := &Node{
		IP:      "127.0.0.1",
		UDPPort: uint16(20332),
		TCPPort: uint16(20333),
	}
	id := ConstructID(node.IP, node.UDPPort)
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, id)
	copy(node.ID[:], b[:])

	requestId, isNewRequest := messagePool.AddRequest(node,
		DHT_PING_REQUEST, nil, nil)
	assert.Equal(t, true, isNewRequest)
	pendingNode, _ := messagePool.GetRequestData(requestId)
	assert.NotNil(t, pendingNode)
	assert.Equal(t, pendingNode.ID, node.ID)
	assert.Equal(t, pendingNode.IP, node.IP)
	assert.Equal(t, pendingNode.UDPPort, node.UDPPort)
	assert.Equal(t, pendingNode.TCPPort, node.TCPPort)

	messagePool.DeleteRequest(requestId)
	pendingNode, _ = messagePool.GetRequestData(requestId)
	assert.Nil(t, pendingNode)

	resultCh := messagePool.GetResultChan()
	assert.NotNil(t, resultCh)
}

func TestReqType(t *testing.T) {
	id := ConstructID("127.0.0.1", uint16(18888))
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, id)
	var nodeID NodeID
	copy(nodeID[:], b[:])
	reqId := ConstructRequestId(nodeID, DHT_FIND_NODE_REQUEST)
	reqType := GetReqTypeFromReqId(reqId)
	assert.Equal(t, reqType, DHT_FIND_NODE_REQUEST)
}
