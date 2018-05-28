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
	requestTimerQueue  map[NodeID]*time.Timer
	onTimeOut          func(id NodeID) // time out event should be handled by dht
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

func (this *PingNodeQueue) start() {
	for {
		select {
		// time out
		case nodeId := <-this.deleteNodeListener:
			this.onTimeOut(nodeId)
		}
	}
}

func (this *PingNodeQueue) AddNode(requestNode, pendingNode *Node) {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.requestNodeQueue[requestNode.ID] = requestNode
	this.pendingNodeQueue[requestNode.ID] = pendingNode
	// start timer
	timer := time.AfterFunc(PING_TIMEOUT, func() {
		this.deleteNodeListener <- requestNode.ID
	})
	this.requestTimerQueue[requestNode.ID] = timer
	go func() {
		<-timer.C
	}()
}

func (this *PingNodeQueue) DeleteNode(node NodeID) {
	this.lock.Lock()
	defer this.lock.Unlock()

	delete(this.requestNodeQueue, node)
	delete(this.pendingNodeQueue, node)
	timer := this.requestTimerQueue[node]
	timer.Stop()
	delete(this.requestTimerQueue, node)
}

func (this *PingNodeQueue) GetRequestNode(node NodeID) (*Node, bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()

	result, ok := this.requestNodeQueue[node]
	return result, ok
}

func (this *PingNodeQueue) GetPendingNode(node NodeID) (*Node, bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()

	result, ok := this.pendingNodeQueue[node]
	return result, ok
}
