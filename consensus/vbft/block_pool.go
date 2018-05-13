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

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
)

type BlockList []*Block

var errDupProposal = errors.New("multi proposal from same proposer")
var errDupEndorse = errors.New("multi endorsement from same endorser")
var errDupCommit = errors.New("multi commit from same committer")

type CandidateEndorseSigInfo struct {
	EndorsedProposer uint32
	Signature        []byte
	ForEmpty         bool
}

type CandidateInfo struct {
	// server endorsed proposals
	EndorsedProposal      *blockProposalMsg
	EndorsedEmptyProposal *blockProposalMsg

	// server committed proposals (one of them must be nil)
	CommittedProposal      *blockProposalMsg
	CommittedEmptyProposal *blockProposalMsg

	commitDone bool

	// server sealed block for this round
	SealedBlock *Block

	// candidate msgs for this round
	Proposals  []*blockProposalMsg
	CommitMsgs []*blockCommitMsg

	// indexed by endorserIndex
	EndorseSigs map[uint32][]*CandidateEndorseSigInfo
}

type BlockPool struct {
	lock       sync.RWMutex
	HistoryLen uint64

	server          *Server
	chainStore      *ChainStore
	candidateBlocks map[uint64]*CandidateInfo // indexed by blockNum
}

func newBlockPool(server *Server, historyLen uint64, store *ChainStore) (*BlockPool, error) {
	pool := &BlockPool{
		server:          server,
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
		pool.candidateBlocks[blkNum] = &CandidateInfo{
			SealedBlock: blk,
		}
	}

	return pool, nil
}

func (pool *BlockPool) clean() {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.candidateBlocks = make(map[uint64]*CandidateInfo)
}

func (pool *BlockPool) getCandidateInfoLocked(blkNum uint64) *CandidateInfo {

	// NOTE: call this function only when pool.lock locked

	if candidate, present := pool.candidateBlocks[blkNum]; !present {
		// new candiateInfo for blockNum
		candidate = &CandidateInfo{
			Proposals:   make([]*blockProposalMsg, 0),
			CommitMsgs:  make([]*blockCommitMsg, 0),
			EndorseSigs: make(map[uint32][]*CandidateEndorseSigInfo),
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

	if c.EndorsedEmptyProposal != nil {
		return c.EndorsedEmptyProposal, true
	} else if c.EndorsedProposal != nil {
		return c.EndorsedProposal, false
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

func (pool *BlockPool) addBlockEndorsementLocked(blkNum uint64, endorser uint32, eSig *CandidateEndorseSigInfo) error {
	candidate := pool.getCandidateInfoLocked(blkNum)

	if eSigs, present := candidate.EndorseSigs[endorser]; present {
		for _, eSig := range eSigs {
			if eSig.ForEmpty {
				// has endorsed for empty, ignore new endorsement
				return nil
			}
		}
		if eSig.ForEmpty {
			// add empty endorsement
			candidate.EndorseSigs[endorser] = append(eSigs, eSig)
			return nil
		}

		// check dup endorsement
		for _, esig := range eSigs {
			if esig.EndorsedProposer == eSig.EndorsedProposer {
				return nil
			}
		}
		candidate.EndorseSigs[endorser] = append(eSigs, eSig)
	} else {
		candidate.EndorseSigs[endorser] = []*CandidateEndorseSigInfo{eSig}
	}

	return nil
}

//
// add endorsement msg to CandidateInfo
//
func (pool *BlockPool) newBlockEndorsement(msg *blockEndorseMsg) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	eSig := &CandidateEndorseSigInfo{
		EndorsedProposer: msg.EndorsedProposer,
		Signature:        msg.EndorserSig,
		ForEmpty:         msg.EndorseForEmpty,
	}
	return pool.addBlockEndorsementLocked(msg.GetBlockNum(), msg.Endorser, eSig)
}

//
// check if has reached consensus for endorse-msg
//
// return
//		@ endorsable proposer
//		@ for empty commit
//		@ endorsable
//
func (pool *BlockPool) endorseDone(blkNum uint64, C uint32) (uint32, bool, bool) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	endorseCount := make(map[uint32]uint32)
	emptyEndorseCount := 0

	candidate := pool.candidateBlocks[blkNum]
	if candidate == nil {
		return math.MaxUint32, false, false
	}

	if uint32(len(candidate.EndorseSigs)) < C+1 {
		return math.MaxUint32, false, false
	}

	for _, eSigs := range candidate.EndorseSigs {
		for _, esig := range eSigs {
			if esig.ForEmpty {
				emptyEndorseCount++
				if emptyEndorseCount > int(C) {
					// FIXME: endorsedProposer need fix
					return esig.EndorsedProposer, true, true
				}
			} else {
				endorseCount[esig.EndorsedProposer] += 1
				// check if endorse-consensus reached
				if endorseCount[esig.EndorsedProposer] > C {
					return esig.EndorsedProposer, false, true
				}
			}
		}
	}

	return math.MaxUint32, false, false
}

func (pool *BlockPool) endorseFailed(blkNum uint64, C uint32) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	proposalCount := make(map[uint32]uint32)
	endorserCount := make(map[uint32]uint32)
	candidate := pool.candidateBlocks[blkNum]
	if candidate == nil {
		return false
	}

	if uint32(len(candidate.EndorseSigs)) < C+1 {
		return false
	}

	var emptyEndorseCnt uint32
	for endorser, eSigs := range candidate.EndorseSigs {
		for _, esig := range eSigs {
			if !esig.ForEmpty {
				proposalCount[esig.EndorsedProposer] += 1
				if proposalCount[esig.EndorsedProposer] > C+1 {
					return false
				}
			} else {
				emptyEndorseCnt++
			}
		}
		endorserCount[endorser] += 1
	}

	if uint32(len(proposalCount)) > C+1 {
		return true
	}
	if emptyEndorseCnt > C {
		return true
	}

	l := 2*C + 1 - uint32(len(endorserCount))
	for _, v := range proposalCount {
		if v+l > C {
			return false
		}
	}

	return true
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
			if bytes.Compare(c.CommitBlockHash[:], msg.CommitBlockHash[:]) == 0 {
				return nil
			}
			// one committer, one commit
			return errDupCommit
		}
	}

	// add all endorse sigs
	for endorser, sig := range msg.EndorsersSig {
		eSig := &CandidateEndorseSigInfo{
			EndorsedProposer: msg.BlockProposer,
			Signature:        sig,
			ForEmpty:         msg.CommitForEmpty,
		}
		if err := pool.addBlockEndorsementLocked(blkNum, endorser, eSig); err != nil {
			return fmt.Errorf("failed to verify endorse sig from %d: %s", endorser, err)
		}
	}

	// add committer sig
	pool.addBlockEndorsementLocked(blkNum, msg.Committer, &CandidateEndorseSigInfo{
		EndorsedProposer: msg.BlockProposer,
		Signature:        msg.CommitterSig,
		ForEmpty:         msg.CommitForEmpty,
	})

	// add msg to commit-msgs
	candidate.CommitMsgs = append(candidate.CommitMsgs, msg)
	return nil
}

//
// check if has reached consensus on block-commit
// return
//		@ consensused proposer
//		@ for empty commit
//		@ consensused
//
func (pool *BlockPool) commitDone(blkNum uint64, C uint32) (uint32, bool, bool) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	candidate := pool.candidateBlocks[blkNum]
	if candidate == nil {
		return math.MaxUint32, false, false
	}

	// check consensus with commit msgs
	proposer, forEmpty := getCommitConsensus(candidate.CommitMsgs, int(C))

	if proposer == math.MaxUint32 {
		// check consensus with endorse sigs
		var emptyCnt uint32
		endorseCnt := make(map[uint32]uint32) // proposer -> endorsed-cnt
		for endorser, eSigs := range candidate.EndorseSigs {
			// check if from endorser
			if !pool.server.isEndorser(blkNum, endorser) {
				for _, sig := range eSigs {
					if sig.ForEmpty {
						emptyCnt++
					}
				}
			}

			for _, sig := range eSigs {
				if sig.ForEmpty {
					emptyCnt++
				} else {
					endorseCnt[sig.EndorsedProposer] += 1
					if endorseCnt[sig.EndorsedProposer] > C+1 {
						proposer = sig.EndorsedProposer
						forEmpty = emptyCnt > C+1
						break
					}
				}
			}

			if proposer != math.MaxUint32 {
				break
			}
		}
	}

	if proposer != math.MaxUint32 {
		candidate.commitDone = true
		return proposer, forEmpty, true
	}

	return math.MaxUint32, false, false
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

func (pool *BlockPool) addSignaturesToBlockLocked(block *Block, forEmpty bool) error {

	blkNum := block.getBlockNum()
	c := pool.getCandidateInfoLocked(blkNum)
	if c == nil {
		panic(fmt.Errorf("non-candidates for block %d yet when seal block", blkNum))
	}

	bookkeepers := make([]keypair.PublicKey, 0)
	sigData := make([][]byte, 0)

	// add proposer sig
	proposer := block.getProposer()
	proposerPk := pool.server.peerPool.GetPeerPubKey(proposer)
	bookkeepers = append(bookkeepers, proposerPk)
	if !forEmpty {
		sigData = append(sigData, block.Block.Header.SigData[0])
	} else {
		sigData = append(sigData, block.Block.Header.SigData[1])
	}

	// add endorsers' sig
	for endorser, eSigs := range c.EndorseSigs {
		for _, sig := range eSigs {
			if sig.EndorsedProposer == proposer && sig.ForEmpty == forEmpty {
				endoresrPk := pool.server.peerPool.GetPeerPubKey(endorser)
				if endoresrPk != nil {
					bookkeepers = append(bookkeepers, endoresrPk)
					sigData = append(sigData, sig.Signature)
				}
				break
			}
		}
	}

	block.Block.Header.Bookkeepers = bookkeepers
	block.Block.Header.SigData = sigData

	return nil
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

	if err := pool.addSignaturesToBlockLocked(block, forEmpty); err != nil {
		return fmt.Errorf("failed to add sig to block: %s", err)
	}

	if !forEmpty {
		c.SealedBlock = block
	} else {
		block.Block.Transactions = nil // remove its payload
		c.SealedBlock = block
	}

	// add block to chain store
	if err := pool.chainStore.AddBlock(c.SealedBlock); err != nil {
		return fmt.Errorf("failed to seal block (%d) to chainstore: %s", blkNum, err)
	}

	return nil
}

func (pool *BlockPool) getSealedBlock(blockNum uint64) (*Block, common.Uint256) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	// get from cached candidate blocks
	c := pool.candidateBlocks[blockNum]
	if c != nil && c.SealedBlock != nil {
		h, _ := HashBlock(c.SealedBlock)
		if bytes.Compare(h[:], common.UINT256_EMPTY[:]) != 0 {
			return c.SealedBlock, h
		}
		log.Errorf("empty hash founded in block pool sealed cache, blk: %d", blockNum)
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
	for _, eSigs := range c.EndorseSigs {
		for _, esig := range eSigs {
			if !esig.ForEmpty {
				continue
			}
			n := msgHashCnt[esig.EndorsedProposer] + 1
			if n > maxCnt {
				maxCnt = n
				maxEndorsedProposer = esig.EndorsedProposer
			}
			msgHashCnt[esig.EndorsedProposer] = n
		}
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
