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
	"github.com/urfave/cli"
	"strconv"
)

var InfoCommand = cli.Command{
	Name:  "info",
	Usage: "Display informations about the chain",
	Subcommands: []cli.Command{
		{
			Action: txInfo,
			Name:   "tx",
			Usage:  "Display transaction information",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
			},
			Description: "Display transaction information",
		},
		{
			Action: blockInfo,
			Name:   "block",
			Usage:  "Display block information",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
			},
			Description: "Display block information",
		},
		{
			Action:      txStatus,
			Name:        "status",
			Usage:       "Display transaction status",
			ArgsUsage:   "<txhash>",
			Description: `Display status of transaction.`,
			Flags: []cli.Flag{
				utils.RPCPortFlag,
			},
		},
		{
			Action:      curBlockHeight,
			Name:        "curblockheight",
			Usage:       "Display the current block height",
			ArgsUsage:   "",
			Description: `Display the current block height.`,
			Flags: []cli.Flag{
				utils.RPCPortFlag,
			},
		},
	},
	Description: ``,
}

func blockInfo(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NArg() < 1 {
		fmt.Println("Missing argument. BlockHash or height expected.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	var data []byte
	var err error
	var height int64
	if len(ctx.Args().First()) > 30 {
		blockHash := ctx.Args().First()
		data, err = utils.GetBlock(blockHash)
	} else {
		height, err = strconv.ParseInt(ctx.Args().First(), 10, 64)
		if err != nil {
			return fmt.Errorf("Arg:%s invalid block hash or block height\n", ctx.Args().First())
		}
		data, err = utils.GetBlock(height)
	}
	if err != nil {
		return err
	}
	var out bytes.Buffer
	err = json.Indent(&out, data, "", "   ")
	if err != nil {
		return err
	}
	fmt.Println(out.String())
	return nil
}

func txInfo(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NArg() < 1 {
		fmt.Println("Missing argument. TxHash expected.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	txInfo, err := utils.GetRawTransaction(ctx.Args().First())
	if err != nil {
		return err
	}
	var out bytes.Buffer
	err = json.Indent(&out, txInfo, "", "   ")
	if err != nil {
		return err
	}
	fmt.Println(out.String())
	return nil
}

func txStatus(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NArg() < 1 {
		fmt.Println("Missing argument. TxHash expected.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	txHash := ctx.Args().First()
	evtInfos, err := utils.GetSmartContractEventInfo(txHash)
	if err != nil {
		return fmt.Errorf("GetSmartContractEvent error:%s", err)
	}

	fmt.Printf("Transaction states:\n")
	var out bytes.Buffer
	err = json.Indent(&out, evtInfos, "", "   ")
	if err != nil {
		return err
	}
	fmt.Println(out.String())
	return nil
}

func curBlockHeight(ctx *cli.Context) error {
	SetRpcPort(ctx)
	count, err := utils.GetBlockCount()
	if err != nil {
		return err
	}
	fmt.Printf("CurrentBlockHeight:%d\n", count-1)
	return nil
}
