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
	"math"
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

func (self TimerEventType) String() string {
	switch self {
	case EventProposeBlockTimeout:
		return "ProposeBlockTimeout"
	case EventProposalBackoff:
		return "EventProposalBackoff"
	case EventRandomBackoff:
		return "EventRandomBackoff"
	case EventPropose2ndBlockTimeout:
		return "EventPropose2ndBlockTimeout"
	case EventEndorseBlockTimeout:
		return "EventEndorseBlockTimeout"
	case EventEndorseEmptyBlockTimeout:
		return "EventEndorseEmptyBlockTimeout"
	case EventCommitBlockTimeout:
		return "EventCommitBlockTimeout"
	case EventPeerHeartbeat:
		return "EventPeerHeartbeat"
	case EventTxPool:
		return "EventTxPool"
	case EventTxBlockTimeout:
		return "EventTxBlockTimeout"
	default:
		panic("unknown timer type")
	}
}

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
}

type EventTimer struct {
	lock        sync.Mutex
	server      *Server
	C           chan *TimerEvent
	eventTimers map[TimerEventType]map[uint32]*time.Timer // bft timers: type -> block -> timer
	peerTickers *time.Timer                               // peer heartbeat tickers
}

func NewEventTimer(server *Server) *EventTimer {
	timer := &EventTimer{
		server:      server,
		C:           make(chan *TimerEvent, 64),
		eventTimers: make(map[TimerEventType]map[uint32]*time.Timer),
	}

	for i := 0; i < int(EventMax); i++ {
		timer.eventTimers[TimerEventType(i)] = make(map[uint32]*time.Timer)
	}

	return timer
}

func (self *EventTimer) stop() {
	self.lock.Lock()
	defer self.lock.Unlock()

	// clear timers by event timer
	for i := 0; i < int(EventMax); i++ {
		for _, t := range self.eventTimers[TimerEventType(i)] {
			t.Stop()
		}
		self.eventTimers[TimerEventType(i)] = make(map[uint32]*time.Timer)
	}

	self.stopPeerTicker()
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
		rank := self.server.GetVbftContext().GetProposerRank(self.server.Index)
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
		panic("unknown event type")
	}
}

func (self *EventTimer) StartEventTimer(evtType TimerEventType, blockNum uint32) {
	log.Infof("server %d started %s timer for blk %d", self.server.Index, evtType, blockNum)
	self.lock.Lock()
	defer self.lock.Unlock()

	timers := self.eventTimers[evtType]
	if t, present := timers[blockNum]; present {
		t.Stop()
		delete(timers, blockNum)
		log.Infof("timer (type: %d) for %d got reset", evtType, blockNum)
	}

	timeout := self.getEventTimeout(evtType)
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

func (self *EventTimer) CancelEventTimer(evtType TimerEventType, blockNum uint32) {
	log.Infof("server %d canceled %s timer for blk %d", self.server.Index, evtType, blockNum)
	self.lock.Lock()
	defer self.lock.Unlock()

	self.cancelEventTimer(evtType, blockNum)
}

func (self *EventTimer) OnBlockSealed(blockNum uint32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for i := 0; i < int(EventMax); i++ {
		self.cancelEventTimer(TimerEventType(i), blockNum)
	}
}

func (self *EventTimer) startPeerTicker() {
	self.lock.Lock()
	defer self.lock.Unlock()

	if p := self.peerTickers; p != nil {
		p.Stop()
		log.Infof("peer ticker got reset")
	}

	timeout := self.getEventTimeout(EventPeerHeartbeat)
	self.peerTickers = time.AfterFunc(timeout, func() {
		self.server.heartbeat(math.MaxUint32)
		self.peerTickers.Reset(timeout)
	})
}

func (self *EventTimer) stopPeerTicker() {
	self.lock.Lock()
	defer self.lock.Unlock()

	if p := self.peerTickers; p != nil {
		p.Stop()
		log.Infof("peer ticker got reset")
	}
	self.peerTickers = nil
}
