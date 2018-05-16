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
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
)

func peerPool() *PeerPool {
	nodeId, _ := vconfig.StringID("120202c924ed1a67fd1719020ce599d723d09d48362376836e04b0be72dfe825e24d81")
	peerconfig := &vconfig.PeerConfig{
		Index: 1,
		ID:    nodeId,
	}
	peerpool := constructPeerPool(false)
	peerpool.addPeer(peerconfig)
	return peerpool
}

func constructServer() *Server {
	statemgr := &StateMgr{
		currentState: Syncing,
	}
	blockparticipantconfig := &BlockParticipantConfig{
		BlockNum:   1,
		L:          uint32(2),
		Proposers:  []uint32{1, 2, 3},
		Endorsers:  []uint32{1, 2, 3},
		Committers: []uint32{1, 2, 3},
	}
	chainconfig := &vconfig.ChainConfig{
		Version:              1,
		View:                 12,
		N:                    4,
		C:                    3,
		BlockMsgDelay:        1000,
		HashMsgDelay:         1000,
		PeerHandshakeTimeout: 10000,
		PosTable:             []uint32{2, 3, 1, 3, 1, 3, 2, 3, 2, 3, 2, 1, 3},
	}
	chainstore := &ChainStore{
		chainedBlockNum: 2,
	}
	server := &Server{
		Index:                    1,
		stateMgr:                 statemgr,
		currentParticipantConfig: blockparticipantconfig,
		config:     chainconfig,
		chainStore: chainstore,
	}
	return server
}
func TestIsPeerAlive(t *testing.T) {
	server := constructServer()
	server.peerPool = peerPool()
	res := server.isPeerAlive(2, 1)
	t.Logf("TestIsPeerAlive: %v", res)
}

func TestIsPeerActive(t *testing.T) {
	server := constructServer()
	server.peerPool = peerPool()
	res := server.isPeerActive(uint32(2), 1)
	t.Logf("TestIsPeerActive: %v", res)
}

func TestIsProposer(t *testing.T) {
	server := constructServer()
	server.peerPool = peerPool()
	res := server.isProposer(1, 1)
	t.Logf("TestIsProposer: %v", res)
}

func TestIs2ndProposer(t *testing.T) {
	server := constructServer()
	server.peerPool = peerPool()
	res := server.is2ndProposer(1, 1)
	t.Logf("TestIs2ndProposer %v", res)
}

func TestIsEndorser(t *testing.T) {
	server := constructServer()
	server.peerPool = peerPool()
	res := server.isEndorser(1, 1)
	t.Logf("TestIsEndorser %v", res)
}

func TestIsCommitter(t *testing.T) {
	server := constructServer()
	server.peerPool = peerPool()
	res := server.isCommitter(1, 1)
	t.Logf("TestIsCommitter %v", res)
}

func TestGetProposerRankLocked(t *testing.T) {
	server := constructServer()
	server.peerPool = peerPool()
	rank := server.getProposerRankLocked(1, 1)
	t.Logf("TestGetProposerRankLocked %v", rank)
}

func TestGetHighestRankProposal(t *testing.T) {
	log.Init(log.PATH, log.Stdout)
	server := constructServer()
	server.peerPool = peerPool()
	block, err := constructBlock()
	if err != nil {
		t.Errorf("constructBlock failed :%v", err)
		return
	}
	blockproposalmsg := &blockProposalMsg{
		Block: block,
	}
	var proposals []*blockProposalMsg
	proposals = append(proposals, blockproposalmsg)
	msg := server.getHighestRankProposal(1, proposals)
	t.Logf("TestGetHighestRankProposal %v", msg)
}

func TestGetCommitConsensus(t *testing.T) {
	blockcommitmsg := &blockCommitMsg{
		Committer:       1,
		BlockProposer:   1,
		BlockNum:        1,
		CommitBlockHash: common.Uint256{},
		CommitForEmpty:  true,
	}
	var commitMsgs []*blockCommitMsg
	commitMsgs = append(commitMsgs, blockcommitmsg)
	blockproposer, flag := getCommitConsensus(commitMsgs, 2)
	t.Logf("TestGetCommitConsensus %d ,%v", blockproposer, flag)
}
