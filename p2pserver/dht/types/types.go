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
	"github.com/ontio/ontology/common"
	"net"
	"sync"
	"time"
)

const (
	BUCKET_NUM        = 256
	BUCKET_SIZE       = 8
	FACTOR            = 3
	MSG_CACHE         = 10240
	PING_TIMEOUT      = 10 * time.Second
	FIND_NODE_TIMEOUT = 20 * time.Second
)

type ptype uint8

const (
	ping_rpc ptype = iota
	pong_rpc
	find_node_rpc
	neighbors_rpc
)

type DHTMessage struct {
	From    *net.UDPAddr
	Payload []byte
}

type Node struct {
	ID      NodeID
	Hash    common.Uint256
	IP      string
	UDPPort uint16
	TCPPort uint16
}

type PingNodeQueue struct {
	lock               sync.RWMutex
	deleteNodeListener chan NodeID
	requestNodeQueue   map[NodeID]*Node
	pendingNodeQueue   map[NodeID]*Node // used to record the node corresponded to request node
	onTimeOut func(id NodeID) // time out event should be handled by dht
}

func NewPingNodeQueue(onTimeOut func(id NodeID)) *PingNodeQueue {
	nodeQueue := new(PingNodeQueue)
	nodeQueue.requestNodeQueue = make(map[NodeID]*Node)
	nodeQueue.pendingNodeQueue = make(map[NodeID]*Node)
	nodeQueue.deleteNodeListener = make(chan NodeID)
	nodeQueue.onTimeOut = onTimeOut
	go nodeQueue.start()
	return nodeQueue
}

func (nodeQueue *PingNodeQueue) start() {
	for {
		select {
		// time out
		case nodeId := <-nodeQueue.deleteNodeListener:
			nodeQueue.onTimeOut(nodeId)
		}
	}
}

func (nodeQueue *PingNodeQueue) AddNode(requestNode, pendingNode *Node, timeout time.Duration) {
	nodeQueue.lock.Lock()
	defer nodeQueue.lock.Unlock()

	nodeQueue.requestNodeQueue[requestNode.ID] = requestNode
	nodeQueue.pendingNodeQueue[requestNode.ID] = pendingNode
	go func(queue *PingNodeQueue) {
		<-time.After(timeout)
		queue.deleteNodeListener <- requestNode.ID
	}(nodeQueue)
}

func (nodeQueue *PingNodeQueue) DeleteNode(node NodeID) {
	nodeQueue.lock.Lock()
	defer nodeQueue.lock.Unlock()

	delete(nodeQueue.requestNodeQueue, node)
	delete(nodeQueue.pendingNodeQueue, node)
}

func (nodeQueue *PingNodeQueue) GetRequestNode(node NodeID) (*Node, bool) {
	nodeQueue.lock.RLock()
	defer nodeQueue.lock.RUnlock()

	result, ok := nodeQueue.requestNodeQueue[node]
	return result, ok
}

func (nodeQueue *PingNodeQueue) GetPendingNode(node NodeID) (*Node, bool) {
	nodeQueue.lock.RLock()
	defer nodeQueue.lock.RUnlock()

	result, ok := nodeQueue.pendingNodeQueue[node]
	return result, ok
}