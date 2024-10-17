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
	"math"
	"sync/atomic"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/signature"
	msgpack "github.com/ontio/ontology/p2pserver/message/msg_pack"
	p2pmsg "github.com/ontio/ontology/p2pserver/message/types"
)

func (self *Server) GetCurrentBlockNo() uint32 {
	return self.GetVbftContext().BlockNum
}

func (self *Server) GetVbftContext() *VbftContext {
	self.metaLock.RLock()
	defer self.metaLock.RUnlock()
	return self.vbftCtx
}

func (self *Server) GetPeerMsgChan(peerIdx uint32) chan *p2pMsgPayload {
	if C, ok := self.msgRecvC.Load(peerIdx); ok {
		return C.(chan *p2pMsgPayload)
	}
	return nil
}

func (self *Server) CreatePeerMsgChan(peerIdx uint32) {
	newC := make(chan *p2pMsgPayload, 1024)
	_, loaded := self.msgRecvC.LoadOrStore(peerIdx, newC)
	if loaded {
		close(newC)
	}
}

func (self *Server) ClosePeerMsgChan(peerIdx uint32) {
	C, loaded := self.msgRecvC.LoadOrStore(peerIdx, nil)
	if loaded {
		close(C.(chan *p2pMsgPayload))
	}
	self.msgRecvC.Delete(peerIdx)
}

func (self *Server) isPeerAlive(peerIdx uint32) bool {
	return peerIdx == self.Index || self.peerPool.IsPeerConnected(peerIdx)
}

// the first proposer as leader-proposer,
// all other proposer as 2nd-proposer
// before propose-timeout, only proposal from leader-proposer is accepted
func (self *Server) isProposer(peerIdx uint32) bool {
	return self.GetActiveProposer() == peerIdx
}

func (self *Server) GetActiveProposer() uint32 {
	// the first active proposer
	for _, id := range self.GetVbftContext().Proposers {
		if self.isPeerAlive(id) {
			if id == self.Index && !self.getState().IsActive() {
				continue
			}
			return id
		}
	}

	// all not active
	return math.MaxUint32
}

func (self *Server) updateTimerParams(config *vconfig.ChainConfig) {
	setTimeout := func(time *int64, value int64) {
		if value > 0 {
			atomic.StoreInt64(time, value)
		}
	}
	setTimeout(&makeProposalTimeout, int64(config.BlockMsgDelay*2))
	setTimeout(&make2ndProposalTimeout, int64(config.BlockMsgDelay))
	setTimeout(&endorseBlockTimeout, int64(config.HashMsgDelay*2))
	setTimeout(&commitBlockTimeout, int64(config.HashMsgDelay*3))
	setTimeout(&peerHandshakeTimeout, int64(config.PeerHandshakeTimeout))
	setTimeout(&zeroTxBlockTimeout, int64(config.BlockMsgDelay*3))
}

func buildPeerRoles(blkNum, proposer uint32, preVrfValue []byte, chainCfg *vconfig.ChainConfig) ([]uint32, []uint32, []uint32) {
	vrfValue := getParticipantSelectionSeed(blkNum, proposer, preVrfValue)
	return calcPeerRoles(vrfValue, chainCfg)
}

func calcPeerRoles(vrfValue vconfig.VRFValue, chain *vconfig.ChainConfig) ([]uint32, []uint32, []uint32) {
	peers := make([]uint32, 0)
	peerMap := make(map[uint32]bool)

	// 1. select peers as many as possible
	c := int(chain.C)
	for i := 0; i < len(chain.PosTable); i++ {
		peerId := calcParticipant(vrfValue, chain.PosTable, uint32(i))
		if peerId == math.MaxUint32 {
			break
		}
		if _, present := peerMap[peerId]; !present {
			peers = append(peers, peerId)
			peerMap[peerId] = true
			if len(peerMap) > (c+1)+((2*c+1)*2) || len(peerMap) == int(chain.N) {
				break
			}
		}
	}

	if len(peerMap) <= c*3 {
		for _, peer := range chain.Peers {
			if _, present := peerMap[peer.Index]; !present {
				peers = append(peers, peer.Index)
				peerMap[peer.Index] = true
			}
			if len(peerMap) > c*3 {
				break
			}
		}
	}

	// [p0, p1, p2, .... p_c+1, ...    .. p_m, ....      pn]
	//  <-- proposer  --><--- endorser --><-- committer -->
	nCommitter := 2*c + 1
	propsers := peers[0 : c+1]
	n1 := (len(peers) - len(propsers)) / 2
	endorsers0 := peers[c+1 : c+1+n1]
	committers := peers[c+1+n1:]

	// copy endorser0 to endorser
	endorsers := make([]uint32, 0)
	endorsers = append(endorsers, endorsers0...)
	if len(endorsers) < nCommitter {
		// not enough endorser, get more from committer/proposer
		// 1. add last empty proposer
		endorsers = append(endorsers, propsers[c])
		// 2. add committers if not enough
		for i := len(committers) - 1; i >= 0 && len(endorsers) < nCommitter; i-- {
			endorsers = append(endorsers, committers[i])
		}
		// 3. add proposers if not enough
		for i := c - 1; i > 0 && len(endorsers) < nCommitter; i-- {
			endorsers = append(endorsers, propsers[i])
		}
	}
	if len(committers) < nCommitter {
		// not enough committer, get more from endorser/proposer
		// 1. add proposers if not enough
		for i := 1; i < len(propsers) && len(committers) < nCommitter; i++ {
			committers = append(committers, propsers[i])
		}
		// 2. add init-endorsers if not enough
		for i := len(endorsers0) - 1; i >= 0 && len(committers) < nCommitter; i-- {
			committers = append(committers, endorsers0[i])
		}
	}

	return propsers, endorsers, committers
}

func calcParticipant(vrf vconfig.VRFValue, dposTable []uint32, k uint32) uint32 {
	var v1, v2 uint32
	bIdx := k / 8
	bits1 := k % 8
	bits2 := 8 + bits1 // L - 8 + bits1
	if k >= 512 {
		return math.MaxUint32
	}
	// Note: take 16bits random variable from vrf, if len(dposTable) is not power of 2,
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

// check if commit msgs has reached consensus
// return
//
//	@ consensused proposer
//	@ consensused for empty commit
func getCommitConsensus(commitMsgs []*blockCommitMsg, C int, N int) (uint32, bool) {
	emptyCommitCount := 0
	emptyCommit := false
	signCount := make(map[uint32]map[uint32]int)
	for _, c := range commitMsgs {
		if c.CommitForEmpty {
			emptyCommitCount++
			if emptyCommitCount > C && !emptyCommit {
				emptyCommit = true
			}
		}
		if _, present := signCount[c.BlockProposer]; !present {
			signCount[c.BlockProposer] = make(map[uint32]int)
		}
		signCount[c.BlockProposer][c.Committer] += 1
		for endorser := range c.EndorsersSig {
			signCount[c.BlockProposer][endorser] += 1
		}
		if len(signCount[c.BlockProposer]) >= N-(N-1)/3 {
			return c.BlockProposer, emptyCommit
		}
	}

	return math.MaxUint32, false
}

func (self *Server) heartbeat(peerIdx uint32) {
	//	build heartbeat msg
	msg, err := self.constructHeartbeatMsg()
	if err != nil {
		log.Errorf("failed to build heartbeat msg: %s", err)
		return
	}

	//	send to peer
	self.msgSendC <- &SendMsgEvent{
		ToPeer: peerIdx,
		Msg:    msg,
	}
}

func (self *Server) receiveFromPeer(peerIdx uint32) (uint32, ConsensusMsg, error) {
	if C := self.GetPeerMsgChan(peerIdx); C != nil {
		select {
		case payload := <-C:
			if payload != nil {
				return payload.fromPeer, payload.Data, nil
			}

		case <-self.quitC:
			return 0, nil, fmt.Errorf("server %d quit", self.Index)
		}
	}

	return 0, nil, fmt.Errorf("nil consensus payload")
}

func (self *Server) sendToPeer(peerIdx uint32, msg ConsensusMsg) error {
	p2pid, present := self.peerPool.GetP2pId(peerIdx)
	if !present {
		return fmt.Errorf("send peer failed: failed to get peer %d", peerIdx)
	}
	go self.p2p.SendTo(p2pid, msgpack.NewConsensus(self.packAndSignP2PMsg(msg)))
	return nil
}

func (self *Server) broadcast(msg ConsensusMsg) {
	self.msgSendC <- &SendMsgEvent{
		ToPeer: math.MaxUint32,
		Msg:    msg,
	}
}

func (self *Server) packAndSignP2PMsg(msg ConsensusMsg) *p2pmsg.ConsensusPayload {
	data := MustSerializeVbftMsg(msg)
	payload := &p2pmsg.ConsensusPayload{
		Data:            data,
		BookkeeperIndex: uint16(self.Index), // TODO: define as uint32, currently uint16 is enough
		Owner:           self.account.PublicKey,
	}

	sink := common.NewZeroCopySink(nil)
	payload.SerializationUnsigned(sink)
	payload.Signature, _ = signature.Sign(self.account, sink.Bytes())

	return payload
}

func (self *Server) broadcastToAll(msg ConsensusMsg) {
	go self.p2p.Broadcast(msgpack.NewConsensus(self.packAndSignP2PMsg(msg)))
}
