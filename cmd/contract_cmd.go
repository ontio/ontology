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
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"
	"strings"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/password"
	"github.com/ontio/ontology/smartcontract/types"
	"github.com/urfave/cli"
)

var (
	ContractCommand = cli.Command{
		Name:         "contract",
		Action:       utils.MigrateFlags(contractCommand),
		Usage:        "Deploy or invoke smart contract",
		ArgsUsage:    " ",
		OnUsageError: contractUsageError,
		Description:  `Deploy or invoke smart contract`,
		Subcommands: []cli.Command{
			{
				Action:       utils.MigrateFlags(invokeContract),
				Name:         "invoke",
				OnUsageError: invokeUsageError,
				Usage:        "Invoke a deployed smart contract",
				ArgsUsage:    " ",
				Flags: []cli.Flag{
					utils.ContractAddrFlag,
					utils.ContractParamsFlag,
					utils.AccountFileFlag,
					utils.AccountPassFlag,
				},
				Description: ``,
			},
			{
				Action:       utils.MigrateFlags(deployContract),
				OnUsageError: deployUsageError,
				Name:         "deploy",
				Usage:        "Deploy a smart contract to the chain",
				ArgsUsage:    " ",
				Flags: []cli.Flag{
					utils.ContractVmTypeFlag,
					utils.ContractStorageFlag,
					utils.ContractCodeFlag,
					utils.ContractNameFlag,
					utils.ContractVersionFlag,
					utils.ContractAuthorFlag,
					utils.ContractEmailFlag,
					utils.ContractDescFlag,
					utils.AccountFileFlag,
					utils.AccountPassFlag,
				},
				Description: ``,
			},
		},
	}
)

func contractCommand(ctx *cli.Context) error {
	cli.ShowSubcommandHelp(ctx)
	return nil
}

func contractUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err.Error(), "\n")
	cli.ShowSubcommandHelp(context)
	return nil
}

func invokeUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err.Error(), "\n")
	cli.ShowSubcommandHelp(context)
	return nil
}

func invokeContract(ctx *cli.Context) error {
	if !ctx.IsSet(utils.ContractAddrFlag.Name) || !ctx.IsSet(utils.ContractParamsFlag.Name) {
		fmt.Println("Missing argument.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	var wallet = account.WALLET_FILENAME
	if ctx.IsSet("file") {
		wallet = ctx.String("file")
	}
	var passwd []byte
	var err error
	if ctx.IsSet("password") {
		passwd = []byte(ctx.String("password"))
	} else {
		passwd, err = password.GetAccountPassword()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return errors.New("input password error")
		}
	}
	client := account.Open(wallet, passwd)
	for i, _ := range passwd {
		passwd[i] = 0
	}
	if client == nil {
		fmt.Println("Can't get local account")
		return errors.New("Get client is nil")
	}

	acct := client.GetDefaultAccount()
	if acct == nil {
		fmt.Println("Can't get default account.")
		return errors.New("Deploy contract failed.")
	}

	contractAddr := ctx.String(utils.ContractAddrFlag.Name)
	params := ctx.String(utils.ContractParamsFlag.Name)
	if "" == contractAddr {
		fmt.Println("contract address does not allow empty")
	}

	addr, err := common.AddressFromBase58(contractAddr)
	if err != nil {
		fmt.Println("Parase contract address error, please use correct smart contract address")
		return err
	}

	txHash, err := ontSdk.Rpc.InvokeNeoVMSmartContract(acct, new(big.Int), addr, []interface{}{params})
	if err != nil {
		fmt.Printf("InvokeSmartContract InvokeNeoVMSmartContract error:%s", err)
		return err
	} else {
		fmt.Printf("invoke transaction hash:%s", common.ToHexString(txHash[:]))
	}

	//WaitForGenerateBlock
	_, err = ontSdk.Rpc.WaitForGenerateBlock(30*time.Second, 1)
	if err != nil {
		fmt.Printf("InvokeSmartContract WaitForGenerateBlock error:%s", err)
	}
	return err
}

func getVmType(vmType string) types.VmType {
	switch vmType {
	case "neovm":
		return types.NEOVM
	case "wasm":
		return types.WASMVM
	default:
		return types.NEOVM
	}
}

func deployUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err.Error(), "\n")
	cli.ShowSubcommandHelp(context)
	return nil
}

func deployContract(ctx *cli.Context) error {
	if !ctx.IsSet(utils.ContractStorageFlag.Name) || !ctx.IsSet(utils.ContractVmTypeFlag.Name) ||
		!ctx.IsSet(utils.ContractCodeFlag.Name) || !ctx.IsSet(utils.ContractNameFlag.Name) ||
		!ctx.IsSet(utils.ContractVersionFlag.Name) || !ctx.IsSet(utils.ContractAuthorFlag.Name) ||
		!ctx.IsSet(utils.ContractEmailFlag.Name) || !ctx.IsSet(utils.ContractDescFlag.Name) {
		fmt.Println("Missing argument.\n")
		cli.ShowSubcommandHelp(ctx)
		return errors.New("Parameter is err")
	}

	var wallet = account.WALLET_FILENAME
	if ctx.IsSet("file") {
		wallet = ctx.String("file")
	}
	var passwd []byte
	var err error
	if ctx.IsSet("password") {
		passwd = []byte(ctx.String("password"))
	} else {
		passwd, err = password.GetAccountPassword()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return errors.New("input password error")
		}
	}
	client := account.Open(wallet, passwd)
	for i, _ := range passwd {
		passwd[i] = 0
	}
	if nil == client {
		fmt.Println("Error load wallet file", wallet)
		return errors.New("Get client return nil")
	}

	acct := client.GetDefaultAccount()
	if acct == nil {
		fmt.Println("Can't get default account.")
		return errors.New("Deploy contract failed.")
	}

	store := ctx.Bool(utils.ContractStorageFlag.Name)
	vmType := getVmType(ctx.String(utils.ContractVmTypeFlag.Name))

	codeDir := ctx.String(utils.ContractCodeFlag.Name)
	if "" == codeDir {
		fmt.Println("Code dir is error, value does not allow null")
		return errors.New("Smart contract code dir does not allow empty")
	}
	code, err := ioutil.ReadFile(codeDir)
	if err != nil {
		fmt.Printf("Error in read file,%s", err.Error())
		return err
	}

	name := ctx.String(utils.ContractNameFlag.Name)
	version := ctx.String(utils.ContractVersionFlag.Name)
	author := ctx.String(utils.ContractAuthorFlag.Name)
	email := ctx.String(utils.ContractEmailFlag.Name)
	desc := ctx.String(utils.ContractDescFlag.Name)

	trHash, err := ontSdk.Rpc.DeploySmartContract(acct, vmType, store, strings.TrimSpace(fmt.Sprintf("%s", code)), name, version, author, email, desc)
	if err != nil {
		fmt.Printf("Deploy smart error: %s", err.Error())
		return err
	}
	//WaitForGenerateBlock
	_, err = ontSdk.Rpc.WaitForGenerateBlock(30*time.Second, 1)
	if err != nil {
		fmt.Printf("DeploySmartContract WaitForGenerateBlock error:%s", err.Error())
		return err
	} else {
		fmt.Printf("Deploy smartContract transaction hash: %s\n", common.ToHexString(trHash[:]))
	}

	return nil
}
