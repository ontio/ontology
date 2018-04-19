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
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/types"
	"github.com/urfave/cli"
)

var (
	ContractCommand = cli.Command{
		Name:         "contract",
		Action:       utils.MigrateFlags(contractCommand),
		Usage:        "ontology contract [invoke|deploy] [OPTION]",
		Category:     "CONTRACT COMMANDS",
		OnUsageError: contractUsageError,
		Description:  `account command`,
		Subcommands: []cli.Command{
			{
				Action:       utils.MigrateFlags(invokeContract),
				Name:         "invoke",
				OnUsageError: invokeUsageError,
				Usage:        "ontology invoke [OPTION]\n",
				Flags:        append(NodeFlags, ContractFlags...),
				Category:     "CONTRACT COMMANDS",
				Description:  ``,
			},
			{
				Action:       utils.MigrateFlags(deployContract),
				OnUsageError: deployUsageError,
				Name:         "deploy",
				Usage:        "ontology deploy [OPTION]\n",
				Flags:        append(NodeFlags, ContractFlags...),
				Category:     "CONTRACT COMMANDS",
				Description:  ``,
			},
		},
	}
)

func contractCommand(ctx *cli.Context) error {
	showContractHelp()
	return nil
}

func contractUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	showContractHelp()
	return nil
}

func invokeUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	showInvokeHelp()
	return nil
}

func invokeContract(ctx *cli.Context) error {
	if !ctx.IsSet(utils.ContractAddrFlag.Name) || !ctx.IsSet(utils.ContractParamsFlag.Name) {
		showInvokeHelp()
		return nil
	}

	client := account.GetClient(ctx)
	if client == nil {
		log.Fatal("Can't get local account.")
		os.Exit(1)
	}

	acct := client.GetDefaultAccount()

	contractAddr := ctx.String(utils.ContractAddrFlag.Name)
	params := ctx.String(utils.ContractParamsFlag.Name)
	if "" == contractAddr {
		log.Fatal("contract address does not allow empty")
	}

	addr, _ := common.AddressFromBase58(contractAddr)
	txHash, err := ontSdk.Rpc.InvokeNeoVMSmartContract(acct, new(big.Int), addr, []interface{}{params})
	if err != nil {
		log.Fatalf("InvokeSmartContract InvokeNeoVMSmartContract error:%s", err)
	} else {
		fmt.Printf("invoke transaction hash:%s", common.ToHexString(txHash[:]))
	}

	//WaitForGenerateBlock
	_, err = ontSdk.Rpc.WaitForGenerateBlock(30*time.Second, 1)
	if err != nil {
		log.Fatalf("InvokeSmartContract WaitForGenerateBlock error:%s", err)
	}
	return nil
}

func getVmType(vmType uint) types.VmType {
	switch vmType {
	case 0:
		return types.Native
	case 1:
		return types.NEOVM
	case 2:
		return types.WASMVM
	default:
		return types.Native
	}
}

func deployUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	showDeployHelp()
	return nil
}

func deployContract(ctx *cli.Context) error {
	if !ctx.IsSet(utils.ContractStorageFlag.Name) || !ctx.IsSet(utils.ContractVmTypeFlag.Name) ||
		!ctx.IsSet(utils.ContractCodeFlag.Name) || !ctx.IsSet(utils.ContractNameFlag.Name) ||
		!ctx.IsSet(utils.ContractVersionFlag.Name) || !ctx.IsSet(utils.ContractAuthorFlag.Name) ||
		!ctx.IsSet(utils.ContractEmailFlag.Name) || !ctx.IsSet(utils.ContractDescFlag.Name) {
		showDeployHelp()
		return nil
	}

	client := account.GetClient(ctx)
	if nil == client {
		log.Fatal("Can't get local account.")
	}

	acct := client.GetDefaultAccount()

	store := ctx.Bool(utils.ContractStorageFlag.Name)
	vmType := getVmType(ctx.Uint(utils.ContractVmTypeFlag.Name))

	codeDir := ctx.String(utils.ContractCodeFlag.Name)
	if "" == codeDir {
		log.Fatal("code dir is error, value does not allow null")
	}
	code, err := ioutil.ReadFile(codeDir)
	if err != nil {
		log.Fatal("error in read file", err.Error())
	}
	name := ctx.String(utils.ContractNameFlag.Name)
	version := ctx.String(utils.ContractVersionFlag.Name)
	author := ctx.String(utils.ContractAuthorFlag.Name)
	email := ctx.String(utils.ContractEmailFlag.Name)
	desc := ctx.String(utils.ContractDescFlag.Name)
	if "" == name || "" == version || "" == author || "" == email || "" == desc {
		log.Fatal("Params are not put completely: which contain code, name, version, author, email, desc")
	}

	trHash, err := ontSdk.Rpc.DeploySmartContract(acct, vmType, store, fmt.Sprintf("%s", code), name, version, author, email, desc)
	if err != nil {
		log.Fatal("Deploy smart error: ", err)
	}
	//WaitForGenerateBlock
	_, err = ontSdk.Rpc.WaitForGenerateBlock(30*time.Second, 1)
	if err != nil {
		log.Fatalf("DeploySmartContract WaitForGenerateBlock error:%s", err)
	} else {
		fmt.Printf("Deploy smartContract transaction hash: %s\n", common.ToHexString(trHash[:]))
	}

	return nil
}
