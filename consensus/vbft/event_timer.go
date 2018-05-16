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

package vbft

import (
	"container/heap"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ontio/ontology/common/log"
)

type TimerEventType int

const (
	EventProposeBlockTimeout TimerEventType = iota
	EventProposalBackoff
	EventRandomBackoff
	EventPropose2ndBlockTimeout
	EventEndorseBlockTimeout
	EventEndorseEmptyBlockTimeout
	EventCommitBlockTimeout
	EventPeerHeartbeat
	EventTxPool
	EventTxBlockTimeout
	EventMax
)

var (
	makeProposalTimeout    = 300 * time.Millisecond
	make2ndProposalTimeout = 300 * time.Millisecond
	endorseBlockTimeout    = 100 * time.Millisecond
	commitBlockTimeout     = 200 * time.Millisecond
	peerHandshakeTimeout   = 10 * time.Second
	txPooltimeout          = 300 * time.Millisecond
	zeroTxBlockTimeout     = 10 * time.Second
)

type SendMsgEvent struct {
	ToPeer uint32 // peer index
	Msg    ConsensusMsg
}

type TimerEvent struct {
	evtType  TimerEventType
	blockNum uint32
	msg      ConsensusMsg
}

type perBlockTimer map[uint32]*time.Timer

type EventTimer struct {
	lock   sync.Mutex
	server *Server
	C      chan *TimerEvent
	stopC  chan interface{}
	//timerQueue TimerQueue

	// bft timers
	eventTimers map[TimerEventType]perBlockTimer

	// peer heartbeat tickers
	peerTickers map[uint32]*time.Timer
	// other timers
	normalTimers map[uint32]*time.Timer
}

func NewEventTimer(server *Server) *EventTimer {
	timer := &EventTimer{
		server:       server,
		C:            make(chan *TimerEvent, 64),
		stopC:        make(chan interface{}),
		eventTimers:  make(map[TimerEventType]perBlockTimer),
		peerTickers:  make(map[uint32]*time.Timer),
		normalTimers: make(map[uint32]*time.Timer),
	}

	for i := 0; i < int(EventMax); i++ {
		timer.eventTimers[TimerEventType(i)] = make(map[uint32]*time.Timer)
	}

	return timer
}

func stopAllTimers(timers map[uint32]*time.Timer) {
	for _, t := range timers {
		t.Stop()
	}
}

func (self *EventTimer) stop() {
	self.lock.Lock()
	defer self.lock.Unlock()

	// clear timers by event timer
	for i := 0; i < int(EventMax); i++ {
		stopAllTimers(self.eventTimers[TimerEventType(i)])
		self.eventTimers[TimerEventType(i)] = make(map[uint32]*time.Timer)
	}

	// clear normal timers
	stopAllTimers(self.normalTimers)
	self.normalTimers = make(map[uint32]*time.Timer)

	self.stopC <- struct{}{}
}

func (self *EventTimer) StartTimer(Idx uint32, timeout time.Duration) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if t, present := self.normalTimers[Idx]; present {
		t.Stop()
		log.Infof("timer for %d got reset", Idx)
	}

	self.normalTimers[Idx] = time.AfterFunc(timeout, func() {
		// remove timer from map
		self.lock.Lock()
		defer self.lock.Unlock()
		delete(self.normalTimers, Idx)

		self.C <- &TimerEvent{
			evtType:  EventMax,
			blockNum: Idx,
		}
	})

	return nil
}

func (self *EventTimer) CancelTimer(idx uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if t, present := self.normalTimers[idx]; present {
		t.Stop()
		delete(self.normalTimers, idx)
	}
	return nil
}

func (self *EventTimer) getEventTimeout(evtType TimerEventType) time.Duration {
	switch evtType {
	case EventProposeBlockTimeout:
		return makeProposalTimeout
	case EventPropose2ndBlockTimeout:
		return make2ndProposalTimeout
	case EventEndorseBlockTimeout:
		return endorseBlockTimeout
	case EventEndorseEmptyBlockTimeout:
		return endorseBlockTimeout
	case EventCommitBlockTimeout:
		return commitBlockTimeout
	case EventPeerHeartbeat:
		return peerHandshakeTimeout
	case EventProposalBackoff:
		rank := self.server.getProposerRank(self.server.GetCurrentBlockNo(), self.server.Index)
		if rank >= 0 {
			d := int64(rank+1) * int64(make2ndProposalTimeout) / 3
			return time.Duration(d)
		}
		return time.Duration(100 * time.Second)
	case EventRandomBackoff:
		d := (rand.Int63n(100) + 50) * int64(endorseBlockTimeout) / 100
		return time.Duration(d)
	case EventTxPool:
		return txPooltimeout
	case EventTxBlockTimeout:
		return zeroTxBlockTimeout
	}

	return 0
}

//
// internal helper, should call with lock held
//
func (self *EventTimer) startEventTimer(evtType TimerEventType, blockNum uint32) error {
	timers := self.eventTimers[evtType]
	if t, present := timers[blockNum]; present {
		t.Stop()
		delete(timers, blockNum)
		log.Infof("timer (type: %d) for %d got reset", evtType, blockNum)
	}

	timeout := self.getEventTimeout(evtType)
	if timeout == 0 {
		panic(fmt.Errorf("invalid timeout for event %d, blkNum %d", evtType, blockNum))
	}
	timers[blockNum] = time.AfterFunc(timeout, func() {
		self.C <- &TimerEvent{
			evtType:  evtType,
			blockNum: blockNum,
		}
	})

	return nil
}

//
// internal helper, should call with lock held
//
func (self *EventTimer) cancelEventTimer(evtType TimerEventType, blockNum uint32) error {
	timers := self.eventTimers[evtType]

	if t, present := timers[blockNum]; present {
		t.Stop()
		delete(timers, blockNum)
	}

	return nil
}

func (self *EventTimer) StartProposalTimer(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	log.Infof("server %d started proposal timer for blk %d", self.server.Index, blockNum)
	return self.startEventTimer(EventProposeBlockTimeout, blockNum)
}

func (self *EventTimer) CancelProposalTimer(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.cancelEventTimer(EventProposeBlockTimeout, blockNum)
}

func (self *EventTimer) StartEndorsingTimer(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	log.Infof("server %d started endorsing timer for blk %d", self.server.Index, blockNum)
	return self.startEventTimer(EventEndorseBlockTimeout, blockNum)
}

func (self *EventTimer) CancelEndorseMsgTimer(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.cancelEventTimer(EventEndorseBlockTimeout, blockNum)
}

func (self *EventTimer) StartEndorseEmptyBlockTimer(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	log.Infof("server %d started empty endorsing timer for blk %d", self.server.Index, blockNum)
	return self.startEventTimer(EventEndorseEmptyBlockTimeout, blockNum)
}

func (self *EventTimer) CancelEndorseEmptyBlockTimer(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.cancelEventTimer(EventEndorseEmptyBlockTimeout, blockNum)
}

func (self *EventTimer) StartCommitTimer(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	log.Infof("server %d started commit timer for blk %d", self.server.Index, blockNum)
	return self.startEventTimer(EventCommitBlockTimeout, blockNum)
}

func (self *EventTimer) CancelCommitMsgTimer(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.cancelEventTimer(EventCommitBlockTimeout, blockNum)
}

func (self *EventTimer) StartProposalBackoffTimer(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.startEventTimer(EventProposalBackoff, blockNum)
}

func (self *EventTimer) CancelProposalBackoffTimer(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.cancelEventTimer(EventProposalBackoff, blockNum)
}

func (self *EventTimer) StartBackoffTimer(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.startEventTimer(EventRandomBackoff, blockNum)
}

func (self *EventTimer) CancelBackoffTimer(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.cancelEventTimer(EventRandomBackoff, blockNum)
}

func (self *EventTimer) Start2ndProposalTimer(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.startEventTimer(EventPropose2ndBlockTimeout, blockNum)
}

func (self *EventTimer) Cancel2ndProposalTimer(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.cancelEventTimer(EventPropose2ndBlockTimeout, blockNum)
}

func (self *EventTimer) onBlockSealed(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	// clear event timers
	for i := 0; i < int(EventMax); i++ {
		if err := self.cancelEventTimer(TimerEventType(i), blockNum); err != nil {
			log.Errorf("server %d, failed to stop timer %d on sealing",
				self.server.Index, i)
		}
	}
}

func (self *EventTimer) StartTxBlockTimeout(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.startEventTimer(EventTxBlockTimeout, blockNum)
}

func (self *EventTimer) CancelTxBlockTimeout(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.cancelEventTimer(EventTxBlockTimeout, blockNum)
}

func (self *EventTimer) startPeerTicker(peerIdx uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if p, present := self.peerTickers[peerIdx]; present {
		p.Stop()
		log.Infof("ticker for %d got reset", peerIdx)
	}

	timeout := self.getEventTimeout(EventPeerHeartbeat)
	self.peerTickers[peerIdx] = time.AfterFunc(timeout, func() {
		self.C <- &TimerEvent{
			evtType:  EventPeerHeartbeat,
			blockNum: peerIdx,
		}
		self.peerTickers[peerIdx].Reset(timeout)
	})

	return nil
}

func (self *EventTimer) stopPeerTicker(peerIdx uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if p, present := self.peerTickers[peerIdx]; present {
		p.Stop()
		delete(self.peerTickers, peerIdx)
	}
	return nil
}

func (self *EventTimer) startTxTicker(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.startEventTimer(EventTxPool, blockNum)
}

func (self *EventTimer) stopTxTicker(blockNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return self.cancelEventTimer(EventTxPool, blockNum)
}

///////////////////////////////////////////////////////////
//
// timer queue
//
///////////////////////////////////////////////////////////

type TimerItem struct {
	due   time.Time
	evt   *TimerEvent
	index int
}

type TimerQueue []*TimerItem

func (tq TimerQueue) Len() int {
	return len(tq)
}

func (tq TimerQueue) Less(i, j int) bool {
	return tq[j].due.After(tq[i].due)
}

func (tq TimerQueue) Swap(i, j int) {
	tq[i], tq[j] = tq[j], tq[i]
	tq[i].index = i
	tq[j].index = j
}

func (tq *TimerQueue) Push(x interface{}) {
	item := x.(*TimerItem)
	item.index = len(*tq)
	*tq = append(*tq, item)
}

func (tq *TimerQueue) Pop() interface{} {
	old := *tq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*tq = old[0 : n-1]
	return item
}

func (tq *TimerQueue) update(item *TimerItem, due time.Time) {
	item.due = due
	heap.Fix(tq, item.index)
}
