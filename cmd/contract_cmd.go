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
		Usage:        "ontology contract [invoke|deploy] [OPTION]\n",
		Category:     "CONTRACT COMMANDS",
		OnUsageError: contractUsageError,
		Description:  `account command`,
		Subcommands: []cli.Command{
			{
				Action:       utils.MigrateFlags(invokeContract),
				Name:         "invoke",
				OnUsageError: contractUsageError,
				Usage:        "ontology invoke [OPTION]\n",
				Flags:        append(NodeFlags, ContractFlags...),
				Category:     "CONTRACT COMMANDS",
				Description:  ``,
			},
			{
				Action:       utils.MigrateFlags(deployContract),
				OnUsageError: contractUsageError,
				Name:         "deploy",
				Usage:        "ontology deploy [OPTION]\n",
				Flags:        append(NodeFlags, ContractFlags...),
				Category:     "CONTRACT COMMANDS",
				Description:  ``,
			},
		},
	}
)

func showContractHelp() {
	var contractUsingHelp = `
   Name:
      ontology contract      deploy or invoke a smart contract by this command
   Usage:
      ontology contract [command options] [args]

   Description:
      With this command, you can transfer asset from one account to another

   Command:
     deploy
       --caddr      value               smart contract address that will be invoke
       --params     value               params will be  
			
     invoke
       --type       value               contract type ,value: NEOVM | NATIVE | SWAM
       --store      value               does this contract will be stored, value: true or false
       --code       value               directory of smart contract that will be deployed
       --cname      value               contract name that will be deployed
       --cversion   value               contract version which will be deployed
       --author     value               owner of deployed smart contract
       --email      value               owner email who deploy the smart contract
       --desc       value               contract description when deploy one
`
	fmt.Println(contractUsingHelp)
}

func contractUsageError(context *cli.Context, err error, isSubcommand bool) error {
	fmt.Println(err.Error())
	showContractHelp()
	return nil
}

func invokeContract(ctx *cli.Context) error {
	if !ctx.IsSet(utils.ContractAddrFlag.Name) || !ctx.IsSet(utils.ContractParamsFlag.Name) {
		showContractHelp()
		return nil
	}

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

	addr, _ := common.AddressFromBase58(contractAddr)
	txHash, err := ontSdk.Rpc.InvokeNeoVMSmartContract(acct, new(big.Int), addr, []interface{}{params})
	if err != nil {
		log.Fatalf("InvokeSmartContract InvokeNeoVMSmartContract error:%s", err)
	} else {
		log.Print("invoke transaction hash:", txHash)
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

func deployContract(ctx *cli.Context) error {
	if !ctx.IsSet(utils.ContractStorageFlag.Name) || !ctx.IsSet(utils.ContractVmTypeFlag.Name) ||
		!ctx.IsSet(utils.ContractCodeFlag.Name) || !ctx.IsSet(utils.ContractNameFlag.Name) ||
		!ctx.IsSet(utils.ContractVersionFlag.Name) || !ctx.IsSet(utils.ContractAuthorFlag.Name) ||
		!ctx.IsSet(utils.ContractEmailFlag.Name) || !ctx.IsSet(utils.ContractDescFlag.Name) {
		showContractHelp()
		return nil
	}

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
		log.Fatal("Params are not put completely: which contain code, name, version, author, email, desc")
	}

	trHash, err := ontSdk.Rpc.DeploySmartContract(acct, vmType, store, fmt.Sprintf("%s", code), name, version, author, email, desc)

	//WaitForGenerateBlock
	_, err = ontSdk.Rpc.WaitForGenerateBlock(30*time.Second, 1)
	if err != nil {
		log.Fatalf("DeploySmartContract WaitForGenerateBlock error:%s", err)
	} else {
		fmt.Printf("Deploy smartContract transaction hash: %+v\n", trHash)
	}

	return nil
}
