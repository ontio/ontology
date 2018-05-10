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
	s "github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/common/password"
	"github.com/ontio/ontology/core/types"
	"github.com/urfave/cli"
	"os"
	"strconv"
	"strings"
)

//map info, to get some information easily
type keyTypeInfo struct {
	name string
	code keypair.KeyType
}

var keyTypeMap = map[string]keyTypeInfo{
	"":  {"ecdsa", keypair.PK_ECDSA},
	"1": {"ecdsa", keypair.PK_ECDSA},
	"2": {"sm2", keypair.PK_SM2},
	"3": {"ed25519", keypair.PK_EDDSA},

	"ecdsa":   {"ecdsa", keypair.PK_ECDSA},
	"sm2":     {"sm2", keypair.PK_SM2},
	"ed25519": {"ed25519", keypair.PK_EDDSA},
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

	"P-224": {"P-224", keypair.P224},
	"P-256": {"P-256", keypair.P256},
	"P-384": {"P-384", keypair.P384},
	"P-521": {"P-521", keypair.P521},

	"224": {"P-224", keypair.P224},
	"256": {"P-256", keypair.P256},
	"384": {"P-384", keypair.P384},
	"521": {"P-521", keypair.P521},

	"SM2P256V1": {"SM2P256V1", keypair.SM2P256V1},
	"ED25519":   {"ED25519", keypair.ED25519},
}

type schemeInfo struct {
	name string
	code s.SignatureScheme
}

var schemeMap = map[string]schemeInfo{
	"":  {"SHA256withECDSA", s.SHA256withECDSA},
	"1": {"SHA224withECDSA", s.SHA224withECDSA},
	"2": {"SHA256withECDSA", s.SHA256withECDSA},
	"3": {"SHA384withECDSA", s.SHA384withECDSA},
	"4": {"SHA512withEDDSA", s.SHA512withEDDSA},
	"5": {"SHA3_224withECDSA", s.SHA3_224withECDSA},
	"6": {"SHA3_256withECDSA", s.SHA3_256withECDSA},
	"7": {"SHA3_384withECDSA", s.SHA3_384withECDSA},
	"8": {"SHA3_512withECDSA", s.SHA3_512withECDSA},
	"9": {"RIPEMD160withECDSA", s.RIPEMD160withECDSA},

	"SHA224withECDSA":    {"SHA224withECDSA", s.SHA224withECDSA},
	"SHA256withECDSA":    {"SHA256withECDSA", s.SHA256withECDSA},
	"SHA384withECDSA":    {"SHA384withECDSA", s.SHA384withECDSA},
	"SHA512withEDDSA":    {"SHA512withEDDSA", s.SHA512withEDDSA},
	"SHA3_224withECDSA":  {"SHA3_224withECDSA", s.SHA3_224withECDSA},
	"SHA3_256withECDSA":  {"SHA3_256withECDSA", s.SHA3_256withECDSA},
	"SHA3_384withECDSA":  {"SHA3_384withECDSA", s.SHA3_384withECDSA},
	"SHA3_512withECDSA":  {"SHA3_512withECDSA", s.SHA3_512withECDSA},
	"RIPEMD160withECDSA": {"RIPEMD160withECDSA", s.RIPEMD160withECDSA},

	"SM3withSM2": {"SM3withSM2", s.SM3withSM2},
}

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

			{
				Action:      accountImport,
				Name:        "import",
				ArgsUsage:   "[sub-command options] <args>",
				Usage:       "Import the account to current wallet.",
				Flags:       importFlags,
				Description: `Import the accounts of source wallet file to current wallet. Import an account by private key(not recommended).`,
			},
		},
	}
)

func accountCreate(ctx *cli.Context) error {

	reader := bufio.NewReader(os.Stdin)
	optionType := ""
	optionCurve := ""
	optionScheme := ""

	optionDefault := ctx.IsSet("default")
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

	wallet := new(account.WalletData)
	err := wallet.Load(optionFile)
	if err != nil {
		fmt.Printf("%v doesn't exist, create file automatically.\n\n", optionFile)
		wallet.Inititalize()
	}

	for i := 0; i < optionNumber; i++ {
		acc := account.CreateAccount(keyTypeMap[optionType].code, curveMap[optionCurve].code, schemeMap[optionScheme].name, &pass)
		acc.SetLabel(optionLabel)
		wallet.AddAccount(acc)

		fmt.Println()
		fmt.Println("Label: ", acc.Label)
		fmt.Println("Address: ", acc.Address)
		fmt.Println("Public key:", acc.PubKey)
		fmt.Println("Signature scheme:", acc.SigSch)
	}
	fmt.Println()

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
				fmt.Printf("* %v\t%v\t%v\n", i+1, acc.Address, acc.Label)
			} else {
				fmt.Printf("  %v\t%v\t%v\n", i+1, acc.Address, acc.Label)
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
			fmt.Printf("	Label: %v\n", acc.Label)
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

	if ctx.Bool("as-default") {
		fmt.Printf("Account <%v> is set as default account.\n", index)
		for _, v := range wallet.Accounts {
			if v.IsDefault {
				v.IsDefault = false
			}
		}
		wallet.Accounts[index-1].IsDefault = true
	}
	if ctx.IsSet("signature-scheme") {
		find := false
		for key, val := range schemeMap {
			if val.name == ctx.String("signature-scheme") {
				inputSchemeInfo := schemeMap[key]
				find = true
				fmt.Printf("Account <%v>'s signature scheme is set to '%s'.\n", index, inputSchemeInfo.name)
				wallet.Accounts[index-1].SigSch = inputSchemeInfo.name
				break
			}
		}
		if !find {
			fmt.Printf("%s is not a valid content for option -s \n", ctx.String("signature-scheme"))
		}
	}

	if ctx.IsSet("label") {
		wallet.Accounts[index-1].Label = checkLabel(ctx)
		fmt.Printf("Account <%v>'s label is set to '%v'.\n", index, wallet.Accounts[index-1].Label)
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
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	index, err := strconv.Atoi(ctx.Args().First())
	if err != nil {
		fmt.Printf("Your input is not an number.\n")
		cli.ShowSubcommandHelp(ctx)
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

	label := wallet.Accounts[index-1].Label
	addr := wallet.DelAccount(index)

	if wallet.Save(optionFile) != nil {
		fmt.Println("Wallet file save failed.")
	}

	fmt.Printf("Delete account successfully.\n")
	fmt.Printf("index = %v, address = %v, label=%v.\n", index, addr, label)
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

func accountImport(ctx *cli.Context) error {
	optionFile := checkFileName(ctx)
	wallet := new(account.WalletData)
	if wallet.Load(optionFile) != nil {
		wallet.Inititalize()
	}
	accQty := 0
	if ctx.IsSet("source") { //check -source
		source := ctx.String("source")
		sourceWallet := new(account.WalletData)
		//TODO: check wallet file format.
		err := sourceWallet.Load(source)
		if err != nil {
			fmt.Printf("%s doesn't exist, import failed.\n\n", source)
			return nil
		}
		accQty = len(sourceWallet.Accounts)
		if accQty == 0 {
			fmt.Printf("%s doesn't have any account, nothing has been imported.\n\n", source)
			return nil
		}
		// remove default account setting in source file
		for _, v := range sourceWallet.Accounts {
			if v.IsDefault {
				v.IsDefault = false
			}
		}

		wallet.Accounts = append(wallet.Accounts, sourceWallet.Accounts...)
	} else if ctx.IsSet("key") { //check -key
		//TODO: wait to discuss what type of key to import.
		fmt.Printf("-key is not supported currentlt, please use -source to import account.\n\n")
		return nil

		accQty = 1

		skHex := ctx.String("key")
		skStr, err := hex.DecodeString(skHex)
		if err != nil {
			fmt.Printf("cannot parse key!\n")
			return nil
		}
		prvkey, _ := keypair.DeserializePrivateKey(skStr)
		pubkey := prvkey.Public()
		ta := types.AddressFromPubKey(pubkey)
		address := ta.ToBase58()

		pass := make([]byte, 0, 0)
		if ctx.IsSet("password") {
			pass = []byte(ctx.String("password"))
		} else {
			//let user enter password with double check
			pass, _ = password.GetConfirmedPassword()
		}

		prvSecret, _ := keypair.EncryptPrivateKey(prvkey, address, pass)
		h := sha256.Sum256(pass)
		for i := 0; i < len(pass); i++ {
			pass[i] = 0
		}

		var acc = new(account.AccountData)
		acc.SetKeyPair(prvSecret)
		acc.SigSch = ""
		acc.PubKey = hex.EncodeToString(keypair.SerializePublicKey(pubkey))
		acc.PassHash = hex.EncodeToString(h[:])

		wallet.Accounts = append(wallet.Accounts, acc)
	}

	if wallet.Save(optionFile) != nil {
		fmt.Println("Wallet file save failed.")
	} else {
		fmt.Printf("Import finished. %d accounts has been imported.\n\n", accQty)
	}
	return nil
}

// wait for user to choose options
func chooseKeyType(reader *bufio.Reader) string {
	common.PrintNotice("key type")
	for true {
		tmp, _ := reader.ReadString('\n')
		tmp = strings.TrimSpace(tmp)
		_, ok := keyTypeMap[tmp]
		if ok {
			fmt.Printf("%s is selected. \n", keyTypeMap[tmp].name)
			return keyTypeMap[tmp].name
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

		_, ok := schemeMap[tmp]
		if ok {
			fmt.Printf("scheme %s is selected.\n", schemeMap[tmp].name)
			return schemeMap[tmp].name
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
		_, ok := curveMap[tmp]
		if ok {
			fmt.Printf("scheme %s is selected.\n", curveMap[tmp].name)
			return curveMap[tmp].name
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
func checkNumber(ctx *cli.Context) int {
	if ctx.IsSet("number") {
		if ctx.Uint("number") < 1 {
			fmt.Println("the minimum number is 1, set to default value(1).")
			return 1
		}
		if ctx.Uint("number") > 100 {
			fmt.Println("the maximum number is 100, set to default value(1).")
			return 1
		}
		return int(ctx.Uint("number"))
	} else {
		return 1
	}
}
func checkLabel(ctx *cli.Context) string {
	if ctx.IsSet("label") {
		return ctx.String("label")
	} else {
		return ""
	}
}
func checkType(ctx *cli.Context, reader *bufio.Reader) string {
	t := ""
	if ctx.IsSet("type") {
		if _, ok := keyTypeMap[ctx.String("type")]; ok {
			t = keyTypeMap[ctx.String("type")].name
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
			if _, ok := curveMap[ctx.String("bit-length")]; ok {
				c = curveMap[ctx.String("bit-length")].name
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
	sch := ""
	switch *t {
	case "ecdsa":
		if ctx.IsSet("signature-scheme") {
			if _, ok := schemeMap[ctx.String("signature-scheme")]; ok {
				sch = schemeMap[ctx.String("signature-scheme")].name
				fmt.Printf("%s is selected. \n", sch)
			} else {
				fmt.Printf("%s is not a valid content for option -s \n", ctx.String("signature-scheme"))
				sch = chooseScheme(reader)
			}
		} else {
			sch = chooseScheme(reader)
		}
		break
	case "sm2":
		fmt.Println("Use SM3withSM2 as the signature scheme.")
		sch = "SM3withSM2"
		break
	case "ed25519":
		fmt.Println("Use Ed25519 as the signature scheme.")
		sch = "SHA512withEDDSA"
		break
	default:
		return ""
	}
	return sch
}
