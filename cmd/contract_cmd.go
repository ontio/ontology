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

	sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
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

func rpcAddress() string {
	//return config.Parameters.SeedList[0] + strconv.Itoa(config.Parameters.HttpJsonPort)
	return "http://139.219.108.204:20336"
}

func invokeContract(ctx *cli.Context) error {
	config.Init(ctx)
	ontSdk := sdk.NewOntologySdk()
	ontSdk.Rpc.SetAddress(rpcAddress())

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
	var addr common.Address
	txHash, err := ontSdk.Rpc.InvokeNeoVMSmartContract(acct, new(big.Int), addr, []interface{}{params})
	if err != nil {
		fmt.Printf("TestInvokeSmartContract InvokeNeoVMSmartContract error:%s", err)
	} else {
		fmt.Println("invoke transaction hash:", txHash)
	}

	//WaitForGenerateBlock
	_, err = ontSdk.Rpc.WaitForGenerateBlock(30*time.Second, 1)
	if err != nil {
		fmt.Printf("TestInvokeSmartContract WaitForGenerateBlock error:%s", err)
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
	config.Init(ctx)
	ontSdk := sdk.NewOntologySdk()
	ontSdk.Rpc.SetAddress(rpcAddress())

	client := account.GetClient(ctx)
	if client == nil {

		log.Fatal("Can't get local account.")
		os.Exit(1)
	}

	acct := client.GetDefaultAccount()

	store := ctx.GlobalBool(utils.ContractStorageFlag.Name)
	vmType := getVmType(ctx.GlobalUint(utils.ContractVmTypeFlag.Name))

	codeDir := ctx.GlobalString(utils.ContractCodeFlag.Name)
	if "" == codeDir {
		fmt.Println("code dir is error")
		return nil

	}
	code, err := ioutil.ReadFile(codeDir)
	if err != nil {
		fmt.Println("error in read file", err.Error())
		return nil
	}
	name := ctx.GlobalString(utils.ContractNameFlag.Name)
	version := ctx.GlobalString(utils.ContractVersionFlag.Name)
	author := ctx.GlobalString(utils.ContractAuthorFlag.Name)
	email := ctx.GlobalString(utils.ContractEmailFlag.Name)
	desc := ctx.GlobalString(utils.ContractDescFlag.Name)
	if "" == name || "" == version || "" == author || "" == email || "" == desc {
		log.Fatal("Params are not put completely: which contain code name version author email desc")
		os.Exit(1)
	}
	ontSdk.Rpc.DeploySmartContract(acct, vmType, store, fmt.Sprintf("%s", code), name, version, author, email, desc)

	//WaitForGenerateBlock
	_, err = ontSdk.Rpc.WaitForGenerateBlock(30*time.Second, 1)
	if err != nil {
		fmt.Printf("TestDeploySmartContract WaitForGenerateBlock error:%s", err)
	}

	return nil
}
