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
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
)

func shuffle_hash(txid uint64, ts uint64, id []byte, idx int) (uint64, error) {
	data, err := json.Marshal(struct {
		InitTxid       uint64 `json:"init_txid"`
		BlockTimestamp uint64 `json:"block_timestamp"`
		NodeID         []byte `json:"node_id"`
		Index          int    `json:"index"`
	}{txid, ts, id, idx})
	if err != nil {
		return 0, err
	}

	hash := fnv.New64a()
	hash.Write(data)
	return hash.Sum64(), nil
}
func genConsensusPayload(cfg *config.VBFTConfig) ([]byte, error) {

	if int(cfg.K) > len(cfg.Peers) {
		return nil, fmt.Errorf("peer count is less than K")
	}
	if cfg.K < 2*cfg.C+1 {
		return nil, fmt.Errorf("invalid config, K: %d, C: %d", cfg.K, cfg.C)
	}
	if cfg.L%cfg.K != 0 || cfg.L < cfg.K*2 {
		return nil, fmt.Errorf("invalid config, K: %d, L: %d", cfg.K, cfg.L)
	}
	chainConfig, err := GenesisChainConfig(cfg, cfg.Peers)
	if err != nil {
		return nil, err
	}
	vbftBlockInfo := &VbftBlockInfo{
		Proposer:           math.MaxUint32,
		LastConfigBlockNum: math.MaxUint64,
		NewChainConfig:     chainConfig,
	}
	return json.Marshal(vbftBlockInfo)
}

//GenesisChainConfig return chainconfig
func GenesisChainConfig(config *config.VBFTConfig, peersinfo []*config.VBFTPeerStakeInfo) (*ChainConfig, error) {

	peers := peersinfo
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].InitPos > peers[j].InitPos
	})
	log.Debugf("sorted peers: %v", peers)
	// get stake sum of top-k peers
	var sum uint64
	for i := 0; i < int(config.K); i++ {
		sum += peers[i].InitPos
		log.Infof("peer: %d, stack: %d", peers[i].Index, peers[i].InitPos)
	}

	log.Debugf("sum of top K stakes: %d", sum)

	// calculate peer ranks
	scale := config.L/config.K - 1
	if scale <= 0 {
		return nil, fmt.Errorf("L is equal or less than K")
	}

	peerRanks := make([]uint64, 0)
	for i := 0; i < int(config.K); i++ {
		if peers[i].InitPos == 0 {
			return nil, fmt.Errorf("peers rank %d, has zero stake", i)
		}
		s := uint64(math.Ceil(float64(peers[i].InitPos) * float64(scale) * float64(config.K) / float64(sum)))
		peerRanks = append(peerRanks, s)
	}

	log.Debugf("peers rank table: %v", peerRanks)

	// calculate pos table
	chainPeers := make(map[uint32]*PeerConfig, 0)
	posTable := make([]uint32, 0)
	for i := 0; i < int(config.K); i++ {
		nodeId, err := StringID(peers[i].PeerPubkey)
		if err != nil {
			return nil, fmt.Errorf("failed to format NodeID, index: %d: %s", peers[i].Index, err)
		}
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
		h, err := shuffle_hash(0, 0, chainPeers[posTable[i]].ID.Bytes(), i)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate hash value: %s", err)
		}
		j := h % uint64(i)
		posTable[i], posTable[j] = posTable[j], posTable[i]
	}
	log.Debugf("init pos table: %v", posTable)

	// generate chain config, and save to ChainConfigFile
	peerCfgs := make([]*PeerConfig, 0)
	for i := 0; i < int(config.K); i++ {
		peerCfgs = append(peerCfgs, chainPeers[peers[i].Index])
	}

	chainConfig := &ChainConfig{
		Version:              1,
		View:                 1,
		N:                    config.N,
		C:                    config.C,
		BlockMsgDelay:        time.Duration(config.BlockMsgDelay) * time.Millisecond,
		HashMsgDelay:         time.Duration(config.HashMsgDelay) * time.Millisecond,
		PeerHandshakeTimeout: time.Duration(config.PeerHandshakeTimeout) * time.Second,
		Peers:                peerCfgs,
		PosTable:             posTable,
		MaxBlockChangeView:   config.MaxBlockChangeView,
	}
	return chainConfig, nil
}

func GenesisConsensusPayload() ([]byte, error) {
	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)

	switch consensusType {
	case "vbft":
		return genConsensusPayload(config.DefConfig.Genesis.VBFT)
	}
	return nil, nil
}
