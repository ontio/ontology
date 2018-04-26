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
	"net/http"
	"strconv"

	"github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/urfave/cli"
)

var (
	InfoCommand = cli.Command{
		Action: infoCommand,
		Name:   "info",
		Usage:  "Display informations about the chain",
		Subcommands: []cli.Command{
			blockCommandSet,
			txCommandSet,
			versionCommand,
		},
		Description: ``,
	}
)

var blockCommandSet = cli.Command{
	Action: utils.MigrateFlags(blockInfoCommand),
	Name:   "block",
	Usage:  "Display block informations",
	Flags: []cli.Flag{
		utils.HashInfoFlag,
		utils.HeightInfoFlag,
	},
	OnUsageError: blockInfoUsageError,
	Description:  ``,
	Subcommands: []cli.Command{
		{
			Action:      utils.MigrateFlags(getCurrentBlockHeight),
			Name:        "count",
			Usage:       "issue asset by command",
			Description: ``,
		},
	},
}

var txCommandSet = cli.Command{
	Action: utils.MigrateFlags(txInfoCommand),
	Name:   "tx",
	Usage:  "Display transaction informations",
	Flags: []cli.Flag{
		utils.HashInfoFlag,
	},
	OnUsageError: txInfoUsageError,
	Description:  ``,
}

var versionCommand = cli.Command{
	Action:       utils.MigrateFlags(versionInfoCommand),
	Name:         "version",
	Usage:        "Display the version",
	OnUsageError: versionInfoUsageError,
	Description:  ``,
}

func infoCommand(context *cli.Context) error {
	cli.ShowSubcommandHelp(context)
	return nil
}

func blockInfoUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println("Error:", err.Error(), "\n")
	cli.ShowSubcommandHelp(context)
	return nil
}

func getCurrentBlockHeight(ctx *cli.Context) error {
	height, err := ontSdk.Rpc.GetBlockCount()
	if nil != err {
		fmt.Printf("Get block height information is error:  %s", err.Error())
		return err
	}
	fmt.Println("Current blockchain height: ", height)
	return nil
}

func txInfoUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println("Error:", err.Error(), "\n")
	cli.ShowSubcommandHelp(context)
	return nil
}

func versionInfoUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println("Error:", err.Error(), "\n")
	cli.ShowSubcommandHelp(context)
	return nil
}

func versionInfoCommand(ctx *cli.Context) error {
	version, err := ontSdk.Rpc.GetVersion()
	if nil != err {
		fmt.Printf("Get version information is error:  %s", err.Error())
		return err
	}
	fmt.Println("Node version: ", version)
	return nil
}

func txInfoCommand(ctx *cli.Context) error {
	if ctx.IsSet(utils.HashInfoFlag.Name) {
		txHash := ctx.String(utils.HashInfoFlag.Name)
		resp, err := request("GET", nil, restfulAddr()+"/api/v1/transaction/"+txHash)
		if err != nil {
			return err
		}
		common.EchoJsonDataGracefully(resp)
		return nil
	}
	cli.ShowSubcommandHelp(ctx)
	return nil
}

func blockInfoCommand(ctx *cli.Context) error {
	if ctx.IsSet(utils.HeightInfoFlag.Name) {
		height := ctx.Int(utils.HeightInfoFlag.Name)
		if height >= 0 {
			resp, err := request("GET", nil, restfulAddr()+"/api/v1/block/details/height/"+strconv.Itoa(height))
			if err != nil {
				return err
			}
			common.EchoJsonDataGracefully(resp)
			return nil
		}
	} else if ctx.IsSet(utils.HashInfoFlag.Name) {
		blockHash := ctx.String(utils.HashInfoFlag.Name)
		if "" != blockHash {
			resp, err := request("GET", nil, restfulAddr()+"/api/v1/block/details/hash/"+blockHash)
			if err != nil {
				return err
			}
			common.EchoJsonDataGracefully(resp)
			return nil
		}
	}

	cli.ShowSubcommandHelp(ctx)
	return nil
}

func request(method string, cmd map[string]interface{}, url string) (map[string]interface{}, error) {
	hClient := &http.Client{}
	var repMsg = make(map[string]interface{})
	var response *http.Response
	var err error
	switch method {
	case "GET":
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return repMsg, err
		}
		response, err = hClient.Do(req)
	case "POST":
		data, err := json.Marshal(cmd)
		if err != nil {
			return repMsg, err
		}
		reqData := bytes.NewBuffer(data)
		req, err := http.NewRequest("POST", url, reqData)
		if err != nil {
			return repMsg, err
		}
		req.Header.Set("Content-type", "application/json")
		response, err = hClient.Do(req)
	default:
		return repMsg, err
	}
	if response != nil {
		defer response.Body.Close()

		body, _ := ioutil.ReadAll(response.Body)
		if err := json.Unmarshal(body, &repMsg); err == nil {
			return repMsg, err
		}
	}
	if err != nil {
		return repMsg, err
	}
	return repMsg, err
}
