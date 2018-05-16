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
			Action:      txInfo,
			Name:        "tx",
			Usage:       "Display transaction information",
			Flags:       []cli.Flag{},
			Description: "Display transaction information",
		},
		{
			Action:      blockInfo,
			Name:        "block",
			Usage:       "Display block information",
			Flags:       []cli.Flag{},
			Description: "Display block information",
		},
	},
	Description: ``,
}

func blockInfo(ctx *cli.Context) error {
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
