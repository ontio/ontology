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

/*
 * https://github.com/valyala/fasthttp
 * The MIT License (MIT)

 *Copyright (c) 2015-present Aliaksandr Valialkin, VertaMedia
 *Copyright (c) 2018-present Kirill Danshin
 *Copyright (c) 2018-present Erik Dubbelboer
 *Copyright (c) 2018-present FastHTTP Authors

 *Permission is hereby granted, free of charge, to any person obtaining a copy
 *of this software and associated documentation files (the "Software"), to deal
 *in the Software without restriction, including without limitation the rights
 *to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *copies of the Software, and to permit persons to whom the Software is
 *furnished to do so, subject to the following conditions:

 *The above copyright notice and this permission notice shall be included in all
 *copies or substantial portions of the Software.
 */

package utils

import (
	"runtime"
	"sync"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/protocol"
)

type msgJobChanList *[]*msgJobChan

type msgWorkerPool struct {
	maxWorkerCount    uint
	maxWorkerIdleTime time.Duration
	curWorkerCount    uint
	stopChan          chan struct{}
	waitingWokers     map[string]msgJobChanList
	workChanPool      sync.Pool
	lock              sync.Mutex
}

type msgJobItem struct {
	msgPayload *types.MsgPayload
	p2p        p2p.P2P
	pid        *actor.PID
	msgHandler MessageHandler
}

type msgJobChan struct {
	lastUseTime time.Time
	jobChan     chan *msgJobItem
}

func newMsgWorkerPool(maxWorkerCount uint) *msgWorkerPool {

	msgWP := &msgWorkerPool{
		maxWorkerCount: maxWorkerCount,
	}
	msgWP.init()

	return msgWP
}

func (this *msgWorkerPool) init() {

	this.waitingWokers = make(map[string]msgJobChanList)
	this.stopChan = make(chan struct{})
	this.maxWorkerIdleTime = 10 * time.Second
}

var msgJobChanCap = func() int {
	// Use blocking msgJobChan if GOMAXPROCS=1.
	// This immediately switches receiveMsg to disposeMsg, which results
	// in higher performance
	if runtime.GOMAXPROCS(0) == 1 {
		return 0
	}

	// Use non-blocking msgJobChan if GOMAXPROCS>1,
	// since otherwise the receiveMsg caller (Acceptor) may lag accepting
	// new message if disposeMsg is CPU-bound.
	return 1
}()

func (this *msgWorkerPool) start() {

	if this.waitingWokers == nil || this.stopChan == nil {
		panic("[p2p]invalid start invoking, the msg worker pool hasn't been initialized")
	}

	go func() {
		for {
			this.clean()
			select {
			case <-this.stopChan:
				return
			default:
				time.Sleep(this.maxWorkerIdleTime)
			}
		}
	}()
}

func (this *msgWorkerPool) stop() {

	if this.stopChan == nil {
		panic("[p2p]invalid stop invoking, the msg worker pool hasn't been initialized!")
	}
	close(this.stopChan)
	this.stopChan = nil
}

func (this *msgWorkerPool) clean() {

	var willCleanJC []*msgJobChan
	curTime := time.Now()

	this.lock.Lock()
	for _, v := range this.waitingWokers {
		jobChs := *v
		n := len(jobChs)
		if n <= 0 {
			continue
		}
		i := 0
		for i < n && jobChs[i] != nil && curTime.Sub(jobChs[i].lastUseTime) > this.maxWorkerIdleTime {
			i++
		}
		willCleanJC = append(willCleanJC[:0], jobChs[:i]...)
		if i > 0 {
			m := copy(*v, jobChs[i:])
			for j := m; j < n; j++ {
				(*v)[j] = nil
			}
			*v = (*v)[:m]
		}
	}
	this.lock.Unlock()

	tmp := willCleanJC
	for i, jCh := range tmp {
		jCh.jobChan <- nil
		tmp[i] = nil
	}
}

//Only receiveMsg can invoke getMsgWorkChan
func (this *msgWorkerPool) getMsgWorkChan(msgType string) *msgJobChan {

	var msgJobCh *msgJobChan = nil
	var willCreateNew = false

	if msgWaitingWorks, ok := this.waitingWokers[msgType]; ok {
		lmWW := len(*msgWaitingWorks)
		if lmWW > 0 {
			msgJobCh = (*msgWaitingWorks)[lmWW-1]
			(*msgWaitingWorks)[lmWW-1] = nil
			*msgWaitingWorks = (*msgWaitingWorks)[:(lmWW - 1)]
		} else {
			if this.curWorkerCount < this.maxWorkerCount {
				willCreateNew = true
			}
		}
	} else {
		msgWaitingWorksT := make([]*msgJobChan, 0)
		this.waitingWokers[msgType] = &msgWaitingWorksT

		if this.curWorkerCount < this.maxWorkerCount {
			willCreateNew = true
		}
	}

	if msgJobCh == nil && willCreateNew {
		mJobCh := this.workChanPool.Get()

		if mJobCh == nil {
			mJobCh = &msgJobChan{
				jobChan: make(chan *msgJobItem, msgJobChanCap),
			}
		}
		msgJobCh = mJobCh.(*msgJobChan)

		this.curWorkerCount++
		go func() {
			this.disposeMsg(msgJobCh)
			this.workChanPool.Put(msgJobCh)
		}()
	}

	return msgJobCh
}

func (this *msgWorkerPool) receiveMsg(msg *msgJobItem) bool {

	this.lock.Lock()
	defer this.lock.Unlock()

	msgJobCh := this.getMsgWorkChan(msg.msgPayload.Payload.CmdType())
	if msgJobCh != nil {
		msgJobCh.jobChan <- msg
		return true
	}

	return false
}

func (this *msgWorkerPool) disposeMsg(msgJobCh *msgJobChan) {

	for msgJob := range msgJobCh.jobChan {
		if msgJob == nil {
			break
		}
		msgJob.msgHandler(msgJob.msgPayload, msgJob.p2p, msgJob.pid)
		this.release(msgJob.msgPayload.Payload.CmdType(), msgJobCh)
	}

	this.lock.Lock()
	this.curWorkerCount--
	this.lock.Unlock()
}

func (this *msgWorkerPool) release(msgType string, msgJobCh *msgJobChan) {

	msgJobCh.lastUseTime = time.Now()

	this.lock.Lock()
	defer this.lock.Unlock()

	if msgWaitingWorks, ok := this.waitingWokers[msgType]; ok {
		*msgWaitingWorks = append(*msgWaitingWorks, msgJobCh)
	}
}
