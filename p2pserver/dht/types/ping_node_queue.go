package types

import (
	"sync"
	"time"
)

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
	// TODO should be invoked in dht.loop
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