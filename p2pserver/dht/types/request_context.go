package types

import (
	"sync"
	"time"
)

type reqContext struct {
	destNode        *Node
	replaceNode     *Node
	syncRequestLock *sync.WaitGroup
	timeOutTimer    *time.Timer

	timeoutListener chan RequestId
}

func newReqContext(requestId RequestId, destNode, replaceNode *Node, syncLock *sync.WaitGroup,
	timeoutListener chan RequestId, timeoutTime time.Duration) *reqContext {
	ctx := &reqContext{
		destNode:    destNode,
		replaceNode: replaceNode,

		timeoutListener: timeoutListener,
	}

	if syncLock != nil {
		syncLock.Add(1)
		ctx.syncRequestLock = syncLock
	}
	timer := time.AfterFunc(timeoutTime, func() {
		ctx.timeoutListener <- requestId
	})
	ctx.timeOutTimer = timer
	go func() {
		<-timer.C
	}()
	return ctx
}

func (this *reqContext) clean() {
	if this.timeOutTimer != nil {
		this.timeOutTimer.Stop()
	}
	if this.syncRequestLock != nil {
		this.syncRequestLock.Done()
	}
}

func (this *reqContext) reset(timeoutTime time.Duration) {
	if this.timeOutTimer != nil {
		this.timeOutTimer.Reset(timeoutTime)
	}
}
