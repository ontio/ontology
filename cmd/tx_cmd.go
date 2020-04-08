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
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	cmdcom "github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	nutils "github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/urfave/cli"
)

var SendTxCommand = cli.Command{
	Name:        "sendtx",
	Usage:       "Send raw transaction to Ontology",
	Description: "Send raw transaction to Ontology.",
	ArgsUsage:   "<rawtx>",
	Action:      sendTx,
	Flags: []cli.Flag{
		utils.RPCPortFlag,
		utils.PrepareExecTransactionFlag,
	},
}

func sendTx(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing raw tx argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	rawTx := ctx.Args().First()

	isPre := ctx.IsSet(utils.GetFlagName(utils.PrepareExecTransactionFlag))
	if isPre {
		preResult, err := utils.PrepareSendRawTransaction(rawTx)
		if err != nil {
			return err
		}
		if preResult.State == 0 {
			return fmt.Errorf("prepare execute transaction failed. %v", preResult)
		}
		PrintInfoMsg("Prepare execute transaction success.")
		PrintInfoMsg("Gas limit:%d", preResult.Gas)
		PrintInfoMsg("Result:%v", preResult.Result)
		return nil
	}
	txHash, err := utils.SendRawTransactionData(rawTx)
	if err != nil {
		return err
	}
	PrintInfoMsg("Send transaction success.")
	PrintInfoMsg("  TxHash:%s", txHash)
	PrintInfoMsg("\nTip:")
	PrintInfoMsg("  Using './ontology info status %s' to query transaction status.", txHash)
	return nil
}

var TxCommond = cli.Command{
	Name:  "buildtx",
	Usage: "Build transaction",
	Subcommands: []cli.Command{
		TransferTxCommond,
		ApproveTxCommond,
		TransferFromTxCommond,
		WithdrawONGTxCommond,
	},
	Description: "Build transaction",
}

var TransferTxCommond = cli.Command{
	Name:        "transfer",
	Usage:       "Build transfer transaction",
	Description: "Build transfer transaction.",
	Action:      transferTx,
	Flags: []cli.Flag{
		utils.WalletFileFlag,
		utils.TransactionGasPriceFlag,
		utils.TransactionGasLimitFlag,
		utils.TransactionPayerFlag,
		utils.TransactionAssetFlag,
		utils.TransactionFromFlag,
		utils.TransactionToFlag,
		utils.TransactionAmountFlag,
	},
}

var ApproveTxCommond = cli.Command{
	Name:        "approve",
	Usage:       "Build approve transaction",
	Description: "Build approve transaction.",
	Action:      approveTx,
	Flags: []cli.Flag{
		utils.WalletFileFlag,
		utils.TransactionGasPriceFlag,
		utils.TransactionGasLimitFlag,
		utils.TransactionPayerFlag,
		utils.ApproveAssetFlag,
		utils.ApproveAssetFromFlag,
		utils.ApproveAssetToFlag,
		utils.ApproveAmountFlag,
	},
}

var TransferFromTxCommond = cli.Command{
	Name:        "transferfrom",
	Usage:       "Build transfer from transaction",
	Description: "Build transfer from transaction.",
	Action:      transferFromTx,
	Flags: []cli.Flag{
		utils.WalletFileFlag,
		utils.TransactionGasPriceFlag,
		utils.TransactionGasLimitFlag,
		utils.ApproveAssetFlag,
		utils.TransactionPayerFlag,
		utils.TransferFromSenderFlag,
		utils.ApproveAssetFromFlag,
		utils.ApproveAssetToFlag,
		utils.TransferFromAmountFlag,
	},
}

var WithdrawONGTxCommond = cli.Command{
	Action:      withdrawONGTx,
	Name:        "withdrawong",
	Usage:       "Build Withdraw ONG transaction",
	Description: "Build Withdraw ONG transaction",
	ArgsUsage:   "<address|label|index>",
	Flags: []cli.Flag{
		utils.RPCPortFlag,
		utils.WalletFileFlag,
		utils.TransactionGasPriceFlag,
		utils.TransactionGasLimitFlag,
		utils.TransactionPayerFlag,
		utils.WithdrawONGAmountFlag,
		utils.WithdrawONGReceiveAccountFlag,
	},
}

func transferTx(ctx *cli.Context) error {
	if !ctx.IsSet(utils.GetFlagName(utils.TransactionToFlag)) ||
		!ctx.IsSet(utils.GetFlagName(utils.TransactionFromFlag)) ||
		!ctx.IsSet(utils.GetFlagName(utils.TransactionAmountFlag)) {
		PrintErrorMsg("Missing %s %s or %s argument.", utils.TransactionToFlag.Name, utils.TransactionFromFlag.Name, utils.TransactionAmountFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)

	asset := ctx.String(utils.GetFlagName(utils.TransactionAssetFlag))
	if asset == "" {
		asset = utils.ASSET_ONT
	}
	from := ctx.String(utils.GetFlagName(utils.TransactionFromFlag))
	fromAddr, err := cmdcom.ParseAddress(from, ctx)
	if err != nil {
		return err
	}
	to := ctx.String(utils.GetFlagName(utils.TransactionToFlag))
	toAddr, err := cmdcom.ParseAddress(to, ctx)
	if err != nil {
		return err
	}

	var payer common.Address
	payerAddr := ctx.String(utils.GetFlagName(utils.TransactionPayerFlag))
	if payerAddr != "" {
		payerAddr, err = cmdcom.ParseAddress(payerAddr, ctx)
		if err != nil {
			return err
		}
	} else {
		payerAddr = fromAddr
	}

	payer, err = common.AddressFromBase58(payerAddr)
	if err != nil {
		return fmt.Errorf("invalid payer address:%s", err)
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

	err = utils.CheckAssetAmount(asset, amount)
	if err != nil {
		return err
	}

	mutTx, err := utils.TransferTx(gasPrice, gasLimit, asset, fromAddr, toAddr, amount)
	if err != nil {
		return err
	}
	mutTx.Payer = payer

	tx, err := mutTx.IntoImmutable()
	if err != nil {
		return fmt.Errorf("IntoImmutable error:%s", err)
	}
	sink := common.ZeroCopySink{}
	tx.Serialization(&sink)
	PrintInfoMsg("Transfer raw tx:")
	PrintInfoMsg(hex.EncodeToString(sink.Bytes()))
	return nil
}

func approveTx(ctx *cli.Context) error {
	asset := ctx.String(utils.GetFlagName(utils.ApproveAssetFlag))
	from := ctx.String(utils.GetFlagName(utils.ApproveAssetFromFlag))
	to := ctx.String(utils.GetFlagName(utils.ApproveAssetToFlag))
	amountStr := ctx.String(utils.GetFlagName(utils.ApproveAmountFlag))
	if asset == "" ||
		from == "" ||
		to == "" ||
		amountStr == "" {
		PrintErrorMsg("Missing %s %s %s or %s argument.", utils.ApproveAssetFlag.Name, utils.ApproveAssetFromFlag.Name, utils.ApproveAssetToFlag.Name, utils.ApproveAmountFlag.Name)
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

	var payer common.Address
	payerAddr := ctx.String(utils.GetFlagName(utils.TransactionPayerFlag))
	if payerAddr != "" {
		payerAddr, err = cmdcom.ParseAddress(payerAddr, ctx)
		if err != nil {
			return err
		}
	} else {
		payerAddr = fromAddr
	}

	payer, err = common.AddressFromBase58(payerAddr)
	if err != nil {
		return fmt.Errorf("invalid payer address:%s", err)
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

	err = utils.CheckAssetAmount(asset, amount)
	if err != nil {
		return err
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)

	mutTx, err := utils.ApproveTx(gasPrice, gasLimit, asset, fromAddr, toAddr, amount)
	if err != nil {
		return err
	}
	mutTx.Payer = payer

	tx, err := mutTx.IntoImmutable()
	if err != nil {
		return fmt.Errorf("IntoImmutable error:%s", err)
	}
	sink := common.ZeroCopySink{}
	tx.Serialization(&sink)
	PrintInfoMsg("Approve raw tx:")
	PrintInfoMsg(hex.EncodeToString(sink.Bytes()))
	return nil
}

func transferFromTx(ctx *cli.Context) error {
	asset := ctx.String(utils.GetFlagName(utils.ApproveAssetFlag))
	from := ctx.String(utils.GetFlagName(utils.ApproveAssetFromFlag))
	to := ctx.String(utils.GetFlagName(utils.ApproveAssetToFlag))
	amountStr := ctx.String(utils.GetFlagName(utils.TransferFromAmountFlag))
	if asset == "" ||
		from == "" ||
		to == "" ||
		amountStr == "" {
		PrintErrorMsg("Missing %s %s %s or %s argument.", utils.ApproveAssetFlag.Name, utils.ApproveAssetFromFlag.Name, utils.ApproveAssetToFlag.Name, utils.TransferFromAmountFlag.Name)
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

	var payer common.Address
	payerAddr := ctx.String(utils.GetFlagName(utils.TransactionPayerFlag))
	if payerAddr != "" {
		payerAddr, err = cmdcom.ParseAddress(payerAddr, ctx)
		if err != nil {
			return err
		}
	} else {
		payerAddr = sendAddr
	}

	payer, err = common.AddressFromBase58(payerAddr)
	if err != nil {
		return fmt.Errorf("invalid payer address:%s", err)
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

	err = utils.CheckAssetAmount(asset, amount)
	if err != nil {
		return err
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)

	mutTx, err := utils.TransferFromTx(gasPrice, gasLimit, asset, sendAddr, fromAddr, toAddr, amount)
	if err != nil {
		return err
	}
	mutTx.Payer = payer

	tx, err := mutTx.IntoImmutable()
	if err != nil {
		return fmt.Errorf("IntoImmutable error:%s", err)
	}
	sink := common.ZeroCopySink{}
	tx.Serialization(&sink)
	PrintInfoMsg("TransferFrom raw tx:")
	PrintInfoMsg(hex.EncodeToString(sink.Bytes()))
	return nil
}

func withdrawONGTx(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing account argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	addrArg := ctx.Args().First()
	accAddr, err := cmdcom.ParseAddress(addrArg, ctx)
	if err != nil {
		return err
	}

	fromAddr := nutils.OntContractAddress.ToBase58()

	var amount uint64
	amountStr := ctx.String(utils.GetFlagName(utils.TransferFromAmountFlag))
	if amountStr == "" {
		balance, err := utils.GetAllowance("ong", fromAddr, accAddr)
		if err != nil {
			return err
		}
		amount, err = strconv.ParseUint(balance, 10, 64)
		if err != nil {
			return err
		}
		if amount <= 0 {
			return fmt.Errorf("haven't unbound ong")
		}
		amountStr = utils.FormatOng(amount)
	} else {
		amount = utils.ParseOng(amountStr)
		if amount <= 0 {
			return fmt.Errorf("haven't unbound ong")
		}
		amountStr = utils.FormatOng(amount)
	}

	var payer common.Address
	payerAddr := ctx.String(utils.GetFlagName(utils.TransactionPayerFlag))
	if payerAddr != "" {
		payerAddr, err = cmdcom.ParseAddress(payerAddr, ctx)
		if err != nil {
			return err
		}
	} else {
		payerAddr = accAddr
	}
	payer, err = common.AddressFromBase58(payerAddr)
	if err != nil {
		return fmt.Errorf("invalid payer address:%s", err)
	}

	var receiveAddr string
	receive := ctx.String(utils.GetFlagName(utils.WithdrawONGReceiveAccountFlag))
	if receive == "" {
		receiveAddr = accAddr
	} else {
		receiveAddr, err = cmdcom.ParseAddress(receive, ctx)
		if err != nil {
			return err
		}
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)

	PrintInfoMsg("Withdraw account:%s", accAddr)
	PrintInfoMsg("Receive account:%s", receiveAddr)
	PrintInfoMsg("Withdraw ONG amount:%v", amount)
	mutTx, err := utils.TransferFromTx(gasPrice, gasLimit, "ong", accAddr, fromAddr, receiveAddr, amount)
	if err != nil {
		return err
	}

	mutTx.Payer = payer
	tx, err := mutTx.IntoImmutable()
	if err != nil {
		return fmt.Errorf("IntoImmutable error:%s", err)
	}
	sink := common.ZeroCopySink{}
	tx.Serialization(&sink)
	PrintInfoMsg("Withdraw raw tx:")
	PrintInfoMsg(hex.EncodeToString(sink.Bytes()))
	return nil
}
