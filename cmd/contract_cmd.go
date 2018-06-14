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
	"encoding/json"
	"fmt"
	cmdcom "github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	httpcom "github.com/ontio/ontology/http/base/common"
	"github.com/urfave/cli"
	"io/ioutil"
	"strings"
)

var (
	ContractCommand = cli.Command{
		Name:        "contract",
		Action:      cli.ShowSubcommandHelp,
		Usage:       "Deploy or invoke smart contract",
		ArgsUsage:   " ",
		Description: `Deploy or invoke smart contract`,
		Subcommands: []cli.Command{
			{
				Action:    deployContract,
				Name:      "deploy",
				Usage:     "Deploy a smart contract to ontolgoy",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					utils.RPCPortFlag,
					utils.TransactionGasPriceFlag,
					utils.TransactionGasLimitFlag,
					utils.ContractStorageFlag,
					utils.ContractCodeFileFlag,
					utils.ContractNameFlag,
					utils.ContractVersionFlag,
					utils.ContractAuthorFlag,
					utils.ContractEmailFlag,
					utils.ContractDescFlag,
					utils.WalletFileFlag,
					utils.AccountAddressFlag,
				},
			},
			{
				Action:    invokeContract,
				Name:      "invoke",
				Usage:     "Invoke smart contract",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					utils.RPCPortFlag,
					utils.TransactionGasPriceFlag,
					utils.TransactionGasLimitFlag,
					utils.ContractAddrFlag,
					utils.ContractParamsFlag,
					utils.ContractVersionFlag,
					utils.ContractPrepareInvokeFlag,
					utils.ContranctReturnTypeFlag,
					utils.WalletFileFlag,
					utils.AccountAddressFlag,
				},
			},
			{
				Action:    invokeCodeContract,
				Name:      "invokeCode",
				Usage:     "Invoke smart contract by code",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					utils.RPCPortFlag,
					utils.ContractCodeFileFlag,
					utils.TransactionGasPriceFlag,
					utils.TransactionGasLimitFlag,
					utils.WalletFileFlag,
					utils.ContractPrepareInvokeFlag,
					utils.AccountAddressFlag,
				},
			},
		},
	}
)

func deployContract(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if !ctx.IsSet(utils.GetFlagName(utils.ContractCodeFileFlag)) ||
		!ctx.IsSet(utils.GetFlagName(utils.ContractNameFlag)) {
		fmt.Errorf("Missing code or name argument\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	signer, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("Get signer account error:%s", err)
	}

	store := ctx.Bool(utils.GetFlagName(utils.ContractStorageFlag))
	codeFile := ctx.String(utils.GetFlagName(utils.ContractCodeFileFlag))
	if "" == codeFile {
		return fmt.Errorf("Please specific code file")
	}
	codeStr, err := ioutil.ReadFile(codeFile)
	if err != nil {
		return fmt.Errorf("Read code:%s error:%s", codeFile, err)
	}

	name := ctx.String(utils.GetFlagName(utils.ContractNameFlag))
	version := ctx.Int(utils.GetFlagName(utils.ContractVersionFlag))
	author := ctx.String(utils.GetFlagName(utils.ContractAuthorFlag))
	email := ctx.String(utils.GetFlagName(utils.ContractEmailFlag))
	desc := ctx.String(utils.GetFlagName(utils.ContractDescFlag))
	code := strings.TrimSpace(string(codeStr))
	gasPrice := ctx.Uint64(utils.GetFlagName(utils.TransactionGasPriceFlag))
	gasLimit := ctx.Uint64(utils.GetFlagName(utils.TransactionGasLimitFlag))
	cversion := fmt.Sprintf("%s", version)

	txHash, err := utils.DeployContract(gasPrice, gasLimit, signer, store, code, name, cversion, author, email, desc)
	if err != nil {
		return fmt.Errorf("DeployContract error:%s", err)
	}
	c, _ := common.HexToBytes(code)
	address := types.AddressFromVmCode(c)
	fmt.Printf("Deploy contract:\n")
	fmt.Printf("  Contract Address:%s\n", address.ToHexString())
	fmt.Printf("  TxHash:%s\n", txHash)
	fmt.Printf("\nTip:\n")
	fmt.Printf("  Using './ontology info status %s' to query transaction status\n", txHash)
	return nil
}

func invokeCodeContract(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if !ctx.IsSet(utils.GetFlagName(utils.ContractCodeFileFlag)) {
		fmt.Printf("Missing code or name argument\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	signer, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("Get signer account error:%s", err)
	}
	codeFile := ctx.String(utils.GetFlagName(utils.ContractCodeFileFlag))
	if "" == codeFile {
		return fmt.Errorf("Please specific code file")
	}
	codeStr, err := ioutil.ReadFile(codeFile)
	if err != nil {
		return fmt.Errorf("Read code:%s error:%s", codeFile, err)
	}
	code := strings.TrimSpace(string(codeStr))
	c, err := common.HexToBytes(code)
	if err != nil {
		return fmt.Errorf("hex to bytes error:%s", err)
	}

	if ctx.IsSet(utils.GetFlagName(utils.ContractPrepareInvokeFlag)) {
		preResult, err := utils.PrepareInvokeCodeNeoVMContract(c)
		if err != nil {
			return fmt.Errorf("PrepareInvokeCodeNeoVMContract error:%s", err)
		}
		if preResult.State == 0 {
			return fmt.Errorf("Contract invoke failed\n")
		}
		fmt.Printf("Contract invoke successfully\n")
		fmt.Printf("Gas consumed:%d\n", preResult.Gas)

		rawReturnTypes := ctx.String(utils.GetFlagName(utils.ContranctReturnTypeFlag))
		if rawReturnTypes == "" {
			fmt.Printf("Return:%s (raw value)\n", preResult.Result)
			return nil
		}
		values, err := utils.ParseReturnValue(preResult.Result, rawReturnTypes)
		if err != nil {
			return fmt.Errorf("parseReturnValue values:%+v types:%s error:%s", values, rawReturnTypes, err)
		}
		switch len(values) {
		case 0:
			fmt.Printf("Return: nil\n")
		case 1:
			fmt.Printf("Return:%+v\n", values[0])
		default:
			fmt.Printf("Return:%+v\n", values)
		}
		return nil
	}
	gasPrice := ctx.Uint64(utils.GetFlagName(utils.TransactionGasPriceFlag))
	gasLimit := ctx.Uint64(utils.GetFlagName(utils.TransactionGasLimitFlag))

	invokeTx, err := httpcom.NewSmartContractTransaction(gasPrice, gasLimit, c)
	if err != nil {
		return err
	}
	err = utils.SignTransaction(signer, invokeTx)
	if err != nil {
		return fmt.Errorf("SignTransaction error:%s", err)
	}
	txHash, err := utils.SendRawTransaction(invokeTx)
	if err != nil {
		return fmt.Errorf("SendTransaction error:%s", err)
	}

	fmt.Printf("TxHash:%s\n", txHash)
	fmt.Printf("\nTip:\n")
	fmt.Printf("  Using './ontology info status %s' to query transaction status\n", txHash)
	return nil
}

func invokeContract(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if !ctx.IsSet(utils.GetFlagName(utils.ContractAddrFlag)) {
		fmt.Printf("Missing contract address argument.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	contractAddrStr := ctx.String(utils.GetFlagName(utils.ContractAddrFlag))
	contractAddr, err := common.AddressFromHexString(contractAddrStr)
	if err != nil {
		return fmt.Errorf("Invalid contract address error:%s", err)
	}

	paramsStr := ctx.String(utils.GetFlagName(utils.ContractParamsFlag))
	params, err := utils.ParseParams(paramsStr)
	if err != nil {
		return fmt.Errorf("parseParams error:%s", err)
	}

	paramData, _ := json.Marshal(params)
	fmt.Printf("Invoke:%x Params:%s\n", contractAddr[:], paramData)

	if ctx.IsSet(utils.GetFlagName(utils.ContractPrepareInvokeFlag)) {
		preResult, err := utils.PrepareInvokeNeoVMContract(contractAddr, params)
		if err != nil {
			return fmt.Errorf("PrepareInvokeNeoVMSmartContact error:%s", err)
		}
		if preResult.State == 0 {
			return fmt.Errorf("Contract invoke failed\n")
		}
		fmt.Printf("Contract invoke successfully\n")
		fmt.Printf("Gas consumed:%d\n", preResult.Gas)

		rawReturnTypes := ctx.String(utils.GetFlagName(utils.ContranctReturnTypeFlag))
		if rawReturnTypes == "" {
			fmt.Printf("Return:%s (raw value)\n", preResult.Result)
			return nil
		}
		values, err := utils.ParseReturnValue(preResult.Result, rawReturnTypes)
		if err != nil {
			return fmt.Errorf("parseReturnValue values:%+v types:%s error:%s", values, rawReturnTypes, err)
		}
		switch len(values) {
		case 0:
			fmt.Printf("Return: nil\n")
		case 1:
			fmt.Printf("Return:%+v\n", values[0])
		default:
			fmt.Printf("Return:%+v\n", values)
		}
		return nil
	}
	signer, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("Get signer account error:%s", err)
	}
	gasPrice := ctx.Uint64(utils.GetFlagName(utils.TransactionGasPriceFlag))
	gasLimit := ctx.Uint64(utils.GetFlagName(utils.TransactionGasLimitFlag))

	txHash, err := utils.InvokeNeoVMContract(gasPrice, gasLimit, signer, contractAddr, params)
	if err != nil {
		return fmt.Errorf("Invoke NeoVM contract error:%s", err)
	}

	fmt.Printf("TxHash:%s\n", txHash)
	fmt.Printf("\nTip:\n")
	fmt.Printf("  Using './ontology info status %s' to query transaction status\n", txHash)
	return nil
}
