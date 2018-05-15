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
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/urfave/cli"
	"io/ioutil"
)

func SetOntologyConfig(ctx *cli.Context) (*config.OntologyConfig, error) {
	cfg := config.DefConfig
	err := setGenesis(ctx, cfg.Genesis)
	if err != nil {
		return nil, fmt.Errorf("setGenesis error:%s", err)
	}
	setCommonConfig(ctx, cfg.Common)
	setP2PNodeConfig(ctx, cfg.P2PNode)
	setRpcConfig(ctx, cfg.Rpc)
	setRestfulConfig(ctx, cfg.Restful)
	setWebSocketConfig(ctx, cfg.Ws)
	return cfg, nil
}

func setGenesis(ctx *cli.Context, cfg *config.GenesisConfig) error {
	if ctx.GlobalBool(utils.EnableTestModeFlag.Name) {
		cfg.ConsensusType = config.CONSENSUS_TYPE_SOLO
		cfg.SOLO.GenBlockTime = ctx.Uint(utils.TestModeGenBlockTimeFlag.Name)
		if cfg.SOLO.GenBlockTime <= 1 {
			cfg.SOLO.GenBlockTime = config.DEFAULT_GEN_BLOCK_TIME
		}
		return nil
	}

	genesisFile := ctx.GlobalString(utils.ConfigFlag.Name)

	if !common.FileExisted(genesisFile) {
		return nil
	}
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
		if len(cfg.VBFT.Peers) < config.VBFT_MIN_NODE_NUM {
			return fmt.Errorf("VBFT consensus at least need %d peers in config", config.VBFT_MIN_NODE_NUM)
		}
	default:
		return fmt.Errorf("Unknow consensus:%s", cfg.ConsensusType)
	}

	return nil
}

func setCommonConfig(ctx *cli.Context, cfg *config.CommonConfig) {
	cfg.MaxTxInBlock = ctx.GlobalUint(utils.MaxTxInBlockFlag.Name)
	cfg.DisableEventLog = ctx.GlobalBool(utils.DisableEventLogFlag.Name)
	cfg.GasLimit = ctx.GlobalUint64(utils.GasLimitFlag.Name)
	cfg.GasPrice = ctx.GlobalUint64(utils.GasPriceFlag.Name)
}

func setP2PNodeConfig(ctx *cli.Context, cfg *config.P2PNodeConfig) {
	cfg.NodePort = ctx.GlobalUint(utils.NodePortFlag.Name)
	cfg.NodeConsensusPort = ctx.GlobalUint(utils.ConsensusPortFlag.Name)
	cfg.DualPortSupport = ctx.GlobalBool(utils.DualPortSupportFlag.Name)
}

func setRpcConfig(ctx *cli.Context, cfg *config.RpcConfig) {
	cfg.EnableHttpJsonRpc = true
	cfg.HttpJsonPort = ctx.GlobalUint(utils.RPCPortFlag.Name)
	cfg.HttpLocalPort = ctx.GlobalUint(utils.RPCLocalProtFlag.Name)
}

func setRestfulConfig(ctx *cli.Context, cfg *config.RestfulConfig) {
	cfg.EnableHttpRestful = ctx.GlobalBool(utils.RestfulEnableFlag.Name)
	cfg.HttpRestPort = ctx.GlobalUint(utils.RestfulPortFlag.Name)
}

func setWebSocketConfig(ctx *cli.Context, cfg *config.WebSocketConfig) {
	cfg.EnableHttpWs = ctx.GlobalBool(utils.WsEnabledFlag.Name)
	cfg.HttpWsPort = ctx.GlobalUint(utils.WsPortFlag.Name)
}
