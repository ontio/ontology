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
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/core/types"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

//map info, to get some information easily
//todo: move to crypto package
type keyTypeInfo struct {
	name string
	code keypair.KeyType
}

var keyTypeMap = map[string]keyTypeInfo{
	"":  {"ecdsa", keypair.PK_ECDSA},
	"1": {"ecdsa", keypair.PK_ECDSA},
	"2": {"sm2", keypair.PK_SM2},
	"3": {"ed25519", keypair.PK_EDDSA},
}

type curveInfo struct {
	name string
	code byte
}

var curveMap = map[string]curveInfo{
	"":  {"P-256", keypair.P256},
	"1": {"P-224", keypair.P224},
	"2": {"P-256", keypair.P256},
	"3": {"P-384", keypair.P384},
	"4": {"P-521", keypair.P521},
}

type schemeInfo struct {
	name string
	code signature.SignatureScheme
}

var schemeMap = map[string]schemeInfo{
	"":  {"SHA256withECDSA", signature.SHA256withECDSA},
	"1": {"SHA224withECDSA", signature.SHA224withECDSA},
	"2": {"SHA256withECDSA", signature.SHA256withECDSA},
	"3": {"SHA384withECDSA", signature.SHA384withECDSA},
	"4": {"SHA512withEDDSA", signature.SHA512withEDDSA},
	"5": {"SHA3_224withECDSA", signature.SHA3_224withECDSA},
	"6": {"SHA3_256withECDSA", signature.SHA3_256withECDSA},
	"7": {"SHA3_384withECDSA", signature.SHA3_384withECDSA},
	"8": {"SHA3_512withECDSA", signature.SHA3_512withECDSA},
	"9": {"RIPEMD160withECDSA", signature.RIPEMD160withECDSA},
}

//account file name
//TODO: change a default file name
//TODO: add an option -f to specify the file
var wFilePath = account.WALLET_FILENAME

//sub commands of 'account'
var (
	AccountCommand = cli.Command{
		Action:      cli.ShowSubcommandHelp,
		Name:        "account",
		Usage:       "Manage accounts",
		Description: `Manage accounts stored in the wallet`,
		Subcommands: []cli.Command{
			{
				Action:      accountCreate,
				Name:        "add",
				Usage:       "Add a new account",
				ArgsUsage:   "",
				Flags:       addFlags,
				Description: `Add a new account`,
			},
			{
				Action:      accountShow,
				Name:        "list",
				Usage:       "List existing accounts",
				ArgsUsage:   " ",
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

//todo: optimise print info.
func accountCreate(ctx *cli.Context) error {
	checkFileName(ctx)
	reader := bufio.NewReader(os.Stdin)

	var prvkey keypair.PrivateKey
	var pubkey keypair.PublicKey

	var inputKeyTypeInfo keyTypeInfo
	var inputCurveInfo curveInfo
	var inputSchemeInfo schemeInfo

	var defaultFlag = false
	if ctx.IsSet("default") {
		fmt.Printf("use default value for all options \n")
		inputKeyTypeInfo = keyTypeMap[""]
		inputCurveInfo = curveMap[""]
		inputSchemeInfo = schemeMap[""]
		defaultFlag = true
	} else {
		find := false
		if ctx.IsSet("type") {
			for key, val := range keyTypeMap {
				if val.name == ctx.String("signature-scheme") {
					inputKeyTypeInfo = keyTypeMap[key]
					find = true
					fmt.Printf("%s is selected. \n", inputKeyTypeInfo.name)
					break
				}
			}
			if !find {
				fmt.Printf("%s is not a valid content for option -t \n", ctx.String("type"))
			}
		}
		if !find {
			inputKeyTypeInfo = chooseKeyType(reader)
		}

		switch inputKeyTypeInfo.name {
		case "ecdsa":
			find := false
			if (!defaultFlag) && ctx.IsSet("bit-length") {
				for key, val := range curveMap {
					if val.name == ctx.String(utils.AccountKeylenFlag.Name) {
						inputCurveInfo = curveMap[key]
						find = true
						fmt.Printf("%s is selected. \n", inputCurveInfo.name)
						break
					}
				}
				if !find {
					fmt.Printf("%s is not a valid content for option -b \n", ctx.String("bit-length"))
				}
			}
			if !find {
				inputCurveInfo = chooseCurve(reader)
			}

			find = false
			if (!defaultFlag) && ctx.IsSet("signature-scheme") {
				for key, val := range schemeMap {
					if val.name == ctx.String(utils.AccountSigSchemeFlag.Name) {
						inputSchemeInfo = schemeMap[key]
						find = true
						fmt.Printf("%s is selected. \n", inputSchemeInfo.name)
						break
					}
				}
				if !find {
					fmt.Printf("%s is not a valid content for option -s \n", ctx.String("signature-scheme"))
				}
			}
			if !find {
				inputSchemeInfo = chooseScheme(reader)
			}

			break
		case "sm2":
			fmt.Println("Use curve sm2p256v1 with key length of 256 bits and SM3withSM2 as the signature scheme.")

			inputCurveInfo.code = keypair.SM2P256V1
			inputCurveInfo.name = "SM2P256V1"

			inputSchemeInfo.code = signature.SM3withSM2
			inputSchemeInfo.name = "SM3withSM2"
			break
		case "ed25519":
			fmt.Println("Use curve 25519 with key length of 256 bits and Ed25519 as the signature scheme.")

			inputCurveInfo.code = keypair.ED25519
			inputCurveInfo.name = "ED25519"

			inputSchemeInfo.code = signature.SHA512withEDDSA
			inputSchemeInfo.name = "SHA512withEDDSA"
			break
		default:
			return errors.New("it shouldn't go here. \n")
		}

	}

	var password []byte = nil
	if ctx.IsSet(utils.AccountPassFlag.Name) {
		password = []byte(ctx.String(utils.AccountPassFlag.Name))
	} else {
		var input0, input1 []byte
		for i := 0; i < 3; i++ {
			fmt.Print("Enter a password for encrypting the private key:")
			input0 = enterPassword(false)
			fmt.Print("Re-enter password:")
			input1 = enterPassword(true)
			if bytes.Equal(input0, input1) {
				password = input0
				break
			} else {
				fmt.Println("Passwords not match, please try again!")
			}
		}

		if password == nil {
			fmt.Println("Input password error")
			return errors.New("Add account failed")
		}
	}
	walletFile := wFilePath
	if ctx.IsSet(utils.WalletFileFlag.Name) {
		walletFile = ctx.String(utils.WalletFileFlag.Name)
	}

	prvkey, pubkey, _ = keypair.GenerateKeyPair(inputKeyTypeInfo.code, inputCurveInfo.code)
	ta := types.AddressFromPubKey(pubkey)
	address := ta.ToBase58()

	prvSectet, _ := keypair.EncryptPrivateKey(prvkey, address, password)
	h := sha256.Sum256(password)
	for i := 0; i < len(password); i++ {
		password[i] = 0
	}

	var acc = new(account.Accountx)
	acc.SetKeyPair(prvSectet)
	acc.SigSch = inputSchemeInfo.name
	acc.PubKey = hex.EncodeToString(keypair.SerializePublicKey(pubkey))
	acc.PassHash = hex.EncodeToString(h[:])

	wallet := new(account.WalletData)
	err := wallet.Load(walletFile)
	if err != nil {
		wallet.Inititalize()
	}

	wallet.AddAccount(acc)

	if wallet.Save(walletFile) != nil {
		fmt.Println("Wallet file save failed.")
	}

	fmt.Println("\nCreate account successfully.")
	fmt.Println("Address: ", address)
	fmt.Println("Public key:", hex.EncodeToString(keypair.SerializePublicKey(pubkey)))
	fmt.Println("Signature scheme:", acc.SigSch)

	return nil
}

func accountShow(ctx *cli.Context) error {
	checkFileName(ctx)
	wallet := new(account.WalletData)
	err := wallet.Load(wFilePath)
	if err != nil {
		wallet.Inititalize()
	}
	if len(wallet.Accounts) == 0 {
		fmt.Println("No account")
		return nil
	}

	if !ctx.Bool("verbose") {
		// look for every account and show address
		for i, acc := range wallet.Accounts {
			if acc.IsDefault {
				fmt.Printf("* %v\t%v\n", i+1, acc.Address)
			} else {
				fmt.Printf("  %v\t%v\n", i+1, acc.Address)
			}
		}
		fmt.Println("\nUse -v or --verbose option to display details.")
	} else {
		// look for every account and show details
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
		fmt.Println("Missing argument.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	index, err := strconv.Atoi(ctx.Args().First())
	if err != nil {
		fmt.Println("Invalid argument. Account index expected.\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	checkFileName(ctx)
	wallet := new(account.WalletData)
	if wallet.Load(wFilePath) != nil {
		wallet.Inititalize()
	}

	index -= 1
	if index < 0 || index >= len(wallet.Accounts) {
		fmt.Printf("Index out of range.\n")
		return nil
	}

	if ctx.Bool("as-default") {
		fmt.Printf("Set account %d as the default account\n", index+1)
		for _, v := range wallet.Accounts {
			if v.IsDefault {
				v.IsDefault = false
			}
		}
		wallet.Accounts[index].IsDefault = true
	}
	if ctx.IsSet("signature-scheme") {
		find := false
		for key, val := range schemeMap {
			if val.name == ctx.String("signature-scheme") {
				inputSchemeInfo := schemeMap[key]
				find = true
				fmt.Printf("%s is selected. \n", inputSchemeInfo.name)
				wallet.Accounts[index].SigSch = inputSchemeInfo.name
				break
			}
		}
		if !find {
			fmt.Printf("Invalid arguments! Nothing changed.\n")
		}
	}

	if wallet.Save(wFilePath) != nil {
		fmt.Println("Wallet file save failed.")
	}
	return nil
}

//delete an account by index from 'list'
func accountDelete(ctx *cli.Context) error {
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
	checkFileName(ctx)
	wallet := new(account.WalletData)
	if wallet.Load(wFilePath) != nil {
		wallet.Inititalize()
	}

	if index < 1 || index > len(wallet.Accounts) {
		fmt.Printf("Your input is out of index range.\n")
		return nil
	}
	addr := wallet.DelAccount(index)

	if wallet.Save(wFilePath) != nil {
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
	checkFileName(ctx)
	wallet := new(account.WalletData)
	if wallet.Load(wFilePath) != nil {
		wallet.Inititalize()
	}
	if index < 1 || index > len(wallet.Accounts) {
		fmt.Printf("Your input is out of index range.\n")
		return nil
	}

	//password enter.
	var old_password []byte
	for wait := true; wait; {
		fmt.Print("Please enter the original password:")
		old_password = enterPassword(false)
		h := sha256.Sum256(old_password)

		//todo: check is pwd is correct by passhash.
		passhash, _ := hex.DecodeString(wallet.Accounts[index-1].PassHash)
		if bytes.Equal(passhash, h[:]) {
			wait = false
		} else {
			fmt.Print("Wrong password!")
		}
	}

	prv, _ := keypair.DecryptPrivateKey(wallet.Accounts[index-1].GetKeyPair(), old_password)

	//password enter.
	var password, repeatPassword []byte
	for wait := true; wait; {
		fmt.Print("Enter a password for encrypting the private key:")
		password = enterPassword(false)
		fmt.Print("Re-enter password:")
		repeatPassword = enterPassword(true)
		if bytes.Equal(password, repeatPassword) {
			break
		} else {
			fmt.Println("inputs not match, please try again!")
		}
	}

	prvSectet, _ := keypair.EncryptPrivateKey(prv, wallet.Accounts[index-1].Address, password)
	h := sha256.Sum256(password)
	for i := 0; i < len(password); i++ {
		password[i] = 0
	}

	wallet.Accounts[index-1].SetKeyPair(prvSectet)
	wallet.Accounts[index-1].PassHash = hex.EncodeToString(h[:])

	if wallet.Save(wFilePath) != nil {
		fmt.Println("Wallet file save failed.")
	}

	fmt.Println("encrypt account successfully.")
	fmt.Println("")

	return nil
}

// wait for user to enter password
func enterPassword(is_repeat bool) []byte {
	for wait := true; wait; {
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println("system cannot read your input, sorry.")
		}
		fmt.Println("")
		if !is_repeat && len(bytePassword) == 0 {
			fmt.Print("password cannot be null, input again:")
		} else {
			return bytePassword
		}
	}
	return nil
}

// wait for user to choose options
func chooseKeyType(reader *bufio.Reader) keyTypeInfo {
	var t keyTypeInfo

	//choose key type process
	common.PrintNotice("key type")
	for true {
		tmp, _ := reader.ReadString('\n')
		tmp = strings.TrimSpace(tmp)
		_, ok := keyTypeMap[tmp]
		if ok {
			t = keyTypeMap[tmp]
			break
		} else {
			fmt.Print("Input error! Please enter a number above: ")
		}
	}

	fmt.Printf("%s is selected. \n", t.name)
	return t
}
func chooseScheme(reader *bufio.Reader) schemeInfo {
	var s schemeInfo
	common.PrintNotice("signature-scheme")
	for true {
		tmp, _ := reader.ReadString('\n')
		tmp = strings.TrimSpace(tmp)

		_, ok := schemeMap[tmp]
		if ok {
			s = schemeMap[tmp]
			break
		} else {
			fmt.Print("Input error! Please enter a number above:")
		}
	}
	fmt.Printf("scheme %s is selected.\n", s.name)
	return s
}
func chooseCurve(reader *bufio.Reader) curveInfo {
	var c curveInfo
	common.PrintNotice("curve")

	for true {
		t, _ := reader.ReadString('\n')
		t = strings.TrimSpace(t)

		_, ok := curveMap[t]
		if ok {
			c = curveMap[t]
			break
		} else {
			fmt.Print("Input error! Please enter a number above:")
		}
	}

	fmt.Printf("curve %s is selected.\n", c.name)

	return c
}

func checkFileName(ctx *cli.Context) error {
	if ctx.IsSet("file") {
		wFilePath = ctx.String("file")
	}
	return nil
}
