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
	nutils "github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/urfave/cli"
	"strconv"
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
				utils.RPCPortFlag,
				utils.TransactionGasPriceFlag,
				utils.TransactionGasLimitFlag,
				utils.TransactionAssetFlag,
				utils.TransactionFromFlag,
				utils.TransactionToFlag,
				utils.TransactionAmountFlag,
				utils.WalletFileFlag,
			},
		},
		{
			Action:    approve,
			Name:      "approve",
			ArgsUsage: " ",
			Usage:     "Approve another user can transfer asset",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
				utils.TransactionGasPriceFlag,
				utils.TransactionGasLimitFlag,
				utils.ApproveAssetFlag,
				utils.ApproveAssetFromFlag,
				utils.ApproveAssetToFlag,
				utils.ApproveAmountFlag,
				utils.WalletFileFlag,
			},
		},
		{
			Action:    transferFrom,
			Name:      "transferfrom",
			ArgsUsage: " ",
			Usage:     "Using to transfer asset after approve",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
				utils.TransactionGasPriceFlag,
				utils.TransactionGasLimitFlag,
				utils.ApproveAssetFlag,
				utils.TransferFromSenderFlag,
				utils.ApproveAssetFromFlag,
				utils.ApproveAssetToFlag,
				utils.TransferFromAmountFlag,
				utils.WalletFileFlag,
			},
		},
		{
			Action:    getBalance,
			Name:      "balance",
			Usage:     "Show balance of ont and ong of specified account",
			ArgsUsage: "<address|label|index>",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
				utils.WalletFileFlag,
			},
		},
		{
			Action: getAllowance,
			Name:   "allowance",
			Usage:  "Show approve balance of ont or ong of specified account",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
				utils.ApproveAssetFlag,
				utils.ApproveAssetFromFlag,
				utils.ApproveAssetToFlag,
				utils.WalletFileFlag,
			},
		},
		{
			Action:    unboundOng,
			Name:      "unboundong",
			Usage:     "Show the balance of unbound ONG",
			ArgsUsage: "<address|label|index>",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
				utils.WalletFileFlag,
			},
		},
		{
			Action:    withdrawOng,
			Name:      "withdrawong",
			Usage:     "Withdraw ONG",
			ArgsUsage: "<address|label|index>",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
				utils.TransactionGasPriceFlag,
				utils.TransactionGasLimitFlag,
				utils.WalletFileFlag,
			},
		},
	},
}

func transfer(ctx *cli.Context) error {
	SetRpcPort(ctx)
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

	var amount uint64
	amountStr := ctx.String(utils.TransactionAmountFlag.Name)
	switch strings.ToLower(asset) {
	case "ont":
		amount = utils.ParseOnt(amountStr)
		amountStr = utils.FormatOnt(amount)
	case "ong":
		amount = utils.ParseOng(amountStr)
		amountStr = utils.FormatOng(amount)
	default:
		return fmt.Errorf("unsupport asset:%s", asset)
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)

	var signer *account.Account
	signer, err = cmdcom.GetAccount(ctx, fromAddr)
	if err != nil {
		return fmt.Errorf("GetAccount error:%s", err)
	}
	txHash, err := utils.Transfer(gasPrice, gasLimit, signer, asset, fromAddr, toAddr, amount)
	if err != nil {
		return fmt.Errorf("Transfer error:%s", err)
	}
	fmt.Printf("Transfer %s\n", strings.ToUpper(asset))
	fmt.Printf("  From:%s\n", fromAddr)
	fmt.Printf("  To:%s\n", toAddr)
	fmt.Printf("  Amount:%s\n", amountStr)
	fmt.Printf("  TxHash:%s\n", txHash)
	fmt.Printf("\nTip:\n")
	fmt.Printf("  Using './ontology info status %s' to query transaction status\n", txHash)
	return nil
}

func getBalance(ctx *cli.Context) error {
	SetRpcPort(ctx)
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

	ong, err := strconv.ParseUint(balance.Ong, 10, 64)
	if err != nil {
		return err
	}
	fmt.Printf("BalanceOf:%s\n", accAddr)
	fmt.Printf("  ONT:%s\n", balance.Ont)
	fmt.Printf("  ONG:%s\n", utils.FormatOng(ong))
	return nil
}

func getAllowance(ctx *cli.Context) error {
	SetRpcPort(ctx)
	from := ctx.String(utils.GetFlagName(utils.ApproveAssetFromFlag))
	to := ctx.String(utils.GetFlagName(utils.ApproveAssetToFlag))
	if from == "" || to == "" {
		fmt.Printf("Missing approve from or to argument\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	asset := ctx.String(utils.GetFlagName(utils.ApproveAssetFlag))
	if asset == "" {
		asset = utils.ASSET_ONT
	}
	fromAddr, err := cmdcom.ParseAddress(from, ctx)
	if err != nil {
		return err
	}
	toAddr, err := cmdcom.ParseAddress(to, ctx)
	if err != nil {
		return err
	}
	balanceStr, err := utils.GetAllowance(asset, fromAddr, toAddr)
	if err != nil {
		return err
	}
	switch strings.ToLower(asset) {
	case "ont":
	case "ong":
		balance, err := strconv.ParseUint(balanceStr, 10, 64)
		if err != nil {
			return err
		}
		balanceStr = utils.FormatOng(balance)
	default:
		return fmt.Errorf("unsupport asset:%s", asset)
	}
	fmt.Printf("Allowance:%s\n", asset)
	fmt.Printf("  From:%s\n", fromAddr)
	fmt.Printf("  To:%s\n", toAddr)
	fmt.Printf("  Balance:%s\n", balanceStr)
	return nil
}

func approve(ctx *cli.Context) error {
	SetRpcPort(ctx)
	asset := ctx.String(utils.GetFlagName(utils.ApproveAssetFlag))
	from := ctx.String(utils.GetFlagName(utils.ApproveAssetFromFlag))
	to := ctx.String(utils.GetFlagName(utils.ApproveAssetToFlag))
	amountStr := ctx.String(utils.GetFlagName(utils.ApproveAmountFlag))
	if asset == "" ||
		from == "" ||
		to == "" ||
		amountStr == "" {
		fmt.Printf("Missing asset, from, to, or amount argument\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	fromAddr, err := cmdcom.ParseAddress(from, ctx)
	if err != nil {
		return err
	}
	toAddr, err := cmdcom.ParseAddress(to, ctx)
	if err != nil {
		return err
	}
	var amount uint64
	switch strings.ToLower(asset) {
	case "ont":
		amount = utils.ParseOnt(amountStr)
		amountStr = utils.FormatOnt(amount)
	case "ong":
		amount = utils.ParseOng(amountStr)
		amountStr = utils.FormatOng(amount)
	default:
		return fmt.Errorf("unsupport asset:%s", asset)
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)

	var signer *account.Account
	signer, err = cmdcom.GetAccount(ctx, fromAddr)
	if err != nil {
		return fmt.Errorf("GetAccount error:%s", err)
	}

	txHash, err := utils.Approve(gasPrice, gasLimit, signer, asset, fromAddr, toAddr, amount)
	if err != nil {
		return fmt.Errorf("approve error:%s", err)
	}

	fmt.Printf("Approve:\n")
	fmt.Printf("  Asset:%s\n", asset)
	fmt.Printf("  From:%s\n", fromAddr)
	fmt.Printf("  To:%s\n", toAddr)
	fmt.Printf("  Amount:%s\n", amountStr)
	fmt.Printf("  TxHash:%s\n", txHash)
	fmt.Printf("\nTip:\n")
	fmt.Printf("  Using './ontology info status %s' to query transaction status\n", txHash)
	return nil
}

func transferFrom(ctx *cli.Context) error {
	SetRpcPort(ctx)
	asset := ctx.String(utils.GetFlagName(utils.ApproveAssetFlag))
	from := ctx.String(utils.GetFlagName(utils.ApproveAssetFromFlag))
	to := ctx.String(utils.GetFlagName(utils.ApproveAssetToFlag))
	amountStr := ctx.String(utils.GetFlagName(utils.TransferFromAmountFlag))
	if asset == "" ||
		from == "" ||
		to == "" ||
		amountStr == "" {
		fmt.Printf("Missing asset, from, to, or amount argument\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	fromAddr, err := cmdcom.ParseAddress(from, ctx)
	if err != nil {
		return err
	}
	toAddr, err := cmdcom.ParseAddress(to, ctx)
	if err != nil {
		return err
	}

	var sendAddr string
	sender := ctx.String(utils.GetFlagName(utils.TransferFromSenderFlag))
	if sender == "" {
		sendAddr = toAddr
	} else {
		sendAddr, err = cmdcom.ParseAddress(sender, ctx)
		if err != nil {
			return err
		}
	}

	var signer *account.Account
	signer, err = cmdcom.GetAccount(ctx, sendAddr)
	if err != nil {
		return fmt.Errorf("GetAccount error:%s", err)
	}

	var amount uint64
	switch strings.ToLower(asset) {
	case "ont":
		amount = utils.ParseOnt(amountStr)
		amountStr = utils.FormatOnt(amount)
	case "ong":
		amount = utils.ParseOng(amountStr)
		amountStr = utils.FormatOng(amount)
	default:
		return fmt.Errorf("unsupport asset:%s", asset)
	}
	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)

	txHash, err := utils.TransferFrom(gasPrice, gasLimit, signer, asset, sendAddr, fromAddr, toAddr, amount)
	if err != nil {
		return err
	}

	fmt.Printf("Transfer from:\n")
	fmt.Printf("  Asset:%s\n", asset)
	fmt.Printf("  Sender:%s\n", sendAddr)
	fmt.Printf("  From:%s\n", fromAddr)
	fmt.Printf("  To:%s\n", toAddr)
	fmt.Printf("  Amount:%s\n", amountStr)
	fmt.Printf("  TxHash:%s\n", txHash)
	fmt.Printf("\nTip:\n")
	fmt.Printf("  Using './ontology info status %s' to query transaction status\n", txHash)
	return nil
}

func unboundOng(ctx *cli.Context) error {
	SetRpcPort(ctx)
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
	fromAddr := nutils.OntContractAddress.ToBase58()
	balanceStr, err := utils.GetAllowance("ong", fromAddr, accAddr)
	if err != nil {
		return err
	}
	balance, err := strconv.ParseUint(balanceStr, 10, 64)
	if err != nil {
		return err
	}
	balanceStr = utils.FormatOng(balance)
	fmt.Printf("Unbound ONG:\n")
	fmt.Printf("  Account:%s\n", accAddr)
	fmt.Printf("  ONG:%s\n", balanceStr)
	return nil
}

func withdrawOng(ctx *cli.Context) error {
	SetRpcPort(ctx)
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
	fromAddr := nutils.OntContractAddress.ToBase58()
	balance, err := utils.GetAllowance("ong", fromAddr, accAddr)
	if err != nil {
		return err
	}

	amount, err := strconv.ParseUint(balance, 10, 64)
	if err != nil {
		return err
	}
	if amount <= 0 {
		return fmt.Errorf("Don't have unbound ong\n")
	}

	var signer *account.Account
	signer, err = cmdcom.GetAccount(ctx, accAddr)
	if err != nil {
		return fmt.Errorf("GetAccount error:%s", err)
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)

	txHash, err := utils.TransferFrom(gasPrice, gasLimit, signer, "ong", accAddr, fromAddr, accAddr, amount)
	if err != nil {
		return err
	}

	fmt.Printf("Withdraw ONG:\n")
	fmt.Printf("  Account:%s\n", accAddr)
	fmt.Printf("  Amount:%s\n", utils.FormatOng(amount))
	fmt.Printf("  TxHash:%s\n", txHash)
	fmt.Printf("\nTip:\n")
	fmt.Printf("  Using './ontology info status %s' to query transaction status\n", txHash)
	return nil
}
