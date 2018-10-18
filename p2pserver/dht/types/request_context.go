package types

import (
	"sync"
	"time"
)

type requestContext struct {
	destNodes    map[RequestId]*Node
	replaceNodes map[RequestId]*Node

	syncRequestLock map[RequestId]*sync.WaitGroup // used to synchronize
	timerQueue      map[RequestId]*time.Timer     // timeout timer

	timeOutChan chan RequestId

	lock sync.RWMutex
}

func newRequestContext() *requestContext {
	return &requestContext{
		destNodes:       make(map[RequestId]*Node),
		replaceNodes:    make(map[RequestId]*Node),
		syncRequestLock: make(map[RequestId]*sync.WaitGroup),
		timerQueue:      make(map[RequestId]*time.Timer),

		timeOutChan: make(chan RequestId, MESSAGE_POOL_BUFFER_SIZE),
	}
}

func (this *requestContext) addRequest(requestId RequestId, destNode, replaceNode *Node,
	waitGroup *sync.WaitGroup, timeOut time.Duration) (isNewRequest bool) {
	this.lock.Lock()
	defer this.lock.Unlock()

	_, ok := this.destNodes[requestId]
	if ok {
		this.timerQueue[requestId].Reset(timeOut)
	} else {
		this.destNodes[requestId] = destNode
		this.replaceNodes[requestId] = replaceNode

		if waitGroup != nil {
			waitGroup.Add(1)
			this.syncRequestLock[requestId] = waitGroup
		}

		timeoutTimer := time.AfterFunc(timeOut, func() {
			this.timeOutChan <- requestId
		})
		this.timerQueue[requestId] = timeoutTimer

		// start timeout timer
		go func() {
			<-timeoutTimer.C
		}()
	}
	return !ok
}

func (this *requestContext) deleteRequest(requestId RequestId) {
	this.lock.Lock()
	defer this.lock.Unlock()

	_, ok := this.destNodes[requestId]
	if ok {
		delete(this.destNodes, requestId)
		delete(this.replaceNodes, requestId)

		syncRequesLock, ok := this.syncRequestLock[requestId]
		if ok {
			syncRequesLock.Done()
			delete(this.syncRequestLock, requestId)
		}

		timer := this.timerQueue[requestId]
		if timer != nil {
			timer.Stop()
		}
		delete(this.timerQueue, requestId)
	} else {
		return
	}
}

func (this *requestContext) getTimeOutRequest() <-chan RequestId {
	return this.timeOutChan
}

func (this *requestContext) getDestNode(requestId RequestId) (*Node, bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()

	node, ok := this.destNodes[requestId]
	return node, ok
}

func (this *requestContext) getReplaceNode(requestId RequestId) (*Node, bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()

	node, ok := this.replaceNodes[requestId]
	return node, ok
}
