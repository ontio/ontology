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
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common/config"
	"github.com/urfave/cli"
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
	"4": {"SHA512withEdDSA", s.SHA512withEDDSA},
	"5": {"SHA3-224withECDSA", s.SHA3_224withECDSA},
	"6": {"SHA3-256withECDSA", s.SHA3_256withECDSA},
	"7": {"SHA3-384withECDSA", s.SHA3_384withECDSA},
	"8": {"SHA3-512withECDSA", s.SHA3_512withECDSA},
	"9": {"RIPEMD160withECDSA", s.RIPEMD160withECDSA},

	"SHA224withECDSA":    {"SHA224withECDSA", s.SHA224withECDSA},
	"SHA256withECDSA":    {"SHA256withECDSA", s.SHA256withECDSA},
	"SHA384withECDSA":    {"SHA384withECDSA", s.SHA384withECDSA},
	"SHA512withEdDSA":    {"SHA512withEdDSA", s.SHA512withEDDSA},
	"SHA3-224withECDSA":  {"SHA3-224withECDSA", s.SHA3_224withECDSA},
	"SHA3-256withECDSA":  {"SHA3-256withECDSA", s.SHA3_256withECDSA},
	"SHA3-384withECDSA":  {"SHA3-384withECDSA", s.SHA3_384withECDSA},
	"SHA3-512withECDSA":  {"SHA3-512withECDSA", s.SHA3_512withECDSA},
	"RIPEMD160withECDSA": {"RIPEMD160withECDSA", s.RIPEMD160withECDSA},

	"SM3withSM2": {"SM3withSM2", s.SM3withSM2},
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
	if ctx.IsSet(utils.GetFlagName(utils.WalletFileFlag)) {
		return ctx.String(utils.GetFlagName(utils.WalletFileFlag))
	} else {
		//default account file name
		return config.DEFAULT_WALLET_FILE_NAME
	}
}
func checkNumber(ctx *cli.Context) int {
	numFlag := utils.GetFlagName(utils.AccountQuantityFlag)
	if ctx.IsSet(numFlag) {
		num := ctx.Uint(numFlag)
		if num < 1 {
			fmt.Println("the minimum number is 1, set to default value(1).")
			return 1
		}
		if num > 100 {
			fmt.Println("the maximum number is 100, set to default value(1).")
			return 1
		}
		return int(num)
	} else {
		return 1
	}
}
func checkLabel(ctx *cli.Context) string {
	if ctx.IsSet(utils.GetFlagName(utils.AccountLabelFlag)) {
		return ctx.String(utils.GetFlagName(utils.AccountLabelFlag))
	} else {
		return ""
	}
}
func checkType(ctx *cli.Context, reader *bufio.Reader) string {
	t := ""
	typeFlag := utils.GetFlagName(utils.AccountTypeFlag)
	if ctx.IsSet(typeFlag) {
		if _, ok := keyTypeMap[ctx.String(typeFlag)]; ok {
			t = keyTypeMap[ctx.String(typeFlag)].name
			fmt.Printf("%s is selected. \n", t)
		} else {
			fmt.Printf("%s is not a valid content for option -t \n", ctx.String(typeFlag))
			t = chooseKeyType(reader)
		}
	} else {
		t = chooseKeyType(reader)
	}
	return t
}
func checkCurve(ctx *cli.Context, reader *bufio.Reader, t *string) string {
	bitFlag := utils.GetFlagName(utils.AccountKeylenFlag)
	c := ""
	switch *t {
	case "ecdsa":
		if ctx.IsSet(bitFlag) {
			if _, ok := curveMap[ctx.String(bitFlag)]; ok {
				c = curveMap[ctx.String(bitFlag)].name
				fmt.Printf("%s is selected. \n", c)
			} else {
				fmt.Printf("%s is not a valid content for option -b \n", ctx.String(bitFlag))
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
	sigFlag := utils.GetFlagName(utils.AccountSigSchemeFlag)
	switch *t {
	case "ecdsa":
		if ctx.IsSet(sigFlag) {
			if _, ok := schemeMap[ctx.String(sigFlag)]; ok {
				sch = schemeMap[ctx.String(sigFlag)].name
				fmt.Printf("%s is selected. \n", sch)
			} else {
				fmt.Printf("%s is not a valid content for option -s \n", ctx.String(sigFlag))
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
