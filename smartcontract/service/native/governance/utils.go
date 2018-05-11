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

package governance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"math/big"

	"github.com/ontio/ontology/common"
	vbftconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func shufflehash(txid common.Uint256, ts uint32, id []byte, idx int) (uint64, error) {
	data, err := json.Marshal(struct {
		Txid           common.Uint256 `json:"txid"`
		BlockTimestamp uint32         `json:"block_timestamp"`
		NodeID         []byte         `json:"node_id"`
		Index          int            `json:"index"`
	}{txid, ts, id, idx})
	if err != nil {
		return 0, err
	}

	hash := fnv.New64a()
	hash.Write(data)
	return hash.Sum64(), nil
}

func calDposTable(native *native.NativeService, config *Configuration,
	peers []*PeerStakeInfo) ([]uint32, map[uint32]*vbftconfig.PeerConfig, error) {
	// get stake sum of top-k peers
	var sum uint64
	for i := 0; i < int(config.K); i++ {
		sum += peers[i].Stake
	}

	// calculate peer ranks
	scale := config.L/config.K - 1
	if scale <= 0 {
		return nil, nil, errors.NewErr("[calDposTable] L is equal or less than K!")
	}

	peerRanks := make([]uint64, 0)
	for i := 0; i < int(config.K); i++ {
		if peers[i].Stake == 0 {
			return nil, nil, errors.NewErr(fmt.Sprintf("[calDposTable] peers rank %d, has zero stake!", i))
		}
		s := uint64(math.Ceil(float64(peers[i].Stake) * float64(scale) * float64(config.K) / float64(sum)))
		peerRanks = append(peerRanks, s)
	}

	// calculate pos table
	chainPeers := make(map[uint32]*vbftconfig.PeerConfig, 0)
	posTable := make([]uint32, 0)
	for i := 0; i < int(config.K); i++ {
		nodeId, err := vbftconfig.StringID(peers[i].PeerPubkey)
		if err != nil {
			return nil, nil, errors.NewDetailErr(err, errors.ErrNoCode,
				fmt.Sprintf("[calDposTable] Failed to format NodeID, index: %d: %s", peers[i].Index, err))
		}
		chainPeers[peers[i].Index] = &vbftconfig.PeerConfig{
			Index: peers[i].Index,
			ID:    nodeId,
		}
		for j := uint64(0); j < peerRanks[i]; j++ {
			posTable = append(posTable, peers[i].Index)
		}
	}

	// shuffle
	for i := len(posTable) - 1; i > 0; i-- {
		h, err := shufflehash(native.Tx.Hash(), native.Height, chainPeers[posTable[i]].ID.Bytes(), i)
		if err != nil {
			return nil, nil, errors.NewDetailErr(err, errors.ErrNoCode, "[calDposTable] Failed to calculate hash value")
		}
		j := h % uint64(i)
		posTable[i], posTable[j] = posTable[j], posTable[i]
	}

	return posTable, chainPeers, nil
}

func GetPeerPoolMap(native *native.NativeService, contract common.Address, view *big.Int) (*PeerPoolMap, error) {
	peerPoolMap := &PeerPoolMap{
		PeerPoolMap: make(map[string]*PeerPool),
	}
	peerPoolMapBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(PEER_POOL), view.Bytes()))
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[GetPeerPoolMap] Get all peerPoolMap error!")
	}
	if peerPoolMapBytes == nil {
		return nil, errors.NewErr("[GetPeerPoolMap] peerPoolMap is nil!")
	}
	peerPoolMapStore, _ := peerPoolMapBytes.(*cstates.StorageItem)
	err = json.Unmarshal(peerPoolMapStore.Value, peerPoolMap)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[GetPeerPoolMap] Unmarshal peerPoolMap error!")
	}
	return peerPoolMap, nil
}

func GetGovernanceView(native *native.NativeService, contract common.Address) (*GovernanceView, error) {
	governanceViewBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, utils.ConcatKey(contract, []byte(GOVERNANCE_VIEW)))
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[GetGovernanceView] Get governanceViewBytes error!")
	}
	governanceView := new(GovernanceView)
	if governanceViewBytes == nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[GetGovernanceView] Get nil governanceViewBytes!")
	} else {
		governanceViewStore, _ := governanceViewBytes.(*cstates.StorageItem)
		err = json.Unmarshal(governanceViewStore.Value, governanceView)
		if err != nil {
			return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[GetGovernanceView] Unmarshal governanceView error!")
		}
	}
	return governanceView, nil
}

func GetView(native *native.NativeService, contract common.Address) (*big.Int, error) {
	governanceView, err := GetGovernanceView(native, contract)
	if err != nil {
		return new(big.Int), errors.NewDetailErr(err, errors.ErrNoCode, "[GetView] GetGovernanceView error!")
	}
	return governanceView.View, nil
}

func AppCallTransferOng(native *native.NativeService, from common.Address, to common.Address, amount uint64) error {
	buf := bytes.NewBuffer(nil)
	var sts []*ont.State
	sts = append(sts, &ont.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := &ont.Transfers{
		States: sts,
	}
	err := transfers.Serialize(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOng] transfers.Serialize error!")
	}

	if _, err := native.ContextRef.AppCall(genesis.OngContractAddress, "transfer", []byte{}, buf.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOng] appCall error!")
	}
	return nil
}

func AppCallTransferOnt(native *native.NativeService, from common.Address, to common.Address, amount uint64) error {
	buf := bytes.NewBuffer(nil)
	var sts []*ont.State
	sts = append(sts, &ont.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := &ont.Transfers{
		States: sts,
	}
	err := transfers.Serialize(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOnt] transfers.Serialize error!")
	}

	if _, err := native.ContextRef.AppCall(genesis.OntContractAddress, "transfer", []byte{}, buf.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOnt] appCall error!")
	}
	return nil
}
