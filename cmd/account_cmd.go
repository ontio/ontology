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
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/common/password"
	"github.com/urfave/cli"
	"os"
	"strconv"
	"strings"
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
				Action:      accountCreate,
				Name:        "add",
				Usage:       "Add a new account",
				ArgsUsage:   "[sub-command options] <args>",
				Flags:       addFlags,
				Description: `Add a new account`,
			},
			{
				Action:      accountShow,
				Name:        "list",
				Usage:       "List existing accounts",
				ArgsUsage:   "[sub-command options] <args>",
				Flags:       listFlags,
				Description: `List existing accounts`,
			},
			{
				Action:      accountSet,
				Name:        "set",
				Usage:       "Modify an account",
				ArgsUsage:   "<index>",
				Flags:       setFlags,
				Description: `Modify settings for an account. Account is specified by index. This can be showed by the 'list' command.`,
			},
			{
				Action:      accountDelete,
				Name:        "del",
				Usage:       "Delete an account",
				ArgsUsage:   "<index>",
				Flags:       fileFlags,
				Description: `Delete an account specified by index. The index can be showed by the 'list' command`,
			},

			{
				Action:      encrypt,
				Name:        "encrypt",
				ArgsUsage:   "<index>",
				Usage:       "Encrypt the specified account",
				Flags:       fileFlags,
				Description: `Encrypt the specified account using new password. Account is specified by index. This can be showed by the 'list' command.`,
			},
		},
	}
)

func accountCreate(ctx *cli.Context) error {

	reader := bufio.NewReader(os.Stdin)
	var optionNumber = 1
	optionType := ""
	optionCurve := ""
	optionScheme := ""

	optionFile := checkFileName(ctx)
	optionDefault := ctx.IsSet("default")
	if !optionDefault {
		optionNumber = int(checkNumber(ctx))
		optionType = checkType(ctx, reader)
		optionCurve = checkCurve(ctx, reader, &optionType)
		optionScheme = checkScheme(ctx, reader, &optionType)
	}

	pass, _ := password.GetConfirmedPassword()

	wallet := new(account.WalletData)
	err := wallet.Load(optionFile)
	if err != nil {
		fmt.Printf("%v doesn't exist, create file automatically.\n\n", optionFile)
		wallet.Inititalize()
	}

	for i := 0; i < optionNumber; i++ {
		acc := account.CreateAccount(&optionType, &optionCurve, &optionScheme, &pass)
		wallet.AddAccount(acc)
		fmt.Println("Address: ", acc.Address)
		fmt.Println("Public key:", acc.PubKey)
		fmt.Println("Signature scheme:", acc.SigSch)
		fmt.Println("")
	}

	for i := 0; i < len(pass); i++ {
		(pass)[i] = 0
	}

	if wallet.Save(optionFile) != nil {
		fmt.Println("Wallet file save failed.")
	}

	fmt.Println("\nCreate account successfully.")

	return nil
}

func accountShow(ctx *cli.Context) error {
	wallet := new(account.WalletData)
	optionFile := checkFileName(ctx)
	err := wallet.Load(optionFile)
	if err != nil {
		fmt.Printf("%v doesn't exist, please enter the right wallet file name or create the file first.\n", optionFile)
		return nil
	}
	if len(wallet.Accounts) == 0 {
		fmt.Println("No account")
		return nil
	}

	if !ctx.Bool("verbose") {
		// look for every account and show details
		for i, acc := range wallet.Accounts {
			if acc.IsDefault {
				fmt.Printf("* %v\t%v\n", i+1, acc.Address)
			} else {
				fmt.Printf("  %v\t%v\n", i+1, acc.Address)
			}
		}
		fmt.Println("\nUse -v or --verbose option to display details.")
	} else {
		// look for every account and show address only
		for i, acc := range wallet.Accounts {
			if acc.IsDefault {
				fmt.Printf("* %v\t%v\n", i+1, acc.Address)
			} else {
				fmt.Printf("  %v\t%v\n", i+1, acc.Address)
			}
			fmt.Printf("	Signature algorithm: %v\n", acc.Alg)
			fmt.Printf("	Curve: %v\n", acc.Param["curve"])
			fmt.Printf("	Key length: %v bits\n", len(acc.Key)*8)
			fmt.Printf("	Public key: %v\n", acc.PubKey)
			fmt.Printf("	Signature scheme: %v\n", acc.SigSch)
			fmt.Println()
		}
	}
	return nil
}

//set signature scheme for an account
func accountSet(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		fmt.Printf("Please enter an index of account, for index list please use 'account list' command.\n")
		return nil
	}
	index, err := strconv.Atoi(ctx.Args().First())
	if err != nil {
		fmt.Printf("Your input is not an number.\n")
		return nil
	}

	optionFile := checkFileName(ctx)
	wallet := new(account.WalletData)
	if wallet.Load(optionFile) != nil {
		fmt.Printf("%v doesn't exist, please enter the right wallet file name or create the file first.\n", optionFile)
		return nil
	}
	if index < 1 || index > len(wallet.Accounts) {
		fmt.Printf("Your input is out of index range.\n")
		return nil
	}

	find := false

	if ctx.IsSet("as-default") {
		fmt.Printf("Set account %d to the default account\n", index)
		for _, v := range wallet.Accounts {
			if v.IsDefault {
				v.IsDefault = false
			}
		}
		wallet.Accounts[index-1].IsDefault = true
	} else {
		if ctx.IsSet("signature-scheme") {
			for key, val := range account.SchemeMap {
				if val.Name == ctx.String("signature-scheme") {
					inputSchemeInfo := account.SchemeMap[key]
					find = true
					fmt.Printf("%s is selected. \n", inputSchemeInfo.Name)
					wallet.Accounts[index-1].SigSch = inputSchemeInfo.Name
					break
				}
			}
			fmt.Printf("%s is not a valid content for option -s \n", ctx.String("signature-scheme"))
		}
		if !find {
			fmt.Printf("Invalid arguments! Nothing changed.\n")
		}
	}

	if wallet.Save(optionFile) != nil {
		fmt.Println("Wallet file save failed.")
	}
	return nil
}

//delete an account by index from 'list'
func accountDelete(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		fmt.Printf("Please enter an index of account, for index list please use 'account list' command.\n")
		return nil
	}
	index, err := strconv.Atoi(ctx.Args().First())
	if err != nil {
		fmt.Printf("Your input is not an number.\n")
		return nil
	}

	optionFile := checkFileName(ctx)
	wallet := new(account.WalletData)
	if wallet.Load(optionFile) != nil {
		fmt.Printf("%v doesn't exist, please enter the right wallet file name or create the file first.\n", optionFile)
		return nil
	}

	if index < 1 || index > len(wallet.Accounts) {
		fmt.Printf("Your input is out of index range.\n")
		return nil
	}

	oldPass, _ := password.GetPassword()
	h := sha256.Sum256(oldPass)

	passHash, _ := hex.DecodeString(wallet.Accounts[index-1].PassHash)
	if !bytes.Equal(passHash, h[:]) {
		fmt.Println("Wrong password! Delete account failed.")
		os.Exit(1)
	}

	addr := wallet.DelAccount(index)

	if wallet.Save(optionFile) != nil {
		fmt.Println("Wallet file save failed.")
	}

	fmt.Printf("Delete account successfully.\n")
	fmt.Printf("index = %v, address = %v.\n", index, addr)
	return nil
}

//change password
func encrypt(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		fmt.Println("Missing argument.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	index, err := strconv.Atoi(ctx.Args().First())
	if err != nil {
		fmt.Println("Invalid aragument. Account index expected.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	optionFile := checkFileName(ctx)
	wallet := new(account.WalletData)
	if wallet.Load(optionFile) != nil {
		wallet.Inititalize()
	}
	if index < 1 || index > len(wallet.Accounts) {
		fmt.Printf("Your input is out of index range.\n")
		return nil
	}

	oldPass, _ := password.GetPassword()
	h := sha256.Sum256(oldPass)

	passHash, _ := hex.DecodeString(wallet.Accounts[index-1].PassHash)
	if !bytes.Equal(passHash, h[:]) {
		fmt.Println("Wrong password! Encrypt account failed.")
		os.Exit(1)
	}

	prv, _ := keypair.DecryptPrivateKey(wallet.Accounts[index-1].GetKeyPair(), oldPass)

	//let user enter password with double check
	pass, _ := password.GetConfirmedPassword()

	prvSecret, _ := keypair.EncryptPrivateKey(prv, wallet.Accounts[index-1].Address, pass)
	_h := sha256.Sum256(pass)
	for i := 0; i < len(pass); i++ {
		pass[i] = 0
	}

	wallet.Accounts[index-1].SetKeyPair(prvSecret)
	wallet.Accounts[index-1].PassHash = hex.EncodeToString(_h[:])

	if wallet.Save(optionFile) != nil {
		fmt.Println("Wallet file save failed.")
	}

	fmt.Println("encrypt account successfully.")
	fmt.Println("")

	return nil
}

// wait for user to choose options
func chooseKeyType(reader *bufio.Reader) string {
	common.PrintNotice("key type")
	for {
		tmp, _ := reader.ReadString('\n')
		tmp = strings.TrimSpace(tmp)
		_, ok := account.KeyTypeMap[tmp]
		if ok {
			fmt.Printf("%s is selected. \n", account.KeyTypeMap[tmp].Name)
			return account.KeyTypeMap[tmp].Name
		} else {
			fmt.Print("Input error! Please enter a number above: ")
		}
	}
	return ""
}
func chooseScheme(reader *bufio.Reader) string {
	common.PrintNotice("signature-scheme")
	for true {
		tmp, _ := reader.ReadString('\n')
		tmp = strings.TrimSpace(tmp)

		_, ok := account.SchemeMap[tmp]
		if ok {
			fmt.Printf("scheme %s is selected.\n", account.SchemeMap[tmp].Name)
			return account.SchemeMap[tmp].Name
		} else {
			fmt.Print("Input error! Please enter a number above:")
		}
	}
	return ""
}
func chooseCurve(reader *bufio.Reader) string {
	common.PrintNotice("curve")
	for true {
		tmp, _ := reader.ReadString('\n')
		tmp = strings.TrimSpace(tmp)
		_, ok := account.CurveMap[tmp]
		if ok {
			fmt.Printf("scheme %s is selected.\n", account.CurveMap[tmp].Name)
			return account.CurveMap[tmp].Name
		} else {
			fmt.Print("Input error! Please enter a number above:")
		}
	}
	return ""
}

func checkFileName(ctx *cli.Context) string {
	if ctx.IsSet("file") {
		return ctx.String("file")
	} else {
		//default account file name
		return account.WALLET_FILENAME
	}
}
func checkNumber(ctx *cli.Context) uint {
	if ctx.IsSet("number") {
		if ctx.Uint("number") < 1 {
			fmt.Println("the minimum number is 1, set to default value(1).")
			return 1
		}
		if ctx.Uint("number") > 100 {
			fmt.Println("the maximum number is 100, set to default value(1).")
			return 1
		}
		return ctx.Uint("number")
	} else {
		return 1
	}
}
func checkType(ctx *cli.Context, reader *bufio.Reader) string {
	t := ""
	if ctx.IsSet("type") {
		if _, ok := account.KeyTypeMap[ctx.String("type")]; ok {
			t = account.KeyTypeMap[ctx.String("type")].Name
			fmt.Printf("%s is selected. \n", t)
		} else {
			fmt.Printf("%s is not a valid content for option -t \n", ctx.String("type"))
			t = chooseKeyType(reader)
		}
	} else {
		t = chooseKeyType(reader)
	}
	return t
}
func checkCurve(ctx *cli.Context, reader *bufio.Reader, t *string) string {
	c := ""
	switch *t {
	case "ecdsa":
		if ctx.IsSet("bit-length") {
			if _, ok := account.CurveMap[ctx.String("bit-length")]; ok {
				c = account.CurveMap[ctx.String("bit-length")].Name
				fmt.Printf("%s is selected. \n", c)
			} else {
				fmt.Printf("%s is not a valid content for option -b \n", ctx.String("bit-length"))
				c = chooseCurve(reader)
			}
		} else {
			c = chooseCurve(reader)
		}
		break
	case "sm2":
		fmt.Println("Use curve sm2p256v1 with key length of 256 bits.")
		c = "SM2P256V1"
		break
	case "ed25519":
		fmt.Println("Use curve 25519 with key length of 256 bits.")
		c = "ED25519"
		break
	default:
		return ""
	}
	return c
}
func checkScheme(ctx *cli.Context, reader *bufio.Reader, t *string) string {
	s := ""
	switch *t {
	case "ecdsa":
		if ctx.IsSet("bit-length") {
			if _, ok := account.SchemeMap[ctx.String("signature-scheme")]; ok {
				s = account.SchemeMap[ctx.String("signature-scheme")].Name
				fmt.Printf("%s is selected. \n", s)
			} else {
				fmt.Printf("%s is not a valid content for option -s \n", ctx.String("signature-scheme"))
				s = chooseScheme(reader)
			}
		} else {
			s = chooseScheme(reader)
		}
		break
	case "sm2":
		fmt.Println("Use SM3withSM2 as the signature scheme.")
		s = "SM3withSM2"
		break
	case "ed25519":
		fmt.Println("Use Ed25519 as the signature scheme.")
		s = "SHA512withEDDSA"
		break
	default:
		return ""
	}
	return s
}
