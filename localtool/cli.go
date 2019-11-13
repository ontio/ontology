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

package main

import (
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"

	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/wasmtest/common"
	"github.com/urfave/cli"
)

type TestConfigElt struct {
	Contract    string               `json:"contract"`
	Balanceaddr []common.BalanceAddr `json:"balanceaddr"`
	Testcase    [][]common.TestCase  `json:"testcase"`
}

type TestConfig struct {
	Testconfigelt []TestConfigElt `json:"testconfig"`
}

func ontologyCLI(ctx *cli.Context) error {
	// Get the deployObject.
	args := ctx.Args()
	if len(args) != 1 {
		cli.ShowAppHelp(ctx)
		return errors.NewErr("[Option Error]: need <input> arg.")
	}
	deployobject := args[0]

	// Load the deployObject. Assert deployObject is directory or not.
	contracts, objIsDir, err := common.GetContact(deployobject)
	if err != nil {
		cli.ShowAppHelp(ctx)
		return err
	}

	// check option.
	config, paramsStr, err := parseOption(ctx, objIsDir)
	if err != nil {
		cli.ShowAppHelp(ctx)
		return err
	}

	acct, database := common.InitOntologyLedger()
	err = common.DeployContract(acct, database, contracts)
	if err != nil {
		return err
	}

	testContext := common.MakeTestContext(acct, contracts)

	if config != nil {
		for _, configelt := range config.Testconfigelt {
			err := common.InitBalanceAddress(configelt.Balanceaddr, acct, database)
			if err != nil {
				return err
			}
			// currently only use the index zero array.
			for _, testcase := range configelt.Testcase[0] {
				err := common.TestWithConfigElt(acct, database, configelt.Contract, testcase, testContext)
				if err != nil {
					return err
				}
			}
		}
	} else {
		err = common.InvokeSpecifiedContract(acct, database, path.Base(deployobject), paramsStr, testContext)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseOption(ctx *cli.Context, objIsDir bool) (*TestConfig, string, error) {
	if ctx.IsSet(utils.GetFlagName(ContractParamsFlag)) && ctx.IsSet(utils.GetFlagName(ConfigFlag)) {
		return nil, "", errors.NewErr("[Option Error]: You can only specify --param or --config")
	}

	LogLevel := ctx.Uint(utils.GetFlagName(LogLevelFlag))
	log.InitLog(int(LogLevel), log.PATH, log.Stdout)

	if objIsDir {
		if ctx.IsSet(utils.GetFlagName(ContractParamsFlag)) {
			return nil, "", errors.NewErr("[Option Error]: Can not specify --param when input is a directory")
		}

		if !ctx.IsSet((utils.GetFlagName(ConfigFlag))) {
			return nil, "", errors.NewErr("[Option Error]: Must specify --config config.json file when input is a directory")
		}
	}

	if ctx.IsSet(utils.GetFlagName(ConfigFlag)) {
		configFileName := ctx.String(utils.GetFlagName(ConfigFlag))
		configBuff, err := ioutil.ReadFile(configFileName)
		if err != nil {
			return nil, "", err
		}

		var config TestConfig
		err = json.Unmarshal([]byte(configBuff), &config)
		if err != nil {
			return nil, "", err
		}

		if len(config.Testconfigelt) == 0 {
			return nil, "", errors.NewErr("No testcase in config file")
		}

		for _, configelt := range config.Testconfigelt {
			if len(configelt.Contract) == 0 {
				return nil, "", errors.NewErr("[Config format error]: Do not specify contract name")
			}

			if len(configelt.Testcase) == 0 {
				return nil, "", errors.NewErr("[Config format error]: Do not specify testcase")
			}
		}

		return &config, "", nil
	}

	if ctx.IsSet(utils.GetFlagName(ContractParamsFlag)) {
		paramsStr := ctx.String(utils.GetFlagName(ContractParamsFlag))
		return nil, paramsStr, nil
	}

	return nil, "", nil
}

func main() {
	if err := setupAPP().Run(os.Args); err != nil {
		os.Exit(1)
	}
}

var (
	ConfigFlag = cli.StringFlag{
		Name:  "config,c",
		Usage: "the contract filename to be tested.",
	}
	ContractParamsFlag = cli.StringFlag{
		Name:  "param,p",
		Usage: "specify contract param when input is a file.",
	}
	LogLevelFlag = cli.UintFlag{
		Name:  "loglevel,l",
		Usage: "set the log levela.",
		Value: log.InfoLog,
	}
)

func setupAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "cli"
	app.UsageText = "cli [option] input"
	app.Action = ontologyCLI
	app.Version = "1.0.0"
	app.Copyright = "Copyright in 2019 The Ontology Authors"
	app.Flags = []cli.Flag{
		ConfigFlag,
		ContractParamsFlag,
		LogLevelFlag,
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}
	app.ExitErrHandler = func(context *cli.Context, err error) {
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
	return app
}
