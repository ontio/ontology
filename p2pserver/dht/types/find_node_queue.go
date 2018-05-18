package types

import (
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
	queue.resultChan = make(chan []*Node)
	queue.timeoutListener = make(chan NodeID)
	queue.onTimeOutEvent = onTimeOutEvent
	// TODO should be invoked in dht.loop
	go queue.start()
	return queue
}

func (nodeQueue *FindNodeQueue) start() {
	for {
		select {
		case requestNodeId := <-nodeQueue.timeoutListener:
			nodeQueue.onTimeOutEvent(requestNodeId)
		}
	}
}

func (nodeQueue *FindNodeQueue) GetResultCh() <-chan []*Node {
	return nodeQueue.resultChan
}

func (nodeQueue *FindNodeQueue) SetResult(results []*Node, resultsFromNode NodeID) {
	nodeQueue.lock.Lock()
	defer nodeQueue.lock.Unlock()

	if _, ok := nodeQueue.requestNodeQueue[resultsFromNode]; ok {
		delete(nodeQueue.requestNodeQueue, resultsFromNode)
		nodeQueue.resultChan <- results
	}
}

func (nodeQueue *FindNodeQueue) Timer(requestNodeId NodeID) {
	<-time.After(FIND_NODE_TIMEOUT)
	nodeQueue.timeoutListener <- requestNodeId
}

func (nodeQueue *FindNodeQueue) AddRequestNode(requestNode *Node) {
	nodeQueue.lock.Lock()
	defer nodeQueue.lock.Unlock()

	nodeQueue.requestNodeQueue[requestNode.ID] = requestNode
}
