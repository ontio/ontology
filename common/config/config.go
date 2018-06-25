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
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
	"io"
)

var Version = "" //Set value when build project

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

	DEFAULT_LOG_LEVEL                       = 1
	DEFAULT_MAX_LOG_SIZE                    = 100 //MByte
	DEFAULT_NODE_PORT                       = uint(20338)
	DEFAULT_CONSENSUS_PORT                  = uint(20339)
	DEFAULT_RPC_PORT                        = uint(20336)
	DEFAULT_RPC_LOCAL_PORT                  = uint(20337)
	DEFAULT_REST_PORT                       = uint(20334)
	DEFAULT_WS_PORT                         = uint(20335)
	DEFAULT_MAX_CONN_IN_BOUND               = uint(1024)
	DEFAULT_MAX_CONN_OUT_BOUND              = uint(1024)
	DEFAULT_MAX_CONN_IN_BOUND_FOR_SINGLE_IP = uint(16)
	DEFAULT_HTTP_INFO_PORT                  = uint(0)
	DEFAULT_MAX_TX_IN_BLOCK                 = 60000
	DEFAULT_MAX_SYNC_HEADER                 = 500
	DEFAULT_ENABLE_CONSENSUS                = true
	DEFAULT_ENABLE_EVENT_LOG                = true
	DEFAULT_CLI_RPC_PORT                    = uint(20000)
	DEFAULT_GAS_LIMIT                       = 30000
	DEFAULT_GAS_PRICE                       = 0

	DEFAULT_DATA_DIR      = "./Chain"
	DEFAULT_RESERVED_FILE = "./peers.rsv"
)

const (
	NETWORK_ID_MAIN_NET      = 1
	NETWORK_ID_POLARIS_NET   = 2
	NETWORK_ID_SOLO_NET      = 3
	NETWORK_NAME_MAIN_NET    = "ontology"
	NETWORK_NAME_POLARIS_NET = "polaris"
	NETWORK_NAME_SOLO_NET    = "testmode"
)

var NETWORK_MAGIC = map[uint32]uint32{
	NETWORK_ID_MAIN_NET:    constants.NETWORK_MAGIC_MAINNET, //Network main
	NETWORK_ID_POLARIS_NET: constants.NETWORK_MAGIC_POLARIS, //Network polaris
	NETWORK_ID_SOLO_NET:    0,                               //Network solo
}

var NETWORK_NAME = map[uint32]string{
	NETWORK_ID_MAIN_NET:    NETWORK_NAME_MAIN_NET,
	NETWORK_ID_POLARIS_NET: NETWORK_NAME_POLARIS_NET,
	NETWORK_ID_SOLO_NET:    NETWORK_NAME_SOLO_NET,
}

func GetNetworkMagic(id uint32) uint32 {
	nid, ok := NETWORK_MAGIC[id]
	if ok {
		return nid
	}
	return id
}

func GetNetworkName(id uint32) string {
	name, ok := NETWORK_NAME[id]
	if ok {
		return name
	}
	return fmt.Sprintf("%d", id)
}

var PolarisConfig = &GenesisConfig{
	SeedList: []string{
		"polaris1.ont.io:20338",
		"polaris2.ont.io:20338",
		"polaris3.ont.io:20338",
		"polaris4.ont.io:20338"},
	ConsensusType: CONSENSUS_TYPE_VBFT,
	VBFT: &VBFTConfig{
		N:                    7,
		C:                    2,
		K:                    7,
		L:                    112,
		BlockMsgDelay:        10000,
		HashMsgDelay:         10000,
		PeerHandshakeTimeout: 10,
		MaxBlockChangeView:   3000,
		AdminOntID:           "did:ont:AVaSGN1ugQJBS7R7ZcVwAoWLVK6onBgfyg",
		MinInitStake:         10000,
		VrfValue:             "1c9810aa9822e511d5804a9c4db9dd08497c31087b0daafa34d768a3253441fa20515e2f30f81741102af0ca3cefc4818fef16adb825fbaa8cad78647f3afb590e",
		VrfProof:             "c57741f934042cb8d8b087b44b161db56fc3ffd4ffb675d36cd09f83935be853d8729f3f5298d12d6fd28d45dde515a4b9d7f67682d182ba5118abf451ff1988",
		Peers: []*VBFTPeerStakeInfo{
			{
				Index:      1,
				PeerPubkey: "037c9e6c6a446b6b296f89b722cbf686b81e0a122444ef05f0f87096777663284b",
				Address:    "AXmQDzzvpEtPkNwBEFsREzApTTDZFW6frD",
				InitPos:    10000,
			},
			{
				Index:      2,
				PeerPubkey: "03dff4c63267ae5e23da44ace1bc47d0da1eb8d36fd71181dcccf0e872cb7b31fa",
				Address:    "AY5W6p4jHeZG2jjW6nS1p4KDUhcqLkU6jz",
				InitPos:    20000,
			},
			{
				Index:      3,
				PeerPubkey: "0205bc592aa9121428c4144fcd669ece1fa73fee440616c75624967f83fb881050",
				Address:    "ALZVrZrFqoSvqyi38n7mpPoeDp7DMtZ9b6",
				InitPos:    30000,
			},
			{
				Index:      4,
				PeerPubkey: "030a34dcb075d144df1f65757b85acaf053395bb47b019970607d2d1cdd222525c",
				Address:    "AMogjmLf2QohTcGST7niV75ekZfj44SKme",
				InitPos:    40000,
			},
			{
				Index:      5,
				PeerPubkey: "021844159f97d81da71da52f84e8451ee573c83b296ff2446387b292e44fba5c98",
				Address:    "AZzQTkZvjy7ih9gjvwU8KYiZZyNoy6jE9p",
				InitPos:    30000,
			},
			{
				Index:      6,
				PeerPubkey: "020cc76feb375d6ea8ec9ff653bab18b6bbc815610cecc76e702b43d356f885835",
				Address:    "AKEqQKmxCsjWJz8LPGryXzb6nN5fkK1WDY",
				InitPos:    20000,
			},
			{
				Index:      7,
				PeerPubkey: "03aa4d52b200fd91ca12deff46505c4608a0f66d28d9ae68a342c8a8c1266de0f9",
				Address:    "AQNpGWz4oHHFBejtBbakeR43DHfen7cm8L",
				InitPos:    10000,
			},
		},
	},
	DBFT: &DBFTConfig{},
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

//
// VBFT genesis config, from local config file
//
type VBFTConfig struct {
	N                    uint32               `json:"n"` // network size
	C                    uint32               `json:"c"` // consensus quorum
	K                    uint32               `json:"k"`
	L                    uint32               `json:"l"`
	BlockMsgDelay        uint32               `json:"block_msg_delay"`
	HashMsgDelay         uint32               `json:"hash_msg_delay"`
	PeerHandshakeTimeout uint32               `json:"peer_handshake_timeout"`
	MaxBlockChangeView   uint32               `json:"max_block_change_view"`
	MinInitStake         uint32               `json:"min_init_stake"`
	AdminOntID           string               `json:"admin_ont_id"`
	VrfValue             string               `json:"vrf_value"`
	VrfProof             string               `json:"vrf_proof"`
	Peers                []*VBFTPeerStakeInfo `json:"peers"`
}

func (this *VBFTConfig) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, this.N); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize n error!")
	}
	if err := serialization.WriteUint32(w, this.C); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize c error!")
	}
	if err := serialization.WriteUint32(w, this.K); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize k error!")
	}
	if err := serialization.WriteUint32(w, this.L); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize l error!")
	}
	if err := serialization.WriteUint32(w, this.BlockMsgDelay); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize block_msg_delay error!")
	}
	if err := serialization.WriteUint32(w, this.HashMsgDelay); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize hash_msg_delay error!")
	}
	if err := serialization.WriteUint32(w, this.PeerHandshakeTimeout); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize peer_handshake_timeout error!")
	}
	if err := serialization.WriteUint32(w, this.MaxBlockChangeView); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize max_block_change_view error!")
	}
	if err := serialization.WriteUint32(w, this.MinInitStake); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize min_init_stake error!")
	}
	if err := serialization.WriteString(w, this.AdminOntID); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize admin_ont_id error!")
	}
	if err := serialization.WriteString(w, this.VrfValue); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize vrf_value error!")
	}
	if err := serialization.WriteString(w, this.VrfProof); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize vrf_proof error!")
	}
	if err := serialization.WriteVarUint(w, uint64(len(this.Peers))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarUint, serialize peer length error!")
	}
	for _, peer := range this.Peers {
		if err := peer.Serialize(w); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "serialize peer error!")
		}
	}
	return nil
}

func (this *VBFTConfig) Deserialize(r io.Reader) error {
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize n error!")
	}
	c, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize c error!")
	}
	k, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize k error!")
	}
	l, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize l error!")
	}
	blockMsgDelay, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize blockMsgDelay error!")
	}
	hashMsgDelay, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize hashMsgDelay error!")
	}
	peerHandshakeTimeout, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize peerHandshakeTimeout error!")
	}
	maxBlockChangeView, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize maxBlockChangeView error!")
	}
	minInitStake, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize minInitStake error!")
	}
	adminOntID, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize adminOntID error!")
	}
	vrfValue, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize vrfValue error!")
	}
	vrfProof, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize vrfProof error!")
	}
	length, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarUint, deserialize peer length error!")
	}
	peers := make([]*VBFTPeerStakeInfo, 0)
	for i := 0; uint64(i) < length; i++ {
		peer := new(VBFTPeerStakeInfo)
		err = peer.Deserialize(r)
		if err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize peer error!")
		}
		peers = append(peers, peer)
	}
	this.N = n
	this.C = c
	this.K = k
	this.L = l
	this.BlockMsgDelay = blockMsgDelay
	this.HashMsgDelay = hashMsgDelay
	this.PeerHandshakeTimeout = peerHandshakeTimeout
	this.MaxBlockChangeView = maxBlockChangeView
	this.MinInitStake = minInitStake
	this.AdminOntID = adminOntID
	this.VrfValue = vrfValue
	this.VrfProof = vrfProof
	this.Peers = peers
	return nil
}

type VBFTPeerStakeInfo struct {
	Index      uint32 `json:"index"`
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
	InitPos    uint64 `json:"initPos"`
}

func (this *VBFTPeerStakeInfo) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, this.Index); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize index error!")
	}
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize peerPubkey error!")
	}
	address, err := common.AddressFromBase58(this.Address)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "common.AddressFromBase58, address format error!")
	}
	if err := address.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize address error!")
	}
	if err := serialization.WriteUint64(w, this.InitPos); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteUint32, serialize initPos error!")
	}
	return nil
}

func (this *VBFTPeerStakeInfo) Deserialize(r io.Reader) error {
	index, err := serialization.ReadUint32(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize index error!")
	}
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize peerPubkey error!")
	}
	address := new(common.Address)
	err = address.Deserialize(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "address.Deserialize, deserialize address error!")
	}
	initPos, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadUint32, deserialize initPos error!")
	}
	this.Index = index
	this.PeerPubkey = peerPubkey
	this.Address = address.ToBase58()
	this.InitPos = initPos
	return nil
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
	LogLevel       uint
	NodeType       string
	EnableEventLog bool
	SystemFee      map[string]int64
	GasLimit       uint64
	GasPrice       uint64
	DataDir        string
}

type ConsensusConfig struct {
	EnableConsensus bool
	MaxTxInBlock    uint
}

type P2PRsvConfig struct {
	ReservedPeers []string `json:"reserved"`
	MaskPeers     []string `json:"mask"`
}

type P2PNodeConfig struct {
	ReservedPeersOnly         bool
	ReservedCfg               *P2PRsvConfig
	NetworkMaigc              uint32
	NetworkId                 uint32
	NetworkName               string
	NodePort                  uint
	NodeConsensusPort         uint
	DualPortSupport           bool
	IsTLS                     bool
	CertPath                  string
	KeyPath                   string
	CAPath                    string
	HttpInfoPort              uint
	MaxHdrSyncReqs            uint
	MaxConnInBound            uint
	MaxConnOutBound           uint
	MaxConnInBoundForSingleIP uint
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
	Genesis   *GenesisConfig
	Common    *CommonConfig
	Consensus *ConsensusConfig
	P2PNode   *P2PNodeConfig
	Rpc       *RpcConfig
	Restful   *RestfulConfig
	Ws        *WebSocketConfig
}

func NewOntologyConfig() *OntologyConfig {
	return &OntologyConfig{
		Genesis: PolarisConfig,
		Common: &CommonConfig{
			LogLevel:       DEFAULT_LOG_LEVEL,
			EnableEventLog: DEFAULT_ENABLE_EVENT_LOG,
			SystemFee:      make(map[string]int64),
			GasLimit:       DEFAULT_GAS_LIMIT,
			DataDir:        DEFAULT_DATA_DIR,
		},
		Consensus: &ConsensusConfig{
			EnableConsensus: true,
			MaxTxInBlock:    DEFAULT_MAX_TX_IN_BLOCK,
		},
		P2PNode: &P2PNodeConfig{
			ReservedCfg:               &P2PRsvConfig{},
			ReservedPeersOnly:         false,
			NetworkId:                 NETWORK_ID_MAIN_NET,
			NetworkName:               GetNetworkName(NETWORK_ID_POLARIS_NET),
			NetworkMaigc:              GetNetworkMagic(NETWORK_ID_POLARIS_NET),
			NodePort:                  DEFAULT_NODE_PORT,
			NodeConsensusPort:         DEFAULT_CONSENSUS_PORT,
			DualPortSupport:           true,
			IsTLS:                     false,
			CertPath:                  "",
			KeyPath:                   "",
			CAPath:                    "",
			HttpInfoPort:              DEFAULT_HTTP_INFO_PORT,
			MaxHdrSyncReqs:            DEFAULT_MAX_SYNC_HEADER,
			MaxConnInBound:            DEFAULT_MAX_CONN_IN_BOUND,
			MaxConnOutBound:           DEFAULT_MAX_CONN_OUT_BOUND,
			MaxConnInBoundForSingleIP: DEFAULT_MAX_CONN_IN_BOUND_FOR_SINGLE_IP,
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
