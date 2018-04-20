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

	"errors"
	"reflect"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/password"
	"github.com/urfave/cli"
)

var (
	WalletCommand = cli.Command{
		Action:      utils.MigrateFlags(walletCommand),
		Name:        "wallet",
		Usage:       "ontology wallet [create|show|balance] [OPTION]",
		ArgsUsage:   "",
		Category:    "WALLET COMMANDS",
		Description: `[create/show/balance]`,
		Subcommands: []cli.Command{
			{
				Action:      utils.MigrateFlags(walletCreate),
				Name:        "create",
				Usage:       "ontology wallet create [OPTION]\n",
				Flags:       append(NodeFlags, ContractFlags...),
				Category:    "WALLET COMMANDS",
				Description: ``,
			},
			{
				Action:      utils.MigrateFlags(walletShow),
				Name:        "show",
				Usage:       "ontology wallet show [OPTION]\n",
				Flags:       append(NodeFlags, ContractFlags...),
				Category:    "WALLET COMMANDS",
				Description: ``,
			},
			{
				Action:      utils.MigrateFlags(walletBalance),
				Name:        "balance",
				Usage:       "ontology wallet balance\n",
				Flags:       append(NodeFlags, ContractFlags...),
				Category:    "WALLET COMMANDS",
				Description: ``,
			},
		},
	}
)

func walletCommand(context *cli.Context) error {
	showWalletHelp()
	return nil
}

func walletCreate(ctx *cli.Context) error {
	encrypt := ctx.String(utils.EncryptTypeFlag.Name)

	name := ctx.String(utils.WalletNameFlag.Name)
	if name == "" {
		fmt.Println("Invalid wallet name.")
		return errors.New("Wallet name is necessary")
	}
	if common.FileExisted(name) {
		fmt.Printf("CAUTION: '%s' already exists!\n", name)
		return errors.New("File alreay exists")
	}
	tmpPassword, err := password.GetConfirmedPassword()
	if err != nil {
		fmt.Println(err)
		return err
	}
	password := string(tmpPassword)
	wallet := account.Create(name, encrypt, []byte(password))
	account := wallet.GetDefaultAccount()

	pubKey := account.PubKey()
	address := account.Address

	pubKeyBytes := keypair.SerializePublicKey(pubKey)
	fmt.Println("public key:     \t", common.ToHexString(pubKeyBytes))
	fmt.Println("hex address:    \t", common.ToHexString(address[:]))
	fmt.Println("base58 address: \t", address.ToBase58())

	return nil
}

func walletShow(ctx *cli.Context) error {
	client := account.GetClient(ctx)
	cli := reflect.ValueOf(client)
	if !cli.IsValid() || cli.IsNil() || nil == client {
		fmt.Println("Can't get local account.")
		return errors.New("Can't get local account. ")
	}
	acct := client.GetDefaultAccount()
	if acct == nil {
		fmt.Println("can not get default account")
		return errors.New("can not get default account")
	}

	pubKey := acct.PubKey()
	address := acct.Address

	pubKeyBytes := keypair.SerializePublicKey(pubKey)
	fmt.Println("public key:     \t", common.ToHexString(pubKeyBytes))
	fmt.Println("hex address:    \t", common.ToHexString(address[:]))
	fmt.Println("base58 address: \t", address.ToBase58())
	return nil
}

func walletBalance(ctx *cli.Context) error {
	client := account.GetClient(ctx)
	if client == nil {
		fmt.Println("Can't get local account.")
		return errors.New("Can't get local account. ")
	}

	acct := client.GetDefaultAccount()
	if acct == nil {
		fmt.Println("can not get default account")
		return errors.New("can not get default account")
	}

	base58Addr := acct.Address.ToBase58()
	balance, err := ontSdk.Rpc.GetBalanceWithBase58(base58Addr)
	if nil != err {
		fmt.Printf("Get Balance with base58 err: %s", err.Error())
		return err
	}
	fmt.Printf("ONT: %d; ONG: %d; ONGAppove: %d\n Address(base58): %s\n", balance.Ont.Int64(), balance.Ong.Int64(), balance.OngAppove.Int64(), base58Addr)
	return nil
}
