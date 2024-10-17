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
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/store"
)

var errDupProposal = errors.New("multi proposal from same proposer")
var errDupEndorse = errors.New("multi endorsement from same endorser")
var errDupCommit = errors.New("multi commit from same committer")

type CandidateEndorseSigInfo struct {
	BlockHash        common.Uint256
	EndorsedProposer uint32
	Signature        []byte
	ForEmpty         bool
	CrossChainMsgSig []byte
}

type CandidateInfo struct {
	// server endorsed proposals
	EndorseMsg      *blockEndorseMsg
	EndorseEmptyMsg *blockEndorseMsg
	SelfCommitMsg   *blockCommitMsg

	// server sealed block for this round
	SealedBlock           *VbftBlock
	SealedBlockExecResult *store.ExecuteResult

	// candidate msgs for this round
	Proposals  []*blockProposalMsg
	CommitMsgs []*blockCommitMsg

	// indexed by endorserIndex
	EndorseSigs map[uint32][]*CandidateEndorseSigInfo
}

func (self *CandidateInfo) GetEndorseSigInfos(blockHash common.Uint256) map[uint32]*CandidateEndorseSigInfo {
	result := make(map[uint32]*CandidateEndorseSigInfo)
	for index, v := range self.EndorseSigs {
		for _, info := range v {
			if info.BlockHash == blockHash {
				result[index] = info
				break
			}
		}
	}
	return result
}

func (pool *BlockPool) GetEndorseSigInfos(blkNum uint32, blockHash common.Uint256) map[uint32]*CandidateEndorseSigInfo {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	return pool.getCandidateInfoLocked(blkNum).GetEndorseSigInfos(blockHash)
}

func (self *CandidateInfo) String() string {
	proposals := make(map[uint32][]uint32)
	for _, p := range self.Proposals {
		proposals[p.Block.Info.Proposer] = nil
	}
	for e, sigs := range self.EndorseSigs {
		for _, s := range sigs {
			proposals[s.EndorsedProposer] = append(proposals[s.EndorsedProposer], e)
		}
	}
	v, _ := json.Marshal(proposals)
	return string(v)
}

type BlockPool struct {
	lock       sync.RWMutex
	HistoryLen uint32

	chainStore      *ChainStore
	candidateBlocks map[uint32]*CandidateInfo // indexed by blockNum
}

func newBlockPool(historyLen uint32, store *ChainStore) (*BlockPool, error) {
	pool := &BlockPool{
		HistoryLen:      historyLen,
		chainStore:      store,
		candidateBlocks: make(map[uint32]*CandidateInfo),
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
	if _, present := pool.candidateBlocks[blkNum]; !present {
		// new candiateInfo for blockNum
		candidate := &CandidateInfo{
			EndorseSigs: make(map[uint32][]*CandidateEndorseSigInfo),
		}
		pool.candidateBlocks[blkNum] = candidate
	}

	return pool.candidateBlocks[blkNum]
}

func (pool *BlockPool) AddBlockProposal(msg *blockProposalMsg) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	return pool.getCandidateInfoLocked(msg.GetBlockNum()).AddBlockProposal(msg)
}

func (candidate *CandidateInfo) AddBlockProposal(msg *blockProposalMsg) error {
	// check dup-proposal from same proposer
	proposer := msg.Block.getProposer()
	for _, p := range candidate.Proposals {
		if p.Block.getProposer() == proposer {
			if bytes.Equal(p.BlockProposerSig, msg.BlockProposerSig) {
				return nil
			}
			return errDupProposal
		}
	}

	// add msg to proposals
	candidate.Proposals = append(candidate.Proposals, msg)

	// add endorse-sig
	eSig := &CandidateEndorseSigInfo{
		BlockHash:        msg.Block.Block.Hash(),
		EndorsedProposer: proposer,
		Signature:        msg.BlockProposerSig,
		ForEmpty:         false,
	}
	if msg.Block.Block.Header.Height > 1 && msg.Block.CrossChainMsg != nil {
		eSig.CrossChainMsgSig = msg.Block.CrossChainMsg.SigData[0]
	}
	candidate.addBlockEndorsementLocked(proposer, eSig, false)
	return nil
}

func (self *BlockPool) Info(blkNum uint32) string {
	return self.getCandidateInfoLocked(blkNum).String()
}

func (pool *BlockPool) GetSelfCommitMsg(blkNum uint32) *blockCommitMsg {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	return pool.getCandidateInfoLocked(blkNum).SelfCommitMsg
}

func (pool *BlockPool) GetBlockProposal(blkNum, proposer uint32) *blockProposalMsg {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	return pool.getCandidateInfoLocked(blkNum).GetBlockProposal(proposer)
}

func (pool *BlockPool) GetBlockProposals(blkNum uint32) []*blockProposalMsg {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	return pool.getCandidateInfoLocked(blkNum).Proposals
}

func (self *CandidateInfo) GetBlockProposal(proposer uint32) *blockProposalMsg {
	for _, p := range self.Proposals {
		if p.Block.getProposer() == proposer {
			return p
		}
	}
	return nil
}

func (pool *BlockPool) HasEndorsedForBlock(blkNum uint32) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	c := pool.candidateBlocks[blkNum]
	return c != nil && (c.EndorseMsg != nil || c.EndorseEmptyMsg != nil)
}

func (pool *BlockPool) GetSelfEndorseMsg(blkNum uint32) *blockEndorseMsg {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	c := pool.getCandidateInfoLocked(blkNum)

	if c.EndorseEmptyMsg != nil {
		return c.EndorseEmptyMsg
	}
	return c.EndorseMsg
}

func (pool *BlockPool) HasEndorsedForEmptyBlock(blkNum uint32) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	return pool.getCandidateInfoLocked(blkNum).EndorseEmptyMsg != nil
}

func (pool *BlockPool) setProposalEndorsed(endorse *blockEndorseMsg) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	c := pool.getCandidateInfoLocked(endorse.GetBlockNum())
	if endorse.EndorseForEmpty {
		c.EndorseEmptyMsg = endorse
		return
	}
	c.EndorseMsg = endorse
}

func (candidate *CandidateInfo) addBlockEndorsementLocked(endorser uint32, eSig *CandidateEndorseSigInfo, commitment bool) {
	if commitment {
		candidate.EndorseSigs[endorser] = []*CandidateEndorseSigInfo{eSig}
		return
	}
	eSigs := candidate.EndorseSigs[endorser]
	for _, eSig := range eSigs {
		if eSig.ForEmpty {
			return // has endorsed for empty, ignore new endorsement
		}
	}
	if !eSig.ForEmpty {
		// check dup endorsement
		for _, esig := range eSigs {
			if esig.EndorsedProposer == eSig.EndorsedProposer {
				return
			}
		}
	}

	candidate.EndorseSigs[endorser] = append(eSigs, eSig)
}

func (pool *BlockPool) AddBlockEndorseMsg(msg *blockEndorseMsg) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	eSig := &CandidateEndorseSigInfo{
		BlockHash:        msg.EndorsedBlockHash,
		EndorsedProposer: msg.EndorsedProposer,
		Signature:        msg.EndorserSig,
		ForEmpty:         msg.EndorseForEmpty,
		CrossChainMsgSig: msg.CrossChainMsgEndorserSig,
	}
	candidate := pool.getCandidateInfoLocked(msg.GetBlockNum())
	err := candidate.CheckBlockHashWithProposal(msg.EndorsedProposer, msg.EndorseForEmpty, msg.EndorsedBlockHash)
	if err != nil {
		return err
	}
	candidate.addBlockEndorsementLocked(msg.Endorser, eSig, false)
	return nil
}

func (self *CandidateInfo) CheckBlockHashWithProposal(proposer uint32, forEmpty bool, hash common.Uint256) error {
	p := self.GetBlockProposal(proposer)
	if p != nil {
		block := p.Block.Block
		if forEmpty {
			block = p.Block.EmptyBlock
		}
		if block == nil || hash != block.Hash() {
			return fmt.Errorf("failed to compare block hash with proposal")
		}
	}
	return nil
}

// check if has reached consensus for endorse-msg
//
// return
//
//	@ endorsable proposer
//	@ for empty commit
//	@ endorsable
func (pool *BlockPool) endorseDone(blkNum uint32, C uint32) (uint32, bool, bool) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	return pool.getCandidateInfoLocked(blkNum).EndorseDone(C)
}

func (candidate *CandidateInfo) EndorseDone(C uint32) (uint32, bool, bool) {
	endorseCount := make(map[uint32]uint32)
	emptyEndorseCount := 0

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
	return pool.getCandidateInfoLocked(blkNum).EndorseFailed(C)
}

func (candidate *CandidateInfo) EndorseFailed(C uint32) bool {
	proposerCount := make(map[uint32]uint32)
	if uint32(len(candidate.EndorseSigs)) < C+1 {
		return false
	}

	var emptyEndorseCnt uint32
	for _, eSigs := range candidate.EndorseSigs {
		for _, esig := range eSigs {
			if !esig.ForEmpty {
				proposerCount[esig.EndorsedProposer] += 1
				if proposerCount[esig.EndorsedProposer] > C+1 {
					return false
				}
			} else {
				emptyEndorseCnt++
			}
		}
	}

	if uint32(len(proposerCount)) > C+1 {
		return true
	}
	if emptyEndorseCnt > C {
		return true
	}

	l := 2*C + 1 - uint32(len(candidate.EndorseSigs))
	for _, v := range proposerCount {
		if v+l > C {
			return false
		}
	}

	return true
}

func (pool *BlockPool) SetProposalCommitted(commit *blockCommitMsg) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	c := pool.getCandidateInfoLocked(commit.BlockNum)
	c.SelfCommitMsg = commit
}

func (pool *BlockPool) AddBlockCommitMsg(msg *blockCommitMsg) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	blkNum := msg.GetBlockNum()
	candidate := pool.getCandidateInfoLocked(blkNum)
	err := candidate.CheckBlockHashWithProposal(msg.BlockProposer, msg.CommitForEmpty, msg.CommitBlockHash)
	if err != nil {
		return err
	}
	return candidate.AddBlockCommitMsg(msg)
}

func (candidate *CandidateInfo) AddBlockCommitMsg(msg *blockCommitMsg) error {
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
			BlockHash:        msg.CommitBlockHash,
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

		candidate.addBlockEndorsementLocked(endorser, eSig, false)
	}

	// add committer sig
	candidate.addBlockEndorsementLocked(msg.Committer, &CandidateEndorseSigInfo{
		BlockHash:        msg.CommitBlockHash,
		EndorsedProposer: msg.BlockProposer,
		Signature:        msg.CommitterSig,
		ForEmpty:         msg.CommitForEmpty,
		CrossChainMsgSig: msg.CrossChainMsgCommitterSig,
	}, true)

	// add msg to commit-msgs
	candidate.CommitMsgs = append(candidate.CommitMsgs, msg)
	return nil
}

// check if has reached consensus on block-commit
// return
//
//	@ consensused proposer
//	@ for empty commit
//	@ consensused
//
// Note: Attentions on lock contention.
// Only shared-lock for this function, because this function will also acquires shared-lock on peer-pool.
func (pool *BlockPool) commitDone(vbftCtx *VbftContext, blkNum uint32) (uint32, bool, bool) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	return pool.getCandidateInfoLocked(blkNum).CommitDone(vbftCtx)
}

func (candidate *CandidateInfo) CommitDone(vbftCtx *VbftContext) (uint32, bool, bool) {
	C := vbftCtx.Config.C
	N := vbftCtx.Config.N
	// check consensus with commit msgs
	proposer, forEmpty := getCommitConsensus(candidate.CommitMsgs, int(C), int(N))

	if proposer == math.MaxUint32 {
		// check consensus with endorse sigs
		// enforce signature quorum if checking commit-consensus base on signature count
		C = N - (N-1)/3
		var emptyCnt uint32
		endorseCnt := make(map[uint32]uint32) // proposer -> endorsed-cnt
		for endorser, eSigs := range candidate.EndorseSigs {
			// check if from endorser
			if !vbftCtx.IsEndorser(endorser) {
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
					if endorseCnt[sig.EndorsedProposer] >= C {
						proposer = sig.EndorsedProposer
						if !forEmpty {
							forEmpty = emptyCnt >= C
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

func (c *CandidateInfo) addSignaturesToBlockLocked(vbftCtx *VbftContext, block *VbftBlock, forEmpty bool) error {
	bookkeepers := make([]keypair.PublicKey, 0)
	sigData := make([][]byte, 0)

	// add proposer sig
	proposer := block.getProposer()
	proposerPk := vbftCtx.GetPeerPubKey(proposer)
	blockToAdd := block.Block
	if forEmpty {
		blockToAdd = block.EmptyBlock
		if block.EmptyBlock == nil {
			return fmt.Errorf("block has no empty candidate")
		}
	}
	bookkeepers = append(bookkeepers, proposerPk)
	sigData = append(sigData, blockToAdd.Header.SigData[0])

	// add endorsers' sig
	for endorser, eSigs := range c.EndorseSigs {
		for _, sig := range eSigs {
			if sig.EndorsedProposer == proposer && sig.ForEmpty == forEmpty && endorser != proposer {
				endoresrPk := vbftCtx.GetPeerPubKey(endorser)
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

	blockToAdd.Header.Bookkeepers = bookkeepers
	blockToAdd.Header.SigData = sigData

	return nil
}

func (pool *BlockPool) checkBlockSign(vbftCtx *VbftContext, block *VbftBlock, forEmpty bool, requiredSigs uint32) bool {
	blkNum := block.getBlockNum()
	c := pool.getCandidateInfoLocked(blkNum)
	proposer := block.getProposer()
	sigData := make([][]byte, 0)
	var blkHash common.Uint256
	if !forEmpty {
		sigData = append(sigData, block.Block.Header.SigData[0])
		blkHash = block.Block.Hash()
	} else {
		if block.EmptyBlock == nil {
			return false
		}
		sigData = append(sigData, block.EmptyBlock.Header.SigData[0])
		blkHash = block.EmptyBlock.Hash()
	}
	for endorser, eSigs := range c.EndorseSigs {
		for _, sig := range eSigs {
			if sig.EndorsedProposer == proposer && sig.BlockHash == blkHash && sig.ForEmpty == forEmpty && endorser != proposer {
				if vbftCtx.GetPeerPubKey(endorser) != nil {
					sigData = append(sigData, sig.Signature)
				}
				break
			}
		}
	}

	return uint32(len(sigData)) >= requiredSigs
}

func (pool *BlockPool) SetBlockSealed(vbftCtx *VbftContext, block *VbftBlock, forEmpty bool, sigdata bool) (*VbftBlock, *store.ExecuteResult, error) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	blkNum := block.getBlockNum()
	c := pool.getCandidateInfoLocked(blkNum)

	if c.SealedBlock != nil {
		if c.SealedBlock.getProposer() == block.getProposer() {
			return c.SealedBlock, c.SealedBlockExecResult, nil
		}
		return nil, nil, fmt.Errorf("double seal for block %d", blkNum)
	}
	if sigdata {
		if err := c.addSignaturesToBlockLocked(vbftCtx, block, forEmpty); err != nil {
			return nil, nil, fmt.Errorf("failed to add sig to block: %s", err)
		}
	}
	sealedBlock := &VbftBlock{
		Info:               block.Info,
		PrevExecMerkleRoot: block.PrevExecMerkleRoot,
		CrossChainMsg:      block.CrossChainMsg,
		Block:              block.Block,
	}
	if forEmpty {
		sealedBlock.Block = block.EmptyBlock
	}

	// add block to chain store
	result, err := pool.chainStore.AddBlock(sealedBlock)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to seal block (%d) to chainstore: %s", blkNum, err)
	}

	for n := range pool.candidateBlocks {
		if n+pool.HistoryLen < blkNum {
			delete(pool.candidateBlocks, n)
		}
	}
	c.SealedBlock = sealedBlock
	c.SealedBlockExecResult = result
	return sealedBlock, result, nil
}

func (pool *BlockPool) getChainedBlock(blockNum uint32) (*VbftBlock, common.Uint256) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	// get from chainstore
	blk, err := pool.chainStore.GetBlock(blockNum)
	if err != nil {
		log.Errorf("getChainedBlock %d err:%v", blockNum, err)
		return nil, common.Uint256{}
	}
	return blk, blk.Block.Hash()
}

func (pool *BlockPool) SubmitBlock(blkNum uint32) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	return pool.chainStore.SubmitBlock(blkNum)
}

func (pool *BlockPool) ReloadFromLedger() {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	pool.chainStore.ReloadFromLedger()
}
