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
