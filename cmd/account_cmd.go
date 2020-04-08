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
	"os"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common/password"
	"github.com/ontio/ontology/core/types"
	"github.com/urfave/cli"
)

var (
	AccountCommand = cli.Command{
		Action:    cli.ShowSubcommandHelp,
		Name:      "account",
		Usage:     "Manage accounts",
		ArgsUsage: "[arguments...]",
		Description: `Wallet management commands can be used to add, view, modify, delete, import account, and so on.
You can use ./Ontology account --help command to view help information of wallet management command.`,
		Subcommands: []cli.Command{
			{
				Action:    accountCreate,
				Name:      "add",
				Usage:     "Add a new account",
				ArgsUsage: "[sub-command options]",
				Flags: []cli.Flag{
					utils.AccountQuantityFlag,
					utils.AccountTypeFlag,
					utils.AccountKeylenFlag,
					utils.AccountSigSchemeFlag,
					utils.AccountDefaultFlag,
					utils.AccountLabelFlag,
					utils.IdentityFlag,
					utils.WalletFileFlag,
				},
				Description: ` Add a new account to wallet.
   Ontology support three type of key: ecdsa, sm2 and ed25519, and support 224、256、384、521 bits length of key in ecdsa, but only support 256 bits length of key in sm2 and ed25519.
   Ontology support multiple signature scheme.
   For ECDSA support SHA224withECDSA、SHA256withECDSA、SHA384withECDSA、SHA512withEdDSA、SHA3-224withECDSA、SHA3-256withECDSA、SHA3-384withECDSA、SHA3-512withECDSA、RIPEMD160withECDSA;
   For SM2 support SM3withSM2, and for SHA512withEdDSA.
   -------------------------------------------------
      Key   |key-length(bits)|  signature-scheme
   ---------|----------------|----------------------
   1 ecdsa  |  1 P-224: 224  | 1 SHA224withECDSA
            |----------------|----------------------
            |  2 P-256: 256  | 2 SHA256withECDSA
            |----------------|----------------------
            |  3 P-384: 384  | 3 SHA384withECDSA
            |----------------|----------------------
            |  4 P-521: 521  | 4 SHA512withEdDSA
            |----------------|----------------------
            |                | 5 SHA3-224withECDSA
            |                |----------------------
            |                | 6 SHA3-256withECDSA
            |                |----------------------
            |                | 7 SHA3-384withECDSA
            |                |----------------------
            |                | 8 SHA3-512withECDSA
            |                |----------------------
            |                | 9 RIPEMD160withECDSA
   ---------|----------------|----------------------
   2 sm2    | sm2p256v1 256  | SM3withSM2
   ---------|----------------|----------------------
   3 ed25519|   25519 256    | SHA512withEdDSA
   -------------------------------------------------`,
			},
			{
				Action:    accountList,
				Name:      "list",
				Usage:     "List existing accounts",
				ArgsUsage: "[sub-command options] <label|address|index>",
				Flags: []cli.Flag{
					utils.WalletFileFlag,
					utils.AccountVerboseFlag,
				},
				Description: `List existing accounts. If specified in args, will list those account. If not specified in args, will list all accouns in wallet`,
			},
			{
				Action:    accountSet,
				Name:      "set",
				Usage:     "Modify an account",
				ArgsUsage: "[sub-command options] <label|address|index>",
				Flags: []cli.Flag{
					utils.AccountSetDefaultFlag,
					utils.WalletFileFlag,
					utils.AccountLabelFlag,
					utils.AccountChangePasswdFlag,
					utils.AccountSigSchemeFlag,
				},
				Description: `Modify settings for an account. Account is specified by address, label of index. Index start from 1. This can be showed by the 'list' command.`,
			},
			{
				Action:    accountDelete,
				Name:      "del",
				Usage:     "Delete an account",
				ArgsUsage: "[sub-command options] <address|label|index>",
				Flags: []cli.Flag{
					utils.WalletFileFlag,
				},
				Description: `Delete an account specified by address, label of index. Index start from 1. This can be showed by the 'list' command`,
			},
			{
				Action:    accountImport,
				Name:      "import",
				Usage:     "Import accounts of wallet to another",
				ArgsUsage: "[sub-command options] <address|label|index>",
				Flags: []cli.Flag{
					utils.WalletFileFlag,
					utils.AccountSourceFileFlag,
					utils.AccountWIFFlag,
				},
				Description: "Import accounts of wallet to another. If not specific accounts in args, all account in source will be import",
			},
			{
				Action:    accountExport,
				Name:      "export",
				Usage:     "Export accounts to a specified wallet file",
				ArgsUsage: "[sub-command options] <filename>",
				Flags: []cli.Flag{
					utils.WalletFileFlag,
					utils.AccountLowSecurityFlag,
				},
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
		PrintInfoMsg("Use default setting '-t ecdsa -b 256 -s SHA256withECDSA'")
		PrintInfoMsg("	signature algorithm: %s", keyTypeMap[optionType].name)
		PrintInfoMsg("	curve: %s", curveMap[optionCurve].name)
		PrintInfoMsg("	signature scheme: %s", schemeMap[optionScheme].name)
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
		return fmt.Errorf("error opening wallet: %s", err)
	}
	defer common.ClearPasswd(pass)
	if ctx.Bool(utils.IdentityFlag.Name) {
		// create ONT ID
		wd := wallet.GetWalletData()
		id, err := account.NewIdentity(optionLabel, keyType, curve, pass)
		if err != nil {
			return fmt.Errorf("error creating ONT ID: %s", err)
		}
		wd.AddIdentity(id)
		err = wd.Save(optionFile)
		if err != nil {
			return fmt.Errorf("error saving to %s: %s", optionFile, err)
		}
		PrintInfoMsg("ONT ID created: %s", id.ID)
		PrintInfoMsg("Bind public key: %s", id.Control[0].Public)
		return nil
	}
	for i := 0; i < optionNumber; i++ {
		label := optionLabel
		if label != "" && optionNumber > 1 {
			label = fmt.Sprintf("%s%d", label, i+1)
		}
		acc, err := wallet.NewAccount(label, keyType, curve, scheme, pass)
		if err != nil {
			return fmt.Errorf("error creating new account: %s", err)
		}
		PrintInfoMsg("Index:%d", wallet.GetAccountNum())
		PrintInfoMsg("Label:%s", label)
		PrintInfoMsg("Address:%s", acc.Address.ToBase58())
		PrintInfoMsg("Public key:%s", hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)))
		PrintInfoMsg("Signature scheme:%s", acc.SigScheme.Name())
	}

	PrintInfoMsg("Create account successfully.")
	return nil
}

func accountList(ctx *cli.Context) error {
	optionFile := checkFileName(ctx)
	wallet, err := account.Open(optionFile)
	if err != nil {
		return fmt.Errorf("error opening wallet %s: %s", optionFile, err)
	}
	accNum := wallet.GetAccountNum()
	if accNum == 0 {
		PrintInfoMsg("No account.")
		return nil
	}
	accList := make(map[string]string, ctx.NArg())
	for i := 0; i < ctx.NArg(); i++ {
		addr := ctx.Args().Get(i)
		accMeta := common.GetAccountMetadataMulti(wallet, addr)
		if accMeta == nil {
			PrintWarnMsg("Cannot find account by %s in wallet: %s", addr, utils.GetFlagName(utils.WalletFileFlag))
			continue
		}
		accList[accMeta.Address] = ""
	}
	for i := 1; i <= accNum; i++ {
		accMeta := wallet.GetAccountMetadataByIndex(i)
		if accMeta == nil {
			continue
		}
		if len(accList) > 0 {
			_, ok := accList[accMeta.Address]
			if !ok {
				continue
			}
		}
		if !ctx.Bool(utils.GetFlagName(utils.AccountVerboseFlag)) {
			if accMeta.IsDefault {
				PrintInfoMsg("Index:%-4d Address:%s  Label:%s (default)", i, accMeta.Address, accMeta.Label)
			} else {
				PrintInfoMsg("Index:%-4d Address:%s  Label:%s", i, accMeta.Address, accMeta.Label)
			}
			continue
		}
		if accMeta.IsDefault {
			PrintInfoMsg("%v\t%v (default)", i, accMeta.Address)
		} else {
			PrintInfoMsg("%v\t%v", i, accMeta.Address)
		}
		PrintInfoMsg("	Label: %v", accMeta.Label)
		PrintInfoMsg("	Signature algorithm: %v", accMeta.KeyType)
		PrintInfoMsg("	Curve: %v", accMeta.Curve)
		PrintInfoMsg("	Key length: %v bits", len(accMeta.Key)*8)
		PrintInfoMsg("	Public key: %v", accMeta.PubKey)
		PrintInfoMsg("	Signature scheme: %v\n", accMeta.SigSch)
	}
	return nil
}

//set signature scheme for an account
func accountSet(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing account argument.")
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
		return fmt.Errorf("cannot find account info by: %s", address)
	}
	address = accMeta.Address
	label := accMeta.Label
	if ctx.Bool(utils.GetFlagName(utils.AccountSetDefaultFlag)) {
		err = wallet.SetDefaultAccount(address)
		if err != nil {
			PrintErrorMsg("Set Label:%s Account:%s as default failed, %s", label, address, err)
		} else {
			PrintInfoMsg("Set Label:%s Account:%s as default successfully", label, address)
		}
	}
	if ctx.IsSet(utils.GetFlagName(utils.AccountLabelFlag)) {
		newLabel := ctx.String(utils.GetFlagName(utils.AccountLabelFlag))
		err = wallet.SetLabel(address, newLabel)
		if err != nil {
			PrintErrorMsg("Set Account:%s Label:%s to %s failed, %s", address, label, newLabel, err)
		} else {
			PrintInfoMsg("Set Account:%s Label:%s to %s successfully.", address, label, newLabel)
			label = newLabel
		}
	}
	if ctx.IsSet(utils.GetFlagName(utils.AccountSigSchemeFlag)) {
		find := false
		sigScheme := ctx.String(utils.GetFlagName(utils.AccountSigSchemeFlag))
		var sigSch signature.SignatureScheme
		for key, val := range schemeMap {
			if key == sigScheme {
				find = true
				sigSch = val.code
				break
			}
			if val.name == sigScheme {
				find = true
				sigSch = val.code
				break
			}
		}
		if find {
			err = wallet.ChangeSigScheme(address, sigSch)
			if err != nil {
				PrintErrorMsg("Set Label:%s Account:%s SigScheme to: %s failed, %s", accMeta.Label, accMeta.Address, sigSch.Name(), err)
			} else {
				PrintInfoMsg("Set Label:%s Account:%s SigScheme to: %s successfully.", accMeta.Label, accMeta.Address, sigSch.Name())
			}
		} else {
			PrintInfoMsg("%s is not a valid content for option -s", sigScheme)
		}
	}

	if ctx.Bool(utils.GetFlagName(utils.AccountChangePasswdFlag)) {
		passwd, err := common.GetPasswd(ctx)
		if err != nil {
			return err
		}
		defer common.ClearPasswd(passwd)
		// verify old password
		if _, err := wallet.GetAccountByAddress(address, passwd); err != nil {
			return fmt.Errorf("invalid old password: %s", err)
		}
		PrintInfoMsg("Please input new password:")
		newPass, err := password.GetConfirmedPassword()
		defer common.ClearPasswd(newPass)
		if err != nil {
			return fmt.Errorf("input password error: %s", err)
		}
		err = wallet.ChangePassword(address, passwd, newPass)
		if err != nil {
			PrintErrorMsg("Change password label:%s account:%s failed, %s", accMeta.Label, address, err)
		} else {
			PrintInfoMsg("Change password label:%s account:%s successfully", accMeta.Label, address)
		}
	}
	return nil
}

//delete an account by index from 'list'
func accountDelete(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing account argument.")
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
		return fmt.Errorf("cannot get account by: %s", address)
	}
	passwd, err := common.GetPasswd(ctx)
	if err != nil {
		return err
	}
	defer common.ClearPasswd(passwd)
	_, err = wallet.DeleteAccount(accMeta.Address, passwd)
	if err != nil {
		PrintErrorMsg("Delete account label:%s address:%s failed, %s", accMeta.Label, accMeta.Address, err)
	} else {
		PrintInfoMsg("Delete account label:%s address:%s successfully.", accMeta.Label, accMeta.Address)
	}
	return nil
}

func accountImport(ctx *cli.Context) error {
	source := ctx.String(utils.GetFlagName(utils.AccountSourceFileFlag))
	if source == "" {
		PrintErrorMsg("Missing source wallet path argument to import.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	target := ctx.String(utils.GetFlagName(utils.WalletFileFlag))
	fn := checkFileName(ctx)
	wallet, err := account.Open(fn)
	if err != nil {
		return err
	}

	succ := 0
	fail := 0
	skip := 0
	total := 0

	if ctx.Bool(utils.GetFlagName(utils.AccountWIFFlag)) {
		// import WIF keys
		file, err := os.Open(source)
		if err != nil {
			return err
		}
		f := bufio.NewScanner(file)
		keys := make([]keypair.PrivateKey, 0)
		for f.Scan() {
			wif := f.Bytes()
			pri, err := keypair.GetP256KeyPairFromWIF(wif)
			common.ClearPasswd(wif)
			if err != nil {
				return err
			}
			keys = append(keys, pri)
		}
		total = len(keys)
		PrintInfoMsg("Please input a password to encrypt the imported key(s)")
		pwd, err := password.GetConfirmedPassword()
		if err != nil {
			return err
		}
		for _, v := range keys {
			pub := v.Public()
			addr := types.AddressFromPubKey(pub)
			b58addr := addr.ToBase58()
			old := wallet.GetAccountMetadataByAddress(b58addr)
			if old != nil {
				PrintWarnMsg("Account %s already exists.", b58addr)
				skip += 1
				continue
			}
			PrintInfoMsg("Import account %s", b58addr)
			k, err := keypair.EncryptPrivateKey(v, b58addr, pwd)
			if err != nil {
				PrintWarnMsg("Import account %s error: %s", b58addr, err)
				fail += 1
				continue
			}
			var accMeta account.AccountMetadata
			accMeta.Address = k.Address
			accMeta.KeyType = k.Alg
			accMeta.EncAlg = k.EncAlg
			accMeta.Hash = k.Hash
			accMeta.Key = k.Key
			accMeta.Curve = k.Param["curve"]
			accMeta.Salt = k.Salt
			accMeta.Label = ""
			accMeta.PubKey = hex.EncodeToString(keypair.SerializePublicKey(v.Public()))
			accMeta.SigSch = signature.SHA256withECDSA.Name()
			err = wallet.ImportAccount(&accMeta)
			if err != nil {
				PrintWarnMsg("Import account %s error: %s", accMeta.Address, err)
				fail += 1
				continue
			}
			succ += 1
		}
		common.ClearPasswd(pwd)
	} else {
		ctx.Set(utils.GetFlagName(utils.WalletFileFlag), source)
		sourceWallet, err := common.OpenWallet(ctx)
		if err != nil {
			return fmt.Errorf("open source wallet to import error: %s", err)
		}
		accountNum := sourceWallet.GetAccountNum()
		if accountNum == 0 {
			PrintWarnMsg("No account to import.")
			return nil
		}
		accList := make(map[string]string, ctx.NArg())
		for i := 0; i < ctx.NArg(); i++ {
			addr := ctx.Args().Get(i)
			accMeta := common.GetAccountMetadataMulti(sourceWallet, addr)
			if accMeta == nil {
				PrintWarnMsg("Cannot find account %s in wallet %s.", addr, source)
				continue
			}
			accList[accMeta.Address] = ""
		}

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
				PrintWarnMsg("Account: %s (label: %s) already exists in wallet, skip.", accMeta.Address, accMeta.Label)
				continue
			}
			err = wallet.ImportAccount(accMeta)
			if err != nil {
				fail++
				PrintWarnMsg("Import account: %s (label: %s) failed, %s", accMeta.Address, accMeta.Label, err)
				continue
			}
			succ++
			PrintInfoMsg("Import account: %s (label: %s) successfully.", accMeta.Address, accMeta.Label)
		}
	}

	PrintInfoMsg("Import from %s to %s complete.", source, target)
	PrintInfoMsg("Total:\t%d", total)
	PrintInfoMsg("Success:%d", succ)
	PrintInfoMsg("Failed:\t%d", fail)
	PrintInfoMsg("Skip:\t%d", skip)
	return nil
}

func accountExport(ctx *cli.Context) error {
	if ctx.NArg() <= 0 {
		PrintErrorMsg("Missing target file argument to export.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	target := ctx.Args().First()
	client, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	wallet := client.GetWalletData()
	if ctx.IsSet(utils.GetFlagName(utils.AccountLowSecurityFlag)) {
		n := client.GetAccountNum()
		passwords := make([][]byte, n)
		for i := 0; i < n; i++ {
			ac := client.GetAccountMetadataByIndex(i + 1)
			PrintInfoMsg("Account %d %s: %s", i+1, ac.Label, ac.Address)
			for j := 0; j < 3; j++ {
				pwd, err := password.GetPassword()
				if err != nil {
					fmt.Println(err)
				} else {
					passwords[i] = pwd
					break
				}
			}
		}
		wallet = wallet.Clone()
		err := wallet.ToLowSecurity(passwords)
		for _, v := range passwords {
			common.ClearPasswd(v)
		}
		if err != nil {
			return fmt.Errorf("export failed: %s", err)
		}
	}
	err = wallet.Save(target)
	if err != nil {
		return fmt.Errorf("save wallet file error: %s", err)
	}

	PrintInfoMsg("Export wallet success.")
	return nil
}
