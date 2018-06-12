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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/native/governance"
	"github.com/urfave/cli"
)

func SetOntologyConfig(ctx *cli.Context) (*config.OntologyConfig, error) {
	cfg := config.DefConfig
	err := setGenesis(ctx, cfg.Genesis)
	if err != nil {
		return nil, fmt.Errorf("setGenesis error:%s", err)
	}
	setCommonConfig(ctx, cfg.Common)
	setConsensusConfig(ctx, cfg.Consensus)
	setP2PNodeConfig(ctx, cfg.P2PNode)
	setRpcConfig(ctx, cfg.Rpc)
	setRestfulConfig(ctx, cfg.Restful)
	setWebSocketConfig(ctx, cfg.Ws)
	setCliConfig(ctx, cfg.Cli)
	if cfg.Genesis.ConsensusType == config.CONSENSUS_TYPE_SOLO {
		cfg.Ws.EnableHttpWs = true
		cfg.Restful.EnableHttpRestful = true
		cfg.P2PNode.NetworkId = config.NETWORK_ID_SOLO_NET
		cfg.P2PNode.NetworkName = config.GetNetworkName(cfg.P2PNode.NetworkId)
		cfg.P2PNode.NetworkMaigc = config.GetNetworkMagic(cfg.P2PNode.NetworkId)
	}
	return cfg, nil
}

func setGenesis(ctx *cli.Context, cfg *config.GenesisConfig) error {
	if ctx.GlobalBool(utils.GetFlagName(utils.EnableTestModeFlag)) {
		cfg.ConsensusType = config.CONSENSUS_TYPE_SOLO
		cfg.SOLO.GenBlockTime = ctx.Uint(utils.GetFlagName(utils.TestModeGenBlockTimeFlag))
		if cfg.SOLO.GenBlockTime <= 1 {
			cfg.SOLO.GenBlockTime = config.DEFAULT_GEN_BLOCK_TIME
		}
		return nil
	}

	if !ctx.IsSet(utils.GetFlagName(utils.ConfigFlag)) {
		//Using Polaris config
		return nil
	}

	genesisFile := ctx.GlobalString(utils.GetFlagName(utils.ConfigFlag))
	if !common.FileExisted(genesisFile) {
		return nil
	}

	log.Infof("Load genesis config:%s", genesisFile)
	data, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile:%s error:%s", genesisFile, err)
	}
	// Remove the UTF-8 Byte Order Mark
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	err = json.Unmarshal(data, cfg)
	if err != nil {
		return fmt.Errorf("json.Unmarshal GenesisConfig:%s error:%s", data, err)
	}

	switch cfg.ConsensusType {
	case config.CONSENSUS_TYPE_DBFT:
		if len(cfg.DBFT.Bookkeepers) < config.DBFT_MIN_NODE_NUM {
			return fmt.Errorf("DBFT consensus at least need %d bookkeepers in config", config.DBFT_MIN_NODE_NUM)
		}
		if cfg.DBFT.GenBlockTime <= 0 {
			cfg.DBFT.GenBlockTime = config.DEFAULT_GEN_BLOCK_TIME
		}
	case config.CONSENSUS_TYPE_VBFT:
		err = governance.CheckVBFTConfig(cfg.VBFT)
		if err != nil {
			return fmt.Errorf("VBFT config error %v", err)
		}
		if len(cfg.VBFT.Peers) < config.VBFT_MIN_NODE_NUM {
			return fmt.Errorf("VBFT consensus at least need %d peers in config", config.VBFT_MIN_NODE_NUM)
		}
	default:
		return fmt.Errorf("Unknow consensus:%s", cfg.ConsensusType)
	}

	return nil
}

func setCommonConfig(ctx *cli.Context, cfg *config.CommonConfig) {
	cfg.LogLevel = ctx.GlobalUint(utils.GetFlagName(utils.LogLevelFlag))
	cfg.EnableEventLog = !ctx.GlobalBool(utils.GetFlagName(utils.DisableEventLogFlag))
	cfg.GasLimit = ctx.GlobalUint64(utils.GetFlagName(utils.GasLimitFlag))
	cfg.GasPrice = ctx.GlobalUint64(utils.GetFlagName(utils.GasPriceFlag))
	cfg.DataDir = ctx.GlobalString(utils.GetFlagName(utils.DataDirFlag))
}

func setConsensusConfig(ctx *cli.Context, cfg *config.ConsensusConfig) {
	cfg.EnableConsensus = !ctx.GlobalBool(utils.GetFlagName(utils.DisableConsensusFlag))
	cfg.MaxTxInBlock = ctx.GlobalUint(utils.GetFlagName(utils.MaxTxInBlockFlag))
}

func setP2PNodeConfig(ctx *cli.Context, cfg *config.P2PNodeConfig) {
	cfg.NetworkId = uint32(ctx.GlobalUint(utils.GetFlagName(utils.NetworkIdFlag)))
	cfg.NetworkMaigc = config.GetNetworkMagic(cfg.NetworkId)
	cfg.NetworkName = config.GetNetworkName(cfg.NetworkId)
	cfg.NodePort = ctx.GlobalUint(utils.GetFlagName(utils.NodePortFlag))
	cfg.NodeConsensusPort = ctx.GlobalUint(utils.GetFlagName(utils.ConsensusPortFlag))
	cfg.DualPortSupport = ctx.GlobalBool(utils.GetFlagName(utils.DualPortSupportFlag))
	cfg.ReservedPeersOnly = ctx.GlobalBool(utils.GetFlagName(utils.ReservedPeersOnlyFlag))
	rsvfile := ctx.GlobalString(utils.GetFlagName(utils.ReservedPeersFileFlag))
	if cfg.ReservedPeersOnly {
		if !common.FileExisted(rsvfile) {
			log.Infof("file %s not exist\n", rsvfile)
			return
		}
		peers, err := ioutil.ReadFile(rsvfile)
		if err != nil {
			log.Errorf("ioutil.ReadFile:%s error:%s", rsvfile, err)
			return
		}
		peers = bytes.TrimPrefix(peers, []byte("\xef\xbb\xbf"))

		err = json.Unmarshal(peers, &cfg.ReservedPeers)
		if err != nil {
			log.Errorf("json.Unmarshal reserved peers:%s error:%s", peers, err)
			return
		}
		for i := 0; i < len(cfg.ReservedPeers); i++ {
			log.Info("reserved addr: " + cfg.ReservedPeers[i])
		}
	}

}

func setRpcConfig(ctx *cli.Context, cfg *config.RpcConfig) {
	cfg.EnableHttpJsonRpc = !ctx.Bool(utils.GetFlagName(utils.RPCDisabledFlag))
	cfg.HttpJsonPort = ctx.GlobalUint(utils.GetFlagName(utils.RPCPortFlag))
	cfg.HttpLocalPort = ctx.GlobalUint(utils.GetFlagName(utils.RPCLocalProtFlag))
}

func setRestfulConfig(ctx *cli.Context, cfg *config.RestfulConfig) {
	cfg.EnableHttpRestful = ctx.GlobalBool(utils.GetFlagName(utils.RestfulEnableFlag))
	cfg.HttpRestPort = ctx.GlobalUint(utils.GetFlagName(utils.RestfulPortFlag))
}

func setWebSocketConfig(ctx *cli.Context, cfg *config.WebSocketConfig) {
	cfg.EnableHttpWs = ctx.GlobalBool(utils.GetFlagName(utils.WsEnabledFlag))
	cfg.HttpWsPort = ctx.GlobalUint(utils.GetFlagName(utils.WsPortFlag))
}

func setCliConfig(ctx *cli.Context, cfg *config.CliConfig) {
	cfg.EnableCliRpcServer = ctx.GlobalBool(utils.GetFlagName(utils.CliEnableRpcFlag))
	cfg.CliRpcPort = ctx.GlobalUint(utils.GetFlagName(utils.CliRpcPortFlag))
}

func SetRpcPort(ctx *cli.Context) {
	if ctx.IsSet(utils.GetFlagName(utils.RPCPortFlag)) {
		config.DefConfig.Rpc.HttpJsonPort = ctx.Uint(utils.GetFlagName(utils.RPCPortFlag))
	}
}
