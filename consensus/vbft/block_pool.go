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
	"bytes"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
)

type BlockList []*Block

var errDupProposal = errors.New("multi proposal from same proposer")
var errDupEndorse = errors.New("multi endorsement from same endorser")
var errDupCommit = errors.New("multi commit from same committer")

type CandidateInfo struct {
	// server endorsed proposals
	EndorsedProposal      *blockProposalMsg
	EndorsedEmptyProposal *blockProposalMsg

	// server committed proposals (one of them must be nil)
	CommittedProposal      *blockProposalMsg
	CommittedEmptyProposal *blockProposalMsg

	commitDone bool

	// server sealed block for this round
	SealedBlock     *Block
	SealedBlockHash common.Uint256 // updated when block sealed
	Participants    []uint32       // updated when block sealed

	// candidate msgs for this round
	Proposals   []*blockProposalMsg
	EndorseMsgs []*blockEndorseMsg
	CommitMsgs  []*blockCommitMsg
}

type BlockPool struct {
	lock       sync.RWMutex
	HistoryLen uint64

	chainStore      *ChainStore
	candidateBlocks map[uint64]*CandidateInfo // indexed by blockNum
}

func newBlockPool(historyLen uint64, store *ChainStore) (*BlockPool, error) {
	pool := &BlockPool{
		HistoryLen:      historyLen,
		chainStore:      store,
		candidateBlocks: make(map[uint64]*CandidateInfo),
	}

	var blkNum uint64
	if store.GetChainedBlockNum() > historyLen {
		blkNum = store.GetChainedBlockNum() - historyLen
	}

	// load history blocks from chainstore
	for ; blkNum <= store.GetChainedBlockNum(); blkNum++ {
		blk, err := store.GetBlock(blkNum)
		if err != nil {
			return nil, fmt.Errorf("failed to load block %d: %s", blkNum, err)
		}
		h, _ := HashBlock(blk)
		pool.candidateBlocks[blkNum] = &CandidateInfo{
			SealedBlock:     blk,
			SealedBlockHash: h,
		}
	}

	return pool, nil
}

func (pool *BlockPool) getCandidateInfoLocked(blkNum uint64) *CandidateInfo {

	// NOTE: call this function only when pool.lock locked

	if candidate, present := pool.candidateBlocks[blkNum]; !present {
		// new candiateInfo for blockNum
		candidate = &CandidateInfo{
			Proposals:    make([]*blockProposalMsg, 0),
			EndorseMsgs:  make([]*blockEndorseMsg, 0),
			CommitMsgs:   make([]*blockCommitMsg, 0),
			Participants: make([]uint32, 0),
		}
		pool.candidateBlocks[blkNum] = candidate
	}

	return pool.candidateBlocks[blkNum]
}

//
// add proposalMsg to CandidateInfo
//
func (pool *BlockPool) newBlockProposal(msg *blockProposalMsg) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	blkNum := msg.GetBlockNum()
	candidate := pool.getCandidateInfoLocked(blkNum)

	// check dup-proposal from same proposer
	for _, p := range candidate.Proposals {
		if p.Block.getProposer() == msg.Block.getProposer() {
			if bytes.Compare(p.Block.Block.Header.SigData[0], msg.Block.Block.Header.SigData[0]) == 0 {
				return nil
			}
			return errDupProposal
		}
	}

	// add msg to proposals
	candidate.Proposals = append(candidate.Proposals, msg)
	return nil
}

func (pool *BlockPool) getBlockProposals(blkNum uint64) []*blockProposalMsg {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	// check if had endorsed for some proposal
	c := pool.candidateBlocks[blkNum]
	if c == nil {
		return []*blockProposalMsg{}
	}

	return c.Proposals
}

func (pool *BlockPool) endorsedForBlock(blkNum uint64) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	// check if has committed for the block
	if pool.chainStore.GetChainedBlockNum() >= blkNum {
		return true
	}

	// check if had endorsed for some proposal
	c := pool.candidateBlocks[blkNum]
	if c == nil {
		return false
	}

	return c.EndorsedProposal != nil || c.EndorsedEmptyProposal != nil
}

func (pool *BlockPool) getEndorsedProposal(blkNum uint64) (*blockProposalMsg, bool) {
	if !pool.endorsedForBlock(blkNum) {
		return nil, false
	}

	pool.lock.RLock()
	defer pool.lock.RUnlock()

	c := pool.candidateBlocks[blkNum]
	if c == nil {
		return nil, false
	}

	if c.EndorsedProposal != nil {
		return c.EndorsedProposal, false
	} else if c.EndorsedEmptyProposal != nil {
		return c.EndorsedEmptyProposal, true
	}

	return nil, false
}

func (pool *BlockPool) endorsedForEmptyBlock(blkNum uint64) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	// check if had endorsed for some proposal
	c := pool.candidateBlocks[blkNum]
	if c == nil {
		return false
	}

	return c.EndorsedEmptyProposal != nil
}

func (pool *BlockPool) setProposalEndorsed(proposal *blockProposalMsg, forEmpty bool) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	// check candidate Info for blkNum
	blkNum := proposal.GetBlockNum()
	if _, present := pool.candidateBlocks[blkNum]; !present {
		return fmt.Errorf("non-candidates for block %d yet when set endorsed", blkNum)
	}

	// check if had endorsed for some proposal
	c := pool.candidateBlocks[blkNum]
	if !forEmpty {
		if c.EndorsedProposal == nil {
			// no endorsedProposal yet
			c.EndorsedProposal = proposal
			return nil
		} else {
			if c.EndorsedProposal.Block.getProposer() == proposal.Block.getProposer() {
				return nil
			}
			return fmt.Errorf("blk %d had endorsed for %d, skip %d", blkNum,
				c.EndorsedProposal.Block.getProposer(), proposal.Block.getProposer())
		}
	}

	// endorse for empty
	if c.EndorsedEmptyProposal != nil {
		return fmt.Errorf("block %d has endorsed for empty", blkNum)
	}
	c.EndorsedEmptyProposal = proposal
	return nil
}

//
// add endorsement msg to CandidateInfo
//
func (pool *BlockPool) newBlockEndorsement(msg *blockEndorseMsg) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	blkNum := msg.GetBlockNum()
	candidate := pool.getCandidateInfoLocked(blkNum)

	// check dup-endorsement
	for i, e := range candidate.EndorseMsgs {
		if e.Endorser == msg.Endorser {
			if bytes.Compare(e.EndorsedBlockHash.ToArray(), msg.EndorsedBlockHash.ToArray()) != 0 &&
				e.EndorseForEmpty == msg.EndorseForEmpty {
				return errDupEndorse
			}
			if e.EndorseForEmpty {
				// had endorsed for empty proposal, reject this endorsement
				return fmt.Errorf("endorser %d had endorsed for empty proposal", e.Endorser)
			}
			if msg.EndorseForEmpty {
				// remove previous non-empty proposal
				candidate.EndorseMsgs = append(candidate.EndorseMsgs[:i], candidate.EndorseMsgs[i+1:]...)
			}
			break
		}
	}

	// add msg to endorses
	candidate.EndorseMsgs = append(candidate.EndorseMsgs, msg)
	return nil
}

//
// check if has reached consensus for endorse-msg
//
func (pool *BlockPool) endorseDone(blkNum uint64, C uint32) (*blockProposalMsg, uint32, bool) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	endorseCount := make(map[uint32]uint32)
	emptyEndorseCount := 0

	candidate := pool.candidateBlocks[blkNum]
	if candidate == nil {
		return nil, math.MaxUint32, false
	}

	for _, e := range candidate.EndorseMsgs {
		if e.EndorseForEmpty {
			emptyEndorseCount++
			if emptyEndorseCount > int(C) {
				return nil, e.EndorsedProposer, true
			}
		} else {
			endorseCount[e.EndorsedProposer] += 1

			// check if endorse-consensus reached
			if endorseCount[e.EndorsedProposer] > C {
				// find proposal
				for _, p := range candidate.Proposals {
					if p.Block.getProposer() == e.EndorsedProposer {
						return p, e.EndorsedProposer, true
					}
				}

				// consensus reached, but we dont have the proposal, set as not done
				// wait timeout, for proposal msg relay
				return nil, e.EndorsedProposer, true
			}
		}
	}

	return nil, math.MaxUint32, false
}

func (pool *BlockPool) committedForBlock(blockNum uint64) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	// check if has committed for the block
	if pool.chainStore.GetChainedBlockNum() >= blockNum {
		return true
	}

	c := pool.candidateBlocks[blockNum]
	if c == nil {
		return false
	}

	return c.CommittedProposal != nil || c.CommittedEmptyProposal != nil
}

func (pool *BlockPool) setProposalCommitted(proposal *blockProposalMsg, forEmpty bool) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	// check candidate Info
	blkNum := proposal.GetBlockNum()
	c := pool.candidateBlocks[blkNum]
	if c == nil {
		return fmt.Errorf("non-candidates for block %d yet when set commit", blkNum)
	}

	// check if has committed
	if c.CommittedProposal != nil || c.CommittedEmptyProposal != nil {
		return fmt.Errorf("had committed for block %d", blkNum)
	}

	if forEmpty {
		if c.CommittedEmptyProposal != nil && c.CommittedEmptyProposal.Block.getProposer() != proposal.Block.getProposer() {
			return fmt.Errorf("had committed for empty block %d (%d)", blkNum, c.CommittedEmptyProposal.Block.getProposer())
		}
		c.CommittedEmptyProposal = proposal
	} else {
		if c.CommittedProposal != nil && c.CommittedProposal.Block.getProposer() != proposal.Block.getProposer() {
			return fmt.Errorf("had committed for block %d (%d)", blkNum, c.CommittedProposal.Block.getProposer())
		}
		c.CommittedProposal = proposal
	}

	return nil
}

//
// add commit msg to CandidateInfo
//
func (pool *BlockPool) newBlockCommitment(msg *blockCommitMsg) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	blkNum := msg.GetBlockNum()
	candidate := pool.getCandidateInfoLocked(blkNum)

	// check dup-commit
	for _, c := range candidate.CommitMsgs {
		if c.Committer == msg.Committer {
			if c.CommitBlockHash.CompareTo(msg.CommitBlockHash) == 0 {
				return nil
			}
			// one committer, one commit
			return errDupCommit
		}
	}

	// add msg to commit-msgs
	candidate.CommitMsgs = append(candidate.CommitMsgs, msg)
	return nil
}

//
// check if has reached consensus on block-commit
// return
//		@ consensused proposal
//		@ for empty commit
//		@ consensused
//
func (pool *BlockPool) commitDone(blkNum uint64, C uint32) (*blockProposalMsg, bool, bool) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	candidate := pool.candidateBlocks[blkNum]
	if candidate == nil {
		return nil, false, false
	}

	proposer, forEmpty := getCommitConsensus(candidate.CommitMsgs, int(C))
	if proposer != math.MaxUint32 {
		for _, p := range candidate.Proposals {
			if p.Block.getProposer() == proposer {
				candidate.commitDone = true
				return p, forEmpty, true
			}
		}
	}

	return nil, false, false
}

func (pool *BlockPool) isCommitHadDone(blkNum uint64) bool {
	pool.lock.RLock()
	pool.lock.RUnlock()
	candidate := pool.candidateBlocks[blkNum]
	if candidate == nil {
		return false
	}

	return candidate.commitDone
}

func (pool *BlockPool) setBlockSealed(block *Block, forEmpty bool) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	blkNum := block.getBlockNum()
	c := pool.getCandidateInfoLocked(blkNum)
	if c == nil {
		panic(fmt.Errorf("non-candidates for block %d yet when seal block", blkNum))
	}

	if c.SealedBlock != nil {
		if c.SealedBlock.getProposer() == block.getProposer() {
			return nil
		}
		return fmt.Errorf("double seal for block %d", blkNum)
	}

	if !forEmpty {
		c.SealedBlock = block
	} else {
		blk := *block                // copy the block block
		blk.Block.Transactions = nil // remove its payload
		c.SealedBlock = &blk
	}
	c.SealedBlockHash, _ = HashBlock(c.SealedBlock)

	// add block to chain store
	if err := pool.chainStore.AddBlock(c.SealedBlock, c.SealedBlockHash); err != nil {
		return fmt.Errorf("failed to seal block (%d) to chainstore: %s", blkNum, err)
	}

	return nil
}

func (pool *BlockPool) getSealedBlock(blockNum uint64) (*Block, common.Uint256) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	// get from cached candidate blocks
	c := pool.candidateBlocks[blockNum]
	if c != nil {
		if c.SealedBlockHash.CompareTo(common.Uint256{}) != 0 {
			return c.SealedBlock, c.SealedBlockHash
		}
		log.Errorf("nil hash founded in block pool sealed cache, blk: %d", blockNum)
	}

	// get from chainstore
	blk, err := pool.chainStore.GetBlock(blockNum)
	if err != nil {
		log.Errorf("getSealedBlock %d err:%v", blockNum, err)
		return nil, common.Uint256{}
	}
	hash, _ := HashBlock(blk)
	return blk, hash
}

func (pool *BlockPool) findConsensusEmptyProposal(blockNum uint64) (*blockProposalMsg, error) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	c := pool.candidateBlocks[blockNum]
	if c == nil {
		return nil, fmt.Errorf("no candidate msgs for block %d", blockNum)
	}

	msgHashCnt := make(map[uint32]int)
	maxCnt := 0
	var maxEndorsedProposer uint32
	for _, p := range c.EndorseMsgs {
		if !p.EndorseForEmpty {
			continue
		}
		n := msgHashCnt[p.EndorsedProposer] + 1
		if n > maxCnt {
			maxCnt = n
			maxEndorsedProposer = p.EndorsedProposer
		}
		msgHashCnt[p.EndorsedProposer] = n
	}

	if maxCnt > 0 {
		for _, p := range c.Proposals {
			if p.Block.getProposer() == maxEndorsedProposer {
				return p, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to get block %d proposal of proposer %d (endorseCnt: %d)",
		blockNum, maxEndorsedProposer, maxCnt)
}

func (pool *BlockPool) onBlockSealed(blockNum uint64) {
	if blockNum <= pool.HistoryLen {
		return
	}

	pool.lock.Lock()
	defer pool.lock.Unlock()

	toFreeCandidates := make([]uint64, 0)
	for n := range pool.candidateBlocks {
		if n < blockNum-pool.HistoryLen {
			toFreeCandidates = append(toFreeCandidates, n)
		}
	}
	for _, n := range toFreeCandidates {
		delete(pool.candidateBlocks, n)
	}
}
