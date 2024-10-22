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
	"github.com/ontio/ontology-crypto/keypair"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
)

type KeyAndTaskId struct {
	Key keypair.PublicKey
	Id  uint32
}

type VbftContext struct {
	BlockNum      uint32
	Config        *vconfig.ChainConfig
	ConfigNum     uint32
	PeerKeys      map[uint32]*KeyAndTaskId
	PrevBlockInfo *BlockAndExecteInfo
	BftStatus     *BftStatus
	Proposers     []uint32
	Endorsers     []uint32
	Committers    []uint32
}

func (self *VbftContext) IsEndorser(peerIdx uint32) bool {
	for _, id := range self.Endorsers {
		if id == peerIdx {
			return true
		}
	}
	return false
}

func (self *VbftContext) IsCommitter(peerIdx uint32) bool {
	for _, id := range self.Committers {
		if id == peerIdx {
			return true
		}
	}

	return false
}

func (self *VbftContext) Is2ndProposer(peerIdx uint32) bool {
	rank := self.GetProposerRank(peerIdx)
	return rank > 0 && rank <= int(self.Config.C)
}

func (self *VbftContext) GetProposerRank(peerIdx uint32) int {
	for rank, id := range self.Proposers {
		if id == peerIdx {
			return rank
		}
	}
	return len(self.Proposers)
}

func (self *VbftContext) GetHighestRankProposal(proposals []*blockProposalMsg) *blockProposalMsg {
	proposerRank := 10000
	var proposal *blockProposalMsg
	for _, p := range proposals {
		if r := self.GetProposerRank(p.Block.getProposer()); r < proposerRank {
			proposerRank = r
			proposal = p
		}
	}

	return proposal
}

func (self *VbftContext) GetPeerPubKey(peerIdx uint32) keypair.PublicKey {
	info := self.PeerKeys[peerIdx]
	if info == nil {
		return nil
	}
	return info.Key
}
