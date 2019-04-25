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
	if revertHeight > 0 {
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
