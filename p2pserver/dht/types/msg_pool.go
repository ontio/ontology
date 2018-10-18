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
	"bytes"
	"sync"
	"time"
)

type DHTRequestType string

const (
	DHT_FIND_NODE_REQUEST DHTRequestType = "find_node"
	DHT_PING_REQUEST      DHTRequestType = "ping"
)

const MESSAGE_POOL_BUFFER_SIZE = 4

type RequestId string

func ConstructRequestId(nodeId NodeID, reqType DHTRequestType) RequestId {
	var buffer bytes.Buffer
	buffer.WriteString(nodeId.String())
	buffer.WriteString(string(reqType))
	return RequestId(buffer.String())
}

func GetReqTypeFromReqId(reqId RequestId) DHTRequestType {
	nodeIdLength := len(NodeID{}.String())
	return DHTRequestType(reqId[nodeIdLength:])
}

type DHTMessagePool struct {
	request         map[RequestId]*reqContext
	timeoutListener chan RequestId
	onTimeOut       func(id RequestId) // time out event should be handled by dht

	resultChan chan []*Node

	lock sync.RWMutex
}

func NewRequestPool(onTimeOut func(id RequestId)) *DHTMessagePool {
	msgPool := new(DHTMessagePool)
	msgPool.onTimeOut = onTimeOut
	msgPool.resultChan = make(chan []*Node, MESSAGE_POOL_BUFFER_SIZE)
	msgPool.request = make(map[RequestId]*reqContext)
	msgPool.timeoutListener = make(chan RequestId)
	go msgPool.start()
	return msgPool
}

func (this *DHTMessagePool) start() {
	for {
		select {
		// time out
		case requestId := <-this.timeoutListener:
			this.onTimeOut(requestId)
		}
	}
}

// AddRequest: when send a ping or find node request, call this function
// destinateNode: request to the node
// replaceNode: should be replaced node
// reqType: request type
// return the request is new request
func (this *DHTMessagePool) AddRequest(destinateNode *Node, reqType DHTRequestType, replaceNode *Node,
	waitGroup *sync.WaitGroup) (id RequestId, isNewRequest bool) {
	this.lock.Lock()
	defer this.lock.Unlock()

	requestId := ConstructRequestId(destinateNode.ID, reqType)
	var timeout time.Duration
	if reqType == DHT_FIND_NODE_REQUEST {
		timeout = FIND_NODE_TIMEOUT
	} else if reqType == DHT_PING_REQUEST {
		timeout = PING_TIMEOUT
	} else {
		timeout = DEFAULT_TIMEOUT
	}
	req, ok := this.request[requestId]
	isNewRequest = !ok
	if ok {
		req.reset(timeout)
	} else {
		this.request[requestId] = newReqContext(requestId, destinateNode, replaceNode, waitGroup,
			this.timeoutListener, timeout)
	}
	return requestId, isNewRequest
}

func (this *DHTMessagePool) GetReplaceNode(id RequestId) (*Node, bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()

	var destNode *Node
	req, ok := this.request[id]
	if ok {
		destNode = req.replaceNode
	}
	return destNode, ok
}

func (this *DHTMessagePool) GetRequestData(id RequestId) (*Node, bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()

	var replaceNode *Node
	req, ok := this.request[id]
	if ok {
		replaceNode = req.destNode
	}
	return replaceNode, ok
}

func (this *DHTMessagePool) DeleteRequest(requestId RequestId) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if req, ok := this.request[requestId]; ok {
		req.clean()
		delete(this.request, requestId)
	}
}

// push result
func (this *DHTMessagePool) SetResults(results []*Node) {
	this.resultChan <- results
}

// get results channel
func (this *DHTMessagePool) GetResultChan() <-chan []*Node {
	return this.resultChan
}
