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

	fmt.Println("SetResult: nodes %d, ", len(results), resultsFromNode.String())
	for id := range this.requestNodeQueue {
		fmt.Println("SetResult: ", id)
	}

	if _, ok := this.requestNodeQueue[resultsFromNode]; ok {
		fmt.Println("SetResult: id %x", resultsFromNode)
		delete(this.requestNodeQueue, resultsFromNode)
		this.resultChan <- results
	}
}

func (this *FindNodeQueue) Timer(requestNodeId NodeID) {
	<-time.After(FIND_NODE_TIMEOUT)
	fmt.Println("timeout: ", requestNodeId.String())
	this.timeoutListener <- requestNodeId
}

func (this *FindNodeQueue) AddRequestNode(requestNode *Node) {
	this.lock.Lock()
	defer this.lock.Unlock()
	fmt.Println("AddRequestNode: ", requestNode.ID.String())
	this.requestNodeQueue[requestNode.ID] = requestNode
}
