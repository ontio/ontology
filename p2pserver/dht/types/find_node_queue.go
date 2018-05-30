package types

import (
	"fmt"
	"sync"
	"time"
)

type FindNodeQueue struct {
	lock              sync.Mutex
	resultChan        chan []*Node
	requestTimerQueue map[NodeID]*time.Timer
	timeoutListener   chan NodeID
	onTimeOutEvent    func(requestNodeId NodeID)
}

func NewFindNodeQueue(onTimeOutEvent func(requestNodeId NodeID)) *FindNodeQueue {
	queue := new(FindNodeQueue)
	queue.resultChan = make(chan []*Node, 4)
	queue.timeoutListener = make(chan NodeID)
	queue.onTimeOutEvent = onTimeOutEvent
	queue.requestTimerQueue = make(map[NodeID]*time.Timer, 0)
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

	if _, ok := this.requestTimerQueue[resultsFromNode]; ok {
		timer := this.requestTimerQueue[resultsFromNode]
		timer.Stop()
		delete(this.requestTimerQueue, resultsFromNode)
		this.resultChan <- results
	}
}

func (this *FindNodeQueue) StartRequestTimer(requestNode *Node) {
	this.lock.Lock()
	defer this.lock.Unlock()
	// start a timer
	timer := time.AfterFunc(FIND_NODE_TIMEOUT, func() {
		fmt.Println("timeout: ", requestNode.ID.String())
		this.timeoutListener <- requestNode.ID
	})
	this.requestTimerQueue[requestNode.ID] = timer
	go func() {
		<-timer.C
	}()
}
