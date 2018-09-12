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
	"sync"
	"sync/atomic"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
)

const (
	RTT_MAX_ESTIMATE   = 20 * time.Second // Maximum round-trip time to target for download requests
	QOS_TUNING_IMPACT  = 0.25             // Impact that a new tuning target has on the previous value
	TTL_SCALING        = 3                // Constant scaling factor for RTT -> TTL conversion
	TTL_LIMIT          = time.Minute      // Maximum TTL allowance to prevent reaching crazy timeouts
	QOS_TUNING_PEERS   = 5                // Number of peers to tune based on (best peers)
	QOS_CONFIDENCE_CAP = 10               // Number of peers above which not to modify RTT confidence
	RTT_MIN_CONFIDENCE = 0.1              // Worse confidence factor in our estimated RTT value
)

type blockTasksInfo struct {
	capacity  int                      // Maximum block task count for a peer
	tasks     map[common.Uint256]uint8 // Hash => State. 0: pend, 1: done
	delivered bool                     // All tasks is delivered
}

func newBlockTasksInfo(cap int) *blockTasksInfo {
	return &blockTasksInfo{
		capacity: cap,
		tasks:    make(map[common.Uint256]uint8),
	}
}

type QosServer struct {
	peers          *peerSet                   // Set of active peers from which download can proceed
	rttEstimate    uint64                     // Round trip time to target for download requests
	rttConfidence  uint64                     // Confidence in the estimated RTT (unit: millionths to allow atomic ops)
	quitCh         chan struct{}              // Quit channel to signal termination
	quitLock       sync.RWMutex               // Lock to prevent double closes
	peerBlockTasks map[uint64]*blockTasksInfo // PeerBlockTasks PeerID => BlockTaskInfo, used for calculate block throughput
	blockTaskLock  sync.RWMutex               // BlockTaskLock for peer
}

// New creates a new downloader to fetch hashes and blocks from remote peers.
func NewQosSever() *QosServer {
	server := &QosServer{
		peers:          newPeerSet(),
		rttEstimate:    uint64(RTT_MAX_ESTIMATE),
		rttConfidence:  uint64(1000000),
		peerBlockTasks: make(map[uint64]*blockTasksInfo, 0),
	}
	go server.qosTuner()
	return server
}

// Stop interrupts the downloader, canceling all pending operations.
// The downloader cannot be reused after calling Terminate.
func (s *QosServer) Stop() {
	// Close the termination channel (make sure double close is allowed)
	s.quitLock.Lock()
	select {
	case <-s.quitCh:
	default:
		close(s.quitCh)
	}
	s.quitLock.Unlock()
}

// RegisterPeer injects a new download peer into the set of block source to be
// used for fetching hashes and blocks from.
func (s *QosServer) RegisterPeer(id uint64) error {
	if err := s.peers.Register(id); err != nil {
		return err
	}
	s.qosReduceConfidence()
	return nil
}

// UnregisterPeer remove a peer from the known list, preventing any action from
// the specified peer. An effort is also made to return any pending fetches into
// the queue.
func (s *QosServer) UnregisterPeer(id uint64) error {
	// Unregister the peer from the active peer set and revoke any fetch tasks
	if err := s.peers.Unregister(id); err != nil {
		return err
	}
	if len(s.peers.AllPeerIds()) == 0 {
		log.Warnf("[p2p] no syncing peer")
	}
	return nil
}

// AllPeerIds get all peer id
func (s *QosServer) AllPeerIds() []uint64 {
	return s.peers.AllPeerIds()
}

// HeaderIdlePeers all header idle peer
func (s *QosServer) HeaderIdlePeers(orderBy IDLE_PEER_TYPE) []uint64 {
	peerConns, len := s.peers.HeaderIdlePeers(orderBy)
	pids := make([]uint64, 0, len)
	for _, peerConn := range peerConns {
		pids = append(pids, peerConn.id)
	}
	return pids
}

// HeaderIdlePeers all block idle peer
func (s *QosServer) BlockIdlePeers(orderBy IDLE_PEER_TYPE) []uint64 {
	peerConns, len := s.peers.BlockIdlePeers(orderBy)
	pids := make([]uint64, 0, len)
	for _, peerConn := range peerConns {
		pids = append(pids, peerConn.id)
	}
	return pids
}

// GetPeerRemainBlockCapacity get the remain sync block task capacity for a peer
// blockCapacity - pending jobs
func (s *QosServer) GetPeerBlockCapacity(id uint64) int {
	s.blockTaskLock.Lock()
	defer s.blockTaskLock.Unlock()
	p := s.peers.Peer(id)
	if p == nil {
		log.Tracef("[p2p] get peer block capacity failed: peer:%d not found", id)
		return 0
	}
	if _, ok := s.peerBlockTasks[id]; !ok {
		s.peerBlockTasks[id] = newBlockTasksInfo(p.BlockCapacity(s.requestRTT()))
	}
	return s.peerBlockTasks[id].capacity
}

// MarkPeerFetchingHeader mark the peer is fetching header
func (s *QosServer) MarkPeerFetchingHeader(id uint64) error {
	p := s.peers.Peer(id)
	if p == nil {
		return fmt.Errorf("[p2p] mark peer fetching header failed: peer:%d not found", id)
	}
	return p.MarkFetchingHeader()
}

// MarkPeerFetchingBlock mark the peer is fetching block
func (s *QosServer) MarkPeerFetchingBlock(id uint64) error {
	p := s.peers.Peer(id)
	if p == nil {
		return fmt.Errorf("[p2p] mark peer fetching block failed: peer:%d not found", id)
	}
	return p.MarkFetchingBlock()
}

// SetPeerHeaderIdle set peer headerIdle with delivered count
func (s *QosServer) SetPeerHeaderIdle(id uint64, count int) error {
	p := s.peers.Peer(id)
	if p == nil {
		return fmt.Errorf("[p2p] SetPeerHeaderIdle peer:%d not found", id)
	}
	return p.SetHeadersIdle(count)
}

// SetPeerBlockIdle set peer blockIdle with delivered count
func (s *QosServer) SetPeerBlockIdle(id uint64, count int) error {
	s.blockTaskLock.Lock()
	defer s.blockTaskLock.Unlock()
	p := s.peers.Peer(id)
	if p == nil {
		return fmt.Errorf("[p2p] SetPeerBlockIdle peer:%d not found", id)
	}
	err := p.SetBlocksIdle(count)
	if err != nil {
		return err
	}
	delete(s.peerBlockTasks, id)
	return nil
}

// IsPeerBlockIdle check peer blockIdle state
func (s *QosServer) IsPeerBlockIdle(id uint64) bool {
	p := s.peers.Peer(id)
	if p == nil {
		log.Tracef("[p2p] IsPeerBlockIdle failed: peer:%d not found", id)
		return false
	}
	return p.IsIdle()
}

// SetPeerHasDelieveredTasks set peer has beed delievered all tasks
// For concurrent request， a peer should setup this flag before setIdle
func (s *QosServer) SetPeerHasDelieveredTasks(id uint64) {
	s.blockTaskLock.Lock()
	defer s.blockTaskLock.Unlock()
	taskInfo := s.peerBlockTasks[id]
	if taskInfo == nil {
		log.Tracef("[p2p] UpdatePeerBlockTaskLength taskInfo not found :%d", id)
		return
	}
	taskInfo.delivered = true
}

// AddPeerBlockTask add peer blocktask, dunplicated task is not allowed
func (s *QosServer) AddPeerBlockTask(id uint64, hash common.Uint256) error {
	s.blockTaskLock.Lock()
	defer s.blockTaskLock.Unlock()
	taskInfo := s.peerBlockTasks[id]
	if taskInfo == nil {
		return fmt.Errorf("[p2p] AddPeerBlockTask failed, no taskInfo for:%d", id)
	}
	if _, ok := taskInfo.tasks[hash]; ok {
		return fmt.Errorf("[p2p] AddPeerBlockTask failed. task:%s exist for peer:%d", hash.ToHexString(), id)
	}
	taskInfo.tasks[hash] = 0
	return nil
}

// DelPeerBlockTask remove a task with hash
func (s *QosServer) DelPeerBlockTask(id uint64, hash common.Uint256) {
	s.blockTaskLock.Lock()
	defer s.blockTaskLock.Unlock()
	taskInfo := s.peerBlockTasks[id]
	if taskInfo == nil {
		return
	}
	delete(taskInfo.tasks, hash)
}

// FinishPeerBlockTask finish job, used for receive response and update the state of a job
func (s *QosServer) FinishPeerBlockTask(id uint64, hash common.Uint256) error {
	s.blockTaskLock.Lock()
	defer s.blockTaskLock.Unlock()
	taskInfo := s.peerBlockTasks[id]
	if taskInfo == nil {
		return fmt.Errorf("[p2p] finish peer blockTask failed. peer:%d taskInfo not found", id)
	}

	if _, ok := taskInfo.tasks[hash]; !ok {
		return fmt.Errorf("[p2p] finish peer blockTask failed. peer:%d task:%s not found", id, hash.ToHexString())
	}
	taskInfo.tasks[hash] = 1
	return nil
}

// GetPeerBlockTaskCount get blocktask total count of a peer
func (s *QosServer) GetPeerBlockTaskCount(id uint64) int {
	s.blockTaskLock.RLock()
	defer s.blockTaskLock.RUnlock()
	taskInfo := s.peerBlockTasks[id]
	if taskInfo == nil {
		return 0
	}
	return len(s.peerBlockTasks[id].tasks)
}

// IsPeerAllBlockTaskDone return if all blockTask is done and the amount of the finished task
func (s *QosServer) IsPeerAllBlockTaskDone(id uint64) (bool, int) {
	s.blockTaskLock.RLock()
	defer s.blockTaskLock.RUnlock()
	taskInfo := s.peerBlockTasks[id]
	if taskInfo == nil {
		// do something
		log.Tracef("[p2p] check blockTask for:%d not found", id)
		return false, 0
	}

	if !taskInfo.delivered {
		// waiting for all jobs delivered
		return false, 0
	}
	count := 0
	for _, state := range taskInfo.tasks {
		if state == 1 {
			count++
		}
	}
	return count == len(taskInfo.tasks), count
}

// SetPeerBlockIdle set peer height
func (s *QosServer) SetPeerHeight(id uint64, height uint32) error {
	p := s.peers.Peer(id)
	if p == nil {
		return fmt.Errorf("[p2p] set peer height failed: peer:%d not found", id)
	}
	p.SetHeight(height)
	return nil
}

// SetPeerBlockIdle set peer error count
func (s *QosServer) SetPeerErrorCount(id uint64, cnt uint32) error {
	p := s.peers.Peer(id)
	if p == nil {
		return fmt.Errorf("[p2p] set peer errorCount failed: peer:%d not found", id)
	}
	p.SetErrorCnt(cnt)
	return nil
}

// GetPeerErrorCount get peer error count
func (s *QosServer) GetPeerErrorCount(id uint64) (uint32, error) {
	p := s.peers.Peer(id)
	if p == nil {
		return 0, fmt.Errorf("[p2p] set peer errorCount failed: peer:%d not found", id)
	}
	return p.ErrorCount(), nil
}

// RequestTTL returns the current timeout allowance for a single download request
// to finish under.
func (s *QosServer) RequestTTL() time.Duration {
	var (
		rtt  = time.Duration(atomic.LoadUint64(&s.rttEstimate))
		conf = float64(atomic.LoadUint64(&s.rttConfidence)) / 1000000.0
	)
	ttl := time.Duration(TTL_SCALING) * time.Duration(float64(rtt)/conf)
	if ttl > TTL_LIMIT {
		ttl = TTL_LIMIT
	}
	return ttl
}

// requestRTT returns the current target round trip time for a download request
// to complete in.
//
// Note, the returned RTT is .9 of the actually estimated RTT. The reason is that
// the downloader tries to adapt queries to the RTT, so multiple RTT values can
// be adapted to, but smaller ones are preferred (stabler download stream).
func (s *QosServer) requestRTT() time.Duration {
	return time.Duration(atomic.LoadUint64(&s.rttEstimate)) * 9 / 10
}

// qosTuner is the quality of service tuning loop that occasionally gathers the
// peer latency statistics and updates the estimated request round trip time.
func (s *QosServer) qosTuner() {
	for {
		// Retrieve the current median RTT and integrate into the previoust target RTT
		rtt := time.Duration((1-QOS_TUNING_IMPACT)*float64(atomic.LoadUint64(&s.rttEstimate)) + QOS_TUNING_IMPACT*float64(s.peers.medianRTT()))
		atomic.StoreUint64(&s.rttEstimate, uint64(rtt))

		// A new RTT cycle passed, increase our confidence in the estimated RTT
		conf := atomic.LoadUint64(&s.rttConfidence)
		conf = conf + (1000000-conf)/2
		atomic.StoreUint64(&s.rttConfidence, conf)

		// Log the new QoS values and sleep until the next RTT
		log.Debug("[p2p] Recalculated downloader QoS values", "rtt", rtt, "confidence", float64(conf)/1000000.0, "ttl", s.RequestTTL())
		select {
		case <-s.quitCh:
			return
		case <-time.After(rtt):
		}
	}
}

// qosReduceConfidence is meant to be called when a new peer joins the downloader's
// peer set, needing to reduce the confidence we have in out QoS estimates.
func (s *QosServer) qosReduceConfidence() {
	// If we have a single peer, confidence is always 1
	peers := uint64(s.peers.Len())
	if peers == 0 {
		// Ensure peer connectivity races don't catch us off guard
		return
	}
	if peers == 1 {
		atomic.StoreUint64(&s.rttConfidence, 1000000)
		return
	}
	// If we have a ton of peers, don't drop confidence)
	if peers >= uint64(QOS_CONFIDENCE_CAP) {
		return
	}
	// Otherwise drop the confidence factor
	conf := atomic.LoadUint64(&s.rttConfidence) * (peers - 1) / peers
	if float64(conf)/1000000 < RTT_MIN_CONFIDENCE {
		conf = uint64(RTT_MIN_CONFIDENCE * 1000000)
	}
	atomic.StoreUint64(&s.rttConfidence, conf)

	rtt := time.Duration(atomic.LoadUint64(&s.rttEstimate))
	log.Debug("[p2p] Relaxed downloader QoS values", "rtt", rtt, "confidence", float64(conf)/1000000.0, "ttl", s.RequestTTL())
}
