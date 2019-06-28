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
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
)

func peerPool() *PeerPool {
	nodeId := "120202c924ed1a67fd1719020ce599d723d09d48362376836e04b0be72dfe825e24d81"
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
		Proposers:  []uint32{1, 2, 3},
		Endorsers:  []uint32{1, 2, 3},
		Committers: []uint32{1, 2, 3},
	}
	chainconfig := &vconfig.ChainConfig{
		Version:              1,
		View:                 12,
		N:                    4,
		C:                    1,
		BlockMsgDelay:        1000,
		HashMsgDelay:         1000,
		PeerHandshakeTimeout: 10000,
		PosTable:             []uint32{2, 3, 1, 3, 1, 3, 2, 3, 2, 3, 2, 1, 3, 0},
	}
	chainstore := &ChainStore{
		chainedBlockNum: 2,
	}
	server := &Server{
		Index:                    1,
		stateMgr:                 statemgr,
		config:                   chainconfig,
		chainStore:               chainstore,
		currentParticipantConfig: blockparticipantconfig,
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
	log.InitLog(log.InfoLog, log.Stdout)
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
	blockproposer, flag := getCommitConsensus(commitMsgs, 2, 7)
	t.Logf("TestGetCommitConsensus %d ,%v", blockproposer, flag)
}

func newTestVrfValue() vconfig.VRFValue {
	v := make([]byte, 1024)
	rand.Read(v[:])
	t := sha512.Sum512(v)
	f := sha512.Sum512(t[:])
	return vconfig.VRFValue(f)
}

func TestCalcParticipantPeers(t *testing.T) {
	for i := 4; i < 100; i++ {
		testCalcParticipantPeers(t, i, (i-1)/3)
	}
	for i := 5; i < 100; i++ {
		testCalcParticipantPeers(t, i, (i-1)/4)
	}
	for i := 0; i < 10000; i++ {
		testCalcParticipantPeers(t, 7, 2)
		testCalcParticipantPeers(t, 4, 1)
	}
}

func testCalcParticipantPeers(t *testing.T, n, c int) {
	server := constructServer()

	pos := make([]uint32, 0)
	for i := 0; i < n; i++ {
		for j := 0; j < 4; j++ {
			pos = append(pos, uint32(i))
		}
	}

	chainCfg := server.config
	chainCfg.N = uint32(n)
	chainCfg.C = uint32(c)
	chainCfg.PosTable = pos
	chainCfg.Peers = make([]*vconfig.PeerConfig, 0)
	for i := 0; i < n; i++ {
		chainCfg.Peers = append(chainCfg.Peers, &vconfig.PeerConfig{
			Index: uint32(i),
			ID:    fmt.Sprintf("test-%d", i)})
	}

	cfg := &BlockParticipantConfig{
		BlockNum:    100,
		Vrf:         newTestVrfValue(),
		ChainConfig: chainCfg,
	}

	pp, pe, pc := calcParticipantPeers(cfg, chainCfg)
	if len(pp) != c+1 {
		t.Fatalf("invalid proposal peer(%d, %d): %v, %v, %v", n, c, pp, pe, pc)
	}
	if len(pe) < 2*c {
		t.Fatalf("invalid endorse peer(%d, %d): %v, %v, %v", n, c, pp, pe, pc)
	}
	if len(pc) < 2*c {
		t.Fatalf("invalid commit peer(%d, %d): %v, %v, %v", n, c, pp, pe, pc)
	}

	// check how many peers are selected
	peers := make(map[uint32]bool)
	for _, p := range pp {
		peers[p] = true
	}
	for _, p := range pe {
		peers[p] = true
	}
	for _, p := range pc {
		peers[p] = true
	}
	if len(peers) < 2*c+1 {
		t.Fatalf("peers(%d, %d, %d, %d, %d, %d): %v, %v, %v", n, c, len(peers), len(pp), len(pe), len(pc), pp, pe, pc)
	}
}
