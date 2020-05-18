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
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/log"
)

var Version = "" //Set value when build project

type VerifyMethod int

const (
	InterpVerifyMethod VerifyMethod = iota
	JitVerifyMethod
	NoneVerifyMethod
)

const (
	DEFAULT_CONFIG_FILE_NAME = "./config.json"
	DEFAULT_WALLET_FILE_NAME = "./wallet.dat"
	MIN_GEN_BLOCK_TIME       = 2
	DEFAULT_GEN_BLOCK_TIME   = 1
	SOLO_MIN_NODE_NUM        = 1 //min node number of solo consensus

	CONSENSUS_TYPE_SOLO = "solo"

	DEFAULT_LOG_LEVEL                       = log.InfoLog
	DEFAULT_MAX_LOG_SIZE                    = 100 //MByte
	DEFAULT_NODE_PORT                       = uint(20338)
	DEFAULT_RPC_PORT                        = uint(20336)
	DEFAULT_RPC_LOCAL_PORT                  = uint(20337)
	DEFAULT_REST_PORT                       = uint(20334)
	DEFAULT_WS_PORT                         = uint(20335)
	DEFAULT_REST_MAX_CONN                   = uint(1024)
	DEFAULT_MAX_CONN_IN_BOUND               = uint(1024)
	DEFAULT_MAX_CONN_OUT_BOUND              = uint(1024)
	DEFAULT_MAX_CONN_IN_BOUND_FOR_SINGLE_IP = uint(16)
	DEFAULT_HTTP_INFO_PORT                  = uint(0)
	DEFAULT_MAX_TX_IN_BLOCK                 = 6000
	DEFAULT_MAX_SYNC_HEADER                 = 500
	DEFAULT_ENABLE_CONSENSUS                = true
	DEFAULT_ENABLE_EVENT_LOG                = true
	DEFAULT_CLI_RPC_PORT                    = uint(20000)
	DEFUALT_CLI_RPC_ADDRESS                 = "127.0.0.1"
	DEFAULT_GAS_LIMIT                       = 20000
	DEFAULT_MIN_ONG_LIMIT                  = 100000000
	DEFAULT_GAS_PRICE                       = 500
	DEFAULT_WASM_GAS_FACTOR                 = uint64(10)
	DEFAULT_WASM_MAX_STEPCOUNT              = uint64(8000000)

	DEFAULT_DATA_DIR      = "./Chain/"
	DEFAULT_RESERVED_FILE = "./peers.rsv"
)

const (
	WASM_GAS_FACTOR = "WASM_GAS_FACTOR"
)

const (
	NETWORK_ID_SOLO_NET      = 3
	NETWORK_NAME_SOLO_NET    = "testmode"
)

var NETWORK_MAGIC = map[uint32]uint32{
	NETWORK_ID_SOLO_NET:    0,                               //Network solo
}

var NETWORK_NAME = map[uint32]string{
	NETWORK_ID_SOLO_NET:    NETWORK_NAME_SOLO_NET,
}

func GetNetworkMagic(id uint32) uint32 {
	nid, ok := NETWORK_MAGIC[id]
	if ok {
		return nid
	}
	return id
}

var STATE_HASH_CHECK_HEIGHT = map[uint32]uint32{
	NETWORK_ID_SOLO_NET:    0,                                   //Network solo
}

func GetStateHashCheckHeight(id uint32) uint32 {
	return STATE_HASH_CHECK_HEIGHT[id]
}

var OPCODE_HASKEY_ENABLE_HEIGHT = map[uint32]uint32{
	NETWORK_ID_SOLO_NET:    0,                                            //Network solo
}

func GetOpcodeUpdateCheckHeight(id uint32) uint32 {
	return OPCODE_HASKEY_ENABLE_HEIGHT[id]
}

var GAS_ROUND_TUNE_HEIGHT = map[uint32]uint32{
	NETWORK_ID_SOLO_NET:    0,                                       //Network solo
}

func GetGasRoundTuneHeight(id uint32) uint32 {
	return GAS_ROUND_TUNE_HEIGHT[id]
}

func GetNetworkName(id uint32) string {
	name, ok := NETWORK_NAME[id]
	if ok {
		return name
	}
	return fmt.Sprintf("%d", id)
}

var DefConfig = NewOntologyConfig()

type GenesisConfig struct {
	SeedList      []string
	ConsensusType string
	SOLO          *SOLOConfig
}

func NewGenesisConfig() *GenesisConfig {
	return &GenesisConfig{
		SeedList:      make([]string, 0),
		ConsensusType: CONSENSUS_TYPE_SOLO,
		SOLO:          &SOLOConfig{},
	}
}

type SOLOConfig struct {
	GenBlockTime uint
	Bookkeepers  []string
}

type CommonConfig struct {
	LogLevel         uint
	NodeType         string
	EnableEventLog   bool
	SystemFee        map[string]int64
	GasLimit         uint64
	GasPrice         uint64
	MinOngLimit      uint64
	DataDir          string
	WasmVerifyMethod VerifyMethod
}

type ConsensusConfig struct {
	EnableConsensus bool
	MaxTxInBlock    uint
}

type RpcConfig struct {
	EnableHttpJsonRpc bool
	HttpJsonPort      uint
	HttpLocalPort     uint
}

type RestfulConfig struct {
	EnableHttpRestful  bool
	HttpRestPort       uint
	HttpMaxConnections uint
	HttpCertPath       string
	HttpKeyPath        string
}

type WebSocketConfig struct {
	EnableHttpWs bool
	HttpWsPort   uint
	HttpCertPath string
	HttpKeyPath  string
}

type OntologyConfig struct {
	Genesis   *GenesisConfig
	Common    *CommonConfig
	Consensus *ConsensusConfig
	Rpc       *RpcConfig
	Restful   *RestfulConfig
	Ws        *WebSocketConfig
}

func NewOntologyConfig() *OntologyConfig {
	return &OntologyConfig{
		Genesis: NewGenesisConfig(),
		Common: &CommonConfig{
			LogLevel:         DEFAULT_LOG_LEVEL,
			EnableEventLog:   DEFAULT_ENABLE_EVENT_LOG,
			SystemFee:        make(map[string]int64),
			GasLimit:         DEFAULT_GAS_LIMIT,
			MinOngLimit:      DEFAULT_MIN_ONG_LIMIT,
			DataDir:          DEFAULT_DATA_DIR,
			WasmVerifyMethod: InterpVerifyMethod,
		},
		Consensus: &ConsensusConfig{
			EnableConsensus: true,
			MaxTxInBlock:    DEFAULT_MAX_TX_IN_BLOCK,
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
	case CONSENSUS_TYPE_SOLO:
		bookKeepers = this.Genesis.SOLO.Bookkeepers
	default:
		return nil, fmt.Errorf("Does not support %s consensus", this.Genesis.ConsensusType)
	}

	pubKeys := make([]keypair.PublicKey, 0, len(bookKeepers))
	for _, key := range bookKeepers {
		pubKey, err := hex.DecodeString(key)
		k, err := keypair.DeserializePublicKey(pubKey)
		if err != nil {
			return nil, fmt.Errorf("Incorrectly book keepers key:%s", key)
		}
		pubKeys = append(pubKeys, k)
	}
	keypair.SortPublicKeys(pubKeys)
	return pubKeys, nil
}

func (this *OntologyConfig) getDefNetworkIDFromGenesisConfig(genCfg *GenesisConfig) (uint32, error) {
	var configData []byte
	var err error
	switch this.Genesis.ConsensusType {
	case CONSENSUS_TYPE_SOLO:
		return NETWORK_ID_SOLO_NET, nil
	default:
		return 0, fmt.Errorf("unknown consensus type:%s", this.Genesis.ConsensusType)
	}
	if err != nil {
		return 0, fmt.Errorf("json.Marshal error:%s", err)
	}
	data := sha256.Sum256(configData)
	return binary.LittleEndian.Uint32(data[0:4]), nil
}
