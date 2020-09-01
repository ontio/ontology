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
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
)

type SyncCheckReq struct {
	msg      ConsensusMsg
	peerIdx  uint32
	blockNum uint32
}

type BlockSyncReq struct {
	targetPeers    []uint32
	startBlockNum  uint32
	targetBlockNum uint32 // targetBlockNum == 0, as one cancel syncing request
}

type PeerSyncer struct {
	lock          sync.Mutex
	peerIdx       uint32
	nextReqBlkNum uint32
	targetBlkNum  uint32
	active        bool

	server *Server
	msgC   chan ConsensusMsg
}

type SyncMsg struct {
	fromPeer uint32
	msg      ConsensusMsg
}

type BlockMsgFromPeer struct {
	fromPeer uint32
	block    *Block
}

type BlockFromPeers map[uint32]*Block // index by peerId

type Syncer struct {
	lock   sync.Mutex
	server *Server

	maxRequestPerPeer int
	nextReqBlkNum     uint32
	targetBlkNum      uint32

	syncCheckReqC  chan *SyncCheckReq
	blockSyncReqC  chan *BlockSyncReq
	syncMsgC       chan *SyncMsg // receive syncmsg from server
	blockFromPeerC chan *BlockMsgFromPeer

	peers         map[uint32]*PeerSyncer
	pendingBlocks map[uint32]BlockFromPeers // index by blockNum
}

func newSyncer(server *Server) *Syncer {
	return &Syncer{
		server:            server,
		maxRequestPerPeer: 4,
		nextReqBlkNum:     1,
		syncCheckReqC:     make(chan *SyncCheckReq, 4),
		blockSyncReqC:     make(chan *BlockSyncReq, 16),
		syncMsgC:          make(chan *SyncMsg, 256),
		blockFromPeerC:    make(chan *BlockMsgFromPeer, 64),
		peers:             make(map[uint32]*PeerSyncer),
		pendingBlocks:     make(map[uint32]BlockFromPeers),
	}
}

func (self *Syncer) stop() {
	self.lock.Lock()
	defer self.lock.Unlock()

	close(self.syncCheckReqC)
	close(self.blockSyncReqC)
	close(self.syncMsgC)
	close(self.blockFromPeerC)

	self.peers = make(map[uint32]*PeerSyncer)
	self.pendingBlocks = make(map[uint32]BlockFromPeers)
}

func (self *Syncer) run() {
	self.server.quitWg.Add(1)
	defer self.server.quitWg.Done()

	for {
		select {
		case <-self.syncCheckReqC:
		case req := <-self.blockSyncReqC:
			if req.targetBlockNum == 0 {
				continue
			}

			log.Infof("server %d, got sync req(%d, %d) to %v",
				self.server.Index, req.startBlockNum, req.targetBlockNum, req.targetPeers)
			if req.startBlockNum <= self.server.GetCommittedBlockNo() {
				req.startBlockNum = self.server.GetCommittedBlockNo() + 1
				log.Infof("server %d, sync req start change to %d",
					self.server.Index, req.startBlockNum)
			}
			for ; req.startBlockNum <= req.targetBlockNum; req.startBlockNum++ {
				blk, _ := self.server.blockPool.getSealedBlock(req.startBlockNum)
				if blk == nil {
					log.Infof("server %d, on starting syncing %d, nil block from ledger",
						self.server.Index, req.startBlockNum)
					break
				}
				if err := self.server.fastForwardBlock(blk); err != nil {
					log.Infof("server %d, on starting syncing %d, %s",
						self.server.Index, req.startBlockNum, err)
					break
				}
			}
			if req.startBlockNum > req.targetBlockNum {
				continue
			}
			self.onNewBlockSyncReq(req)

		case syncMsg := <-self.syncMsgC:
			if p, present := self.peers[syncMsg.fromPeer]; present {
				if p.active {
					p.msgC <- syncMsg.msg
				} else {
					// report err
					p.msgC <- nil
				}
			} //  else {
			// 	// report error
			// }

		case blkMsgFromPeer := <-self.blockFromPeerC:
			blkNum := blkMsgFromPeer.block.getBlockNum()
			if blkNum < self.nextReqBlkNum {
				continue
			}

			log.Infof("server %d, next: %d, target: %d,  from syncer %d, blk %d, proposer %d",
				self.server.Index, self.nextReqBlkNum, self.targetBlkNum, blkMsgFromPeer.fromPeer, blkNum, blkMsgFromPeer.block.getProposer())
			if _, present := self.pendingBlocks[blkNum]; !present {
				self.pendingBlocks[blkNum] = make(BlockFromPeers)
			}
			self.pendingBlocks[blkNum][blkMsgFromPeer.fromPeer] = blkMsgFromPeer.block

			if blkNum != self.nextReqBlkNum {
				continue
			}
			for self.nextReqBlkNum <= self.targetBlkNum {
				// FIXME: compete with ledger syncing
				var blk *Block
				if self.nextReqBlkNum <= ledger.DefLedger.GetCurrentBlockHeight() {
					blk, _ = self.server.blockPool.getSealedBlock(self.nextReqBlkNum)
				}
				if blk == nil {
					blk = self.blockConsensusDone(self.pendingBlocks[self.nextReqBlkNum])
					merkBlk := self.blockCheckMerkleRoot(self.pendingBlocks[self.nextReqBlkNum])
					if blk == nil || merkBlk == nil {
						break
					}
					if blk.getPrevExecMerkleRoot() != merkBlk.getPrevExecMerkleRoot() {
						break
					}
				} else {
					merkleRoot, err := self.server.blockPool.getExecMerkleRoot(blkNum - 1)
					if err != nil {
						log.Errorf("failed to GetExecMerkleRoot: %s,blkNum:%d", err, blkNum-1)
						break
					}
					if blk.getPrevExecMerkleRoot() != merkleRoot {
						break
					}
				}
				if blk == nil {
					break
				}
				prevHash := blk.getPrevBlockHash()
				log.Debugf("server %d syncer, sealed block %d, proposer %d, prevhash: %s",
					self.server.Index, self.nextReqBlkNum, blk.getProposer(), prevHash.ToHexString())
				if err := self.server.fastForwardBlock(blk); err != nil {
					log.Errorf("server %d syncer, fastforward block %d failed %s",
						self.server.Index, self.nextReqBlkNum, err)
					break
				}
				delete(self.pendingBlocks, self.nextReqBlkNum)
				self.nextReqBlkNum++
			}
			if self.nextReqBlkNum > self.targetBlkNum {
				self.server.stateMgr.StateEventC <- &StateEvent{
					Type:     SyncDone,
					blockNum: self.targetBlkNum,
				}

				// stop all sync-peers
				for _, syncPeer := range self.peers {
					syncPeer.stop(true)
				}

				// reset to default
				self.nextReqBlkNum = 1
				self.targetBlkNum = 0
			}

		case <-self.server.quitC:
			log.Infof("server %d, syncer quit", self.server.Index)
			return
		}
	}
}

func (self *Syncer) blockConsensusDone(blks BlockFromPeers) *Block {
	// TODO: also check blockhash
	proposers := make(map[uint32]int)
	for _, blk := range blks {
		proposers[blk.getProposer()] += 1
	}

	chainCfg := self.server.GetChainConfig()
	for proposerId, cnt := range proposers {
		if cnt > int(chainCfg.C) {
			// find the block
			for _, blk := range blks {
				if blk.getProposer() == proposerId {
					return blk
				}
			}
		}
	}
	return nil
}

func (self *Syncer) getCurrentTargetBlockNum() uint32 {
	return self.targetBlkNum
}

func (self *Syncer) blockCheckMerkleRoot(blks BlockFromPeers) *Block {
	merkleRoot := make(map[common.Uint256]int)
	for _, blk := range blks {
		merkleRoot[blk.getPrevExecMerkleRoot()] += 1
	}

	chainCfg := self.server.GetChainConfig()
	for merklerootvalue, cnt := range merkleRoot {
		if cnt > int(chainCfg.C) {
			// find the block
			for _, blk := range blks {
				if blk.getPrevExecMerkleRoot() == merklerootvalue {
					return blk
				}
			}
		}
	}
	return nil
}

func (self *Syncer) isActive() bool {
	return self.nextReqBlkNum <= self.targetBlkNum
}

func (self *Syncer) startPeerSyncer(peerSyncer *PeerSyncer, targetBlkNum uint32) {

	peerSyncer.lock.Lock()
	defer peerSyncer.lock.Unlock()

	if targetBlkNum > peerSyncer.targetBlkNum {
		peerSyncer.targetBlkNum = targetBlkNum
	}
	if peerSyncer.targetBlkNum >= peerSyncer.nextReqBlkNum && !peerSyncer.active {
		peerSyncer.active = true
		go func() {
			peerSyncer.run()
		}()
	}
}

func (self *Syncer) onNewBlockSyncReq(req *BlockSyncReq) {
	if req.startBlockNum < self.nextReqBlkNum {
		log.Errorf("server %d new blockSyncReq startblkNum %d vs %d",
			self.server.Index, req.startBlockNum, self.nextReqBlkNum)
	}
	if req.targetBlockNum <= self.targetBlkNum {
		return
	}
	if self.nextReqBlkNum == 1 {
		self.nextReqBlkNum = req.startBlockNum
	}
	self.targetBlkNum = req.targetBlockNum
	// }

	for _, peerIdx := range req.targetPeers {
		if p, present := self.peers[peerIdx]; !present || !p.active {
			nextBlkNum := self.nextReqBlkNum
			if p != nil && p.nextReqBlkNum > nextBlkNum {
				log.Infof("server %d, syncer with peer %d start from %d, vs %d",
					self.server.Index, peerIdx, p.nextReqBlkNum, self.nextReqBlkNum)
				nextBlkNum = p.nextReqBlkNum
			}
			self.peers[peerIdx] = &PeerSyncer{
				peerIdx:       peerIdx,
				nextReqBlkNum: nextBlkNum,
				targetBlkNum:  self.targetBlkNum,
				active:        false,
				server:        self.server,
				msgC:          make(chan ConsensusMsg, 4),
			}
		}
		p := self.peers[peerIdx]
		self.startPeerSyncer(p, self.targetBlkNum)
	}
}

/////////////////////////////////////////////////////////////////////
//
// peer syncer
//
/////////////////////////////////////////////////////////////////////

func (self *PeerSyncer) run() {
	// send blockinfo fetch req to peer
	// wait blockinfo fetch rep
	// if have the proposal in msgpool, get proposal from msg pool, notify syncer
	// if not have the proposal in msgpool,
	// 				send block fetch req to peer
	//				wait block fetch rsp from peer
	//				notify syncer

	log.Infof("server %d, syncer %d started, start %d, target %d",
		self.server.Index, self.peerIdx, self.nextReqBlkNum, self.targetBlkNum)

	errQuit := true
	defer func() {
		log.Infof("server %d, syncer %d quit, start %d, target %d",
			self.server.Index, self.peerIdx, self.nextReqBlkNum, self.targetBlkNum)
		self.stop(errQuit)
	}()

	var err error
	blkProposers := make(map[uint32]uint32)
	for self.nextReqBlkNum <= self.targetBlkNum {
		blkNum := self.nextReqBlkNum
		if _, present := blkProposers[blkNum]; !present {
			blkInfos, err := self.requestBlockInfo(blkNum)
			if err != nil {
				log.Errorf("server %d failed to construct blockinfo fetch msg to peer %d: %s",
					self.server.Index, self.peerIdx, err)
				return
			}
			for _, p := range blkInfos {
				blkProposers[p.BlockNum] = p.Proposer
			}
		}
		if _, present := blkProposers[blkNum]; !present {
			log.Errorf("server %d failed to get block %d proposer from %d", self.server.Index,
				blkNum, self.peerIdx)
			return
		}

		var proposalBlock *Block
		proposalBlock, _ = self.server.blockPool.getSealedBlock(blkNum)
		if proposalBlock == nil {
			if proposalBlock, err = self.requestBlock(blkNum); err != nil {
				log.Errorf("failed to get block %d from peer %d: %s", blkNum, self.peerIdx, err)
				return
			}
		}
		if err := self.fetchedBlock(blkNum, proposalBlock); err != nil {
			log.Errorf("failed to commit block %d from peer syncer %d to syncer: %s",
				blkNum, self.peerIdx, err)
		}
		delete(blkProposers, blkNum)
	}
	errQuit = false
}

func (self *PeerSyncer) stop(force bool) bool {
	self.lock.Lock()
	defer self.lock.Unlock()
	if force || self.nextReqBlkNum > self.targetBlkNum {
		self.active = false
		return true
	}

	return false
}

func (self *PeerSyncer) requestBlock(blkNum uint32) (*Block, error) {
	msg := self.server.constructBlockFetchMsg(blkNum)
	self.server.msgSendC <- &SendMsgEvent{
		ToPeer: self.peerIdx,
		Msg:    msg,
	}

	t := time.NewTimer(time.Duration(atomic.LoadInt64(&makeProposalTimeout) * 2))
	defer t.Stop()

	select {
	case msg := <-self.msgC:
		if msg == nil {
			return nil, fmt.Errorf("nil block fetch rsp msg received")
		}
		switch msg.Type() {
		case BlockFetchRespMessage:
			pMsg, ok := msg.(*BlockFetchRespMsg)
			if !ok {
				return nil, fmt.Errorf("expect request type: BlockFetchMessage")
			}
			return pMsg.BlockData, nil
		}
	case <-t.C:
		return nil, fmt.Errorf("timeout fetch block %d from peer %d", blkNum, self.peerIdx)
	case <-self.server.quitC:
		return nil, fmt.Errorf("peer syncing %d quit, failed fetching Block %d", self.peerIdx, blkNum)
	}
	return nil, fmt.Errorf("failed to get Block %d from peer %d", blkNum, self.peerIdx)
}

func (self *PeerSyncer) requestBlockInfo(startBlkNum uint32) ([]*BlockInfo_, error) {
	msg := self.server.constructBlockInfoFetchMsg(startBlkNum)
	self.server.msgSendC <- &SendMsgEvent{
		ToPeer: self.peerIdx,
		Msg:    msg,
	}

	t := time.NewTimer(time.Duration(atomic.LoadInt64(&makeProposalTimeout) * 2))
	defer t.Stop()

	select {
	case msg := <-self.msgC:
		if msg == nil {
			return nil, fmt.Errorf("nil blockinfo fetch rsp msg received")
		}
		switch msg.Type() {
		case BlockInfoFetchRespMessage:
			pMsg, ok := msg.(*BlockInfoFetchRespMsg)
			if !ok {
				return nil, fmt.Errorf("expect request type: BlockInfoFetchRespMessage")
			}
			return pMsg.Blocks, nil
		}
	case <-t.C:
		return nil, fmt.Errorf("timeout fetch blockInfo %d from peer %d", startBlkNum, self.peerIdx)
	case <-self.server.quitC:
		return nil, fmt.Errorf("peer syncer %d - %d quit, failed fetching BlockInfo %d",
			self.server.Index, self.peerIdx, startBlkNum)
	}
	return nil, nil
}

func (self *PeerSyncer) fetchedBlock(blkNum uint32, block *Block) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if blkNum == self.nextReqBlkNum {
		self.server.syncer.blockFromPeerC <- &BlockMsgFromPeer{
			fromPeer: self.peerIdx,
			block:    block,
		}
		self.nextReqBlkNum++
	}

	return nil
}
