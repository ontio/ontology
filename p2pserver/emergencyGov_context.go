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

package p2pserver

import (
	"time"

	"github.com/ontio/ontology/common"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/types"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	gov "github.com/ontio/ontology/smartcontract/service/native/governance"
)

type EmergencyGovPeer struct {
	PubKey  string
	Address common.Address
	Status  gov.Status
}

type EmergencyGovStatus uint8

const (
	EmergencyGovInit EmergencyGovStatus = iota
	EmergencyGovStart
	EmergencyGovComplete
)

type emergencyGovContext struct {
	EmergencyReqCache *mt.EmergencyActionRequest                     // Current emergency governance request
	EmergencyRspCache map[vconfig.NodeID]*mt.EmergencyActionResponse // Cache current response from peers
	Signatures        map[vconfig.NodeID][]byte                      // Cache current signatures from peers
	Status            EmergencyGovStatus                             // Current emergency governance status
	Height            uint32                                         // Current emergency governance height
	peers             map[vconfig.NodeID]*EmergencyGovPeer
	timer             *time.Timer
	done              chan struct{}
}

func (this *emergencyGovContext) reset() {
	this.EmergencyReqCache = nil
	this.EmergencyRspCache = make(map[vconfig.NodeID]*mt.EmergencyActionResponse, 0)
	this.Signatures = make(map[vconfig.NodeID][]byte, 0)
	this.Status = EmergencyGovInit
	this.peers = make(map[vconfig.NodeID]*EmergencyGovPeer, 0)
	this.done = make(chan struct{}, 1)
}

func (this *emergencyGovContext) setStatus(status EmergencyGovStatus) {
	this.Status = status
}

func (this *emergencyGovContext) getStatus() EmergencyGovStatus {
	return this.Status
}

func (this *emergencyGovContext) getSig(id vconfig.NodeID) []byte {
	return this.Signatures[id]
}

func (this *emergencyGovContext) setSig(id vconfig.NodeID, sig []byte) {
	this.Signatures[id] = sig
}

func (this *emergencyGovContext) setPeers(peers []*EmergencyGovPeer) {
	for _, peer := range peers {
		id, err := vconfig.StringID(peer.PubKey)
		if err != nil {
			continue
		}
		this.peers[id] = peer
	}
}

func (this *emergencyGovContext) setEmergencyReqCache(msg *mt.EmergencyActionRequest) {
	this.EmergencyReqCache = msg
}

func (this *emergencyGovContext) getEmergencyReqCache() *mt.EmergencyActionRequest {
	return this.EmergencyReqCache
}

func (this *emergencyGovContext) appendEmergencyRsp(id vconfig.NodeID, msg *mt.EmergencyActionResponse) {
	this.EmergencyRspCache[id] = msg
}

func (this *emergencyGovContext) getEmergencyRspCache() map[vconfig.NodeID]*mt.EmergencyActionResponse {
	return this.EmergencyRspCache
}

func (this *emergencyGovContext) clearEmergencyRspCache() {
	this.EmergencyRspCache = make(map[vconfig.NodeID]*mt.EmergencyActionResponse, 0)
}

func (this *emergencyGovContext) getEmergencyBlock() *types.Block {
	if this.EmergencyReqCache == nil {
		return nil
	}
	return this.EmergencyReqCache.ProposalBlk
}

func (this *emergencyGovContext) getEmergencyGovHeight() uint32 {
	return this.Height
}

func (this *emergencyGovContext) setEmergencyGovHeight(height uint32) {
	this.Height = height
}

func (this *emergencyGovContext) getSignatureCount() int {
	count := 0
	for _, sig := range this.Signatures {
		if sig != nil {
			count++
		}
	}
	return count
}

func (this *emergencyGovContext) threshold() int {
	return len(this.peers) - (len(this.peers)-1)/3
}
