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
		Name:        "contract",
		Usage:       "controll contract,just as deploy/invoke.etc",
		Category:    "CONTRACT COMMANDS",
		Description: `account command`,
		Subcommands: []cli.Command{
			{
				Action:      utils.MigrateFlags(invokeContract),
				Name:        "invoke",
				Usage:       "./ontology invoke [OPTION]",
				Flags:       append(append(NodeFlags, RpcFlags...), WhisperFlags...),
				Category:    "CONTRACT COMMANDS",
				Description: ``,
			},
			{
				Action:      utils.MigrateFlags(deployContract),
				Name:        "deploy",
				Usage:       "./ontology deploy [OPTION]",
				Flags:       append(append(NodeFlags, RpcFlags...), WhisperFlags...),
				Category:    "CONTRACT COMMANDS",
				Description: ``,
			},
		},
	}
)

func invokeContract(ctx *cli.Context) error {
	client := account.GetClient(ctx)
	if client == nil {
		log.Fatal("Can't get local account.")
		os.Exit(1)
	}

	acct := client.GetDefaultAccount()

	contractAddr := ctx.GlobalString(utils.ContractAddrFlag.Name)
	params := ctx.GlobalString(utils.ContractParamsFlag.Name)
	if "" == contractAddr {
		log.Fatal("contract address does not allow empty")
	}

	addr58, _ := common.AddressFromBase58(contractAddr)
	txHash, err := ontSdk.Rpc.InvokeNeoVMSmartContract(acct, new(big.Int), addr58, []interface{}{params})
	if err != nil {
		log.Fatalf("TestInvokeSmartContract InvokeNeoVMSmartContract error:%s", err)
	} else {
		log.Print("invoke transaction hash:", txHash)
	}

	//WaitForGenerateBlock
	_, err = ontSdk.Rpc.WaitForGenerateBlock(30*time.Second, 1)
	if err != nil {
		log.Fatalf("TestInvokeSmartContract WaitForGenerateBlock error:%s", err)
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

func deployContract(ctx *cli.Context) error {
	client := account.GetClient(ctx)
	if nil == client {
		log.Fatal("Can't get local account.")
	}

	acct := client.GetDefaultAccount()

	store := ctx.GlobalBool(utils.ContractStorageFlag.Name)
	vmType := getVmType(ctx.GlobalUint(utils.ContractVmTypeFlag.Name))

	codeDir := ctx.GlobalString(utils.ContractCodeFlag.Name)
	if "" == codeDir {
		log.Fatal("code dir is error, value does not allow null")
	}
	code, err := ioutil.ReadFile(codeDir)
	if err != nil {
		log.Fatal("error in read file", err.Error())
	}
	name := ctx.GlobalString(utils.ContractNameFlag.Name)
	version := ctx.GlobalString(utils.ContractVersionFlag.Name)
	author := ctx.GlobalString(utils.ContractAuthorFlag.Name)
	email := ctx.GlobalString(utils.ContractEmailFlag.Name)
	desc := ctx.GlobalString(utils.ContractDescFlag.Name)
	if "" == name || "" == version || "" == author || "" == email || "" == desc {
		log.Fatal("Params are not put completely: which contain code name version author email desc")
	}
	ontSdk.Rpc.DeploySmartContract(acct, vmType, store, fmt.Sprintf("%s", code), name, version, author, email, desc)

	//WaitForGenerateBlock
	_, err = ontSdk.Rpc.WaitForGenerateBlock(30*time.Second, 1)
	if err != nil {
		log.Fatalf("TestDeploySmartContract WaitForGenerateBlock error:%s", err)
	}

	return nil
}
