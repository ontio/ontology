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
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common/password"
	"github.com/urfave/cli"
	"os"
)

var (
	AccountCommand = cli.Command{
		Action:      cli.ShowSubcommandHelp,
		Name:        "account",
		Usage:       "Manage accounts",
		ArgsUsage:   " ",
		Description: `Manage accounts stored in the wallet`,
		Subcommands: []cli.Command{
			{
				Action:    accountCreate,
				Name:      "add",
				Usage:     "Add a new account",
				ArgsUsage: "[sub-command options] <args>",
				Flags: []cli.Flag{
					utils.AccountQuantityFlag,
					utils.AccountTypeFlag,
					utils.AccountKeylenFlag,
					utils.AccountSigSchemeFlag,
					utils.AccountDefaultFlag,
					utils.AccountLabelFlag,
					utils.WalletFileFlag,
				},
				Description: `Add a new account`,
			},
			{
				Action:    accountList,
				Name:      "list",
				Usage:     "List existing accounts",
				ArgsUsage: "[sub-command options] <args>",
				Flags: []cli.Flag{
					utils.WalletFileFlag,
					utils.AccountVerboseFlag,
				},
				Description: `List existing accounts`,
			},
			{
				Action:    accountSet,
				Name:      "set",
				Usage:     "Modify an account",
				ArgsUsage: "<index>",
				Flags: []cli.Flag{
					utils.AccountSetDefaultFlag,
					utils.WalletFileFlag,
					utils.AccountLabelFlag,
					utils.AccountChangePasswdFlag,
					utils.AccountSigSchemeFlag,
				},
				Description: `Modify settings for an account. Account is specified by index. This can be showed by the 'list' command.`,
			},
			{
				Action:    accountDelete,
				Name:      "del",
				Usage:     "Delete an account",
				ArgsUsage: "<address|label|index>",
				Flags: []cli.Flag{
					utils.WalletFileFlag,
				},
				Description: `Delete an account specified by index. The index can be showed by the 'list' command`,
			},
			{
				Action:    accountImport,
				Name:      "import",
				Usage:     "Import accounts of wallet to another",
				ArgsUsage: "<address|label|index>",
				Flags: []cli.Flag{
					utils.WalletFileFlag,
					utils.AccountSourceFileFlag,
				},
				Description: "Import accounts of wallet to another. If not specific accounts in args, all account in source will be import",
			},
		},
	}
)

func accountCreate(ctx *cli.Context) error {
	reader := bufio.NewReader(os.Stdin)
	optionType := ""
	optionCurve := ""
	optionScheme := ""

	optionDefault := ctx.IsSet(utils.GetFlagName(utils.AccountDefaultFlag))
	if !optionDefault {
		optionType = checkType(ctx, reader)
		optionCurve = checkCurve(ctx, reader, &optionType)
		optionScheme = checkScheme(ctx, reader, &optionType)
	} else {
		fmt.Printf("Use default setting '-t ecdsa -b 256 -s SHA256withECDSA' \n")
		fmt.Printf("	signature algorithm: %s \n", keyTypeMap[optionType].name)
		fmt.Printf("	curve: %s \n", curveMap[optionCurve].name)
		fmt.Printf("	signature scheme: %s \n", schemeMap[optionScheme].name)
	}
	optionFile := checkFileName(ctx)
	optionNumber := checkNumber(ctx)
	optionLabel := checkLabel(ctx)
	pass, _ := password.GetConfirmedPassword()
	keyType := keyTypeMap[optionType].code
	curve := curveMap[optionCurve].code
	scheme := schemeMap[optionScheme].code
	wallet, err := account.Open(optionFile)
	if err != nil {
		return fmt.Errorf("Open wallet error:%s", err)
	}
	defer common.ClearPasswd(pass)
	for i := 0; i < optionNumber; i++ {
		label := optionLabel
		if label != "" && optionNumber > 1 {
			label = fmt.Sprintf("%s%d", label, i+1)
		}
		acc, err := wallet.NewAccount(label, keyType, curve, scheme, pass)
		if err != nil {
			return fmt.Errorf("new account error:%s", err)
		}
		fmt.Println()
		fmt.Println("Index:", wallet.GetAccountNum())
		fmt.Println("Label:", label)
		fmt.Println("Address:", acc.Address.ToBase58())
		fmt.Println("Public key:", hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)))
		fmt.Println("Signature scheme:", acc.SigScheme.Name())
	}

	fmt.Println("\nCreate account successfully.")
	return nil
}

func accountList(ctx *cli.Context) error {
	optionFile := checkFileName(ctx)
	wallet, err := account.Open(optionFile)
	if err != nil {
		return fmt.Errorf("Open wallet:%s error:%s", optionFile, err)
	}
	accNum := wallet.GetAccountNum()
	if accNum == 0 {
		fmt.Println("No account")
		return nil
	}
	for i := 1; i <= accNum; i++ {
		accMeta := wallet.GetAccountMetadataByIndex(i)
		if accMeta == nil {
			continue
		}
		if !ctx.Bool(utils.GetFlagName(utils.AccountVerboseFlag)) {
			if accMeta.IsDefault {
				fmt.Printf("Index:%-4d Address:%s  Label:%s (default)\n", i, accMeta.Address, accMeta.Label)
			} else {
				fmt.Printf("Index:%-4d Address:%s  Label:%s\n", i, accMeta.Address, accMeta.Label)
			}
			continue
		}
		if accMeta.IsDefault {
			fmt.Printf("%v\t%v (default)\n", i, accMeta.Address)
		} else {
			fmt.Printf("%v\t%v\n", i, accMeta.Address)
		}
		fmt.Printf("	Label: %v\n", accMeta.Label)
		fmt.Printf("	Signature algorithm: %v\n", accMeta.KeyType)
		fmt.Printf("	Curve: %v\n", accMeta.Curve)
		fmt.Printf("	Key length: %v bits\n", len(accMeta.Key)*8)
		fmt.Printf("	Public key: %v\n", accMeta.PubKey)
		fmt.Printf("	Signature scheme: %v\n", accMeta.SigSch)
		fmt.Println()
	}
	return nil
}

//set signature scheme for an account
func accountSet(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		fmt.Println("Missing account argument. Account address, lable or index expected.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	address := ctx.Args().First()
	wallet, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	accMeta := common.GetAccountMetadataMulti(wallet, address)
	address = accMeta.Address
	label := accMeta.Label
	passwd, err := common.GetPasswd(ctx)
	if err != nil {
		return err
	}
	defer common.ClearPasswd(passwd)
	if ctx.Bool(utils.GetFlagName(utils.AccountSetDefaultFlag)) {
		err = wallet.SetDefaultAccount(address, passwd)
		if err != nil {
			fmt.Printf("Set Label:%s Account:%s as default failed, %s\n", label, address, err)
		} else {
			fmt.Printf("Set Label:%s Account:%s as default successfully\n", label, address)
		}
	}
	if ctx.IsSet(utils.GetFlagName(utils.AccountLabelFlag)) {
		newLabel := ctx.String(utils.GetFlagName(utils.AccountLabelFlag))
		err = wallet.SetLabel(address, newLabel, passwd)
		if err != nil {
			fmt.Printf("Set Account:%s Label:%s to %s failed, %s\n", address, label, newLabel, err)
		} else {
			fmt.Printf("Set Account:%s Label:%s to %s successfully.\n", address, label, newLabel)
			label = newLabel
		}
	}
	if ctx.IsSet(utils.GetFlagName(utils.AccountSigSchemeFlag)) {
		find := false
		sigScheme := ctx.String(utils.GetFlagName(utils.AccountSigSchemeFlag))
		for _, val := range schemeMap {
			if val.name == sigScheme {
				find = true
				err = wallet.ChangeSigScheme(address, val.code, passwd)
				if err != nil {
					fmt.Printf("Set Label:%s Account:%s SigScheme to: %s failed, %s\n", accMeta.Label, accMeta.Address, val.name, err)
				} else {
					fmt.Printf("Set Label:%s Account:%s SigScheme to: %s successfully\n", accMeta.Label, accMeta.Address, val.name)
				}
				break
			}
		}
		if !find {
			fmt.Printf("%s is not a valid content for option -s \n", sigScheme)
		}
	}
	if ctx.Bool(utils.GetFlagName(utils.AccountChangePasswdFlag)) {
		fmt.Printf("Please input new password:\n")
		newPass, err := password.GetConfirmedPassword()
		if err != nil {
			return fmt.Errorf("Input password error:%s", err)
		}
		defer common.ClearPasswd(newPass)
		err = wallet.ChangePassword(address, passwd, newPass)
		if err != nil {
			fmt.Printf("Change password label:%s account:%s failed, %s\n", accMeta.Label, address, err)
		} else {
			fmt.Printf("Change password label:%s account:%s successfully\n", accMeta.Label, address)
		}
	}
	return nil
}

//delete an account by index from 'list'
func accountDelete(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		fmt.Println("Missing account argument. Account address, lable or index expected.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	address := ctx.Args().First()

	wallet, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	accMeta := common.GetAccountMetadataMulti(wallet, address)
	if accMeta == nil {
		return fmt.Errorf("Cannot get account by address:%s", address)
	}
	passwd, err := common.GetPasswd(ctx)
	if err != nil {
		return err
	}
	defer common.ClearPasswd(passwd)
	_, err = wallet.DeleteAccount(accMeta.Address, passwd)
	if err != nil {
		fmt.Printf("Delete account label:%s address:%s failed, %s\n", accMeta.Label, accMeta.Address, err)
	} else {
		fmt.Printf("Delete account label:%s address:%s successfully.\n", accMeta.Label, accMeta.Address)
	}
	return nil
}

func accountImport(ctx *cli.Context) error {
	source := ctx.String(utils.GetFlagName(utils.AccountSourceFileFlag))
	if source == "" {
		fmt.Printf("Missing source wallet path to import!\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	target := ctx.String(utils.GetFlagName(utils.WalletFileFlag))
	wallet, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}

	ctx.Set(utils.GetFlagName(utils.WalletFileFlag), source)
	sourceWallet, err := common.OpenWallet(ctx)
	if err != nil {
		return fmt.Errorf("Open source wallet to import error:%s", err)
	}
	accountNum := sourceWallet.GetAccountNum()
	if accountNum == 0 {
		fmt.Printf("No account to import\n")
		return nil
	}
	accList := make(map[string]string, ctx.NArg())
	for i := 0; i < ctx.NArg(); i++ {
		addr := ctx.Args().Get(i)
		accMeta := common.GetAccountMetadataMulti(sourceWallet, addr)
		if accMeta == nil {
			fmt.Printf("Cannot find account by:%s in wallet:%s\n", addr, source)
			continue
		}
		accList[accMeta.Address] = ""
	}

	succ := 0
	fail := 0
	skip := 0
	total := 0
	for i := 1; i <= accountNum; i++ {
		accMeta := sourceWallet.GetAccountMetadataByIndex(i)
		if accMeta == nil {
			continue
		}
		if len(accList) > 0 {
			_, ok := accList[accMeta.Address]
			if !ok {
				//Account not in import list, skip
				continue
			}
		}
		total++
		old := wallet.GetAccountMetadataByAddress(accMeta.Address)
		if old != nil {
			skip++
			fmt.Printf("Account:%s label:%s has already in wallet, skip.\n", accMeta.Address, accMeta.Label)
			continue
		}
		err = wallet.ImportAccount(accMeta)
		if err != nil {
			fail++
			fmt.Printf("Import account:%s label:%s failed, %s\n", accMeta.Address, accMeta.Label, err)
			continue
		}
		succ++
		fmt.Printf("Import account:%s label:%s successfully.\n", accMeta.Address, accMeta.Label)
	}
	fmt.Printf("\nImport wallet:%s to %s complete, total:%d success:%d failed:%d skip:%d\n", source, target, total, succ, fail, skip)
	return nil
}
