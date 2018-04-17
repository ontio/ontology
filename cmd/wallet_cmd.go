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
	"os"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/password"
	"github.com/urfave/cli"
)

var (
	WalletCommand = cli.Command{
		Name:        "wallet",
		Usage:       "controll wallet,just as create delete action .etc",
		ArgsUsage:   "",
		Category:    "WALLET COMMANDS",
		Description: `[create/show]`,
		Subcommands: []cli.Command{
			{
				Action:      utils.MigrateFlags(walletCreate),
				Name:        "create",
				Usage:       "./ontology wallet create [OPTION]",
				Flags:       append(append(NodeFlags, RpcFlags...), ContractFlags...),
				Category:    "WALLET COMMANDS",
				Description: ``,
			},
			{
				Action:      utils.MigrateFlags(walletShow),
				Name:        "show",
				Usage:       "./ontology wallet show [OPTION]",
				Flags:       append(append(NodeFlags, RpcFlags...), ContractFlags...),
				Category:    "WALLET COMMANDS",
				Description: ``,
			},
			{
				Action:      utils.MigrateFlags(walletBalance),
				Name:        "balance",
				Usage:       "./ontology wallet balance",
				Flags:       append(append(NodeFlags, RpcFlags...), ContractFlags...),
				Category:    "WALLET COMMANDS",
				Description: ``,
			},
		},
	}
)

func walletCreate(ctx *cli.Context) error {
	encrypt := ctx.GlobalString(utils.EncryptTypeFlag.Name)

	name := ctx.GlobalString(utils.WalletNameFlag.Name)
	if name == "" {
		fmt.Println("Invalid wallet name.")
		os.Exit(1)
	}
	if common.FileExisted(name) {
		fmt.Printf("CAUTION: '%s' already exists!\n", name)
		os.Exit(1)
	}
	tmppasswd, err := password.GetConfirmedPassword()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	passwd := string(tmppasswd)
	wallet := account.Create(name, encrypt, []byte(passwd))
	account := wallet.GetDefaultAccount()

	pubKey := account.PubKey()
	address := account.Address

	pubKeyBytes := keypair.SerializePublicKey(pubKey)
	fmt.Println("public key:   ", common.ToHexString(pubKeyBytes))
	fmt.Println("hex address: ", common.ToHexString(address[:]))
	fmt.Println("base58 address:      ", address.ToBase58())

	return nil
}

func walletShow(ctx *cli.Context) error {
	client := account.GetClient(ctx)
	if client == nil {
		log.Fatal("Can't get local account.")
	}

	acct := client.GetDefaultAccount()
	if acct == nil {
		log.Fatal("can not get default account")
	}

	pubKey := acct.PubKey()
	address := acct.Address

	pubKeyBytes := keypair.SerializePublicKey(pubKey)
	fmt.Println("public key:   ", common.ToHexString(pubKeyBytes))
	fmt.Println("hex address: ", common.ToHexString(address[:]))
	fmt.Println("base58 address:      ", address.ToBase58())
	return nil
}

func walletBalance(ctx *cli.Context) error {
	addr := ctx.GlobalString(utils.WalletAddrFlag.Name)
	balance, err := ontSdk.Rpc.GetBalanceWithBase58(addr)
	if nil != err {
		log.Fatal("Get Balance with base58 err: ", err.Error())
	}
	fmt.Printf("ONT: %d; ONG: %d; ONGAppove: %d\n", balance.Ont.Int64(), balance.Ong.Int64(), balance.OngAppove.Int64())
	return nil
}
