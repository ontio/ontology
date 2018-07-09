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

func (self *Server) GetCurrentBlockNo() uint32 {
	return self.currentBlockNum
}

func (self *Server) GetCommittedBlockNo() uint32 {
	return self.chainStore.GetChainedBlockNum()
}

func (self *Server) isPeerAlive(peerIdx uint32, blockNum uint32) bool {

	// TODO
	if peerIdx == self.Index {
		return true
	}

	return self.peerPool.isPeerAlive(peerIdx)
}

func (self *Server) isPeerActive(peerIdx uint32, blockNum uint32) bool {
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
func (self *Server) isProposer(blockNum uint32, peerIdx uint32) bool {
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

func (self *Server) is2ndProposer(blockNum uint32, peerIdx uint32) bool {
	rank := self.getProposerRank(blockNum, peerIdx)
	return rank > 0 && rank <= int(self.config.C)
}

func (self *Server) getProposerRank(blockNum uint32, peerIdx uint32) int {
	self.metaLock.RLock()
	defer self.metaLock.RUnlock()

	return self.getProposerRankLocked(blockNum, peerIdx)
}

func (self *Server) isEndorser(blockNum uint32, peerIdx uint32) bool {
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

func (self *Server) isCommitter(blockNum uint32, peerIdx uint32) bool {
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

func (self *Server) getProposerRankLocked(blockNum uint32, peerIdx uint32) int {
	if blockNum == self.currentParticipantConfig.BlockNum {
		for rank, id := range self.currentParticipantConfig.Proposers {
			if id == peerIdx {
				return rank
			}
		}
	} else {
		log.Errorf("todo: get proposer config for non-current blocknum:%d, current.BlockNum%d,peerIdx:%d", blockNum, self.currentParticipantConfig.BlockNum, peerIdx)
	}
	return len(self.currentParticipantConfig.Proposers)
}

func (self *Server) getHighestRankProposal(blockNum uint32, proposals []*blockProposalMsg) *blockProposalMsg {
	self.metaLock.RLock()
	defer self.metaLock.RUnlock()

	proposerRank := 10000
	var proposal *blockProposalMsg
	for _, p := range proposals {
		if p.GetBlockNum() != blockNum {
			log.Errorf("server %d, diff blockNum found when get highest rank proposal,blockNum:%d", self.Index, blockNum)
			continue
		}

		if r := self.getProposerRankLocked(blockNum, p.Block.getProposer()); r < proposerRank {
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

//
//  call this method with metaLock locked
//
func (self *Server) buildParticipantConfig(blkNum uint32, block *Block, chainCfg *vconfig.ChainConfig) (*BlockParticipantConfig, error) {

	if blkNum == 0 {
		return nil, fmt.Errorf("not participant config for genesis block")
	}

	vrfValue := getParticipantSelectionSeed(block)
	if vrfValue.IsNil() {
		return nil, fmt.Errorf("failed to calculate participant SelectionSeed")
	}

	cfg := &BlockParticipantConfig{
		BlockNum:    blkNum,
		Vrf:         vrfValue,
		ChainConfig: chainCfg, // TODO: copy chain config
	}

	s := 0
	cfg.Proposers = calcParticipantPeers(cfg, chainCfg, s, s+vconfig.MAX_PROPOSER_COUNT)
	if uint32(len(cfg.Proposers)) < chainCfg.C {
		return nil, fmt.Errorf("cfg Proposers length less than chainCfg.C:%d,%d", uint32(len(cfg.Proposers)), chainCfg.C)
	}
	s += vconfig.MAX_PROPOSER_COUNT
	cfg.Endorsers = calcParticipantPeers(cfg, chainCfg, s, s+vconfig.MAX_ENDORSER_COUNT)
	if uint32(len(cfg.Endorsers)) < 2*chainCfg.C {
		return nil, fmt.Errorf("cfg.Endorsers length less than double chainCfg.C:%d,%d", uint32(len(cfg.Endorsers)), chainCfg.C)
	}
	s += vconfig.MAX_ENDORSER_COUNT
	cfg.Committers = calcParticipantPeers(cfg, chainCfg, s, s+vconfig.MAX_COMMITTER_COUNT)
	if uint32(len(cfg.Committers)) < 2*chainCfg.C {
		return nil, fmt.Errorf("cfg.Committers length less than double chainCfg.C:%d,%d", uint32(len(cfg.Committers)), chainCfg.C)
	}
	log.Infof("server %d, blkNum: %d, state: %d, participants config: %v, %v, %v", self.Index, blkNum,
		self.getState(), cfg.Proposers, cfg.Endorsers, cfg.Committers)

	return cfg, nil
}

func calcParticipantPeers(cfg *BlockParticipantConfig, chain *vconfig.ChainConfig, start, end int) []uint32 {

	peers := make([]uint32, 0)
	peerMap := make(map[uint32]bool)
	var cnt uint32

	for i := start; ; i++ {
		peerId := calcParticipant(cfg.Vrf, chain.PosTable, uint32(i))
		if peerId == math.MaxUint32 {
			return []uint32{}
		}
		if _, present := peerMap[peerId]; !present {
			// got new peer
			peers = append(peers, peerId)
			peerMap[peerId] = true
			cnt++
			if cnt >= chain.N {
				return peers
			}
		}
		if end == vconfig.MAX_PROPOSER_COUNT {
			if i >= end && uint32(len(peers)) > chain.C {
				return peers
			}
		}
		if end == vconfig.MAX_ENDORSER_COUNT+vconfig.MAX_PROPOSER_COUNT ||
			end == vconfig.MAX_PROPOSER_COUNT+vconfig.MAX_ENDORSER_COUNT+vconfig.MAX_COMMITTER_COUNT {
			if uint32(len(peers)) > chain.C*2 {
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
	if k >= 512 {
		return math.MaxUint32
	}
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

//
// check if commit msgs has reached consensus
// return
//		@ consensused proposer
//		@ consensused for empty commit
//
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

		for endorser := range c.EndorsersSig {
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

func (self *Server) findBlockProposal(blkNum uint32, proposer uint32, forEmpty bool) *blockProposalMsg {
	for _, p := range self.blockPool.getBlockProposals(blkNum) {
		if p.Block.getProposer() == proposer {
			return p
		}
	}

	for _, p := range self.msgPool.GetProposalMsgs(blkNum) {
		if pMsg := p.(*blockProposalMsg); pMsg != nil {
			if pMsg.Block.getProposer() == proposer {
				return pMsg
			}
		}
	}

	return nil
}

func (self *Server) validateTxsInProposal(proposal *blockProposalMsg) error {
	// TODO: add VBFT specific verifications
	return nil
}

func (self *Server) heartbeat() {
	//	build heartbeat msg
	msg, err := self.constructHeartbeatMsg()
	if err != nil {
		log.Errorf("failed to build heartbeat msg: %s", err)
		return
	}

	//	send to peer
	self.msgSendC <- &SendMsgEvent{
		ToPeer: math.MaxUint32,
		Msg:    msg,
	}
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

func (self *Server) sendToPeer(peerIdx uint32, data []byte, msgType MsgType) error {
	peer := self.peerPool.getPeer(peerIdx)
	if peer == nil {
		return fmt.Errorf("send peer failed: failed to get peer %d", peerIdx)
	}
	msg := &p2pmsg.ConsensusPayload{
		Data:     data,
		Owner:    self.account.PublicKey,
		DataType: uint8(msgType),
	}

	buf := new(bytes.Buffer)
	if err := msg.SerializeUnsigned(buf); err != nil {
		return fmt.Errorf("failed to serialize consensus msg: %s", err)
	}
	msg.Signature, _ = signature.Sign(self.account, buf.Bytes())

	cons := msgpack.NewConsensus(msg)
	p2pid, present := self.peerPool.getP2pId(peerIdx)
	if present {
		self.p2p.Transmit(p2pid, cons)
	} else {
		log.Errorf("sendToPeer transmit failed index:%d", peerIdx)
	}
	return nil
}

func (self *Server) broadcast(msg ConsensusMsg) error {
	self.msgSendC <- &SendMsgEvent{
		ToPeer: math.MaxUint32,
		Msg:    msg,
	}
	return nil
}

func (self *Server) broadcastToAll(data []byte, msgType MsgType) error {
	msg := &p2pmsg.ConsensusPayload{
		Data:     data,
		Owner:    self.account.PublicKey,
		DataType: uint8(msgType),
	}

	buf := new(bytes.Buffer)
	if err := msg.SerializeUnsigned(buf); err != nil {
		return fmt.Errorf("failed to serialize consensus msg: %s", err)
	}
	msg.Signature, _ = signature.Sign(self.account, buf.Bytes())

	self.p2p.Broadcast(msg)
	return nil
}
