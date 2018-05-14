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

package config

import (
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/ontio/ontology-crypto/keypair"
)

const (
	DEFAULT_CONFIG_FILE_NAME = "./config.json"
	DEFAULT_WALLET_FILE_NAME = "./wallet.dat"
	MIN_GEN_BLOCK_TIME       = 2
	DEFAULT_GEN_BLOCK_TIME   = 6
	DBFT_MIN_NODE_NUM        = 4 //min node number of dbft consensus
	SOLO_MIN_NODE_NUM        = 1 //min node number of solo consensus
	VBFT_MIN_NODE_NUM        = 4 //min node number of vbft consensus

	CONSENSUS_TYPE_DBFT = "dbft"
	CONSENSUS_TYPE_SOLO = "solo"
	CONSENSUS_TYPE_VBFT = "vbft"

	DEFAULT_LOG_LEVEL         = 1
	DEFAULT_MAX_LOG_SIZE      = 100 //MByte
	DEFAULT_NODE_PORT         = uint(20338)
	DEFAULT_CONSENSUS_PORT    = uint(20339)
	DEFAULT_RPC_PORT          = uint(20336)
	DEFAULT_RPC_LOCAL_PORT    = uint(20337)
	DEFAULT_REST_PORT         = uint(20334)
	DEFAULT_WS_PORT           = uint(20335)
	DEFAULT_HTTP_INFO_PORT    = uint(0)
	DEFAULT_MAX_TX_IN_BLOCK   = 60000
	DEFAULT_MAX_SYNC_HEADER   = 500
	DEFAULT_ENABLE_CONSENSUS  = true
	DEFAULT_DISABLE_EVENT_LOG = false
	DEFAULT_GAS_LIMIT         = 30000
)

var PolarisConfig = &GenesisConfig{
	SeedList: []string{
		"polaris1.ont.io:20338",
		"polaris2.ont.io:20338",
		"polaris3.ont.io:20338",
		"polaris4.ont.io:20338"},
	ConsensusType: "dbft",
	VBFT:          &VBFTConfig{},
	DBFT: &DBFTConfig{
		GenBlockTime: DEFAULT_GEN_BLOCK_TIME,
		Bookkeepers: []string{
			"12020384d843c02ecef233d3dd3bc266ee0d1a67cf2a1666dc1b2fb455223efdee7452",
			"120203fab19438e18d8a5bebb6cd3ede7650539e024d7cc45c88b95ab13f8266ce9570",
			"120203c43f136596ee666416fedb90cde1e0aee59a79ec18ab70e82b73dd297767eddf",
			"120202a76a434b18379e3bda651b7c04e972dadc4760d1156b5c86b3c4d27da48c91a1"},
	},
	SOLO: &SOLOConfig{},
}

var DefConfig = NewOntologyConfig()

type GenesisConfig struct {
	SeedList      []string
	ConsensusType string
	VBFT          *VBFTConfig
	DBFT          *DBFTConfig
	SOLO          *SOLOConfig
}

func NewGenesisConfig() *GenesisConfig {
	return &GenesisConfig{
		SeedList:      make([]string, 0),
		ConsensusType: CONSENSUS_TYPE_DBFT,
		VBFT:          &VBFTConfig{},
		DBFT:          &DBFTConfig{},
		SOLO:          &SOLOConfig{},
	}
}

type VBFTConfig struct {
	N                    uint32               `json:"n"` // network size
	C                    uint32               `json:"c"` // consensus quorum
	K                    uint32               `json:"k"`
	L                    uint32               `json:"l"`
	BlockMsgDelay        uint32               `json:"block_msg_delay"`
	HashMsgDelay         uint32               `json:"hash_msg_delay"`
	PeerHandshakeTimeout uint32               `json:"peer_handshake_timeout"`
	MaxBlockChangeView   uint32               `json:"max_block_change_view"`
	Peers                []*VBFTPeerStakeInfo `json:"peers"`
}

type VBFTPeerStakeInfo struct {
	Index      uint32 `json:"index"`
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
	InitPos    uint64 `json:"initPos"`
}

type DBFTConfig struct {
	GenBlockTime uint
	Bookkeepers  []string
}

type SOLOConfig struct {
	GenBlockTime uint
	Bookkeepers  []string
}

type CommonConfig struct {
	MaxTxInBlock    uint
	NodeType        string
	EnableConsensus bool
	DisableEventLog bool
	SystemFee       map[string]int64
	GasLimit        uint64
	GasPrice        uint64
}

type P2PNodeConfig struct {
	NodePort          uint
	NodeConsensusPort uint
	DualPortSupport   bool
	IsTLS             bool
	CertPath          string
	KeyPath           string
	CAPath            string
	HttpInfoPort      uint
	MaxHdrSyncReqs    uint
}

type RpcConfig struct {
	EnableHttpJsonRpc bool
	HttpJsonPort      uint
	HttpLocalPort     uint
}

type RestfulConfig struct {
	EnableHttpRestful bool
	HttpRestPort      uint
	HttpCertPath      string
	HttpKeyPath       string
}

type WebSocketConfig struct {
	EnableHttpWs bool
	HttpWsPort   uint
	HttpCertPath string
	HttpKeyPath  string
}

type OntologyConfig struct {
	Genesis *GenesisConfig
	Common  *CommonConfig
	P2PNode *P2PNodeConfig
	Rpc     *RpcConfig
	Restful *RestfulConfig
	Ws      *WebSocketConfig
}

func NewOntologyConfig() *OntologyConfig {
	return &OntologyConfig{
		Genesis: PolarisConfig,
		Common: &CommonConfig{
			MaxTxInBlock:    DEFAULT_MAX_TX_IN_BLOCK,
			EnableConsensus: DEFAULT_ENABLE_CONSENSUS,
			DisableEventLog: DEFAULT_DISABLE_EVENT_LOG,
			SystemFee:       make(map[string]int64),
			GasLimit:        DEFAULT_GAS_LIMIT,
		},
		P2PNode: &P2PNodeConfig{
			NodePort:          DEFAULT_NODE_PORT,
			NodeConsensusPort: DEFAULT_CONSENSUS_PORT,
			DualPortSupport:   true,
			IsTLS:             false,
			CertPath:          "",
			KeyPath:           "",
			CAPath:            "",
			HttpInfoPort:      DEFAULT_HTTP_INFO_PORT,
			MaxHdrSyncReqs:    DEFAULT_MAX_SYNC_HEADER,
		},
		Rpc: &RpcConfig{
			EnableHttpJsonRpc: true,
			HttpJsonPort:      DEFAULT_RPC_PORT,
			HttpLocalPort:     DEFAULT_RPC_LOCAL_PORT,
		},
		Restful: &RestfulConfig{
			EnableHttpRestful: true,
			HttpRestPort:      DEFAULT_REST_PORT,
		},
		Ws: &WebSocketConfig{
			EnableHttpWs: true,
			HttpWsPort:   DEFAULT_WS_PORT,
		},
	}
}

func (this *OntologyConfig) GetBookkeepers() ([]keypair.PublicKey, error) {
	var bookKeepers []string
	switch this.Genesis.ConsensusType {
	case CONSENSUS_TYPE_VBFT:
		for _, peer := range this.Genesis.VBFT.Peers {
			bookKeepers = append(bookKeepers, peer.PeerPubkey)
		}
	case CONSENSUS_TYPE_DBFT:
		bookKeepers = this.Genesis.DBFT.Bookkeepers
	case CONSENSUS_TYPE_SOLO:
		bookKeepers = this.Genesis.SOLO.Bookkeepers
	default:
		return nil, fmt.Errorf("Does not support %s consensus", this.Genesis.ConsensusType)
	}

	sort.Strings(bookKeepers)
	pubKeys := make([]keypair.PublicKey, 0, len(bookKeepers))
	for _, key := range bookKeepers {
		pubKey, err := hex.DecodeString(key)
		k, err := keypair.DeserializePublicKey(pubKey)
		if err != nil {
			return nil, fmt.Errorf("Incorrectly book keepers key:%s", key)
		}
		pubKeys = append(pubKeys, k)
	}
	return pubKeys, nil
}

var Version = ""
