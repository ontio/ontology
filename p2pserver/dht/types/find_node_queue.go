package types

import (
	"fmt"
	"sync"
	"time"
)

type FindNodeQueue struct {
	lock             sync.Mutex
	resultChan       chan []*Node
	requestNodeQueue map[NodeID]*Node
	timeoutListener  chan NodeID
	onTimeOutEvent   func(requestNodeId NodeID)
}

func NewFindNodeQueue(onTimeOutEvent func(requestNodeId NodeID)) *FindNodeQueue {
	queue := new(FindNodeQueue)
	queue.resultChan = make(chan []*Node, 4)
	queue.timeoutListener = make(chan NodeID)
	queue.onTimeOutEvent = onTimeOutEvent
	queue.requestNodeQueue = make(map[NodeID]*Node, 0)
	// TODO should be invoked in dht.loop
	go queue.start()
	return queue
}

func (this *FindNodeQueue) start() {
	for {
		select {
		case requestNodeId := <-this.timeoutListener:
			this.onTimeOutEvent(requestNodeId)
		}
	}
}

func (this *FindNodeQueue) GetResultCh() <-chan []*Node {
	return this.resultChan
}

func (this *FindNodeQueue) SetResult(results []*Node, resultsFromNode NodeID) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if _, ok := this.requestNodeQueue[resultsFromNode]; ok {
		delete(this.requestNodeQueue, resultsFromNode)
		this.resultChan <- results
	}
}

func (this *FindNodeQueue) Timer(requestNodeId NodeID) {
	<-time.After(FIND_NODE_TIMEOUT)
	this.timeoutListener <- requestNodeId
}

func (this *FindNodeQueue) AddRequestNode(requestNode *Node) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.requestNodeQueue[requestNode.ID] = requestNode
}
