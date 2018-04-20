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
	"os"
	"strconv"

	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common/log"
	"github.com/urfave/cli"
)

var (
	InfoCommand = cli.Command{
		Action:   utils.MigrateFlags(infoCommand),
		Name:     "info",
		Usage:    "ontology info [block|chain|transaction|version] [OPTION]",
		Flags:    append(NodeFlags, InfoFlags...),
		Category: "INFO COMMANDS",
		Subcommands: []cli.Command{
			blockCommandSet,
			txCommandSet,
			versionCommand,
		},
		Description: ``,
	}
)

func infoCommand(context *cli.Context) error {
	showInfoHelp()
	return nil
}

func blockInfoUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println("Error:", err.Error())
	showBlockInfoHelp()
	return nil
}

func getCurrentBlockHeight(ctx *cli.Context) error {
	height, err := ontSdk.Rpc.GetBlockCount()
	if nil != err {
		log.Fatalf("Get block height information is error:  %s", err.Error())
		return err
	}
	fmt.Println("Current blockchain height: ", height)
	return nil
}

var blockCommandSet = cli.Command{
	Action:       utils.MigrateFlags(blockInfoCommand),
	Name:         "block",
	Usage:        "./ontology info block [OPTION]",
	Flags:        append(NodeFlags, InfoFlags...),
	OnUsageError: blockInfoUsageError,
	Category:     "INFO COMMANDS",
	Description:  ``,
	Subcommands: []cli.Command{
		{
			Action:      utils.MigrateFlags(getCurrentBlockHeight),
			Name:        "count",
			Usage:       "issue asset by command",
			Category:    "INFO COMMANDS",
			Description: ``,
		},
	},
}

func txInfoUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println("Error:", err.Error())
	showTxInfoHelp()
	return nil
}

var txCommandSet = cli.Command{
	Action:       utils.MigrateFlags(txInfoCommand),
	Name:         "tx",
	Usage:        "ontology info tx [OPTION]\n",
	Flags:        append(NodeFlags, InfoFlags...),
	OnUsageError: txInfoUsageError,
	Category:     "INFO COMMANDS",
	Description:  ``,
}

func versionInfoUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println("Error:", err.Error())
	showVersionInfoHelp()
	return nil
}

var versionCommand = cli.Command{
	Action:       utils.MigrateFlags(versionInfoCommand),
	Name:         "version",
	Usage:        "ontology info version\n",
	OnUsageError: versionInfoUsageError,
	Category:     "INFO COMMANDS",
	Description:  ``,
}

func versionInfoCommand(ctx *cli.Context) error {
	version, err := ontSdk.Rpc.GetVersion()
	if nil != err {
		log.Fatalf("Get version information is error:  %s", err.Error())
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
		echoBlockGracefully(resp)
		return nil
	}
	showTxInfoHelp()
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
			echoBlockGracefully(resp)
			return nil
		}
	} else if ctx.IsSet(utils.HashInfoFlag.Name) {
		blockHash := ctx.String(utils.HashInfoFlag.Name)
		if "" != blockHash {
			resp, err := request("GET", nil, restfulAddr()+"/api/v1/block/details/hash/"+blockHash)
			if err != nil {
				return err
			}
			echoBlockGracefully(resp)
			return nil
		}
	}
	showBlockInfoHelp()
	return nil
}

func echoBlockGracefully(block interface{}) {
	jsons, errs := json.Marshal(block)
	if errs != nil {
		log.Fatalf("Marshal json err:%s", errs.Error())
		return
	}

	var out bytes.Buffer
	err := json.Indent(&out, jsons, "", "\t")
	if err != nil {
		log.Fatalf("Gracefully format json err: %s", err.Error())
		return
	}
	out.WriteTo(os.Stdout)
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
