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

package block_sync

import (
	"math"
	"sort"
	"sync"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	p2pComm "github.com/ontio/ontology/p2pserver/common"
	msgpack "github.com/ontio/ontology/p2pserver/message/msg_pack"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/peer"
)

const (
	SYNC_MAX_HEADER_FORWARD_SIZE = 5000       //keep CurrentHeaderHeight - CurrentBlockHeight <= SYNC_MAX_HEADER_FORWARD_SIZE
	SYNC_MAX_FLIGHT_HEADER_SIZE  = 1          //Number of headers on flight
	SYNC_MAX_FLIGHT_BLOCK_SIZE   = 50         //Number of blocks on flight
	SYNC_MAX_BLOCK_CACHE_SIZE    = 500        //Cache size of block wait to commit to ledger
	SYNC_HEADER_REQUEST_TIMEOUT  = 2          //s, Request header timeout time. If header haven't receive after SYNC_HEADER_REQUEST_TIMEOUT second, retry
	SYNC_BLOCK_REQUEST_TIMEOUT   = 2          //s, Request block timeout time. If block haven't received after SYNC_BLOCK_REQUEST_TIMEOUT second, retry
	SYNC_NEXT_BLOCK_TIMES        = 3          //Request times of next height block
	SYNC_NEXT_BLOCKS_HEIGHT      = 2          //for current block height plus next
	SYNC_NODE_RECORD_SPEED_CNT   = 3          //Record speed count for accuracy
	SYNC_NODE_RECORD_TIME_CNT    = 3          //Record request time  for accuracy
	SYNC_NODE_SPEED_INIT         = 100 * 1024 //Init a big speed (100MB/s) for every node in first round
	SYNC_MAX_ERROR_RESP_TIMES    = 5          //Max error headers/blocks response times, if reaches, delete it
	SYNC_MAX_HEIGHT_OFFSET       = 5          //Offset of the max height and current height
)

//NodeWeight record some params of node, using for sort
type NodeWeight struct {
	id           p2pComm.PeerId //NodeID
	speed        []float32      //Record node request-response speed, using for calc the avg speed, unit kB/s
	timeoutCnt   int            //Node response timeout count
	errorRespCnt int            //Node response error data count
	reqTime      []int64        //Record request time, using for calc the avg req time interval, unit millisecond
}

//NewNodeWeight new a nodeweight
func NewNodeWeight(id p2pComm.PeerId) *NodeWeight {
	s := make([]float32, 0, SYNC_NODE_RECORD_SPEED_CNT)
	for i := 0; i < SYNC_NODE_RECORD_SPEED_CNT; i++ {
		s = append(s, float32(SYNC_NODE_SPEED_INIT))
	}
	r := make([]int64, 0, SYNC_NODE_RECORD_TIME_CNT)
	now := time.Now().UnixNano() / int64(time.Millisecond)
	for i := 0; i < SYNC_NODE_RECORD_TIME_CNT; i++ {
		r = append(r, now)
	}
	return &NodeWeight{
		id:           id,
		speed:        s,
		timeoutCnt:   0,
		errorRespCnt: 0,
		reqTime:      r,
	}
}

//AddTimeoutCnt incre timeout count
func (this *NodeWeight) AddTimeoutCnt() {
	this.timeoutCnt++
}

//AddErrorRespCnt incre receive error header/block count
func (this *NodeWeight) AddErrorRespCnt() {
	this.errorRespCnt++
}

//GetErrorRespCnt get the error response count
func (this *NodeWeight) GetErrorRespCnt() int {
	return this.errorRespCnt
}

//AppendNewReqTime append new request time
func (this *NodeWeight) AppendNewReqtime() {
	copy(this.reqTime[0:SYNC_NODE_RECORD_TIME_CNT-1], this.reqTime[1:])
	this.reqTime[SYNC_NODE_RECORD_TIME_CNT-1] = time.Now().UnixNano() / int64(time.Millisecond)
}

//addNewSpeed apend the new speed to tail, remove the oldest one
func (this *NodeWeight) AppendNewSpeed(s float32) {
	copy(this.speed[0:SYNC_NODE_RECORD_SPEED_CNT-1], this.speed[1:])
	this.speed[SYNC_NODE_RECORD_SPEED_CNT-1] = s
}

//Weight calculate node's weight for sort. Highest weight node will be accessed first for next request.
func (this *NodeWeight) Weight() float32 {
	avgSpeed := float32(0.0)
	for _, s := range this.speed {
		avgSpeed += s
	}
	avgSpeed = avgSpeed / float32(len(this.speed))

	avgInterval := float32(0.0)
	now := time.Now().UnixNano() / int64(time.Millisecond)
	for _, t := range this.reqTime {
		avgInterval += float32(now - t)
	}
	avgInterval = avgInterval / float32(len(this.reqTime))
	w := avgSpeed + avgInterval
	return w
}

//NodeWeights implement sorting
type NodeWeights []*NodeWeight

func (nws NodeWeights) Len() int {
	return len(nws)
}

func (nws NodeWeights) Swap(i, j int) {
	nws[i], nws[j] = nws[j], nws[i]
}
func (nws NodeWeights) Less(i, j int) bool {
	ni := nws[i]
	nj := nws[j]
	return ni.Weight() < nj.Weight() && ni.errorRespCnt >= nj.errorRespCnt && ni.timeoutCnt >= nj.timeoutCnt
}

//SyncFlightInfo record the info of fight object(header or block)
type SyncFlightInfo struct {
	Height      uint32                 //BlockHeight of HeaderHeight
	nodeId      p2pComm.PeerId         //The current node to send msg
	startTime   time.Time              //Request start time
	failedNodes map[p2pComm.PeerId]int //Map nodeId => timeout times
	totalFailed int                    //Total timeout times
	lock        sync.RWMutex
}

//NewSyncFlightInfo return a new SyncFlightInfo instance
func NewSyncFlightInfo(height uint32, nodeId p2pComm.PeerId) *SyncFlightInfo {
	return &SyncFlightInfo{
		Height:      height,
		nodeId:      nodeId,
		startTime:   time.Now(),
		failedNodes: make(map[p2pComm.PeerId]int),
	}
}

//GetNodeId return current node id for sending msg
func (this *SyncFlightInfo) GetNodeId() p2pComm.PeerId {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.nodeId
}

//SetNodeId set a new node id
func (this *SyncFlightInfo) SetNodeId(nodeId p2pComm.PeerId) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.nodeId = nodeId
}

//MarkFailedNode mark node failed, after request timeout
func (this *SyncFlightInfo) MarkFailedNode() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.failedNodes[this.nodeId] += 1
	this.totalFailed++
}

//GetFailedTimes return failed times of a node
func (this *SyncFlightInfo) GetFailedTimes(nodeId p2pComm.PeerId) int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	times, ok := this.failedNodes[nodeId]
	if !ok {
		return 0
	}
	return times
}

//GetTotalFailedTimes return the total failed times of request
func (this *SyncFlightInfo) GetTotalFailedTimes() int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.totalFailed
}

//ResetStartTime
func (this *SyncFlightInfo) ResetStartTime() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.startTime = time.Now()
}

//GetStartTime return the start time of request
func (this *SyncFlightInfo) GetStartTime() time.Time {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.startTime
}

//BlockInfo is used for saving block information in cache
type BlockInfo struct {
	nodeID        p2pComm.PeerId
	block         *types.Block
	crossChainMsg *types.CrossChainMsg
	merkleRoot    common.Uint256
}

//BlockSyncMgr is the manager class to deal with block sync
type BlockSyncMgr struct {
	flightBlocks   map[common.Uint256][]*SyncFlightInfo //Map BlockHash => []SyncFlightInfo, using for manager all of those block flights
	flightHeaders  map[uint32]*SyncFlightInfo           //Map HeaderHeight => SyncFlightInfo, using for manager all of those header flights
	blocksCache    *BlockCache                          //Map BlockHash => BlockInfo, using for cache the blocks receive from net, and waiting for commit to ledger
	server         p2p.P2P                              //Pointer to the local node
	syncBlockLock  bool                                 //Help to avoid send block sync request duplicate
	syncHeaderLock bool                                 //Help to avoid send header sync request duplicate
	saveBlockLock  bool                                 //Help to avoid saving block concurrently
	exitCh         chan interface{}                     //ExitCh to receive exit signal
	ledger         *ledger.Ledger                       //ledger
	lock           sync.RWMutex                         //lock
	nodeWeights    map[p2pComm.PeerId]*NodeWeight       //Map NodeID => NodeStatus, using for getNextNode
}

//NewBlockSyncMgr return a BlockSyncMgr instance
func NewBlockSyncMgr(server p2p.P2P, ld *ledger.Ledger) *BlockSyncMgr {
	return &BlockSyncMgr{
		flightBlocks:  make(map[common.Uint256][]*SyncFlightInfo),
		flightHeaders: make(map[uint32]*SyncFlightInfo),
		blocksCache:   NewBlockCache(),
		server:        server,
		ledger:        ld,
		exitCh:        make(chan interface{}, 1),
		nodeWeights:   make(map[p2pComm.PeerId]*NodeWeight),
	}
}

type BlockCache struct {
	emptyBlockAmount int
	blocksCache      map[uint32]*BlockInfo //Map BlockHeight => BlockInfo, using for cache the blocks receive from net, and waiting for commit to ledger
}

func NewBlockCache() *BlockCache {
	return &BlockCache{
		emptyBlockAmount: 0,
		blocksCache:      make(map[uint32]*BlockInfo),
	}
}

func (this *BlockCache) addBlock(nodeID p2pComm.PeerId, block *types.Block, ccMsg *types.CrossChainMsg,
	merkleRoot common.Uint256) bool {
	this.delBlockLocked(block.Header.Height)
	blockInfo := &BlockInfo{
		nodeID:        nodeID,
		block:         block,
		crossChainMsg: ccMsg,
		merkleRoot:    merkleRoot,
	}
	this.blocksCache[block.Header.Height] = blockInfo
	if block.Header.TransactionsRoot == common.UINT256_EMPTY {
		this.emptyBlockAmount += 1
	}
	return true
}

func (this *BlockSyncMgr) clearBlocks(curBlockHeight uint32) {
	this.lock.Lock()
	this.blocksCache.clearBlocks(curBlockHeight)
	this.lock.Unlock()
}

func (this *BlockCache) clearBlocks(curBlockHeight uint32) {
	for height := range this.blocksCache {
		if height < curBlockHeight {
			this.delBlockLocked(height)
		}
	}
}

func (this *BlockCache) getBlock(blockHeight uint32) (p2pComm.PeerId, *types.Block, *types.CrossChainMsg,
	common.Uint256) {
	blockInfo, ok := this.blocksCache[blockHeight]
	if !ok {
		return p2pComm.PeerId{}, nil, nil, common.UINT256_EMPTY
	}
	return blockInfo.nodeID, blockInfo.block, blockInfo.crossChainMsg, blockInfo.merkleRoot
}

func (this *BlockCache) delBlockLocked(blockHeight uint32) {
	blockInfo, ok := this.blocksCache[blockHeight]
	if ok {
		if blockInfo.block.Header.TransactionsRoot == common.UINT256_EMPTY {
			this.emptyBlockAmount -= 1
		}
	}
	delete(this.blocksCache, blockHeight)
}

func (this *BlockCache) isInBlockCache(blockHeight uint32) bool {
	_, ok := this.blocksCache[blockHeight]
	return ok
}

func (this *BlockCache) getNonEmptyBlockCount() int {
	return len(this.blocksCache) - this.emptyBlockAmount
}
func (this *BlockSyncMgr) getNonEmptyBlockCount() int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.blocksCache.getNonEmptyBlockCount()
}

//Start to sync
func (this *BlockSyncMgr) Start() {
	go this.sync()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-this.exitCh:
			return
		case <-ticker.C:
			go this.checkTimeout()
			go this.sync()
			go this.saveBlock()
		}
	}
}

func (this *BlockSyncMgr) checkTimeout() {
	now := time.Now()
	headerTimeoutFlights := make(map[uint32]*SyncFlightInfo)
	blockTimeoutFlights := make(map[common.Uint256][]*SyncFlightInfo)
	this.lock.RLock()
	for height, flightInfo := range this.flightHeaders {
		if int(now.Sub(flightInfo.startTime).Seconds()) >= SYNC_HEADER_REQUEST_TIMEOUT {
			headerTimeoutFlights[height] = flightInfo
		}
	}
	for blockHash, flightInfos := range this.flightBlocks {
		for _, flightInfo := range flightInfos {
			if int(now.Sub(flightInfo.startTime).Seconds()) >= SYNC_BLOCK_REQUEST_TIMEOUT {
				blockTimeoutFlights[blockHash] = append(blockTimeoutFlights[blockHash], flightInfo)
			}
		}
	}
	this.lock.RUnlock()

	curHeaderHeight := this.ledger.GetCurrentHeaderHeight()
	curBlockHeight := this.ledger.GetCurrentBlockHeight()

	for height, flightInfo := range headerTimeoutFlights {
		this.addTimeoutCnt(flightInfo.GetNodeId())
		if height <= curHeaderHeight {
			this.delFlightHeader(height)
			continue
		}
		flightInfo.ResetStartTime()
		flightInfo.MarkFailedNode()
		log.Tracef("[block-sync] checkTimeout sync headers from id:%d :%d timeout after:%d s Times:%d", flightInfo.GetNodeId(), height, SYNC_HEADER_REQUEST_TIMEOUT, flightInfo.GetTotalFailedTimes())
		reqNode := this.getNodeWithMinFailedTimes(flightInfo, curBlockHeight)
		if reqNode == nil {
			break
		}
		flightInfo.SetNodeId(reqNode.GetID())

		headerHash := this.ledger.GetCurrentHeaderHash()
		msg := msgpack.NewHeadersReq(headerHash)
		err := this.server.Send(reqNode, msg)
		if err != nil {
			log.Warn("[block-sync] checkTimeout failed to send a new headersReq:s", err)
		} else {
			this.appendReqTime(reqNode.GetID())
		}
	}
	for blockHash, flightInfos := range blockTimeoutFlights {
		for _, flightInfo := range flightInfos {
			this.addTimeoutCnt(flightInfo.GetNodeId())
			if flightInfo.Height <= curBlockHeight {
				this.delFlightBlock(blockHash)
				continue
			}
			flightInfo.ResetStartTime()
			flightInfo.MarkFailedNode()
			log.Tracef("[block-sync] checkTimeout sync height:%d block:0x%x timeout after:%d s times:%d", flightInfo.Height, blockHash, SYNC_BLOCK_REQUEST_TIMEOUT, flightInfo.GetTotalFailedTimes())
			reqNode := this.getNodeWithMinFailedTimes(flightInfo, curBlockHeight)
			if reqNode == nil {
				break
			}
			flightInfo.SetNodeId(reqNode.GetID())

			msg := msgpack.NewBlkDataReq(blockHash)
			err := this.server.Send(reqNode, msg)
			if err != nil {
				log.Warnf("[block-sync] checkTimeout reqNode ID:%d Send error:%s", reqNode.GetID(), err)
				continue
			} else {
				this.appendReqTime(reqNode.GetID())
			}
		}
	}
}

func (this *BlockSyncMgr) sync() {
	this.syncHeader()
	this.syncBlock()
}

func (this *BlockSyncMgr) syncHeader() {
	//if !this.server.reachMinConnection() {
	//	return
	//}
	if this.tryGetSyncHeaderLock() {
		return
	}
	defer this.releaseSyncHeaderLock()

	if this.getFlightHeaderCount() >= SYNC_MAX_FLIGHT_HEADER_SIZE {
		return
	}
	curBlockHeight := this.ledger.GetCurrentBlockHeight()

	curHeaderHeight := this.ledger.GetCurrentHeaderHeight()
	//Waiting for block catch up header
	if curHeaderHeight-curBlockHeight >= SYNC_MAX_HEADER_FORWARD_SIZE {
		return
	}
	NextHeaderId := curHeaderHeight + 1
	reqNode := this.getNextNode(NextHeaderId)
	if reqNode == nil {
		return
	}
	this.addFlightHeader(reqNode.GetID(), NextHeaderId)

	headerHash := this.ledger.GetCurrentHeaderHash()
	msg := msgpack.NewHeadersReq(headerHash)
	err := this.server.Send(reqNode, msg)
	if err != nil {
		log.Warn("[block-sync] syncHeader failed to send a new headersReq")
	} else {
		this.appendReqTime(reqNode.GetID())
	}

	log.Infof("Header sync request height:%d", NextHeaderId)
}

func (this *BlockSyncMgr) syncBlock() {
	if this.tryGetSyncBlockLock() {
		return
	}
	defer this.releaseSyncBlockLock()

	availCount := SYNC_MAX_FLIGHT_BLOCK_SIZE - this.getFlightBlockCount()
	if availCount <= 0 {
		return
	}
	curBlockHeight := this.ledger.GetCurrentBlockHeight()
	curHeaderHeight := this.ledger.GetCurrentHeaderHeight()
	count := int(curHeaderHeight - curBlockHeight)
	if count <= 0 {
		return
	}
	if count > availCount {
		count = availCount
	}
	cacheCap := SYNC_MAX_BLOCK_CACHE_SIZE - this.getNonEmptyBlockCount()
	if count > cacheCap {
		count = cacheCap
	}

	counter := 1
	i := uint32(0)
	reqTimes := 1
	for {
		if counter > count {
			break
		}
		i++
		nextBlockHeight := curBlockHeight + i
		nextBlockHash := this.ledger.GetBlockHash(nextBlockHeight)
		if nextBlockHash == common.UINT256_EMPTY {
			return
		}
		if this.isBlockOnFlight(nextBlockHash) {
			if nextBlockHeight <= curBlockHeight+SYNC_NEXT_BLOCKS_HEIGHT {
				//request more nodes for next block height
				reqTimes = SYNC_NEXT_BLOCK_TIMES
			} else {
				continue
			}
		}
		if this.isInBlockCache(nextBlockHeight) {
			continue
		}
		if nextBlockHeight <= curBlockHeight+SYNC_NEXT_BLOCKS_HEIGHT {
			reqTimes = SYNC_NEXT_BLOCK_TIMES
		}
		for t := 0; t < reqTimes; t++ {
			reqNode := this.getNextNode(nextBlockHeight)
			if reqNode == nil {
				return
			}
			this.addFlightBlock(reqNode.GetID(), nextBlockHeight, nextBlockHash)
			msg := msgpack.NewBlkDataReq(nextBlockHash)
			err := this.server.Send(reqNode, msg)
			if err != nil {
				log.Warnf("[block-sync] syncBlock Height:%d ReqBlkData error:%s", nextBlockHeight, err)
				return
			} else {
				this.appendReqTime(reqNode.GetID())
			}
		}
		counter++
		reqTimes = 1
	}
}

//OnHeaderReceive receive header from net
func (this *BlockSyncMgr) OnHeaderReceive(fromID p2pComm.PeerId, headers []*types.Header) {
	if len(headers) == 0 {
		return
	}
	log.Infof("Header receive height:%d - %d", headers[0].Height, headers[len(headers)-1].Height)
	height := headers[0].Height
	curHeaderHeight := this.ledger.GetCurrentHeaderHeight()

	//Means another gorountinue is adding header
	if height <= curHeaderHeight {
		return
	}
	if !this.isHeaderOnFlight(height) {
		return
	}
	err := this.ledger.AddHeaders(headers)
	this.delFlightHeader(height)
	if err != nil {
		this.addErrorRespCnt(fromID)
		n := this.getNodeWeight(fromID)
		if n != nil && n.GetErrorRespCnt() >= SYNC_MAX_ERROR_RESP_TIMES {
			this.delNode(fromID)
		}
		log.Warnf("[block-sync] OnHeaderReceive AddHeaders error:%s", err)
		return
	}
	sort.Slice(headers, func(i, j int) bool {
		return headers[i].Height < headers[j].Height
	})
	curHeaderHeight = this.ledger.GetCurrentHeaderHeight()
	curBlockHeight := this.ledger.GetCurrentBlockHeight()
	for _, header := range headers {
		prevHeader, err := this.ledger.GetHeaderByHeight(header.Height - 1)
		if err != nil {
			log.Debugf("[block-sync] OnHeaderReceive GetHeaderByHeight error:%s", err)
			continue
		}
		log.Debugf("[block-sync] OnHeaderReceive GetHeaderByHeight height:%d, prevHeader transaction root:%+v", header.Height-1, prevHeader.TransactionsRoot)
		//handle empty block
		if header.TransactionsRoot == common.UINT256_EMPTY && prevHeader.TransactionsRoot == common.UINT256_EMPTY {
			log.Trace("[block-sync] OnHeaderReceive empty block Height:%d", header.Height)
			height := header.Height
			blockHash := header.Hash()
			this.delFlightBlock(blockHash)
			nextHeader := curHeaderHeight + 1
			if height > nextHeader {
				break
			}
			if height <= curBlockHeight {
				continue
			}
			block := &types.Block{
				Header: header,
			}
			this.addBlockCache(fromID, block, nil, common.UINT256_EMPTY)
		}
	}
	go this.saveBlock()
	this.syncHeader()
}

// OnBlockReceive receive block from net
func (this *BlockSyncMgr) OnBlockReceive(fromID p2pComm.PeerId, blockSize uint32, block *types.Block, ccMsg *types.CrossChainMsg,
	merkleRoot common.Uint256) {
	height := block.Header.Height
	blockHash := block.Hash()
	log.Tracef("[block-sync] OnBlockReceive Height:%d", height)
	flightInfo := this.getFlightBlock(blockHash, fromID)
	if flightInfo != nil {
		t := (time.Now().UnixNano() - flightInfo.GetStartTime().UnixNano()) / int64(time.Millisecond)
		s := float32(blockSize) / float32(t) * 1000.0 / 1024.0
		this.addNewSpeed(fromID, s)
	}

	this.delFlightBlock(blockHash)
	curHeaderHeight := this.ledger.GetCurrentHeaderHeight()
	nextHeader := curHeaderHeight + 1
	if height > nextHeader {
		return
	}
	curBlockHeight := this.ledger.GetCurrentBlockHeight()
	if height <= curBlockHeight {
		return
	}

	this.addBlockCache(fromID, block, ccMsg, merkleRoot)
	go this.saveBlock()
	this.syncBlock()
}

//OnAddPeer to node list when a new node added
func (this *BlockSyncMgr) OnAddNode(nodeId p2pComm.PeerId) {
	log.Debugf("[block-sync] OnAddNode:%s", nodeId.ToHexString())
	this.lock.Lock()
	defer this.lock.Unlock()
	w := NewNodeWeight(nodeId)
	this.nodeWeights[nodeId] = w
}

//OnDelNode remove from node list. When the node disconnect
func (this *BlockSyncMgr) OnDelNode(nodeId p2pComm.PeerId) {
	this.delNode(nodeId)
}

//delNode remove from node list
func (this *BlockSyncMgr) delNode(nodeId p2pComm.PeerId) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.nodeWeights, nodeId)
	if len(this.nodeWeights) == 0 {
		log.Warnf("no sync nodes")
	}
	log.Infof("[block-sync] delete node: %s", nodeId.ToHexString())
}

func (this *BlockSyncMgr) tryGetSyncHeaderLock() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.syncHeaderLock {
		return true
	}
	this.syncHeaderLock = true
	return false
}

func (this *BlockSyncMgr) releaseSyncHeaderLock() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.syncHeaderLock = false
}

func (this *BlockSyncMgr) tryGetSyncBlockLock() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.syncBlockLock {
		return true
	}
	this.syncBlockLock = true
	return false
}

func (this *BlockSyncMgr) releaseSyncBlockLock() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.syncBlockLock = false
}

func (this *BlockSyncMgr) addBlockCache(nodeID p2pComm.PeerId, block *types.Block, ccMsg *types.CrossChainMsg,
	merkleRoot common.Uint256) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.blocksCache.addBlock(nodeID, block, ccMsg, merkleRoot)
}

func (this *BlockSyncMgr) getBlockCache(blockHeight uint32) (p2pComm.PeerId, *types.Block, *types.CrossChainMsg,
	common.Uint256) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.blocksCache.getBlock(blockHeight)
}

func (this *BlockSyncMgr) delBlockCache(blockHeight uint32) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.blocksCache.delBlockLocked(blockHeight)
}

func (this *BlockSyncMgr) tryGetSaveBlockLock() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.saveBlockLock {
		return true
	}
	this.saveBlockLock = true
	return false
}

func (this *BlockSyncMgr) releaseSaveBlockLock() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.saveBlockLock = false
}

func (this *BlockSyncMgr) saveBlock() {
	if this.tryGetSaveBlockLock() {
		return
	}
	defer this.releaseSaveBlockLock()
	curBlockHeight := this.ledger.GetCurrentBlockHeight()
	nextBlockHeight := curBlockHeight + 1
	this.clearBlocks(curBlockHeight)
	for {
		fromID, nextBlock, ccMsg, merkleRoot := this.getBlockCache(nextBlockHeight)
		if nextBlock == nil {
			return
		}
		err := this.ledger.AddBlock(nextBlock, ccMsg, merkleRoot)
		this.delBlockCache(nextBlockHeight)
		if err != nil {
			this.addErrorRespCnt(fromID)
			n := this.getNodeWeight(fromID)
			if n != nil && n.GetErrorRespCnt() >= SYNC_MAX_ERROR_RESP_TIMES {
				this.delNode(fromID)
			}
			log.Warnf("[block-sync] saveBlock Height:%d AddBlock error:%s", nextBlockHeight, err)
			reqNode := this.getNextNode(nextBlockHeight)
			if reqNode == nil {
				return
			}
			this.addFlightBlock(reqNode.GetID(), nextBlockHeight, nextBlock.Hash())
			msg := msgpack.NewBlkDataReq(nextBlock.Hash())
			err := this.server.Send(reqNode, msg)
			if err != nil {
				log.Warn("[block-sync] require new block error:", err)
				return
			} else {
				this.appendReqTime(reqNode.GetID())
			}
			return
		}
		nextBlockHeight++
		this.pingOutsyncNodes(nextBlockHeight - 1)
	}
}

func (this *BlockSyncMgr) isInBlockCache(blockHeight uint32) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.blocksCache.isInBlockCache(blockHeight)
}

func (this *BlockSyncMgr) addFlightHeader(nodeId p2pComm.PeerId, height uint32) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.flightHeaders[height] = NewSyncFlightInfo(height, nodeId)
}

func (this *BlockSyncMgr) getFlightHeader(height uint32) *SyncFlightInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	info, ok := this.flightHeaders[height]
	if !ok {
		return nil
	}
	return info
}

func (this *BlockSyncMgr) delFlightHeader(height uint32) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.flightHeaders[height]
	if !ok {
		return false
	}
	delete(this.flightHeaders, height)
	return true
}

func (this *BlockSyncMgr) getFlightHeaderCount() int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return len(this.flightHeaders)
}

func (this *BlockSyncMgr) isHeaderOnFlight(height uint32) bool {
	flightInfo := this.getFlightHeader(height)
	return flightInfo != nil
}

func (this *BlockSyncMgr) addFlightBlock(nodeId p2pComm.PeerId, height uint32, blockHash common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.flightBlocks[blockHash] = append(this.flightBlocks[blockHash], NewSyncFlightInfo(height, nodeId))
}

func (this *BlockSyncMgr) getFlightBlocks(blockHash common.Uint256) []*SyncFlightInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	info, ok := this.flightBlocks[blockHash]
	if !ok {
		return nil
	}
	return info
}

func (this *BlockSyncMgr) getFlightBlock(blockHash common.Uint256, nodeId p2pComm.PeerId) *SyncFlightInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	infos, ok := this.flightBlocks[blockHash]
	if !ok {
		return nil
	}
	for _, info := range infos {
		if info.GetNodeId() == nodeId {
			return info
		}
	}
	return nil
}

func (this *BlockSyncMgr) delFlightBlock(blockHash common.Uint256) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.flightBlocks[blockHash]
	if !ok {
		return false
	}
	delete(this.flightBlocks, blockHash)
	return true
}

func (this *BlockSyncMgr) getFlightBlockCount() int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	cnt := 0
	for hash := range this.flightBlocks {
		cnt += len(this.flightBlocks[hash])
	}
	return cnt
}

func (this *BlockSyncMgr) isBlockOnFlight(blockHash common.Uint256) bool {
	flightInfos := this.getFlightBlocks(blockHash)
	return len(flightInfos) != 0
}

func (this *BlockSyncMgr) getNextNode(nextBlockHeight uint32) *peer.Peer {
	weights := this.getAllNodeWeights()
	sort.Sort(sort.Reverse(weights))
	nodelist := make([]p2pComm.PeerId, 0)
	for _, n := range weights {
		nodelist = append(nodelist, n.id)
	}
	nextNodeIndex := 0
	triedNode := make(map[p2pComm.PeerId]bool)
	for {
		var nextNodeId p2pComm.PeerId
		nextNodeIndex, nextNodeId = getNextNodeId(nextNodeIndex, nodelist)
		if nextNodeId.IsEmpty() {
			return nil
		}
		_, ok := triedNode[nextNodeId]
		if ok {
			return nil
		}
		triedNode[nextNodeId] = true
		n := this.server.GetPeer(nextNodeId)
		if n == nil {
			continue
		}
		nodeBlockHeight := n.GetHeight()
		if nextBlockHeight <= uint32(nodeBlockHeight) {
			return n
		}
	}
}

func (this *BlockSyncMgr) getNodeWithMinFailedTimes(flightInfo *SyncFlightInfo, curBlockHeight uint32) *peer.Peer {
	var minFailedTimes = math.MaxInt64
	var minFailedTimesNode *peer.Peer
	triedNode := make(map[p2pComm.PeerId]bool)
	for {
		nextNode := this.getNextNode(curBlockHeight + 1)
		if nextNode == nil {
			return nil
		}
		failedTimes := flightInfo.GetFailedTimes(nextNode.GetID())
		if failedTimes == 0 {
			return nextNode
		}
		_, ok := triedNode[nextNode.GetID()]
		if ok {
			return minFailedTimesNode
		}
		triedNode[nextNode.GetID()] = true
		if failedTimes < minFailedTimes {
			minFailedTimes = failedTimes
			minFailedTimesNode = nextNode
		}
	}
}

//Stop to sync
func (this *BlockSyncMgr) Stop() {
	close(this.exitCh)
}

//getNodeWeight get nodeweight by id
func (this *BlockSyncMgr) getNodeWeight(nodeId p2pComm.PeerId) *NodeWeight {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.nodeWeights[nodeId]
}

//getAllNodeWeights get all nodeweight and return a slice
func (this *BlockSyncMgr) getAllNodeWeights() NodeWeights {
	this.lock.RLock()
	defer this.lock.RUnlock()
	weights := make(NodeWeights, 0, len(this.nodeWeights))
	for _, w := range this.nodeWeights {
		weights = append(weights, w)
	}
	return weights
}

//addTimeoutCnt incre a node's timeout count
func (this *BlockSyncMgr) addTimeoutCnt(nodeId p2pComm.PeerId) {
	n := this.getNodeWeight(nodeId)
	if n != nil {
		n.AddTimeoutCnt()
	}
}

//addErrorRespCnt incre a node's error resp count
func (this *BlockSyncMgr) addErrorRespCnt(nodeId p2pComm.PeerId) {
	n := this.getNodeWeight(nodeId)
	if n != nil {
		n.AddErrorRespCnt()
	}
}

//appendReqTime append a node's request time
func (this *BlockSyncMgr) appendReqTime(nodeId p2pComm.PeerId) {
	n := this.getNodeWeight(nodeId)
	if n != nil {
		n.AppendNewReqtime()
	}
}

//addNewSpeed apend the new speed to tail, remove the oldest one
func (this *BlockSyncMgr) addNewSpeed(nodeId p2pComm.PeerId, speed float32) {
	n := this.getNodeWeight(nodeId)
	if n != nil {
		n.AppendNewSpeed(speed)
	}
}

//pingOutsyncNodes send ping msg to lower height nodes for syncing
func (this *BlockSyncMgr) pingOutsyncNodes(curHeight uint32) {
	peers := make([]*peer.Peer, 0)
	this.lock.RLock()
	maxHeight := curHeight
	for id := range this.nodeWeights {
		peer := this.server.GetPeer(id)
		if peer == nil {
			continue
		}
		peerHeight := uint32(peer.GetHeight())
		if peerHeight >= maxHeight {
			maxHeight = peerHeight
		}
		if peerHeight < curHeight {
			peers = append(peers, peer)
		}
	}
	this.lock.RUnlock()
	if curHeight > maxHeight-SYNC_MAX_HEIGHT_OFFSET && len(peers) > 0 {
		pingTo(this.server, curHeight, peers)
	}
}

//Using polling for load balance
func getNextNodeId(nextNodeIndex int, nodeList []p2pComm.PeerId) (int, p2pComm.PeerId) {
	num := len(nodeList)
	if num == 0 {
		return 0, p2pComm.PeerId{}
	}
	if nextNodeIndex >= num {
		nextNodeIndex = 0
	}
	index := nextNodeIndex
	nextNodeIndex++
	return nextNodeIndex, nodeList[index]
}

func pingTo(net p2p.P2P, height uint32, peers []*peer.Peer) {
	ping := msgpack.NewPingMsg(uint64(height))
	for _, p := range peers {
		go net.Send(p, ping)
	}
}
