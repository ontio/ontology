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
	"bytes"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/states"
	params "github.com/ontio/ontology/smartcontract/service/native/global_params"
	gov "github.com/ontio/ontology/smartcontract/service/native/governance"
	nutils "github.com/ontio/ontology/smartcontract/service/native/utils"
)

// getAdmin returns current admin
func getAdmin() (common.Address, error) {
	storageKey := &states.StorageKey{
		CodeHash: nutils.ParamContractAddress,
		Key:      append([]byte(params.ADMIN)),
	}

	data, err := ledger.DefLedger.GetStorageItem(storageKey.CodeHash, storageKey.Key)
	if err != nil {
		return common.Address{}, err
	}
	adminAddress := new(common.Address)
	err = adminAddress.Deserialize(bytes.NewBuffer(data))
	if err != nil {
		return common.Address{}, err
	}

	return *adminAddress, nil
}

// getGovernanceView returns current governance view
func getGovernanceView() (*gov.GovernanceView, error) {
	storageKey := &states.StorageKey{
		CodeHash: nutils.GovernanceContractAddress,
		Key:      append([]byte(gov.GOVERNANCE_VIEW)),
	}
	data, err := ledger.DefLedger.GetStorageItem(storageKey.CodeHash, storageKey.Key)
	if err != nil {
		return nil, err
	}
	governanceView := new(gov.GovernanceView)
	err = governanceView.Deserialize(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return governanceView, nil
}

// getPeers returns the emergency governance peers
func getPeers() ([]*EmergencyGovPeer, error) {
	goveranceview, err := getGovernanceView()
	if err != nil {
		return nil, err
	}

	viewBytes, err := gov.GetUint32Bytes(goveranceview.View)
	if err != nil {
		return nil, err
	}

	storageKey := &states.StorageKey{
		CodeHash: nutils.GovernanceContractAddress,
		Key:      append([]byte(gov.PEER_POOL), viewBytes...),
	}
	data, err := ledger.DefLedger.GetStorageItem(storageKey.CodeHash, storageKey.Key)
	if err != nil {
		return nil, err
	}
	peerMap := &gov.PeerPoolMap{
		PeerPoolMap: make(map[string]*gov.PeerPoolItem),
	}
	err = peerMap.Deserialize(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	peers := make([]*EmergencyGovPeer, 0, len(peerMap.PeerPoolMap))

	for _, id := range peerMap.PeerPoolMap {
		if id.Status == gov.ConsensusStatus || id.Status == gov.CandidateStatus {
			peer := &EmergencyGovPeer{
				PubKey: id.PeerPubkey,
				Status: id.Status,
			}
			peers = append(peers, peer)
		}
	}

	return peers, nil
}
