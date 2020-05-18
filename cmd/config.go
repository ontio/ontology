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
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/urfave/cli"
)

func SetOntologyConfig(ctx *cli.Context) (*config.OntologyConfig, error) {
	cfg := config.DefConfig
	err := setGenesis(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("setGenesis error:%s", err)
	}
	setCommonConfig(ctx, cfg.Common)
	setConsensusConfig(ctx, cfg.Consensus)
	setRpcConfig(ctx, cfg.Rpc)
	setRestfulConfig(ctx, cfg.Restful)
	setWebSocketConfig(ctx, cfg.Ws)
	if cfg.Genesis.ConsensusType == config.CONSENSUS_TYPE_SOLO {
		cfg.Ws.EnableHttpWs = true
		cfg.Restful.EnableHttpRestful = true
		cfg.Consensus.EnableConsensus = true
		cfg.Common.GasPrice = 0
	}
	return cfg, nil
}

func setGenesis(ctx *cli.Context, cfg *config.OntologyConfig) error {
	log.Infof("This is layer2 mode")
	cfg.Genesis.ConsensusType = config.CONSENSUS_TYPE_SOLO
	cfg.Genesis.SOLO.GenBlockTime = ctx.Uint(utils.GetFlagName(utils.TestModeGenBlockTimeFlag))
	if cfg.Genesis.SOLO.GenBlockTime <= 1 {
		cfg.Genesis.SOLO.GenBlockTime = config.DEFAULT_GEN_BLOCK_TIME
	}
	return nil
}

func setCommonConfig(ctx *cli.Context, cfg *config.CommonConfig) {
	cfg.LogLevel = ctx.Uint(utils.GetFlagName(utils.LogLevelFlag))
	cfg.EnableEventLog = !ctx.Bool(utils.GetFlagName(utils.DisableEventLogFlag))
	cfg.GasLimit = ctx.Uint64(utils.GetFlagName(utils.GasLimitFlag))
	cfg.GasPrice = ctx.Uint64(utils.GetFlagName(utils.GasPriceFlag))
	cfg.MinOngLimit = ctx.Uint64(utils.GetFlagName(utils.MinOngLimitFlag))
	cfg.DataDir = ctx.String(utils.GetFlagName(utils.DataDirFlag))
}

func setConsensusConfig(ctx *cli.Context, cfg *config.ConsensusConfig) {
	cfg.EnableConsensus = ctx.Bool(utils.GetFlagName(utils.EnableConsensusFlag))
	cfg.MaxTxInBlock = ctx.Uint(utils.GetFlagName(utils.MaxTxInBlockFlag))
}

func setRpcConfig(ctx *cli.Context, cfg *config.RpcConfig) {
	cfg.EnableHttpJsonRpc = !ctx.Bool(utils.GetFlagName(utils.RPCDisabledFlag))
	cfg.HttpJsonPort = ctx.Uint(utils.GetFlagName(utils.RPCPortFlag))
	cfg.HttpLocalPort = ctx.Uint(utils.GetFlagName(utils.RPCLocalProtFlag))
}

func setRestfulConfig(ctx *cli.Context, cfg *config.RestfulConfig) {
	cfg.EnableHttpRestful = ctx.Bool(utils.GetFlagName(utils.RestfulEnableFlag))
	cfg.HttpRestPort = ctx.Uint(utils.GetFlagName(utils.RestfulPortFlag))
	cfg.HttpMaxConnections = ctx.Uint(utils.GetFlagName(utils.RestfulMaxConnsFlag))
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
