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
	cmdcom "github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/smartcontract/types"
	"github.com/urfave/cli"
	"io/ioutil"
	"strings"
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
					utils.ContractVmTypeFlag,
					utils.ContractParamsFlag,
					utils.WalletFileFlag,
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
					utils.ContractCodeFileFlag,
					utils.ContractNameFlag,
					utils.ContractVersionFlag,
					utils.ContractAuthorFlag,
					utils.ContractEmailFlag,
					utils.ContractDescFlag,
					utils.WalletFileFlag,
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

//
func invokeContract(ctx *cli.Context) error {
	//	if !ctx.IsSet(utils.ContractAddrFlag.Name) {
	//		return fmt.Errorf("Missing contract address argument.\n")
	//	}
	//
	//	wallet, err := cmdcom.OpenWallet(ctx)
	//	if err != nil {
	//		return fmt.Errorf("OpenWallet error:%s", err)
	//	}
	//
	//	acc := wallet.GetDefaultAccount()
	//	if acc == nil {
	//		return fmt.Errorf("Cannot GetDefaultAccount")
	//	}
	//
	//	vmType := ctx.String(utils.ContractVmTypeFlag.Name)
	//	contractAddr := ctx.String(utils.ContractAddrFlag.Name)
	//
	//	addr, err := common.AddressFromBase58(contractAddr)
	//	if err != nil {
	//		return fmt.Errorf("Invalid contract address")
	//	}
	//
	//	paramsStr := ctx.String(utils.ContractParamsFlag.Name)
	//	ps := strings.Split(paramsStr)
	//
	//
	//	txHash, err := ontSdk.Rpc.InvokeNeoVMSmartContract(acct, new(big.Int), addr, []interface{}{params})
	//	if err != nil {
	//		fmt.Printf("InvokeSmartContract InvokeNeoVMSmartContract error:%s", err)
	//		return err
	//	} else {
	//		fmt.Printf("invoke transaction hash:%s", common.ToHexString(txHash[:]))
	//	}
	//
	//	//WaitForGenerateBlock
	//	_, err = ontSdk.Rpc.WaitForGenerateBlock(30*time.Second, 1)
	//	if err != nil {
	//		fmt.Printf("InvokeSmartContract WaitForGenerateBlock error:%s", err)
	//	}
	return nil
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
	if !ctx.IsSet(utils.ContractCodeFileFlag.Name) ||
		!ctx.IsSet(utils.ContractNameFlag.Name) {
		return fmt.Errorf("Missing code or name argument")
	}

	wallet, err := cmdcom.OpenWallet(ctx)
	if err != nil {
		return fmt.Errorf("OpenWallet error:%s", err)
	}

	acc := wallet.GetDefaultAccount()
	if acc == nil {
		return fmt.Errorf("Cannot get default account")
	}

	store := ctx.Bool(utils.ContractStorageFlag.Name)
	vmType := getVmType(ctx.String(utils.ContractVmTypeFlag.Name))
	codeFile := ctx.String(utils.ContractCodeFileFlag.Name)
	if "" == codeFile {
		return fmt.Errorf("Please specific code file")
	}
	data, err := ioutil.ReadFile(codeFile)
	if err != nil {
		return fmt.Errorf("Read code:%s error:%s", codeFile, err)
	}
	code := strings.TrimSpace(string(data))
	name := ctx.String(utils.ContractNameFlag.Name)
	version := ctx.String(utils.ContractVersionFlag.Name)
	author := ctx.String(utils.ContractAuthorFlag.Name)
	email := ctx.String(utils.ContractEmailFlag.Name)
	desc := ctx.String(utils.ContractDescFlag.Name)

	txHash, err := utils.DeployContract(acc, vmType, store, code, name, version, author, email, desc)
	if err != nil {
		return fmt.Errorf("DeployContract error:%s", err)
	}
	address := utils.GetContractAddress(code, vmType)
	fmt.Printf("Deploy TxHash:%s\n", txHash)
	fmt.Printf("Contract Address:%x\n", address)
	return nil
}
