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

package p2pserver

import (
	"math"
	"sync"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	p2pComm "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	"github.com/ontio/ontology/p2pserver/peer"
)

const (
	SYNC_MAX_HEADER_FORWARD_SIZE = 5000 //keep CurrentHeaderHeight - CurrentBlockHeight <= SYNC_MAX_HEADER_FORWARD_SIZE
	SYNC_MAX_FLIGHT_HEADER_SIZE  = 1    //Number of headers on flight
	SYNC_MAX_FLIGHT_BLOCK_SIZE   = 50   //Number of blocks on flight
	SYNC_MAX_BLOCK_CACHE_SIZE    = 500  //Cache size of block wait to commit to ledger
	SYNC_HEADER_REQUEST_TIMEOUT  = 10   //s, Request header timeout time. If header haven't receive after SYNC_HEADER_REQUEST_TIMEOUT second, retry
	SYNC_BLOCK_REQUEST_TIMEOUT   = 15   //s, Request block timeout time. If block haven't received after SYNC_BLOCK_REQUEST_TIMEOUT second, retry
)

//SyncFlightInfo record the info of fight object(header or block)
type SyncFlightInfo struct {
	Height      uint32         //BlockHeight of HeaderHeight
	nodeId      uint64         //The current node to send msg
	startTime   time.Time      //Request start time
	failedNodes map[uint64]int //Map nodeId => timeout times
	totalFailed int            //Total timeout times
	lock        sync.RWMutex
}

//NewSyncFlightInfo return a new SyncFlightInfo instance
func NewSyncFlightInfo(height uint32, nodeId uint64) *SyncFlightInfo {
	return &SyncFlightInfo{
		Height:      height,
		nodeId:      nodeId,
		startTime:   time.Now(),
		failedNodes: make(map[uint64]int, 0),
	}
}

//GetNodeId return current node id for sending msg
func (this *SyncFlightInfo) GetNodeId() uint64 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.nodeId
}

//SetNodeId set a new node id
func (this *SyncFlightInfo) SetNodeId(nodeId uint64) {
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
func (this *SyncFlightInfo) GetFailedTimes(nodeId uint64) int {
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

//BlockSyncMgr is the manager class to deal with block sync
type BlockSyncMgr struct {
	flightBlocks   map[common.Uint256]*SyncFlightInfo //Map BlockHash => SyncFlightInfo, using for manager all of those block flights
	flightHeaders  map[uint32]*SyncFlightInfo         //Map HeaderHeight => SyncFlightInfo, using for manager all of those header flights
	blocksCache    map[uint32]*types.Block            //Map BlockHash => block, using for cache the blocks receive from net, and waiting for commit to ledger
	nodeList       []uint64                           //Holder all of nodes that can be used
	nextNodeIndex  int                                //Index for polling nodes
	server         *P2PServer                         //Pointer to the local node
	syncBlockLock  bool                               //Help to avoid send block sync request duplicate
	syncHeaderLock bool                               //Help to avoid send header sync request duplicate
	saveBlockLock  bool                               //Help to avoid saving block concurrently
	exitCh         chan interface{}                   //ExitCh to receive exit signal
	ledger         *ledger.Ledger
	lock           sync.RWMutex
}

//NewBlockSyncMgr return a BlockSyncMgr instance
func NewBlockSyncMgr(server *P2PServer) *BlockSyncMgr {
	return &BlockSyncMgr{
		flightBlocks:  make(map[common.Uint256]*SyncFlightInfo, 0),
		flightHeaders: make(map[uint32]*SyncFlightInfo, 0),
		blocksCache:   make(map[uint32]*types.Block, 0),
		nodeList:      make([]uint64, 0),
		server:        server,
		ledger:        server.ledger,
		exitCh:        make(chan interface{}, 1),
	}
}

//Start to sync
func (this *BlockSyncMgr) Start() {
	go this.sync()
	ticker := time.NewTicker(time.Second)
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
	headerTimeoutFlights := make(map[uint32]*SyncFlightInfo, 0)
	blockTimeoutFlights := make(map[common.Uint256]*SyncFlightInfo, 0)
	this.lock.RLock()
	for height, flightInfo := range this.flightHeaders {
		if int(now.Sub(flightInfo.startTime).Seconds()) >= SYNC_HEADER_REQUEST_TIMEOUT {
			headerTimeoutFlights[height] = flightInfo
		}
	}
	for blockHash, flightInfo := range this.flightBlocks {
		if int(now.Sub(flightInfo.startTime).Seconds()) >= SYNC_BLOCK_REQUEST_TIMEOUT {
			blockTimeoutFlights[blockHash] = flightInfo
		}
	}
	this.lock.RUnlock()

	curHeaderHeight := this.ledger.GetCurrentHeaderHeight()
	curBlockHeight := this.ledger.GetCurrentBlockHeight()

	for height, flightInfo := range headerTimeoutFlights {
		if height <= curHeaderHeight {
			this.delFlightHeader(height)
			continue
		}
		flightInfo.ResetStartTime()
		flightInfo.MarkFailedNode()
		log.Infof("BlockSyncMgr checkTimeout sync headers:%d timeout after:%d s Times:%d", height, SYNC_HEADER_REQUEST_TIMEOUT, flightInfo.GetTotalFailedTimes())
		reqNode := this.getNodeWithMinFailedTimes(flightInfo, curBlockHeight)
		if reqNode == nil {
			break
		}
		flightInfo.SetNodeId(reqNode.GetID())

		headerHash := this.ledger.GetCurrentHeaderHash()
		buf, err := msgpack.NewHeadersReq(headerHash)
		if err != nil {
			log.Error("failed build a new headersReq")
		} else {
			this.server.Send(reqNode, buf, false)
		}
	}
	for blockHash, flightInfo := range blockTimeoutFlights {
		if flightInfo.Height <= curBlockHeight {
			this.delFlightBlock(blockHash)
			continue
		}
		flightInfo.ResetStartTime()
		flightInfo.MarkFailedNode()
		log.Infof("BlockSyncMgr checkTimeout sync height:%d block:%x timeout after:%d s times:%d", flightInfo.Height, blockHash, SYNC_BLOCK_REQUEST_TIMEOUT, flightInfo.GetTotalFailedTimes())
		reqNode := this.getNodeWithMinFailedTimes(flightInfo, curBlockHeight)
		if reqNode == nil {
			break
		}
		flightInfo.SetNodeId(reqNode.GetID())

		reqBuf, err := msgpack.NewBlkDataReq(blockHash)
		if err != nil {
			log.Error("syncBlock error:", err)
		}
		err = this.server.Send(reqNode, reqBuf, false)

		if err != nil {
			log.Errorf("BlockSyncMgr checkTimeout Height:%d Hash:%x ReqBlkData error:%s", flightInfo.Height, blockHash, err)
			continue
		}
	}
}

func (this *BlockSyncMgr) sync() {
	this.syncHeader()
	this.syncBlock()
}

func (this *BlockSyncMgr) syncHeader() {
	if !this.server.reachMinConnection() {
		return
	}
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
	buf, err := msgpack.NewHeadersReq(headerHash)
	if err != nil {
		log.Error("failed build a new headersReq")
	} else {
		this.server.Send(reqNode, buf, false)
	}

	log.Infof("SendMsgSyncHeaders Request Height:%d", NextHeaderId)
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
	cacheCap := SYNC_MAX_BLOCK_CACHE_SIZE - this.getBlockCacheSize()
	if count > cacheCap {
		count = cacheCap
	}

	counter := 1
	i := uint32(0)
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
			continue
		}
		if this.isInBlockCache(nextBlockHeight) {
			continue
		}
		reqNode := this.getNextNode(nextBlockHeight)
		if reqNode == nil {
			return
		}
		this.addFlightBlock(reqNode.GetID(), nextBlockHeight, nextBlockHash)
		reqBuf, err := msgpack.NewBlkDataReq(nextBlockHash)
		if err != nil {
			log.Error("syncBlock error:", err)
		}
		err = this.server.Send(reqNode, reqBuf, false)
		if err != nil {
			log.Errorf("BlockSyncMgr syncBlock Height:%d ReqBlkData error:%s", nextBlockHeight, err)
			return
		}
		counter++
	}
}

//OnHeaderReceive receive header from net
func (this *BlockSyncMgr) OnHeaderReceive(headers []*types.Header) {
	if len(headers) == 0 {
		return
	}
	log.Infof("OnHeaderReceive Height:%d - %d", headers[0].Height, headers[len(headers)-1].Height)
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
		log.Errorf("BlockSyncMgr AddHeaders error:%s", err)
		return
	}
	this.syncHeader()
}

//OnBlockReceive receive block from net
func (this *BlockSyncMgr) OnBlockReceive(block *types.Block) {
	height := block.Header.Height
	blockHash := block.Hash()
	log.Debugf("OnBlockReceive Height:%d", height)

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
	this.addBlockCache(block)
	go this.saveBlock()
	this.syncBlock()
}

//OnAddNode to node list when a new node added
func (this *BlockSyncMgr) OnAddNode(nodeId uint64) {
	log.Infof("BlockSyncMgr OnAddNode:%d", nodeId)
	this.lock.Lock()
	defer this.lock.Unlock()
	this.nodeList = append(this.nodeList, nodeId)
}

//OnDelNode remove from node list. When the node disconnect
func (this *BlockSyncMgr) OnDelNode(nodeId uint64) {
	this.lock.Lock()
	defer this.lock.Unlock()
	index := -1
	for i, id := range this.nodeList {
		if nodeId == id {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	this.nodeList = append(this.nodeList[:index], this.nodeList[index+1:]...)
	log.Infof("BlockSyncMgr OnDelNode:%d", nodeId)
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

func (this *BlockSyncMgr) addBlockCache(block *types.Block) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.blocksCache[block.Header.Height] = block
	return true
}

func (this *BlockSyncMgr) getBlockCache(blockHeight uint32) *types.Block {
	this.lock.RLock()
	defer this.lock.RUnlock()
	block, ok := this.blocksCache[blockHeight]
	if !ok {
		return nil
	}
	return block
}

func (this *BlockSyncMgr) delBlockCache(blockHeight uint32) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.blocksCache, blockHeight)
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
	this.lock.Lock()
	for height := range this.blocksCache {
		if height <= curBlockHeight {
			delete(this.blocksCache, height)
		}
	}
	this.lock.Unlock()
	for {
		nextBlock := this.getBlockCache(nextBlockHeight)
		if nextBlock == nil {
			return
		}
		err := this.ledger.AddBlock(nextBlock)
		this.delBlockCache(nextBlockHeight)
		if err != nil {
			log.Warnf("BlockSyncMgr saveBlock Height:%d AddBlock error:%s", nextBlockHeight, err)
			reqNode := this.getNextNode(nextBlockHeight)
			if reqNode == nil {
				return
			}
			this.addFlightBlock(reqNode.GetID(), nextBlockHeight, nextBlock.Hash())
			reqBuf, err := msgpack.NewBlkDataReq(nextBlock.Hash())
			if err != nil {
				log.Error("syncBlock error:", err)
				return
			}
			err = this.server.Send(reqNode, reqBuf, false)
			if err != nil {
				log.Errorf("BlockSyncMgr saveBlock Height:%d ReqBlkData error:%s", nextBlockHeight, err)
			}
			return
		}
		nextBlockHeight++
	}
}

func (this *BlockSyncMgr) isInBlockCache(blockHeight uint32) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	_, ok := this.blocksCache[blockHeight]
	return ok
}

func (this *BlockSyncMgr) getBlockCacheSize() int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return len(this.blocksCache)
}

func (this *BlockSyncMgr) addFlightHeader(nodeId uint64, height uint32) {
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

func (this *BlockSyncMgr) addFlightBlock(nodeId uint64, height uint32, blockHash common.Uint256) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.flightBlocks[blockHash] = NewSyncFlightInfo(height, nodeId)
}

func (this *BlockSyncMgr) getFlightBlock(blockHash common.Uint256) *SyncFlightInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	info, ok := this.flightBlocks[blockHash]
	if !ok {
		return nil
	}
	return info
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
	return len(this.flightBlocks)
}

func (this *BlockSyncMgr) isBlockOnFlight(blockHash common.Uint256) bool {
	flightInfo := this.getFlightBlock(blockHash)
	return flightInfo != nil
}

//Using polling for load balance
func (this *BlockSyncMgr) getNextNodeId() uint64 {
	this.lock.Lock()
	defer this.lock.Unlock()
	num := len(this.nodeList)
	if num == 0 {
		return 0
	}
	if this.nextNodeIndex >= num {
		this.nextNodeIndex = 0
	}
	index := this.nextNodeIndex
	this.nextNodeIndex++
	return this.nodeList[index]
}

func (this *BlockSyncMgr) getNextNode(curBlockHeight uint32) *peer.Peer {
	triedNode := make(map[uint64]bool, 0)
	for {
		nextNodeId := this.getNextNodeId()
		if nextNodeId == 0 {
			return nil
		}
		_, ok := triedNode[nextNodeId]
		if ok {
			return nil
		}
		triedNode[nextNodeId] = true
		n := this.server.getNode(nextNodeId)
		if n == nil {
			continue
		}
		if n.GetSyncState() != p2pComm.ESTABLISH {
			continue
		}
		nodeBlockHeight := n.GetHeight()
		if curBlockHeight < uint32(nodeBlockHeight) {
			return n
		}
	}
}

func (this *BlockSyncMgr) getNodeWithMinFailedTimes(flightInfo *SyncFlightInfo, curBlockHeight uint32) *peer.Peer {
	var minFailedTimes = math.MaxInt64
	var minFailedTimesNode *peer.Peer
	triedNode := make(map[uint64]bool, 0)
	for {
		nextNode := this.getNextNode(curBlockHeight)
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
func (this *BlockSyncMgr) Close() {
	close(this.exitCh)
}
