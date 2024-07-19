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
	"math/rand"
	"sync"
	"sync/atomic"
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
	makeProposalTimeout    = int64(300 * time.Millisecond)
	make2ndProposalTimeout = int64(300 * time.Millisecond)
	endorseBlockTimeout    = int64(100 * time.Millisecond)
	commitBlockTimeout     = int64(200 * time.Millisecond)
	peerHandshakeTimeout   = int64(10 * time.Second)
	txPooltimeout          = int64(1 * time.Second)
	zeroTxBlockTimeout     = int64(10 * time.Second)
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
	//timerQueue TimerQueue

	// bft timers
	eventTimers map[TimerEventType]perBlockTimer

	// peer heartbeat tickers
	peerTickers map[uint32]*time.Timer
}

func NewEventTimer(server *Server) *EventTimer {
	timer := &EventTimer{
		server:      server,
		C:           make(chan *TimerEvent, 64),
		eventTimers: make(map[TimerEventType]perBlockTimer),
		peerTickers: make(map[uint32]*time.Timer),
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
}

func (self *EventTimer) getEventTimeout(evtType TimerEventType) time.Duration {
	switch evtType {
	case EventProposeBlockTimeout:
		return time.Duration(atomic.LoadInt64(&makeProposalTimeout))
	case EventPropose2ndBlockTimeout:
		return time.Duration(atomic.LoadInt64(&make2ndProposalTimeout))
	case EventEndorseBlockTimeout:
		return time.Duration(atomic.LoadInt64(&endorseBlockTimeout))
	case EventEndorseEmptyBlockTimeout:
		return time.Duration(atomic.LoadInt64(&endorseBlockTimeout))
	case EventCommitBlockTimeout:
		return time.Duration(atomic.LoadInt64(&commitBlockTimeout))
	case EventPeerHeartbeat:
		return time.Duration(atomic.LoadInt64(&peerHandshakeTimeout))
	case EventProposalBackoff:
		rank := self.server.getProposerRank(self.server.GetCurrentBlockNo(), self.server.Index)
		if rank >= 0 {
			d := int64(rank+1) * atomic.LoadInt64(&make2ndProposalTimeout) / 3
			return time.Duration(d)
		}
		return time.Duration(100 * time.Second)
	case EventRandomBackoff:
		d := (rand.Int63n(100) + 50) * atomic.LoadInt64(&endorseBlockTimeout) / 10
		return time.Duration(d)
	case EventTxPool:
		return time.Duration(txPooltimeout)
	case EventTxBlockTimeout:
		return time.Duration(atomic.LoadInt64(&zeroTxBlockTimeout))
	default:
		panic("unknown timer event type")
	}
}

// internal helper, should call with lock held
func (self *EventTimer) startEventTimer(evtType TimerEventType, blockNum uint32) {
	timers := self.eventTimers[evtType]
	if t, present := timers[blockNum]; present {
		t.Stop()
		delete(timers, blockNum)
		log.Infof("timer (type: %d) for %d got reset", evtType, blockNum)
	}

	timeout := self.getEventTimeout(evtType)
	if timeout == 0 {
		// never happen when config correctly
		timeout = time.Second
	}
	timers[blockNum] = time.AfterFunc(timeout, func() {
		self.C <- &TimerEvent{
			evtType:  evtType,
			blockNum: blockNum,
		}
	})
}

// internal helper, should call with lock held
func (self *EventTimer) cancelEventTimer(evtType TimerEventType, blockNum uint32) {
	timers := self.eventTimers[evtType]

	if t, present := timers[blockNum]; present {
		t.Stop()
		delete(timers, blockNum)
	}
}

func (self *EventTimer) StartProposalTimer(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	log.Infof("server %d started proposal timer for blk %d", self.server.Index, blockNum)
	self.startEventTimer(EventProposeBlockTimeout, blockNum)
}

func (self *EventTimer) CancelProposalTimer(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.cancelEventTimer(EventProposeBlockTimeout, blockNum)
}

func (self *EventTimer) StartEndorsingTimer(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	log.Infof("server %d started endorsing timer for blk %d", self.server.Index, blockNum)
	self.startEventTimer(EventEndorseBlockTimeout, blockNum)
}

func (self *EventTimer) CancelEndorseMsgTimer(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.cancelEventTimer(EventEndorseBlockTimeout, blockNum)
}

func (self *EventTimer) StartEndorseEmptyBlockTimer(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	log.Infof("server %d started empty endorsing timer for blk %d", self.server.Index, blockNum)
	self.startEventTimer(EventEndorseEmptyBlockTimeout, blockNum)
}

func (self *EventTimer) CancelEndorseEmptyBlockTimer(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.cancelEventTimer(EventEndorseEmptyBlockTimeout, blockNum)
}

func (self *EventTimer) StartCommitTimer(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	log.Infof("server %d started commit timer for blk %d", self.server.Index, blockNum)
	self.startEventTimer(EventCommitBlockTimeout, blockNum)
}

func (self *EventTimer) CancelCommitMsgTimer(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.cancelEventTimer(EventCommitBlockTimeout, blockNum)
}

func (self *EventTimer) StartProposalBackoffTimer(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.startEventTimer(EventProposalBackoff, blockNum)
}

func (self *EventTimer) CancelProposalBackoffTimer(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.cancelEventTimer(EventProposalBackoff, blockNum)
}

func (self *EventTimer) StartBackoffTimer(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.startEventTimer(EventRandomBackoff, blockNum)
}

func (self *EventTimer) CancelBackoffTimer(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.cancelEventTimer(EventRandomBackoff, blockNum)
}

func (self *EventTimer) Start2ndProposalTimer(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.startEventTimer(EventPropose2ndBlockTimeout, blockNum)
}

func (self *EventTimer) Cancel2ndProposalTimer(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.cancelEventTimer(EventPropose2ndBlockTimeout, blockNum)
}

func (self *EventTimer) onBlockSealed(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	// clear event timers
	for i := 0; i < int(EventMax); i++ {
		self.cancelEventTimer(TimerEventType(i), blockNum)
	}
}

func (self *EventTimer) StartTxBlockTimeout(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.startEventTimer(EventTxBlockTimeout, blockNum)
}

func (self *EventTimer) CancelTxBlockTimeout(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.cancelEventTimer(EventTxBlockTimeout, blockNum)
}

func (self *EventTimer) startPeerTicker(peerIdx uint32) {
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
}

func (self *EventTimer) stopPeerTicker(peerIdx uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if p, present := self.peerTickers[peerIdx]; present {
		p.Stop()
		delete(self.peerTickers, peerIdx)
	}
}

func (self *EventTimer) startTxPoolTicker(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.startEventTimer(EventTxPool, blockNum)
}

func (self *EventTimer) stopTxPoolTicker(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.cancelEventTimer(EventTxPool, blockNum)
}
