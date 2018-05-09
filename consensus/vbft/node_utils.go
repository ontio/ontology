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
	"fmt"
	"math"

	"github.com/ontio/ontology/common/log"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/signature"
	msgpack "github.com/ontio/ontology/p2pserver/message/msg_pack"
	p2pmsg "github.com/ontio/ontology/p2pserver/message/types"
)

func (self *Server) GetCurrentBlockNo() uint64 {
	return self.currentBlockNum
}

func (self *Server) GetCommittedBlockNo() uint64 {
	return self.chainStore.GetChainedBlockNum()
}

func (self *Server) isPeerAlive(peerIdx uint32, blockNum uint64) bool {

	// TODO
	if peerIdx == self.Index {
		return true
	}

	return self.peerPool.isPeerAlive(peerIdx)
}

func (self *Server) isPeerActive(peerIdx uint32, blockNum uint64) bool {
	if self.isPeerAlive(peerIdx, blockNum) {
		p := self.peerPool.getPeer(peerIdx)
		if p == nil {
			return false
		}

		if p.LatestInfo != nil {
			return p.LatestInfo.CommittedBlockNumber+MAX_SYNCING_CHECK_BLK_NUM*4 > self.GetCommittedBlockNo()
		}
		return true
	}

	return false
}

//
// the first proposer as leader-proposer,
// all other proposer as 2nd-proposer
// before propose-timeout, only proposal from leader-proposer is accepted
//
func (self *Server) isProposer(blockNum uint64, peerIdx uint32) bool {
	self.metaLock.RLock()
	defer self.metaLock.RUnlock()

	{
		if peerIdx == self.Index && !isActive(self.getState()) {
			return false
		}
		// the first active proposer
		for _, id := range self.currentParticipantConfig.Proposers {
			if self.isPeerAlive(id, blockNum) {
				return peerIdx == id
			}
		}
	}

	// TODO: proposer check for non-current block
	return false
}

func (self *Server) is2ndProposer(blockNum uint64, peerIdx uint32) bool {
	rank := self.getProposerRank(blockNum, peerIdx)
	return rank > 0 && rank <= int(self.config.C)
}

func (self *Server) getProposerRank(blockNum uint64, peerIdx uint32) int {
	self.metaLock.RLock()
	defer self.metaLock.RUnlock()

	return self.getProposerRankLocked(blockNum, peerIdx)
}

func (self *Server) isEndorser(blockNum uint64, peerIdx uint32) bool {
	self.metaLock.RLock()
	defer self.metaLock.RUnlock()

	// the first 2C+1 active endorsers
	var activeN uint32
	{
		for _, id := range self.currentParticipantConfig.Endorsers {
			if id == peerIdx {
				return true
			}
			if self.isPeerActive(id, blockNum) {
				activeN++
				if activeN > self.config.C*2 {
					break
				}
			}
		}
	}

	return false
}

func (self *Server) isCommitter(blockNum uint64, peerIdx uint32) bool {
	self.metaLock.RLock()
	defer self.metaLock.RUnlock()

	// the first 2C+1 active committers
	var activeN uint32
	{
		for _, id := range self.currentParticipantConfig.Committers {
			if id == peerIdx {
				return true
			}
			if self.isPeerActive(id, blockNum) {
				activeN++
				if activeN > self.config.C*2 {
					break
				}
			}
		}
	}

	return false
}

func (self *Server) getProposerRankLocked(blockNum uint64, peerIdx uint32) int {
	if blockNum == self.currentParticipantConfig.BlockNum {
		for rank, id := range self.currentParticipantConfig.Proposers {
			if id == peerIdx {
				return rank
			}
		}
	} else {
		log.Errorf("todo: get proposer config for non-current blocknum:%d, current.BlockNum%d,peerIdx:%d", blockNum, self.currentParticipantConfig.BlockNum, peerIdx)
	}
	return -1
}

func (self *Server) getHighestRankProposal(blockNum uint64, proposals []*blockProposalMsg) *blockProposalMsg {
	self.metaLock.RLock()
	defer self.metaLock.RUnlock()

	proposerRank := 10000
	var proposal *blockProposalMsg
	for _, p := range proposals {
		if p.GetBlockNum() != blockNum {
			log.Errorf("server %d, diff blockNum found when get highest rank proposal,blockNum:%d", self.Index, blockNum)
			continue
		}

		if r := self.getProposerRankLocked(blockNum, p.Block.getProposer()); r < 0 {
			continue
		} else if r < proposerRank {
			proposerRank = r
			proposal = p
		}
	}

	if proposal == nil && len(proposals) > 0 {
		for _, p := range proposals {
			log.Errorf("blk %d, proposer %d", p.Block.getBlockNum(), p.Block.getProposer())
		}
		panic("ERR")
	}

	return proposal
}

func isEmptyProposal(proposal *blockProposalMsg) bool {
	if proposal == nil || proposal.Block == nil {
		return false
	}

	return proposal.Block.isEmpty()
}

//
//  call this method with metaLock locked
//
func (self *Server) buildParticipantConfig(blkNum uint64, chainCfg *vconfig.ChainConfig) (*BlockParticipantConfig, error) {

	if blkNum == 0 {
		return nil, fmt.Errorf("not participant config for genesis block")
	}

	block, blockHash := self.blockPool.getSealedBlock(blkNum - 1)
	if block == nil {
		return nil, fmt.Errorf("failed to get sealed block (%d)", blkNum-1)
	}

	vrfValue := vrf(block, blockHash)
	if vrfValue.IsNil() {
		return nil, fmt.Errorf("failed to calculate vrf")
	}

	cfg := &BlockParticipantConfig{
		BlockNum:    blkNum,
		Vrf:         vrfValue,
		ChainConfig: chainCfg, // TODO: copy chain config
	}

	s := 0
	cfg.Proposers = calcParticipantPeers(cfg, chainCfg, s, s+vconfig.MAX_PROPOSER_COUNT)
	s += vconfig.MAX_PROPOSER_COUNT
	cfg.Endorsers = calcParticipantPeers(cfg, chainCfg, s, s+vconfig.MAX_ENDORSER_COUNT)
	s += vconfig.MAX_ENDORSER_COUNT
	cfg.Committers = calcParticipantPeers(cfg, chainCfg, s, s+vconfig.MAX_COMMITTER_COUNT)

	log.Infof("server %d, blkNum: %d, state: %d, participants config: %v, %v, %v", self.Index, blkNum,
		self.getState(), cfg.Proposers, cfg.Endorsers, cfg.Committers)

	return cfg, nil
}

func calcParticipantPeers(cfg *BlockParticipantConfig, chain *vconfig.ChainConfig, start, end int) []uint32 {

	peers := make([]uint32, 0)
	peerMap := make(map[uint32]bool)
	var cnt uint32

	for i := start; i < end; i++ {
		peerId := calcParticipant(cfg.Vrf, chain.PosTable, uint32(i))
		if _, present := peerMap[peerId]; !present {
			// got new peer
			peers = append(peers, peerId)
			peerMap[peerId] = true
			cnt++

			if cnt > chain.C*3 || cnt >= chain.N {
				return peers
			}
		}
	}

	return peers
}

func calcParticipant(vrf vconfig.VRFValue, dposTable []uint32, k uint32) uint32 {
	var v1, v2 uint32
	bIdx := k / 8
	bits1 := k % 8
	bits2 := 8 + bits1 // L - 8 + bits1

	// FIXME:
	// take 16bits random variable from vrf, if len(dposTable) is not power of 2,
	// this algorithm will break the fairness of vrf. to be fixed

	v1 = uint32(vrf[bIdx]) >> bits1
	if bIdx+1 < uint32(len(vrf)) {
		v2 = uint32(vrf[bIdx+1])
	} else {
		v2 = uint32(vrf[0])
	}

	v2 = v2 & ((1 << bits2) - 1)
	v := (v2 << (8 - bits1)) + v1
	v = v % uint32(len(dposTable))
	return dposTable[v]
}

func getCommitConsensus(commitMsgs []*blockCommitMsg, C int) (uint32, bool) {
	commitCount := make(map[uint32]int)                  // proposer -> #commit-msg
	endorseCount := make(map[uint32]map[uint32]struct{}) // proposer -> []endorsers
	emptyCommitCount := 0
	for _, c := range commitMsgs {
		if c.CommitForEmpty {
			emptyCommitCount++
		}

		commitCount[c.BlockProposer] += 1
		if commitCount[c.BlockProposer] > C {
			return c.BlockProposer, emptyCommitCount > C
		}

		for endorser, _ := range c.EndorsersSig {
			if _, present := endorseCount[c.BlockProposer]; !present {
				endorseCount[c.BlockProposer] = make(map[uint32]struct{})
			}

			endorseCount[c.BlockProposer][endorser] = struct{}{}
			if len(endorseCount[c.BlockProposer]) > C+1 {
				return c.BlockProposer, c.CommitForEmpty
			}
		}
	}

	return math.MaxUint32, false
}

func (self *Server) validateTxsInProposal(proposal *blockProposalMsg) error {
	// TODO: add VBFT specific verifications
	return nil
}

func (self *Server) receiveFromPeer(peerIdx uint32) (uint32, []byte, error) {
	if C, present := self.msgRecvC[peerIdx]; present {
		select {
		case payload := <-C:
			if payload != nil {
				return payload.fromPeer, payload.payload.Data, nil
			}

		case <-self.quitC:
			return 0, nil, fmt.Errorf("server %d quit", self.Index)
		}
	}

	return 0, nil, fmt.Errorf("nil consensus payload")
}

func (self *Server) sendToPeer(peerIdx uint32, data []byte) error {
	peer := self.peerPool.getPeer(peerIdx)
	if peer == nil {
		return fmt.Errorf("send peer failed: failed to get peer %d", peerIdx)
	}
	msg := &p2pmsg.ConsensusPayload{
		Data:  data,
		Owner: self.account.PublicKey,
	}

	buf := new(bytes.Buffer)
	if err := msg.SerializeUnsigned(buf); err != nil {
		return fmt.Errorf("failed to serialize consensus msg: %s", err)
	}
	msg.Signature, _ = signature.Sign(self.account, buf.Bytes())

	buffer, err := msgpack.NewConsensus(msg)
	if err != nil {
		log.Error("Error NewConsensus: ", err)
		return err
	}
	self.p2p.Transmit(peer.PubKey, buffer)
	return nil
}

func (self *Server) broadcast(msg ConsensusMsg) error {
	self.msgSendC <- &SendMsgEvent{
		ToPeer: math.MaxUint32,
		Msg:    msg,
	}
	return nil
}

func (self *Server) broadcastToAll(data []byte) error {
	msg := &p2pmsg.ConsensusPayload{
		Data:  data,
		Owner: self.account.PublicKey,
	}

	buf := new(bytes.Buffer)
	if err := msg.SerializeUnsigned(buf); err != nil {
		return fmt.Errorf("failed to serialize consensus msg: %s", err)
	}
	msg.Signature, _ = signature.Sign(self.account, buf.Bytes())

	self.p2p.Broadcast(msg)
	return nil
}
