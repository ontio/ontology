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
	"github.com/ontio/ontology/blockrelayer"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/urfave/cli"
)

var RevertCommand = cli.Command{
	Name:   "revert",
	Usage:  "revert current info to height",
	Action: revertToHeight,
	Flags: []cli.Flag{
		utils.RevertToHeightFlag,
	},
}

func revertToHeight(ctx *cli.Context) error {
	logLevel := ctx.GlobalInt(utils.GetFlagName(utils.LogLevelFlag))
	log.InitLog(logLevel, log.PATH, log.Stdout)

	revertHeight := ctx.Uint64(utils.GetFlagName(utils.RevertToHeightFlag))
	if revertHeight >= 0 {
		return blockrelayer.RevertToHeight(config.DefConfig.Common.DataDir, uint32(revertHeight))
	}
	return nil
}

func initConfig(ctx *cli.Context) (*config.OntologyConfig, error) {
	//init ontology config from cli
	cfg, err := SetOntologyConfig(ctx)
	if err != nil {
		return nil, err
	}
	log.Infof("Config init success")
	return cfg, nil
}
