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

package blocksync

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	MEASUREMENT_IMPACT = 0.1             // The impact a single measurement has on a peer's final throughput value.
	RTT_MIN_ESTIMATE   = 2 * time.Second // Minimum round-trip time to target for download requests
	MAX_BLOCK_FETCH    = 10              // MaxBlockFetch the maximum fetch block request of a peer
)

// Idle peers type
type IDLE_PEER_TYPE uint8

const (
	IDLE_PEER_THROUGHPUT_FIRST IDLE_PEER_TYPE = iota // Sort the idle peers by throughput
	IDLE_PEER_HEIGHT_FIRST                           // Sort the idle peers by height
)

// peerConnection represents an active peer from which hashes and blocks are retrieved.
type peerConnection struct {
	id               uint64        // Unique identifier of the peer
	headerIdle       int32         // Current header activity state of the peer (idle = 0, active = 1)
	blockIdle        int32         // Current block activity state of the peer (idle = 0, active = 1)
	headerThroughput float64       // Number of headers measured to be retrievable per second
	blockThroughput  float64       // Number of blocks (bodies) measured to be retrievable per second
	height           uint32        // Block Height of peer
	errorCnt         uint32        // Error count of bad behaviour
	rtt              time.Duration // Request round trip time to track responsiveness (QoS)
	headerStarted    time.Time     // Time instance when the last header fetch was started
	blockStarted     time.Time     // Time instance when the last block (body) fetch was started
	lock             sync.RWMutex
}

// newPeerConnection creates a new downloader peer.
func newPeerConnection(id uint64) *peerConnection {
	return &peerConnection{
		id: id,
	}
}

// Reset clears the internal state of a peer entity.
func (p *peerConnection) Reset() {
	p.lock.Lock()
	defer p.lock.Unlock()

	atomic.StoreInt32(&p.headerIdle, 0)
	atomic.StoreInt32(&p.blockIdle, 0)

	p.headerThroughput = 0
	p.blockThroughput = 0
	p.height = 0
	p.errorCnt = 0
}

// BlockCapacity retrieves the peers block download allowance based on its
// previously discovered throughput.
func (p *peerConnection) BlockCapacity(targetRTT time.Duration) int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return int(math.Min(1+math.Max(1, p.blockThroughput*float64(targetRTT)/float64(time.Second)), float64(MAX_BLOCK_FETCH)))
}

// SetHeadersIdle sets the peer to idle, allowing it to execute new header retrieval
// requests. Its estimated header retrieval throughput is updated with that measured
// just now.
func (p *peerConnection) SetHeadersIdle(delivered int) error {
	if atomic.LoadInt32(&p.headerIdle) == 0 {
		return fmt.Errorf("[p2p] SetHeadersIdle for a idle peer:%d", p.id)
	}
	p.setIdle(p.headerStarted, delivered, &p.headerThroughput, &p.headerIdle)

	return nil
}

// SetBlocksIdle sets the peer to idle, allowing it to execute new block retrieval
// requests. Its estimated block retrieval throughput is updated with that measured
// just now.
func (p *peerConnection) SetBlocksIdle(delivered int) error {
	if atomic.LoadInt32(&p.blockIdle) == 0 {
		return fmt.Errorf("[p2p] SetBlocksIdle for a idle peer:%d", p.id)
	}
	p.setIdle(p.blockStarted, delivered, &p.blockThroughput, &p.blockIdle)
	return nil
}

// MarkFetchingHeader mark the peer is active of fetching header
func (p *peerConnection) MarkFetchingHeader() error {
	if !atomic.CompareAndSwapInt32(&p.headerIdle, 0, 1) {
		return fmt.Errorf("[p2p] MarkFetchingHeader the peer is already fetching")
	}
	p.headerStarted = time.Now()
	return nil
}

// MarkFetchingBlock mark the peer is active of fetching block
func (p *peerConnection) MarkFetchingBlock() error {
	if !atomic.CompareAndSwapInt32(&p.blockIdle, 0, 1) {
		return fmt.Errorf("[p2p] MarkFetchingBlock the peer is already fetching")
	}
	p.blockStarted = time.Now()
	return nil
}

// SetHeight set peer current blockheight
func (p *peerConnection) SetHeight(height uint32) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.height = height
}

// SetHeight set peer error bad behaviour count
func (p *peerConnection) SetErrorCnt(errCnt uint32) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.errorCnt = errCnt
}

// ErrorCount return peer error bad behaviour count
func (p *peerConnection) ErrorCount() uint32 {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.errorCnt
}

// IsIdle is peer idle
func (p *peerConnection) IsIdle() bool {
	return atomic.LoadInt32(&p.blockIdle) == 0
}

// setIdle sets the peer to idle, allowing it to execute new retrieval requests.
// Its estimated retrieval throughput is updated with that measured just now.
func (p *peerConnection) setIdle(started time.Time, delivered int, throughput *float64, idle *int32) {
	// Irrelevant of the scaling, make sure the peer ends up idle
	defer atomic.StoreInt32(idle, 0)

	p.lock.Lock()
	defer p.lock.Unlock()

	// If nothing was delivered (hard timeout / unavailable data), reduce throughput to minimum
	if delivered == 0 {
		*throughput = 0
		return
	}
	// Otherwise update the throughput with a new measurement
	elapsed := time.Since(started) + 1 // +1 (ns) to ensure non-zero divisor
	measured := float64(delivered) / (float64(elapsed) / float64(time.Second))

	*throughput = (1-MEASUREMENT_IMPACT)*(*throughput) + MEASUREMENT_IMPACT*measured
	p.rtt = time.Duration((1-MEASUREMENT_IMPACT)*float64(p.rtt) + MEASUREMENT_IMPACT*float64(elapsed))
}

// peerSet represents the collection of active peer participating in the chain
// download procedure.
type peerSet struct {
	peers map[uint64]*peerConnection
	lock  sync.RWMutex
}

// newPeerSet creates a new peer set top track the active download sources.
func newPeerSet() *peerSet {
	return &peerSet{
		peers: make(map[uint64]*peerConnection),
	}
}

// Reset iterates over the current peer set, and resets each of the known peers
// to prepare for a next batch of block retrieval.
func (ps *peerSet) Reset() {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	for _, peer := range ps.peers {
		peer.Reset()
	}
}

// Register injects a new peer into the working set, or returns an error if the
// peer is already known.
//
// The method also sets the starting throughput values of the new peer to the
// average of all existing peers, to give it a realistic chance of being used
// for data retrievals.
func (ps *peerSet) Register(id uint64) error {
	// Retrieve the current median RTT as a sane default
	p := newPeerConnection(id)
	p.rtt = ps.medianRTT()

	// Register the new peer with some meaningful defaults
	ps.lock.Lock()
	if _, ok := ps.peers[p.id]; ok {
		ps.lock.Unlock()
		return fmt.Errorf("[p2p] peer already register")
	}
	if len(ps.peers) > 0 {
		p.headerThroughput, p.blockThroughput = 0, 0

		for _, peer := range ps.peers {
			peer.lock.RLock()
			p.headerThroughput += peer.headerThroughput
			p.blockThroughput += peer.blockThroughput
			peer.lock.RUnlock()
		}
		p.headerThroughput /= float64(len(ps.peers))
		p.blockThroughput /= float64(len(ps.peers))
	}
	ps.peers[p.id] = p
	ps.lock.Unlock()

	return nil
}

// Unregister removes a remote peer from the active set, disabling any further
// actions to/from that particular entity.
func (ps *peerSet) Unregister(id uint64) error {
	ps.lock.Lock()
	_, ok := ps.peers[id]
	if !ok {
		defer ps.lock.Unlock()
		return fmt.Errorf("[p2p]peer not register")
	}
	delete(ps.peers, id)
	ps.lock.Unlock()

	return nil
}

// Peer retrieves the registered peer with the given id.
func (ps *peerSet) Peer(id uint64) *peerConnection {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return ps.peers[id]
}

// Len returns if the current number of peers in the set.
func (ps *peerSet) Len() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return len(ps.peers)
}

// AllPeers retrieves a flat list of all the peers within the set.
func (ps *peerSet) AllPeers() []*peerConnection {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peerConnection, 0, len(ps.peers))
	for _, p := range ps.peers {
		list = append(list, p)
	}
	return list
}

// AllPeers retrieves a flat list of all the peer ids within the set.
func (ps *peerSet) AllPeerIds() []uint64 {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]uint64, 0, len(ps.peers))
	for _, p := range ps.peers {
		list = append(list, p.id)
	}
	return list
}

// HeaderIdlePeers retrieves a flat list of all the currently header-idle peers
// within the active peer set, ordered by their reputation.
func (ps *peerSet) HeaderIdlePeers(orderdBy IDLE_PEER_TYPE) ([]*peerConnection, int) {
	idle := func(p *peerConnection) bool {
		return atomic.LoadInt32(&p.headerIdle) == 0
	}
	less := func(p1 *peerConnection, p2 *peerConnection) bool {
		p1.lock.RLock()
		defer p1.lock.RUnlock()
		p2.lock.RLock()
		defer p2.lock.RUnlock()
		switch orderdBy {
		case IDLE_PEER_THROUGHPUT_FIRST:
			return p1.headerThroughput < p2.headerThroughput
		case IDLE_PEER_HEIGHT_FIRST:
			return p1.height < p2.height
		default:
			return p1.headerThroughput < p2.headerThroughput
		}
	}

	return ps.idlePeers(idle, less)
}

// BlockIdlePeers retrieves a flat list of all the currently body-idle peers within
// the active peer set, ordered by their reputation.
func (ps *peerSet) BlockIdlePeers(orderdBy IDLE_PEER_TYPE) ([]*peerConnection, int) {
	idle := func(p *peerConnection) bool {
		return atomic.LoadInt32(&p.blockIdle) == 0
	}
	less := func(p1 *peerConnection, p2 *peerConnection) bool {
		p1.lock.RLock()
		defer p1.lock.RUnlock()
		p2.lock.RLock()
		defer p2.lock.RUnlock()
		switch orderdBy {
		case IDLE_PEER_THROUGHPUT_FIRST:
			return p1.blockThroughput < p2.blockThroughput
		case IDLE_PEER_HEIGHT_FIRST:
			return p1.height < p2.height
		default:
			return p1.blockThroughput < p2.blockThroughput
		}
	}
	return ps.idlePeers(idle, less)
}

// idlePeers retrieves a flat list of all currently idle peers satisfying the
// protocol version constraints, using the provided function to check idleness.
// The resulting set of peers are sorted by their measure throughput.
func (ps *peerSet) idlePeers(idleCheck func(*peerConnection) bool, less func(*peerConnection, *peerConnection) bool) ([]*peerConnection, int) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	total := len(ps.peers)
	idle := make([]*peerConnection, 0, total)
	for _, p := range ps.peers {
		if idleCheck(p) {
			idle = append(idle, p)
		}
	}
	for i := 0; i < len(idle); i++ {
		for j := i + 1; j < len(idle); j++ {
			if less(idle[i], idle[j]) {
				idle[i], idle[j] = idle[j], idle[i]
			}
		}
	}
	return idle, total
}

// medianRTT returns the median RTT of the peerset, considering only the tuning
// peers if there are more peers available.
func (ps *peerSet) medianRTT() time.Duration {
	// Gather all the currently measured round trip times
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	rtts := make([]float64, 0, len(ps.peers))
	for _, p := range ps.peers {
		p.lock.RLock()
		rtts = append(rtts, float64(p.rtt))
		p.lock.RUnlock()
	}
	sort.Float64s(rtts)

	median := RTT_MIN_ESTIMATE
	if QOS_TUNING_PEERS <= len(rtts) {
		median = time.Duration(rtts[QOS_TUNING_PEERS/2]) // Median of our tuning peers
	} else if len(rtts) > 0 {
		median = time.Duration(rtts[len(rtts)/2]) // Median of our connected peers (maintain even like this some baseline qos)
	}
	// Restrict the RTT into some QoS defaults, irrelevant of true RTT
	if median < RTT_MIN_ESTIMATE {
		median = RTT_MIN_ESTIMATE
	}
	if median > RTT_MIN_ESTIMATE {
		median = RTT_MIN_ESTIMATE
	}
	return median
}
