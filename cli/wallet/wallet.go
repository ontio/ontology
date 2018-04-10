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

package wallet

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	cliCommon "github.com/ontio/ontology/cli/common"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/password"
	"github.com/ontio/ontology/http/base/rpc"
	"github.com/urfave/cli"
)

func walletAction(c *cli.Context) error {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	// wallet name is wallet.dat by default
	name := c.String("name")
	create := c.Bool("create")
	list := c.Bool("list")
	passwd := c.String("password")
	if name == "" {
		fmt.Println("Invalid wallet name.")
		os.Exit(1)
	}
	if common.FileExisted(name) && create {
		fmt.Printf("CAUTION: '%s' already exists!\n", name)
		os.Exit(1)
	}
	// need to input password when password is not specified from command line
	if passwd == "" {
		var err error
		var tmppasswd []byte
		if create {
			tmppasswd, err = password.GetConfirmedPassword()
		} else {
			tmppasswd, err = password.GetPassword()
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		passwd = string(tmppasswd)
	}
	var wallet *account.ClientImpl
	if create {
		encrypt := c.String("encrypt")
		wallet = account.Create(name, encrypt, []byte(passwd))
	} else {
		// list wallet or change wallet password
		wallet = account.Open(name, []byte(passwd))
	}
	if wallet == nil {
		fmt.Println("Failed to open wallet: ", name)
		os.Exit(1)
	}
	fmt.Printf("Wallet File: '%s'\n", name)
	if c.Bool("changepassword") {
		fmt.Println("# input new password #")
		newPassword, _ := password.GetConfirmedPassword()
		if ok := wallet.ChangePassword([]byte(passwd), newPassword); !ok {
			fmt.Println("error: failed to change password")
			os.Exit(1)
		}
		fmt.Println("password changed")
		return nil
	}
	account, _ := wallet.GetDefaultAccount()
	pubKey := account.PubKey()
	address := account.Address

	pubKeyBytes := keypair.SerializePublicKey(pubKey)
	fmt.Println("public key:   ", common.ToHexString(pubKeyBytes))
	fmt.Println("hex address: ", common.ToHexString(address[:]))
	fmt.Println("base58 address:      ", address.ToBase58())
	balance := c.Bool("balance")
	if list && balance {
		resp, err := rpc.Call(cliCommon.RpcAddress(), "getbalance", 0,
			[]interface{}{address.ToBase58()})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		r := make(map[string]interface{})
		err = json.Unmarshal(resp, &r)
		if err != nil {
			fmt.Println("Unmarshal JSON failed")
			return err
		}

		switch res := r["result"].(type) {
		case map[string]interface{}:
			for k, v := range res {
				fmt.Printf("%s: %v\n", k, v)
			}
			return nil
		case string:
			fmt.Println(res)
			return nil
		}
	}
	return nil
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "wallet",
		Usage:       "user wallet operation",
		Description: "With nodectl wallet, you could control your asset.",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "create, c",
				Usage: "create wallet",
			},
			cli.BoolFlag{
				Name:  "list, l",
				Usage: "list wallet information",
			},
			cli.BoolFlag{
				Name:  "changepassword",
				Usage: "change wallet password",
			},
			cli.BoolFlag{
				Name:  "balance, b",
				Usage: "get ont/ong balance",
			},
			cli.StringFlag{
				Name: "encrypt, e",
				Usage: `encrypt type,just as:
				SHA224withECDSA, SHA256withECDSA,
				SHA384withECDSA, SHA512withECDSA,
				SHA3-224withECDSA, SHA3-256withECDSA,
				SHA3-384withECDSA, SHA3-512withECDSA,
				RIPEMD160withECDSA, SM3withSM2, SHA512withEdDSA`,
			},
			cli.StringFlag{
				Name:  "name, n",
				Usage: "wallet name",
				Value: account.WALLET_FILENAME,
			},
			cli.StringFlag{
				Name:  "password, p",
				Usage: "wallet password",
			},
		},
		Action: walletAction,
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			cliCommon.PrintError(c, err, "wallet")
			return cli.NewExitError("", 1)
		},
	}
}
