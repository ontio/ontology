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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/ontio/ontology/cmd/actor"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	ldgactor "github.com/ontio/ontology/core/ledger/actor"
	bcomn "github.com/ontio/ontology/http/base/common"
	"github.com/urfave/cli"
)

var blockCommandSet = cli.Command{
	Action:      utils.MigrateFlags(blockInfoCommand),
	Name:        "block",
	Usage:       "register asset",
	Flags:       append(NodeFlags, InfoFlags...),
	Category:    "INFO COMMANDS",
	Description: ``,
	Subcommands: []cli.Command{
		{
			Action:      utils.MigrateFlags(getCurrentBlockHeight),
			Name:        "height",
			Usage:       "issue asset by command",
			Category:    "INFO COMMANDS",
			Description: ``,
		},
	},
}

var chainCommandSet = cli.Command{
	Action:      utils.MigrateFlags(chainInfoCommand),
	Name:        "chain",
	Usage:       "./ontology info chain [OPTION]",
	Flags:       append(NodeFlags, InfoFlags...),
	Category:    "INFO COMMANDS",
	Description: ``,
}

var trxCommandSet = cli.Command{
	Action:      utils.MigrateFlags(trxInfoCommand),
	Name:        "trx",
	Usage:       "./ontology info trx [OPTION]",
	Flags:       append(NodeFlags, InfoFlags...),
	Category:    "INFO COMMANDS",
	Description: ``,
}

var versionCommand = cli.Command{
	Action:      utils.MigrateFlags(versionInfoCommand),
	Name:        "version",
	Usage:       "./ontology version",
	Category:    "INFO COMMANDS",
	Description: ``,
}

var (
	InfoCommand = cli.Command{
		//Action:   utils.MigrateFlags(infoCommand),
		Name:     "info",
		Usage:    "show block/chain/transaction info",
		Category: "INFO COMMANDS",
		Subcommands: []cli.Command{
			blockCommandSet,
			chainCommandSet,
			trxCommandSet,
			versionCommand,
		},
		Description: ``,
	}
)

func versionInfoCommand(ctx *cli.Context) error {
	fmt.Println("Node version: ", config.Parameters.Version)
	return nil
}

func getCurrentBlockHeight(ctx *cli.Context) error {
	ledger.DefLedger, _ = ledger.NewLedger()
	ldgerActor := ldgactor.NewLedgerActor()
	ledgerPID := ldgerActor.Start()
	actor.SetLedgerPid(ledgerPID)
	height, _ := actor.CurrentBlockHeight()
	fmt.Println("Current blockchain height: ", height)
	return nil
}

func blockInfoCommand(ctx *cli.Context) error {
	/*init ledger actor for info command*/
	ledger.DefLedger, _ = ledger.NewLedger()
	ldgerActor := ldgactor.NewLedgerActor()
	ledgerPID := ldgerActor.Start()
	actor.SetLedgerPid(ledgerPID)

	val := ctx.GlobalString(utils.HeightInfoFlag.Name)
	height, _ := strconv.Atoi(val)
	blockHash := ctx.GlobalString(utils.BHashInfoFlag.Name)

	switch {
	case height >= 0:
		hash, _ := actor.GetBlockHashFromStore(uint32(height))
		block, err := actor.GetBlockFromStore(hash)

		if err != nil {
			log.Errorf("GetBlock GetBlockFromStore BlockHash:%x error:%s", hash, err)
			return nil
		}
		if block == nil || block.Header == nil {
			return nil
		}
		jsons, errs := json.Marshal(bcomn.GetBlockInfo(block))
		if errs != nil {
			fmt.Println(errs.Error())
		}

		var out bytes.Buffer
		err = json.Indent(&out, jsons, "", "\t")
		if err != nil {
			return nil
		}
		out.WriteTo(os.Stdout)
		return nil

	case "" != blockHash:
		var hash common.Uint256
		hex, err := hex.DecodeString(blockHash)
		if err != nil {
			log.Errorf("")
		}
		if err := hash.Deserialize(bytes.NewReader(hex)); err != nil {
			log.Errorf("")
		}
		block, err := actor.GetBlockFromStore(hash)
		if err != nil {
			log.Errorf("GetBlock GetBlockFromStore BlockHash:%x error:%s", hash, err)
			return nil
		}
		if block == nil || block.Header == nil {
			return nil
		}
		jsons, errs := json.Marshal(bcomn.GetBlockInfo(block))
		if errs != nil {
			fmt.Println(errs.Error())
		}

		var out bytes.Buffer
		err = json.Indent(&out, jsons, "", "\t")
		if err != nil {
			return nil
		}
		out.WriteTo(os.Stdout)
		return nil

	default:
		return nil
	}
}

func trxInfoCommand(ctx *cli.Context) error {
	ledger.DefLedger, _ = ledger.NewLedger()
	ldgerActor := ldgactor.NewLedgerActor()
	ledgerPID := ldgerActor.Start()
	actor.SetLedgerPid(ledgerPID)

	trxHash := ctx.GlobalString(utils.BTrxInfoFlag.Name)

	hex, err := hex.DecodeString(trxHash)
	if err != nil {
		log.Errorf("error for trxHash")
	}
	var hash common.Uint256
	err = hash.Deserialize(bytes.NewReader(hex))
	if err != nil {
	}
	trx, err := actor.GetTransaction(hash)
	bcomn.TransArryByteToHexString(trx)

	jsons, errs := json.Marshal(bcomn.TransArryByteToHexString(trx))
	if errs != nil {
		fmt.Println(errs.Error())
	}

	var out bytes.Buffer
	err = json.Indent(&out, jsons, "", "\t")
	if err != nil {
		return nil
	}
	out.WriteTo(os.Stdout)
	return nil
}

func chainInfoCommand(ctx *cli.Context) error {
	return nil
}
