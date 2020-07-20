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
	"github.com/ontio/ontology/core/store/overlaydb"
)

type BlockList []*Block

var errDupProposal = errors.New("multi proposal from same proposer")
var errDupEndorse = errors.New("multi endorsement from same endorser")
var errDupCommit = errors.New("multi commit from same committer")

type CandidateEndorseSigInfo struct {
	EndorsedProposer uint32
	Signature        []byte
	ForEmpty         bool
	CrossChainMsgSig []byte
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
	HistoryLen uint32

	server          *Server
	chainStore      *ChainStore
	candidateBlocks map[uint32]*CandidateInfo // indexed by blockNum
}

func newBlockPool(server *Server, historyLen uint32, store *ChainStore) (*BlockPool, error) {
	pool := &BlockPool{
		server:          server,
		HistoryLen:      historyLen,
		chainStore:      store,
		candidateBlocks: make(map[uint32]*CandidateInfo),
	}

	var blkNum uint32
	if store.GetChainedBlockNum() > historyLen {
		blkNum = store.GetChainedBlockNum() - historyLen
	}

	// load history blocks from chainstore
	for ; blkNum <= store.GetChainedBlockNum(); blkNum++ {
		blk, err := store.getBlock(blkNum)
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

	pool.candidateBlocks = make(map[uint32]*CandidateInfo)
}

func (pool *BlockPool) getCandidateInfoLocked(blkNum uint32) *CandidateInfo {
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
			if bytes.Compare(p.BlockProposerSig, msg.BlockProposerSig) == 0 {
				return nil
			}
			return errDupProposal
		}
	}

	// add msg to proposals
	candidate.Proposals = append(candidate.Proposals, msg)

	// add endorse-sig
	proposer := msg.Block.getProposer()
	eSig := &CandidateEndorseSigInfo{
		EndorsedProposer: proposer,
		Signature:        msg.BlockProposerSig,
		ForEmpty:         false,
	}
	if msg.Block.Block.Header.Height > 1 && msg.Block.CrossChainMsg != nil {
		eSig.CrossChainMsgSig = msg.Block.CrossChainMsg.SigData[0]
	}
	pool.addBlockEndorsementLocked(msg.GetBlockNum(), proposer, eSig, false)
	return nil
}

func (pool *BlockPool) getBlockProposals(blkNum uint32) []*blockProposalMsg {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	// check if had endorsed for some proposal
	c := pool.candidateBlocks[blkNum]
	if c == nil {
		return []*blockProposalMsg{}
	}

	return c.Proposals
}

func (pool *BlockPool) endorsedForBlock(blkNum uint32) bool {
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

func (pool *BlockPool) getEndorsedProposal(blkNum uint32) (*blockProposalMsg, bool) {
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

func (pool *BlockPool) endorsedForEmptyBlock(blkNum uint32) bool {
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

func (pool *BlockPool) addBlockEndorsementLocked(blkNum uint32, endorser uint32, eSig *CandidateEndorseSigInfo, commitment bool) {
	candidate := pool.getCandidateInfoLocked(blkNum)

	if eSigs, present := candidate.EndorseSigs[endorser]; present && !commitment {
		for _, eSig := range eSigs {
			if eSig.ForEmpty {
				// has endorsed for empty, ignore new endorsement
				return
			}
		}
		if eSig.ForEmpty {
			// add empty endorsement
			candidate.EndorseSigs[endorser] = append(eSigs, eSig)
			return
		}

		// check dup endorsement
		for _, esig := range eSigs {
			if esig.EndorsedProposer == eSig.EndorsedProposer {
				return
			}
		}
		candidate.EndorseSigs[endorser] = append(eSigs, eSig)
	} else {
		candidate.EndorseSigs[endorser] = []*CandidateEndorseSigInfo{eSig}
	}
	return
}

//
// add endorsement msg to CandidateInfo
//
func (pool *BlockPool) newBlockEndorsement(msg *blockEndorseMsg) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	eSig := &CandidateEndorseSigInfo{
		EndorsedProposer: msg.EndorsedProposer,
		Signature:        msg.EndorserSig,
		ForEmpty:         msg.EndorseForEmpty,
		CrossChainMsgSig: msg.CrossChainMsgEndorserSig,
	}
	pool.addBlockEndorsementLocked(msg.GetBlockNum(), msg.Endorser, eSig, false)
}

//
// check if has reached consensus for endorse-msg
//
// return
//		@ endorsable proposer
//		@ for empty commit
//		@ endorsable
//
func (pool *BlockPool) endorseDone(blkNum uint32, C uint32) (uint32, bool, bool) {
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

func (pool *BlockPool) endorseFailed(blkNum uint32, C uint32) bool {
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

func (pool *BlockPool) committedForBlock(blockNum uint32) bool {
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
		// old version of committer msg is nil, compatible old version
		if msg.CrossChainMsgCommitterSig != nil {
			if crossChainMsgSig, present := msg.CrossChainMsgEndorserSig[endorser]; present {
				eSig.CrossChainMsgSig = crossChainMsgSig
			}
		}

		pool.addBlockEndorsementLocked(blkNum, endorser, eSig, false)
	}

	// add committer sig
	pool.addBlockEndorsementLocked(blkNum, msg.Committer, &CandidateEndorseSigInfo{
		EndorsedProposer: msg.BlockProposer,
		Signature:        msg.CommitterSig,
		ForEmpty:         msg.CommitForEmpty,
		CrossChainMsgSig: msg.CrossChainMsgCommitterSig,
	}, true)

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
// Note: Attentions on lock contention.
// Only shared-lock for this function, because this function will also acquires shared-lock on peer-pool.
//
func (pool *BlockPool) commitDone(blkNum uint32, C uint32, N uint32) (uint32, bool, bool) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	candidate := pool.candidateBlocks[blkNum]
	if candidate == nil {
		return math.MaxUint32, false, false
	}

	// check consensus with commit msgs
	proposer, forEmpty := getCommitConsensus(candidate.CommitMsgs, int(C), int(N))

	if proposer == math.MaxUint32 {
		// check consensus with endorse sigs
		// enforce signature quorum if checking commit-consensus base on signature count
		C = N - (N-1)/3 - 1
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
					if endorseCnt[sig.EndorsedProposer] > C {
						proposer = sig.EndorsedProposer
						if !forEmpty {
							forEmpty = emptyCnt > C
						}
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
		return proposer, forEmpty, true
	}

	return math.MaxUint32, false, false
}

//
// @ set BlockPool as committed for given BlockNum
//
// Note: setCommitDone supposed to be called after commitDone.
// Because setCommitDone requires exclusive lock, this function is provided separately.
//
func (pool *BlockPool) setCommitDone(blkNum uint32) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	candidate := pool.candidateBlocks[blkNum]
	if candidate != nil {
		candidate.commitDone = true
	}
}

func (pool *BlockPool) isCommitHadDone(blkNum uint32) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
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
	if !forEmpty {
		bookkeepers = append(bookkeepers, proposerPk)
		sigData = append(sigData, block.Block.Header.SigData[0])
	} else {
		if block.EmptyBlock == nil {
			return fmt.Errorf("block has no empty candidate")
		}
		bookkeepers = append(bookkeepers, proposerPk)
		sigData = append(sigData, block.EmptyBlock.Header.SigData[0])
	}

	// add endorsers' sig
	for endorser, eSigs := range c.EndorseSigs {
		for _, sig := range eSigs {
			if sig.EndorsedProposer == proposer && sig.ForEmpty == forEmpty && endorser != proposer {
				endoresrPk := pool.server.peerPool.GetPeerPubKey(endorser)
				if endoresrPk != nil {
					bookkeepers = append(bookkeepers, endoresrPk)
					sigData = append(sigData, sig.Signature)
					if block.CrossChainMsg != nil {
						block.CrossChainMsg.SigData = append(block.CrossChainMsg.SigData, sig.CrossChainMsgSig)
					}
				}
				break
			}
		}
	}
	if !forEmpty {
		block.Block.Header.Bookkeepers = bookkeepers
		block.Block.Header.SigData = sigData
	} else {
		block.EmptyBlock.Header.Bookkeepers = bookkeepers
		block.EmptyBlock.Header.SigData = sigData
	}

	return nil
}

func (pool *BlockPool) setBlockSealed(block *Block, forEmpty bool, sigdata bool) error {
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
	if sigdata {
		if err := pool.addSignaturesToBlockLocked(block, forEmpty); err != nil {
			return fmt.Errorf("failed to add sig to block: %s", err)
		}
	}
	sealedBlock := &Block{
		Info:               block.Info,
		PrevExecMerkleRoot: block.PrevExecMerkleRoot,
		CrossChainMsg:      block.CrossChainMsg,
	}
	if !forEmpty {
		// remove empty block
		sealedBlock.Block = block.Block
	} else {
		// replace with empty block
		sealedBlock.Block = block.EmptyBlock
	}
	c.SealedBlock = sealedBlock
	// add block to chain store
	if err := pool.chainStore.AddBlock(sealedBlock); err != nil {
		return fmt.Errorf("failed to seal block (%d) to chainstore: %s", blkNum, err)
	}
	stateRoot, err := pool.chainStore.getExecMerkleRoot(pool.chainStore.GetChainedBlockNum())
	if err != nil {
		log.Errorf("setBlockSealed blk %d failed:%s", blkNum, err)
		return nil
	}
	if blocksubmitMsg, _ := pool.server.constructBlockSubmitMsg(pool.chainStore.GetChainedBlockNum(), stateRoot); blocksubmitMsg != nil {
		pool.server.broadcast(blocksubmitMsg)
		pool.server.makeBlockSubmit(pool.chainStore.GetChainedBlockNum())
	}
	return nil
}

func (pool *BlockPool) getSealedBlock(blockNum uint32) (*Block, common.Uint256) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	// get from cached candidate blocks
	c := pool.candidateBlocks[blockNum]
	if c != nil && c.SealedBlock != nil {
		h := c.SealedBlock.Block.Hash()
		if bytes.Compare(h[:], common.UINT256_EMPTY[:]) != 0 {
			return c.SealedBlock, h
		}
		log.Errorf("empty hash founded in block pool sealed cache, blk: %d", blockNum)
	}

	// get from chainstore
	blk, err := pool.chainStore.getBlock(blockNum)
	if err != nil {
		log.Errorf("getSealedBlock %d err:%v", blockNum, err)
		return nil, common.Uint256{}
	}
	return blk, blk.Block.Hash()
}

func (pool *BlockPool) getChainedBlock(blockNum uint32) (*Block, common.Uint256) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	// get from chainstore
	blk, err := pool.chainStore.getBlock(blockNum)
	if err != nil {
		log.Errorf("getSealedBlock %d err:%v", blockNum, err)
		return nil, common.Uint256{}
	}
	return blk, blk.Block.Hash()
}

func (pool *BlockPool) findConsensusEmptyProposal(blockNum uint32) (*blockProposalMsg, error) {
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

func (pool *BlockPool) onBlockSealed(blockNum uint32) {
	if blockNum <= pool.HistoryLen {
		return
	}

	pool.lock.Lock()
	defer pool.lock.Unlock()

	toFreeCandidates := make([]uint32, 0)
	for n := range pool.candidateBlocks {
		if n < blockNum-pool.HistoryLen {
			toFreeCandidates = append(toFreeCandidates, n)
		}
	}
	for _, n := range toFreeCandidates {
		delete(pool.candidateBlocks, n)
	}
}

func (pool *BlockPool) getExecMerkleRoot(blkNum uint32) (common.Uint256, error) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	return pool.chainStore.getExecMerkleRoot(blkNum)
}

func (pool *BlockPool) getCrossStatesRoot(blkNum uint32) (common.Uint256, error) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	return pool.chainStore.getCrossStatesRoot(blkNum)
}

func (pool *BlockPool) getExecWriteSet(blkNum uint32) *overlaydb.MemDB {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	return pool.chainStore.getExecWriteSet(blkNum)
}

func (pool *BlockPool) submitBlock(blkNum uint32) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	return pool.chainStore.submitBlock(blkNum)
}

func (pool *BlockPool) ReloadFromLedger() {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	pool.chainStore.ReloadFromLedger()
}
