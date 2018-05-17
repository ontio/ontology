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
	"github.com/ontio/ontology/account"
	cmdcom "github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/urfave/cli"
	"strings"
)

var AssetCommand = cli.Command{
	Name:        "asset",
	Usage:       "Handle assets",
	Description: `asset control`,
	Subcommands: []cli.Command{
		{
			Action:      transfer,
			Name:        "transfer",
			Usage:       "Transfer ont or ong to another account",
			ArgsUsage:   " ",
			Description: "Transfer ont or ong to another account. If from address does not specified, using default account",
			Flags: []cli.Flag{
				utils.TransactionGasPrice,
				utils.TransactionGasLimit,
				utils.TransactionAssetFlag,
				utils.TransactionFromFlag,
				utils.TransactionToFlag,
				utils.TransactionAmountFlag,
				utils.WalletFileFlag,
				utils.AccountAddressFlag,
			},
		},
		{
			Action:    getBalance,
			Name:      "balance",
			Usage:     "Show balance of ont and ong of specified account",
			ArgsUsage: "<address|label|index>",
			Flags: []cli.Flag{
				utils.WalletFileFlag,
			},
		},
		{
			Action:      queryTransferStatus,
			Name:        "status",
			Usage:       "Display asset status",
			ArgsUsage:   "<txhash>",
			Description: `Display asset transfer status of transfer transaction.`,
			Flags:       []cli.Flag{},
		},
	},
}

func transfer(ctx *cli.Context) error {
	if !ctx.IsSet(utils.GetFlagName(utils.TransactionToFlag)) ||
		!ctx.IsSet(utils.GetFlagName(utils.TransactionFromFlag)) ||
		!ctx.IsSet(utils.GetFlagName(utils.TransactionAmountFlag)) {
		fmt.Println("Missing from, to or amount flag\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	asset := ctx.String(utils.GetFlagName(utils.TransactionAssetFlag))
	if asset == "" {
		asset = utils.ASSET_ONT
	}
	from := ctx.String(utils.TransactionFromFlag.Name)
	fromAddr, err := cmdcom.ParseAddress(from, ctx)
	if err != nil {
		return fmt.Errorf("Parse from address:%s error:%s", from, err)
	}
	to := ctx.String(utils.TransactionToFlag.Name)
	toAddr, err := cmdcom.ParseAddress(to, ctx)
	if err != nil {
		return fmt.Errorf("Parse to address:%s error:%s", to, err)
	}
	amount := ctx.Uint64(utils.TransactionAmountFlag.Name)
	gasPrice := ctx.Uint64(utils.TransactionGasPrice.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimit.Name)

	ctx.Set(utils.AccountAddressFlag.Name, from)
	var signer *account.Account
	signer, err = cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("GetAccount error:%s", err)
	}
	txHash, err := utils.Transfer(gasPrice, gasLimit, signer, asset, fromAddr, toAddr, amount)
	if err != nil {
		return fmt.Errorf("Transfer error:%s", err)
	}
	fmt.Printf("Transfer %s\n", strings.ToUpper(asset))
	fmt.Printf("From:%s\n", fromAddr)
	fmt.Printf("To:%s\n", toAddr)
	fmt.Printf("Amount:%d\n", amount)
	fmt.Printf("TxHash:%s\n", txHash)
	return nil
}

func getBalance(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		fmt.Println("Missing argument. Account address, label or index expected.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	addrArg := ctx.Args().First()
	accAddr, err := cmdcom.ParseAddress(addrArg, ctx)
	if err != nil {
		return err
	}
	balance, err := utils.GetBalance(accAddr)
	if err != nil {
		return err
	}
	fmt.Printf("BalanceOf:%s\n", accAddr)
	fmt.Printf("ONT:%s\n", balance.Ont)
	fmt.Printf("ONG:%s\n", balance.Ong)
	fmt.Printf("ONGApprove:%s\n", balance.OngAppove)
	return nil
}

func queryTransferStatus(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		fmt.Println("Missing argument. TxHash expected.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	txHash := ctx.Args().First()
	evtInfos, err := utils.GetSmartContractEvent(txHash)
	if err != nil {
		return fmt.Errorf("GetSmartContractEvent error:%s", err)
	}
	if len(evtInfos) == 0 {
		fmt.Println("Cannot find event log")
		return nil
	}
	for _, eventInfo := range evtInfos {
		states := eventInfo.States.([]interface{})
		fmt.Printf("Transaction:%s success\n", states[0])
		fmt.Printf("From:%s\n", states[1])
		fmt.Printf("To:%s\n", states[2])
		fmt.Printf("Amount:%v\n", states[3])
	}
	return nil
}
