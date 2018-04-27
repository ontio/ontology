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

package node

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/net/actor"
	"github.com/ontio/ontology/net/message"
	"github.com/ontio/ontology/net/protocol"
	"sync"
	"time"
)

const (
	SYNC_MAX_HEADER_FORWARD_SIZE = 5000 //keep CurrentHeaderHeight - CurrentBlockHeight <= SYNC_MAX_HEADER_FORWARD_SIZE
	SYNC_MAX_FLIGHT_HEADER_SIZE  = 1    //Number of headers on flight
	SYNC_MAX_FLIGHT_BLOCK_SIZE   = 15   //Number of blocks on flight
	SYNC_MAX_BLOCK_CACHE_SIZE    = 200  //Cache size of block wait to commit to ledger
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
	blockCache     map[common.Uint256]*types.Block    //Map BlockHash => block, using for cache the blocks receive from net, and waiting for commit to ledger
	blockCacheList []*types.Block                     //Using slice to holder blocks to keep order
	nodeList       []uint64                           //Holder all of nodes that can be used
	nextNodeIndex  int                                //Index for polling nodes
	localNode      *node                              //Pointer to the local node
	isSyncing      bool                               //Helper to avoid send block sync request duplicate
	exitCh         chan interface{}                   //ExitCh to receive exit signal
	lock           sync.RWMutex
}

//NewBlockSyncMgr return a BlockSyncMgr instance
func NewBlockSyncMgr(localNode *node) *BlockSyncMgr {
	return &BlockSyncMgr{
		flightBlocks:   make(map[common.Uint256]*SyncFlightInfo, 0),
		flightHeaders:  make(map[uint32]*SyncFlightInfo, 0),
		blockCache:     make(map[common.Uint256]*types.Block, 0),
		blockCacheList: make([]*types.Block, 0),
		nodeList:       make([]uint64, 0),
		localNode:      localNode,
		exitCh:         make(chan interface{}, 1),
	}
}

//Start to sync
func (this *BlockSyncMgr) Start() {
	go this.sync()
	reqCheckTicker := time.NewTicker(time.Second)
	defer reqCheckTicker.Stop()
	syncTicker := time.NewTicker(time.Second * 5)
	defer syncTicker.Stop()
	for {
		select {
		case <-this.exitCh:
			return
		case <-reqCheckTicker.C:
			go this.checkTimeout()
		case <-syncTicker.C:
			go this.sync()
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

	curHeaderHeight, err := actor.GetCurrentHeaderHeight()
	if err != nil {
		log.Errorf("BlockSyncMgr checkTimeout GetCurrentHeaderHeight error:%s", err)
		return
	}
	curBlockHeight, err := actor.GetCurrentBlockHeight()
	if err != nil {
		log.Errorf("BlockSyncMgr checkTimeout GetCurrentBlockHeight error:%s", err)
		return
	}
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
		message.SendMsgSyncHeaders(reqNode)
	}
	for blockHash, flightInfo := range blockTimeoutFlights {
		if flightInfo.Height <= curBlockHeight {
			this.delFlightBlock(blockHash)
			continue
		}
		flightInfo.ResetStartTime()
		flightInfo.MarkFailedNode()
		log.Debugf("BlockSyncMgr checkTimeout sync height:%x block:%x timeout after:%d s times:%d", blockHash, flightInfo.Height, SYNC_BLOCK_REQUEST_TIMEOUT, flightInfo.GetTotalFailedTimes())
		reqNode := this.getNodeWithMinFailedTimes(flightInfo, curBlockHeight)
		if reqNode == nil {
			break
		}
		flightInfo.SetNodeId(reqNode.GetID())
		err = message.ReqBlkData(reqNode, blockHash)
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
	if !this.localNode.IsUptoMinNodeCount() {
		return
	}
	if this.getFlightHeaderCount() >= SYNC_MAX_FLIGHT_HEADER_SIZE {
		return
	}
	curBlockHeight, err := actor.GetCurrentBlockHeight()
	if err != nil {
		log.Errorf("BlockSyncMgr syncHeader GetCurrentBlockHeight error:%s", err)
		return
	}
	curHeaderHeight, err := actor.GetCurrentHeaderHeight()
	if err != nil {
		log.Errorf("BlockSyncMgr syncHeader GetCurrentHeaderHeight error:%s", err)
		return
	}
	//Waiting for block catch up header
	if curHeaderHeight-curBlockHeight >= SYNC_MAX_HEADER_FORWARD_SIZE {
		return
	}
	NextHeaderId := curHeaderHeight + 1
	reqNode := this.getNextNode(NextHeaderId)
	if reqNode == nil {
		return
	}
	this.addFlightHeader(reqNode.id, NextHeaderId)
	message.SendMsgSyncHeaders(reqNode)
	log.Infof("SendMsgSyncHeaders Request Height:%d", NextHeaderId)
}

func (this *BlockSyncMgr) syncBlock() {
	if this.isSyncingBlock() {
		return
	}
	defer this.resetIsSyncingBlock()

	availCount := SYNC_MAX_FLIGHT_BLOCK_SIZE - this.getFlightBlockCount()
	if availCount <= 0 {
		return
	}
	curBlockHeight, err := actor.GetCurrentBlockHeight()
	if err != nil {
		log.Errorf("BlockSyncMgr syncBlock GetCurrentBlockHeight error:%s", err)
		return
	}
	curHeaderHeight, err := actor.GetCurrentHeaderHeight()
	if err != nil {
		log.Errorf("BlockSyncMgr syncBlock GetCurrentHeaderHeight error:%s", err)
		return
	}
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
		height := curBlockHeight + i
		blockHash, err := actor.GetBlockHashByHeight(height)
		if err != nil {
			log.Errorf("BlockSyncMgr syncBlock GetBlockHashByHeight:%d error:%s", height, err)
			return
		}
		if blockHash == common.UINT256_EMPTY {
			return
		}
		if this.isBlockOnFlight(blockHash) {
			continue
		}
		if this.isInBlockCache(blockHash) {
			continue
		}
		reqNode := this.getNextNode(height)
		if reqNode == nil {
			return
		}
		this.addFlightBlock(reqNode.id, height, blockHash)
		err = message.ReqBlkData(reqNode, blockHash)
		if err != nil {
			log.Errorf("BlockSyncMgr syncBlock Height:%d ReqBlkData error:%s", height, err)
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
	res := this.delFlightHeader(headers[0].Height)
	if !res {
		return
	}
	err := actor.AddHeaders(headers)
	if err != nil {
		headerHeight, errInternal := actor.GetCurrentHeaderHeight()
		if errInternal != nil {
			log.Error("BlockSyncMgr GetCurrentHeaderHeight  failed:%s", err)
			return
		}
		if headers[len(headers)-1].Height <= headerHeight  {
			log.Warnf("BlockSyncMgr AddHeaders error:%s", err)
		}
		return
	}
	this.syncHeader()
}

//OnBlockReceive receive block from net
func (this *BlockSyncMgr) OnBlockReceive(block *types.Block) {
	height := block.Header.Height
	curHeaderHeight, err := actor.GetCurrentHeaderHeight()
	if err != nil {
		log.Errorf("BlockSyncMgr OnBlockReceive GetCurrentHeaderHeight error:%s", err)
		return
	}
	nextHeader := curHeaderHeight + 1
	if height > nextHeader {
		return
	}
	log.Debugf("OnBlockReceive Height:%d", block.Header.Height)
	blockHash := block.Hash()
	this.addToBlockCache(block)
	this.checkBlockCache()
	this.delFlightBlock(blockHash)
	go func() {
		time.After(time.Millisecond * 10)
		this.syncBlock()
	}()
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

func (this *BlockSyncMgr) isSyncingBlock() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.isSyncing {
		return true
	}
	this.isSyncing = true
	return false
}

func (this *BlockSyncMgr) resetIsSyncingBlock() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.isSyncing = false
}

func (this *BlockSyncMgr) addToBlockCache(block *types.Block) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	blockHash := block.Hash()
	blockHeight := block.Header.Height
	_, ok := this.blockCache[blockHash]
	if ok {
		return false
	}
	insertIndex := -1
	for index, b := range this.blockCacheList {
		if b.Header.Height > blockHeight {
			insertIndex = index
			break
		}
	}
	if insertIndex < 0 {
		this.blockCacheList = append(this.blockCacheList, block)
	} else {
		//Insert block to keep slice order
		rear := append([]*types.Block{}, this.blockCacheList[insertIndex:]...)
		this.blockCacheList = append(this.blockCacheList[0:insertIndex], block)
		this.blockCacheList = append(this.blockCacheList, rear...)
	}
	this.blockCache[blockHash] = block
	return true
}

func (this *BlockSyncMgr) isInBlockCache(blockHash common.Uint256) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	_, ok := this.blockCache[blockHash]
	return ok
}

func (this *BlockSyncMgr) checkBlockCache() {
	curBlockHeight, err := actor.GetCurrentBlockHeight()
	if err != nil {
		log.Errorf("BlockSyncMgr checkBlockCache GetCurrentBlockHeight error:%s", err)
		return
	}
	nextBlockHeight := curBlockHeight + 1

	clearSize := 0
	blockSize := 0
	this.lock.Lock()
	for _, block := range this.blockCacheList {
		height := block.Header.Height
		blockHash := block.Hash()
		if block.Header.Height <= curBlockHeight {
			clearSize++
			delete(this.blockCache, blockHash)
			continue
		}
		if height != nextBlockHeight {
			break
		}
		delete(this.blockCache, blockHash)
		nextBlockHeight++
		blockSize++
	}
	if clearSize > 0 {
		this.blockCacheList = this.blockCacheList[clearSize:]
	}

	blocks := this.blockCacheList[:blockSize]
	this.blockCacheList = this.blockCacheList[blockSize:]
	this.lock.Unlock()

	for _, block := range blocks {
		err := actor.AddBlock(block)
		if err != nil {
			log.Errorf("BlockSyncMgr checkBlockCache AddBlock Height:%d Hash:%x error:%s", block.Header.Height, block.Hash(), err)
			return
		}
	}
}

func (this *BlockSyncMgr) getBlockCacheSize() int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return len(this.blockCacheList)
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
	this.lock.RLock()
	defer this.lock.RUnlock()
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

func (this *BlockSyncMgr) getNextNode(curBlockHeight uint32) *node {
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
		n := this.localNode.GetNodeById(nextNodeId)
		if n == nil {
			continue
		}
		if n.GetState() != protocol.ESTABLISH {
			continue
		}
		nodeBlockHeight := n.GetHeight()
		if curBlockHeight < uint32(nodeBlockHeight) {
			return n
		}
	}
}

func (this *BlockSyncMgr) getNodeWithMinFailedTimes(flightInfo *SyncFlightInfo, curBlockHeight uint32) *node {
	var minFailedTimes int
	var minFailedTimesNode *node
	triedNode := make(map[uint64]bool, 0)
	for {
		nextNode := this.getNextNode(curBlockHeight)
		if nextNode == nil {
			return nil
		}
		failedTimes := flightInfo.GetFailedTimes(nextNode.id)
		if failedTimes == 0 {
			return nextNode
		}
		_, ok := triedNode[nextNode.id]
		if ok {
			return minFailedTimesNode
		}
		triedNode[nextNode.id] = true
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
