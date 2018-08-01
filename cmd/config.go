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

package cmd

import (
	"fmt"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/native/governance"
	"github.com/urfave/cli"
	"io/ioutil"
	"bytes"
	"encoding/json"
)

func SetOntologyConfig(ctx *cli.Context) (*config.OntologyConfig, error) {
	cfg := config.DefConfig
	err := setGenesis(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("setGenesis error:%s", err)
	}
	setCommonConfig(ctx, cfg.Common)
	setConsensusConfig(ctx, cfg.Consensus)
	setP2PNodeConfig(ctx, cfg.P2PNode)
	setRpcConfig(ctx, cfg.Rpc)
	setRestfulConfig(ctx, cfg.Restful)
	setWebSocketConfig(ctx, cfg.Ws)
	if cfg.Genesis.ConsensusType == config.CONSENSUS_TYPE_SOLO {
		cfg.Ws.EnableHttpWs = true
		cfg.Restful.EnableHttpRestful = true
		cfg.Consensus.EnableConsensus = true
		cfg.P2PNode.NetworkId = config.NETWORK_ID_SOLO_NET
		cfg.P2PNode.NetworkName = config.GetNetworkName(cfg.P2PNode.NetworkId)
		cfg.P2PNode.NetworkMagic = config.GetNetworkMagic(cfg.P2PNode.NetworkId)
		cfg.Common.GasPrice = 0
	}
	if cfg.P2PNode.NetworkId == config.NETWORK_ID_MAIN_NET ||
		cfg.P2PNode.NetworkId == config.NETWORK_ID_POLARIS_NET {
		defNetworkId, err := cfg.GetDefaultNetworkId()
		if err != nil {
			return nil, fmt.Errorf("GetDefaultNetworkId error:%s", err)
		}
		if defNetworkId != cfg.P2PNode.NetworkId {
			cfg.P2PNode.NetworkId = defNetworkId
			cfg.P2PNode.NetworkMagic = config.GetNetworkMagic(defNetworkId)
			cfg.P2PNode.NetworkName = config.GetNetworkName(defNetworkId)
		}
	}
	return cfg, nil
}

func setGenesis(ctx *cli.Context, cfg *config.OntologyConfig) error {
	netWorkId := ctx.Int(utils.GetFlagName(utils.NetworkIdFlag))
	switch netWorkId {
	case config.NETWORK_ID_MAIN_NET:
		cfg.Genesis = config.MainNetConfig
	case config.NETWORK_ID_POLARIS_NET:
		cfg.Genesis = config.PolarisConfig
	}

	if ctx.Bool(utils.GetFlagName(utils.EnableTestModeFlag)) {
		cfg.Genesis.ConsensusType = config.CONSENSUS_TYPE_SOLO
		cfg.Genesis.SOLO.GenBlockTime = ctx.Uint(utils.GetFlagName(utils.TestModeGenBlockTimeFlag))
		if cfg.Genesis.SOLO.GenBlockTime <= 1 {
			cfg.Genesis.SOLO.GenBlockTime = config.DEFAULT_GEN_BLOCK_TIME
		}
		return nil
	}

	if !ctx.IsSet(utils.GetFlagName(utils.ConfigFlag)) {
		return nil
	}

	genesisFile := ctx.String(utils.GetFlagName(utils.ConfigFlag))
	if !common.FileExisted(genesisFile) {
		return nil
	}

	newGenesisCfg := config.NewGenesisConfig()
	err := utils.GetJsonObjectFromFile(genesisFile, newGenesisCfg)
	if err != nil {
		return err
	}
	cfg.Genesis = newGenesisCfg
	log.Infof("Load genesis config:%s", genesisFile)

	switch cfg.Genesis.ConsensusType {
	case config.CONSENSUS_TYPE_DBFT:
		if len(cfg.Genesis.DBFT.Bookkeepers) < config.DBFT_MIN_NODE_NUM {
			return fmt.Errorf("DBFT consensus at least need %d bookkeepers in config", config.DBFT_MIN_NODE_NUM)
		}
		if cfg.Genesis.DBFT.GenBlockTime <= 0 {
			cfg.Genesis.DBFT.GenBlockTime = config.DEFAULT_GEN_BLOCK_TIME
		}
	case config.CONSENSUS_TYPE_VBFT:
		err = governance.CheckVBFTConfig(cfg.Genesis.VBFT)
		if err != nil {
			return fmt.Errorf("VBFT config error %v", err)
		}
		if len(cfg.Genesis.VBFT.Peers) < config.VBFT_MIN_NODE_NUM {
			return fmt.Errorf("VBFT consensus at least need %d peers in config", config.VBFT_MIN_NODE_NUM)
		}
	default:
		return fmt.Errorf("Unknow consensus:%s", cfg.Genesis.ConsensusType)
	}

	return nil
}

func setCommonConfig(ctx *cli.Context, cfg *config.CommonConfig) {
	cfg.LogLevel = ctx.Uint(utils.GetFlagName(utils.LogLevelFlag))
	cfg.EnableEventLog = !ctx.Bool(utils.GetFlagName(utils.DisableEventLogFlag))
	cfg.GasLimit = ctx.Uint64(utils.GetFlagName(utils.GasLimitFlag))
	cfg.GasPrice = ctx.Uint64(utils.GetFlagName(utils.GasPriceFlag))
	cfg.DataDir = ctx.String(utils.GetFlagName(utils.DataDirFlag))
}

func setConsensusConfig(ctx *cli.Context, cfg *config.ConsensusConfig) {
	cfg.EnableConsensus = ctx.Bool(utils.GetFlagName(utils.EnableConsensusFlag))
	cfg.MaxTxInBlock = ctx.Uint(utils.GetFlagName(utils.MaxTxInBlockFlag))
}

func setP2PNodeConfig(ctx *cli.Context, cfg *config.P2PNodeConfig) {
	cfg.NetworkId = uint32(ctx.Uint(utils.GetFlagName(utils.NetworkIdFlag)))
	cfg.NetworkMagic = config.GetNetworkMagic(cfg.NetworkId)
	cfg.NetworkName = config.GetNetworkName(cfg.NetworkId)
	cfg.NodePort = ctx.Uint(utils.GetFlagName(utils.NodePortFlag))
	cfg.NodeConsensusPort = ctx.Uint(utils.GetFlagName(utils.ConsensusPortFlag))
	cfg.DualPortSupport = ctx.Bool(utils.GetFlagName(utils.DualPortSupportFlag))
	cfg.ReservedPeersOnly = ctx.Bool(utils.GetFlagName(utils.ReservedPeersOnlyFlag))
	cfg.MaxConnInBound = ctx.Uint(utils.GetFlagName(utils.MaxConnInBoundFlag))
	cfg.MaxConnOutBound = ctx.Uint(utils.GetFlagName(utils.MaxConnOutBoundFlag))
	cfg.MaxConnInBoundForSingleIP = ctx.Uint(utils.GetFlagName(utils.MaxConnInBoundForSingleIPFlag))

	rsvfile := ctx.String(utils.GetFlagName(utils.ReservedPeersFileFlag))
	if cfg.ReservedPeersOnly {
		if !common.FileExisted(rsvfile) {
			log.Infof("file %s not exist\n", rsvfile)
			return
		}
		err := utils.GetJsonObjectFromFile(rsvfile, &cfg.ReservedCfg)
		if err != nil {
			log.Errorf("Get ReservedCfg error:%s", err)
			return
		}
		for i := 0; i < len(cfg.ReservedCfg.ReservedPeers); i++ {
			log.Info("reserved addr: " + cfg.ReservedCfg.ReservedPeers[i])
		}
		for i := 0; i < len(cfg.ReservedCfg.MaskPeers); i++ {
			log.Info("mask addr: " + cfg.ReservedCfg.MaskPeers[i])
		}
	}

	// load network manage config
	networkMgrFile := ctx.GlobalString(utils.GetFlagName(utils.NetworkMgrFlag))
	if !common.FileExisted(networkMgrFile) {
		log.Infof("file %s not exist\n", networkMgrFile)
		return
	}
	data, err := ioutil.ReadFile(networkMgrFile)
	if err != nil {
		log.Errorf("ioutil.ReadFile:%s error:%s", networkMgrFile, err)
		return
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	err = json.Unmarshal(data, &cfg.NetworkMgrCfg)
	if err != nil {
		log.Errorf("json.Unmarshal network mgr config:%s error:%s", data, err)
		return
	}
	for i := 0; i < len(cfg.NetworkMgrCfg.Peers); i++ {
		log.Infof("Peer nodeId %d, pubkey %s", cfg.NetworkMgrCfg.Peers[i].NodeId,
			cfg.NetworkMgrCfg.Peers[i].PubKey)
	}
	log.Infof("Local node IP: %s, UDPPort: %d", cfg.NetworkMgrCfg.DHT.IP,
		cfg.NetworkMgrCfg.DHT.UDPPort)

	for i := 0; i < len(cfg.NetworkMgrCfg.DHT.Seeds); i++ {
		seed := cfg.NetworkMgrCfg.DHT.Seeds[i]
		log.Infof("seed IP: %s, udp: %d, tcp: %d", seed.IP, seed.UDPPort, seed.TCPPort)
	}
}

func setRpcConfig(ctx *cli.Context, cfg *config.RpcConfig) {
	cfg.EnableHttpJsonRpc = !ctx.Bool(utils.GetFlagName(utils.RPCDisabledFlag))
	cfg.HttpJsonPort = ctx.Uint(utils.GetFlagName(utils.RPCPortFlag))
	cfg.HttpLocalPort = ctx.Uint(utils.GetFlagName(utils.RPCLocalProtFlag))
}

func setRestfulConfig(ctx *cli.Context, cfg *config.RestfulConfig) {
	cfg.EnableHttpRestful = ctx.Bool(utils.GetFlagName(utils.RestfulEnableFlag))
	cfg.HttpRestPort = ctx.Uint(utils.GetFlagName(utils.RestfulPortFlag))
}

func setWebSocketConfig(ctx *cli.Context, cfg *config.WebSocketConfig) {
	cfg.EnableHttpWs = ctx.Bool(utils.GetFlagName(utils.WsEnabledFlag))
	cfg.HttpWsPort = ctx.Uint(utils.GetFlagName(utils.WsPortFlag))
}

func SetRpcPort(ctx *cli.Context) {
	if ctx.IsSet(utils.GetFlagName(utils.RPCPortFlag)) {
		config.DefConfig.Rpc.HttpJsonPort = ctx.Uint(utils.GetFlagName(utils.RPCPortFlag))
	}
}
