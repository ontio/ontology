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
	"bytes"
	"encoding/json"
	"hash/fnv"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
)

type PeerStakeInfo struct {
	Index  uint32 `json:"index"`
	NodeID string `json:"node_id"`
	Stake  uint64 `json:"stake"`
}

type Configuration struct {
	View                 uint32           `json:"view"`
	N                    uint32           `json:"n"`
	C                    uint32           `json:"c"`
	K                    uint32           `json:"k"`
	L                    uint32           `json:"l"`
	InitTxid             uint64           `json:"init_txid"`
	GenesisTimestamp     uint64           `json:"genesis_timestamp"`
	BlockMsgDelay        uint32           `json:"block_msg_delay"`
	HashMsgDelay         uint32           `json:"hash_msg_delay"`
	PeerHandshakeTimeout uint32           `json:"peer_handshake_timeout"`
	Peers                []*PeerStakeInfo `json:"peers"`
}

func shuffle_hash(txid uint64, ts uint64, id string, idx int) (uint64, error) {
	data, err := json.Marshal(struct {
		InitTxid       uint64 `json:"init_txid"`
		BlockTimestamp uint64 `json:"block_timestamp"`
		NodeID         string `json:"node_id"`
		Index          int    `json:"index"`
	}{txid, ts, id, idx})
	if err != nil {
		return 0, err
	}

	hash := fnv.New64a()
	hash.Write(data)
	return hash.Sum64(), nil
}

func genConsensusPayload(configFilename string) ([]byte, error) {
	// load pos config
	file, err := ioutil.ReadFile(configFilename)
	if err != nil {
		log.Errorf("Failed to open config file %s: %s", configFilename, err)
		os.Exit(1)
	}

	// Remove the UTF-8 Byte Order Mark
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

	config := Configuration{}
	if err := json.Unmarshal(file, &config); err != nil {
		log.Errorf("Failed to unmarshal json file: %s", err)
		os.Exit(1)
	}
	// pos config sanity checks
	if int(config.K) > len(config.Peers) {
		log.Error("peer count is less than K")
		os.Exit(1)
	}
	if config.K < 2*config.C+1 {
		log.Errorf("invalid config, K: %d, C: %d", config.K, config.C)
		os.Exit(1)
	}
	if config.L%config.K != 0 || config.L < config.K*2 {
		log.Errorf("invalid config, K: %d, L: %d", config.K, config.L)
		os.Exit(1)
	}

	// sort peers by stake
	peers := config.Peers
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].Stake > peers[j].Stake
	})

	log.Debugf("sorted peers: %v", peers)

	// get stake sum of top-k peers
	var sum uint64
	for i := 0; i < int(config.K); i++ {
		sum += peers[i].Stake
		log.Debugf("peer: %d, stack: %d", peers[i].Index, peers[i].Stake)
	}

	log.Debugf("sum of top K stakes: %d", sum)

	// calculate peer ranks
	scale := config.L/config.K - 1
	if scale <= 0 {
		log.Error("L is equal or less than K")
		os.Exit(1)
	}

	peerRanks := make([]uint64, 0)
	for i := 0; i < int(config.K); i++ {
		if peers[i].Stake == 0 {
			log.Errorf("peers rank %d, has zero stake", i)
			os.Exit(1)
		}
		s := uint64(math.Ceil(float64(peers[i].Stake) * float64(scale) * float64(config.K) / float64(sum)))
		peerRanks = append(peerRanks, s)
	}

	log.Debugf("peers rank table: %v", peerRanks)

	// calculate pos table
	chainPeers := make([]*PeerConfig, 0)
	posTable := make([]uint32, 0)
	for i := 0; i < int(config.K); i++ {
		nodeId, err := StringID(peers[i].NodeID)
		if err != nil {
			log.Errorf("Failed to format NodeID, index: %d", peers[i].Index)
			os.Exit(1)
		}
		chainPeers = append(chainPeers, &PeerConfig{
			Index: peers[i].Index,
			ID:    nodeId,
		})
		for j := uint64(0); j < peerRanks[i]; j++ {
			posTable = append(posTable, peers[i].Index)
		}
	}

	log.Debugf("init pos table: %v", posTable)

	// shuffle
	for i := len(posTable) - 1; i > 0; i-- {
		h, err := shuffle_hash(config.InitTxid, config.GenesisTimestamp, peers[posTable[i]].NodeID, i)
		if err != nil {
			log.Errorf("Failed to calculate hash value: %s", err)
			os.Exit(1)
		}
		j := h % uint64(i)
		posTable[i], posTable[j] = posTable[j], posTable[i]
	}

	log.Debugf("shuffled pos table: %v", posTable)
	// generate chain config, and save to ChainConfigFile
	chainConfig := &ChainConfig{
		Version:              Version,
		View:                 config.View,
		N:                    config.N,
		C:                    config.C,
		BlockMsgDelay:        time.Duration(config.BlockMsgDelay) * time.Millisecond,
		HashMsgDelay:         time.Duration(config.HashMsgDelay) * time.Millisecond,
		PeerHandshakeTimeout: time.Duration(config.PeerHandshakeTimeout) * time.Second,
		Peers:                chainPeers,
		PosTable:             posTable,
	}

	vbftBlockInfo := &VbftBlockInfo{
		Proposer:           math.MaxUint32,
		LastConfigBlockNum: math.MaxUint64,
		NewChainConfig:     chainConfig,
	}

	return json.Marshal(vbftBlockInfo)
}

func GenesisConsensusPayload() ([]byte, error) {
	consensusType := strings.ToLower(config.Parameters.ConsensusType)
	consensusConfigFile := config.Parameters.ConsensusConfigPath

	switch consensusType {
	case "vbft":
		return genConsensusPayload(consensusConfigFile)
	}

	return nil, nil
}
