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
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
)

type SyncCheckReq struct {
	msg      ConsensusMsg
	peerIdx  uint32
	blockNum uint64
}

type BlockSyncReq struct {
	targetPeers    []uint32
	startBlockNum  uint64
	targetBlockNum uint64 // targetBlockNum == 0, as one cancel syncing request
}

type PeerSyncer struct {
	lock          sync.Mutex
	peerIdx       uint32
	nextReqBlkNum uint64
	targetBlkNum  uint64
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
	nextReqBlkNum     uint64
	targetBlkNum      uint64

	syncCheckReqC  chan *SyncCheckReq
	blockSyncReqC  chan *BlockSyncReq
	syncMsgC       chan *SyncMsg // receive syncmsg from server
	blockFromPeerC chan *BlockMsgFromPeer

	peers         map[uint32]*PeerSyncer
	pendingBlocks map[uint64]BlockFromPeers // index by blockNum
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
		pendingBlocks:     make(map[uint64]BlockFromPeers),
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
	self.pendingBlocks = make(map[uint64]BlockFromPeers)
}

func (self *Syncer) run() {
	self.server.quitWg.Add(1)
	defer self.server.quitWg.Done()

	for {
		select {
		case <-self.syncCheckReqC:
		case req := <-self.blockSyncReqC:
			if req.targetBlockNum == 0 {
				// cancel fetcher for peer
				for _, id := range req.targetPeers {
					self.cancelFetcherForPeer(self.peers[id])
				}
				continue
			}

			log.Infof("server %d, got sync req(%d, %d) to %d",
				self.server.Index, req.startBlockNum, req.targetBlockNum, req.targetPeers)
			if req.startBlockNum <= self.server.GetCommittedBlockNo() {
				req.startBlockNum = self.server.GetCommittedBlockNo() + 1
				log.Infof("server %d, sync req start change to %d",
					self.server.Index, req.startBlockNum)
			}
			if err := self.onNewBlockSyncReq(req); err != nil {
				log.Errorf("server %d failed to handle new block sync req: %s", self.server.Index, err)
			}

		case syncMsg := <-self.syncMsgC:
			if p, present := self.peers[syncMsg.fromPeer]; present {
				if p.active {
					p.msgC <- syncMsg.msg
				} else {
					// report err
					p.msgC <- nil
				}
			} else {
				// report error
			}

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
				if self.nextReqBlkNum <= uint64(ledger.DefLedger.GetCurrentBlockHeight()) {
					blk, _ = self.server.chainStore.GetBlock(self.nextReqBlkNum)
				}
				if blk == nil {
					blk = self.blockConsensusDone(self.pendingBlocks[self.nextReqBlkNum])
				}

				if blk == nil {
					break
				}
				prevHash := blk.getPrevBlockHash()
				log.Debugf("server %d syncer, sealed block %d, proposer %d, prevhash: %s",
					self.server.Index, self.nextReqBlkNum, blk.getProposer(), hex.EncodeToString(prevHash.ToArray()[:4]))
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
	for proposerId, cnt := range proposers {
		if cnt > int(self.server.config.C) {
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

func (self *Syncer) isActive() bool {
	return self.nextReqBlkNum <= self.targetBlkNum
}

func (self *Syncer) startPeerSyncer(syncer *PeerSyncer, targetBlkNum uint64) error {

	syncer.lock.Lock()
	defer syncer.lock.Unlock()

	if targetBlkNum > syncer.targetBlkNum {
		syncer.targetBlkNum = targetBlkNum
	}
	if syncer.targetBlkNum > syncer.nextReqBlkNum && !syncer.active {
		syncer.active = true
		go func() {
			syncer.run()
		}()
	}

	return nil
}

func (self *Syncer) cancelFetcherForPeer(peer *PeerSyncer) error {
	if peer == nil {
		return nil
	}

	peer.lock.Lock()
	defer peer.lock.Unlock()

	// TODO

	return nil
}

func (self *Syncer) onNewBlockSyncReq(req *BlockSyncReq) error {
	if req.startBlockNum < self.nextReqBlkNum {
		log.Errorf("server %d new blockSyncReq startblkNum %d vs %d",
			self.server.Index, req.startBlockNum, self.nextReqBlkNum)
	}
	if req.targetBlockNum <= self.targetBlkNum {
		return nil
	}
	if self.nextReqBlkNum == 1 {
		self.nextReqBlkNum = req.startBlockNum
	}
	self.targetBlkNum = req.targetBlockNum
	peers := req.targetPeers
	if len(peers) == 0 {
		for p := range self.peers {
			peers = append(peers, p)
		}
	}

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

	return nil
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
	blkProposers := make(map[uint64]uint32)
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
		for _, p := range self.server.msgPool.GetProposalMsgs(blkNum) {
			m, ok := p.(*blockProposalMsg)
			if !ok {
				panic("")
			}
			if m.Block.getProposer() == blkProposers[blkNum] {
				proposalBlock = m.Block
				break
			}
		}

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

func (self *PeerSyncer) requestBlock(blkNum uint64) (*Block, error) {
	msg, err := self.server.constructBlockFetchMsg(blkNum)
	if err != nil {
		return nil, err
	}
	self.server.msgSendC <- &SendMsgEvent{
		ToPeer: self.peerIdx,
		Msg:    msg,
	}

	t := time.NewTimer(makeProposalTimeout * 2)
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
				// log error
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

func (self *PeerSyncer) requestBlockInfo(startBlkNum uint64) ([]*BlockInfo_, error) {
	msg, err := self.server.constructBlockInfoFetchMsg(startBlkNum)
	if err != nil {
		return nil, err
	}

	self.server.msgSendC <- &SendMsgEvent{
		ToPeer: self.peerIdx,
		Msg:    msg,
	}

	t := time.NewTimer(makeProposalTimeout * 2)
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
				// log error
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

func (self *PeerSyncer) fetchedBlock(blkNum uint64, block *Block) error {
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
