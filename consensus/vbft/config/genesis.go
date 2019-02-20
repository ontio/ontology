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

package vconfig

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
)

func shuffle_hash(txid common.Uint256, height uint32, id string, idx int) (uint64, error) {
	data, err := json.Marshal(struct {
		Txid   common.Uint256 `json:"txid"`
		Height uint32         `json:"height"`
		NodeID string         `json:"node_id"`
		Index  int            `json:"index"`
	}{txid, height, id, idx})
	if err != nil {
		return 0, err
	}

	hash := fnv.New64a()
	hash.Write(data)
	return hash.Sum64(), nil
}

func deepCopy(peersInfo []*config.VBFTPeerStakeInfo) ([]*config.VBFTPeerStakeInfo, error) {
	var peers []*config.VBFTPeerStakeInfo
	buf, err := json.Marshal(peersInfo)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf, &peers)
	if err != nil {
		return nil, err
	}

	return peers, nil
}

func genConsensusPayload(cfg *config.VBFTConfig, txhash common.Uint256, height uint32) ([]byte, error) {
	if cfg.C == 0 {
		return nil, fmt.Errorf("C must larger than zero")
	}
	if int(cfg.K) > len(cfg.Peers) {
		return nil, fmt.Errorf("peer count is less than K")
	}
	if cfg.K < 2*cfg.C+1 {
		return nil, fmt.Errorf("invalid config, K: %d, C: %d", cfg.K, cfg.C)
	}
	if cfg.L%cfg.K != 0 || cfg.L < cfg.K*2 {
		return nil, fmt.Errorf("invalid config, K: %d, L: %d", cfg.K, cfg.L)
	}
	// deep copy to avoid modify global config
	peers, err := deepCopy(cfg.Peers)
	if err != nil {
		return nil, err
	}

	chainConfig, err := GenesisChainConfig(cfg, peers, txhash, height)
	if err != nil {
		return nil, err
	}

	// save VRF in genesis config file, to genesis block
	vrfValue, err := hex.DecodeString(cfg.VrfValue)
	if err != nil {
		return nil, fmt.Errorf("invalid config, vrf_value: %s", err)
	}

	vrfProof, err := hex.DecodeString(cfg.VrfProof)
	if err != nil {
		return nil, fmt.Errorf("invalid config, vrf_proof: %s", err)
	}

	// Notice:
	// take genesis msg as random source,
	// don't need verify (genesisProposer, vrfValue, vrfProof)

	vbftBlockInfo := &VbftBlockInfo{
		Proposer:           math.MaxUint32,
		VrfValue:           vrfValue,
		VrfProof:           vrfProof,
		LastConfigBlockNum: math.MaxUint32,
		NewChainConfig:     chainConfig,
	}
	return json.Marshal(vbftBlockInfo)
}

//GenesisChainConfig return chainconfig
func GenesisChainConfig(conf *config.VBFTConfig, peers []*config.VBFTPeerStakeInfo, txhash common.Uint256, height uint32) (*ChainConfig, error) {
	sort.SliceStable(peers, func(i, j int) bool {
		if peers[i].InitPos > peers[j].InitPos {
			return true
		} else if peers[i].InitPos == peers[j].InitPos {
			return peers[i].PeerPubkey > peers[j].PeerPubkey
		}
		return false
	})
	log.Debugf("sorted peers: %v", peers)
	// get stake sum of top-k peers
	var sum uint64
	for i := 0; i < int(conf.K); i++ {
		sum += peers[i].InitPos
		log.Debugf("peer: %d, stack: %d", peers[i].Index, peers[i].InitPos)
	}

	log.Debugf("sum of top K stakes: %d", sum)

	// calculate peer ranks
	scale := conf.L/conf.K - 1
	if scale <= 0 {
		return nil, fmt.Errorf("L is equal or less than K")
	}

	peerRanks := make([]uint64, 0)
	for i := 0; i < int(conf.K); i++ {
		var s uint64 = 1
		if sum > 0 && peers[i].InitPos > 0 {
			s = uint64(math.Ceil(float64(peers[i].InitPos) * float64(scale) * float64(conf.K) / float64(sum)))
		}
		peerRanks = append(peerRanks, s)
	}

	log.Debugf("peers rank table: %v", peerRanks)

	// calculate pos table
	chainPeers := make(map[uint32]*PeerConfig, 0)
	posTable := make([]uint32, 0)
	for i := 0; i < int(conf.K); i++ {
		nodeId := peers[i].PeerPubkey
		chainPeers[peers[i].Index] = &PeerConfig{
			Index: peers[i].Index,
			ID:    nodeId,
		}
		for j := uint64(0); j < peerRanks[i]; j++ {
			posTable = append(posTable, peers[i].Index)
		}
	}
	// shuffle
	for i := len(posTable) - 1; i > 0; i-- {
		h, err := shuffle_hash(txhash, height, chainPeers[posTable[i]].ID, i)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate hash value: %s", err)
		}
		j := h % uint64(i)
		posTable[i], posTable[j] = posTable[j], posTable[i]
	}
	log.Debugf("init pos table: %v", posTable)

	// generate chain conf, and save to ChainConfigFile
	peerCfgs := make([]*PeerConfig, 0)
	for i := 0; i < int(conf.K); i++ {
		peerCfgs = append(peerCfgs, chainPeers[peers[i].Index])
	}

	chainConfig := &ChainConfig{
		Version:              1,
		View:                 1,
		N:                    conf.K,
		C:                    conf.C,
		BlockMsgDelay:        time.Duration(conf.BlockMsgDelay) * time.Millisecond,
		HashMsgDelay:         time.Duration(conf.HashMsgDelay) * time.Millisecond,
		PeerHandshakeTimeout: time.Duration(conf.PeerHandshakeTimeout) * time.Second,
		Peers:                peerCfgs,
		PosTable:             posTable,
		MaxBlockChangeView:   conf.MaxBlockChangeView,
	}
	return chainConfig, nil
}

func GenesisConsensusPayload(txhash common.Uint256, height uint32) ([]byte, error) {
	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)

	switch consensusType {
	case "vbft":
		return genConsensusPayload(config.DefConfig.Genesis.VBFT, txhash, height)
	}
	return nil, nil
}
