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

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/store"
)

var errDupProposal = errors.New("multi proposal from same proposer")
var errDupCommit = errors.New("multi commit from same committer")

type EndorseSigInfo struct {
	BlockHash        common.Uint256
	EndorsedProposer uint32
	Signature        []byte
	ForEmpty         bool
}

type BftStatus struct {
	// server signed message
	EndorseMsg      *blockEndorseMsg
	EndorseEmptyMsg *blockEndorseMsg
	SelfCommitMsg   *blockCommitMsg

	Proposals   []*blockProposalMsg
	CommitMsgs  []*blockCommitMsg
	EndorseSigs map[uint32][]*EndorseSigInfo // indexed by endorser index
}

func (self *BftStatus) GetEndorseSigInfos(blockHash common.Uint256) map[uint32]*EndorseSigInfo {
	result := make(map[uint32]*EndorseSigInfo)
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

func (self *BftStatus) String() string {
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

func NewCandidateInfo() *BftStatus {
	return &BftStatus{
		EndorseSigs: make(map[uint32][]*EndorseSigInfo),
	}
}

func (candidate *BftStatus) AddBlockProposal(msg *blockProposalMsg) error {
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
	eSig := &EndorseSigInfo{
		BlockHash:        msg.Block.Block.Hash(),
		EndorsedProposer: proposer,
		Signature:        msg.BlockProposerSig,
		ForEmpty:         false,
	}
	candidate.addBlockEndorsementLocked(proposer, eSig, false)
	return nil
}

func (self *BftStatus) GetBlockProposal(proposer uint32) *blockProposalMsg {
	for _, p := range self.Proposals {
		if p.Block.getProposer() == proposer {
			return p
		}
	}
	return nil
}

func (c *BftStatus) HasEndorsedForBlock() bool {
	return c.EndorseMsg != nil || c.EndorseEmptyMsg != nil
}

func (c *BftStatus) GetSelfEndorseMsg() *blockEndorseMsg {
	if c.EndorseEmptyMsg != nil {
		return c.EndorseEmptyMsg
	}
	return c.EndorseMsg
}

func (c *BftStatus) HasEndorsedForEmptyBlock() bool {
	return c.EndorseEmptyMsg != nil
}

func (c *BftStatus) setProposalEndorsed(endorse *blockEndorseMsg) {
	if endorse.EndorseForEmpty {
		c.EndorseEmptyMsg = endorse
		return
	}
	c.EndorseMsg = endorse
}

func (candidate *BftStatus) addBlockEndorsementLocked(endorser uint32, eSig *EndorseSigInfo, commitment bool) {
	if commitment {
		candidate.EndorseSigs[endorser] = []*EndorseSigInfo{eSig}
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

func (candidate *BftStatus) AddBlockEndorseMsg(msg *blockEndorseMsg) error {
	eSig := &EndorseSigInfo{
		BlockHash:        msg.EndorsedBlockHash,
		EndorsedProposer: msg.EndorsedProposer,
		Signature:        msg.EndorserSig,
		ForEmpty:         msg.EndorseForEmpty,
	}
	err := candidate.CheckBlockHashWithProposal(msg.EndorsedProposer, msg.EndorseForEmpty, msg.EndorsedBlockHash)
	if err != nil {
		return err
	}
	candidate.addBlockEndorsementLocked(msg.Endorser, eSig, false)
	return nil
}

func (self *BftStatus) CheckBlockHashWithProposal(proposer uint32, forEmpty bool, hash common.Uint256) error {
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

func (candidate *BftStatus) EndorseDone(C uint32) (uint32, bool, bool) {
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

func (candidate *BftStatus) EndorseFailed(C uint32) bool {
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

func (candidate *BftStatus) AddBlockCommitMsg(msg *blockCommitMsg) error {
	err := candidate.CheckBlockHashWithProposal(msg.BlockProposer, msg.CommitForEmpty, msg.CommitBlockHash)
	if err != nil {
		return err
	}
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
		eSig := &EndorseSigInfo{
			BlockHash:        msg.CommitBlockHash,
			EndorsedProposer: msg.BlockProposer,
			Signature:        sig,
			ForEmpty:         msg.CommitForEmpty,
		}
		candidate.addBlockEndorsementLocked(endorser, eSig, false)
	}

	// add committer sig
	candidate.addBlockEndorsementLocked(msg.Committer, &EndorseSigInfo{
		BlockHash:        msg.CommitBlockHash,
		EndorsedProposer: msg.BlockProposer,
		Signature:        msg.CommitterSig,
		ForEmpty:         msg.CommitForEmpty,
	}, true)

	// add msg to commit-msgs
	candidate.CommitMsgs = append(candidate.CommitMsgs, msg)
	return nil
}

func (candidate *BftStatus) CommitDone(vbftCtx *VbftContext) (uint32, bool, bool) {
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

func (c *BftStatus) AddSignaturesToBlock(vbftCtx *VbftContext, block *VbftBlock, forEmpty bool) error {
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
				}
				break
			}
		}
	}

	blockToAdd.Header.Bookkeepers = bookkeepers
	blockToAdd.Header.SigData = sigData

	return nil
}

func (c *BftStatus) checkBlockSign(vbftCtx *VbftContext, block *VbftBlock, forEmpty bool, requiredSigs uint32) bool {
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

func (pool *Server) SetBlockSealed(vbftCtx *VbftContext, block *VbftBlock, forEmpty bool, sigdata bool) (*VbftBlock, *store.ExecuteResult, error) {
	if sigdata {
		if err := vbftCtx.BftStatus.AddSignaturesToBlock(vbftCtx, block, forEmpty); err != nil {
			return nil, nil, fmt.Errorf("failed to add sig to block: %s", err)
		}
	}
	sealedBlock := &VbftBlock{
		Info:               block.Info,
		PrevExecMerkleRoot: block.PrevExecMerkleRoot,
		Block:              block.Block,
	}
	if forEmpty {
		sealedBlock.Block = block.EmptyBlock
	}

	pool.lock.Lock()
	defer pool.lock.Unlock()
	result, err := pool.chainStore.AddBlock(sealedBlock)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to seal block (%d) to chainstore: %s", vbftCtx.BlockNum, err)
	}

	return sealedBlock, result, nil
}

func (pool *Server) getChainedBlock(blockNum uint32) (*VbftBlock, common.Uint256) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	blk, err := pool.chainStore.GetBlock(blockNum)
	if err != nil {
		log.Errorf("getChainedBlock %d err:%v", blockNum, err)
		return nil, common.Uint256{}
	}
	return blk, blk.Block.Hash()
}

func (self *Server) SubmitBlock(blkNum uint32) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	return self.chainStore.SubmitBlock(blkNum)
}
