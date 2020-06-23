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

package utils

import (
	"bytes"
	"encoding/json"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/smartcontract/service/native/governance"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const GovNodeCacheTime = time.Minute * 10

type GovNodeResolver interface {
	IsGovNodePubKey(key keypair.PublicKey) bool
	IsGovNode(key string) bool
}

type GovNodeMockResolver struct {
	govNode map[string]struct{}
}

func NewGovNodeMockResolver(gov []string) *GovNodeMockResolver {
	govNode := make(map[string]struct{}, len(gov))
	for _, node := range gov {
		govNode[node] = struct{}{}
	}

	return &GovNodeMockResolver{govNode}
}

func (self *GovNodeMockResolver) IsGovNode(key string) bool {
	_, ok := self.govNode[key]

	return ok
}

func (self *GovNodeMockResolver) IsGovNodePubKey(key keypair.PublicKey) bool {
	pubKey := vconfig.PubkeyID(key)
	_, ok := self.govNode[pubKey]

	return ok
}

type GovNodeLedgerResolver struct {
	db *ledger.Ledger

	cache unsafe.Pointer // atomic pointer to GovCache, avoid read&write data race
}

type GovCache struct {
	view        uint32
	refreshTime time.Time
	govNodeNum  uint32
	pubkeys     map[string]struct{}
}

func NewGovNodeResolver(db *ledger.Ledger) *GovNodeLedgerResolver {
	return &GovNodeLedgerResolver{
		db:    db,
		cache: unsafe.Pointer(&GovCache{pubkeys: make(map[string]struct{})}),
	}
}

func (self *GovNodeLedgerResolver) IsGovNodePubKey(key keypair.PublicKey) bool {
	pubKey := vconfig.PubkeyID(key)
	return self.IsGovNode(pubKey)
}

func (self *GovNodeLedgerResolver) IsGovNode(pubKey string) bool {
	view, err := GetGovernanceView(self.db)
	if err != nil {
		log.Warnf("[subnet] gov node resolver failed to load view from ledger, err: %v", err)
		return false
	}
	cached := (*GovCache)(atomic.LoadPointer(&self.cache))
	if cached != nil && view.View == cached.view && cached.refreshTime.Add(GovNodeCacheTime).After(time.Now()) {
		_, ok := cached.pubkeys[pubKey]
		return ok
	}

	govNode := false
	peers, count, err := GetPeersConfig(self.db, view.View)
	if err != nil {
		log.Warnf("[subnet] gov node resolver failed to load peers from ledger, err: %v", err)
		return false
	}

	pubkeys := make(map[string]struct{}, len(peers))
	for _, peer := range peers {
		pubkeys[peer.PeerPubkey] = struct{}{}
		if peer.PeerPubkey == pubKey {
			govNode = true
		}
	}

	jsonPeers, _ := json.Marshal(peers)
	log.Infof("[subnet] reloading gov node: %s", string(jsonPeers))

	atomic.StorePointer(&self.cache, unsafe.Pointer(&GovCache{
		govNodeNum:  count,
		pubkeys:     pubkeys,
		refreshTime: time.Now(),
		view:        view.View,
	}))

	return govNode
}

func GetGovernanceView(backend *ledger.Ledger) (*governance.GovernanceView, error) {
	value, err := backend.GetStorageItem(utils.GovernanceContractAddress, []byte(governance.GOVERNANCE_VIEW))
	if err != nil {
		return nil, err
	}
	governanceView := new(governance.GovernanceView)
	err = governanceView.Deserialize(bytes.NewBuffer(value))
	if err != nil {
		return nil, err
	}
	return governanceView, nil
}

func GetPeersConfig(backend *ledger.Ledger, view uint32) ([]*config.VBFTPeerStakeInfo, uint32, error) {
	viewBytes := governance.GetUint32Bytes(view)
	key := append([]byte(governance.PEER_POOL), viewBytes...)
	data, err := backend.GetStorageItem(utils.GovernanceContractAddress, key)
	if err != nil {
		return nil, 0, err
	}
	peerMap := &governance.PeerPoolMap{
		PeerPoolMap: make(map[string]*governance.PeerPoolItem),
	}
	err = peerMap.Deserialization(common.NewZeroCopySource(data))
	if err != nil {
		return nil, 0, err
	}

	govCount := uint32(0)
	var peerstakes []*config.VBFTPeerStakeInfo
	for _, id := range peerMap.PeerPoolMap {
		switch id.Status {
		case governance.CandidateStatus, governance.ConsensusStatus, governance.QuitConsensusStatus:
			conf := &config.VBFTPeerStakeInfo{
				Index:      uint32(id.Index),
				PeerPubkey: id.PeerPubkey,
				InitPos:    id.InitPos + id.TotalPos,
			}
			peerstakes = append(peerstakes, conf)
		}
		if id.Status == governance.ConsensusStatus || id.Status == governance.QuitConsensusStatus {
			govCount += 1
		}
	}
	return peerstakes, govCount, nil
}
