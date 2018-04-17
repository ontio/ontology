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
	"math/big"
	"os"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/actor"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/password"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	ldgactor "github.com/ontio/ontology/core/ledger/actor"
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
				Flags:       append(append(NodeFlags, RpcFlags...), WhisperFlags...),
				Category:    "WALLET COMMANDS",
				Description: ``,
			},
			{
				Action:      utils.MigrateFlags(walletShow),
				Name:        "show",
				Usage:       "./ontology wallet show [OPTION]",
				Flags:       append(append(NodeFlags, RpcFlags...), WhisperFlags...),
				Category:    "WALLET COMMANDS",
				Description: ``,
			},
			{
				Action:      utils.MigrateFlags(walletBalance),
				Name:        "balance",
				Usage:       "./ontology wallet balance",
				Flags:       append(append(NodeFlags, RpcFlags...), WhisperFlags...),
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
	fmt.Println("Hello show wallet")
	return nil
}

func walletBalance(ctx *cli.Context) error {
	ledger.DefLedger, _ = ledger.NewLedger()
	ldgerActor := ldgactor.NewLedgerActor()
	ledgerPID := ldgerActor.Start()
	actor.SetLedgerPid(ledgerPID)
	addrBase58 := ctx.GlobalString(utils.WalletAddrFlag.Name)

	address, err := common.AddressFromBase58(addrBase58)
	if err != nil {
		//
	}
	ont := new(big.Int)
	ong := new(big.Int)
	appove := big.NewInt(0)

	ontBalance, err := actor.GetStorageItem(genesis.OntContractAddress, address[:])
	if err != nil {
		log.Errorf("GetOntBalanceOf GetStorageItem ont address:%s error:%s", addrBase58, err)
	}
	if ontBalance != nil {
		ont.SetBytes(ontBalance)
	}

	appoveKey := append(genesis.OntContractAddress[:], address[:]...)
	ongappove, err := actor.GetStorageItem(genesis.OngContractAddress, appoveKey[:])

	if ongappove != nil {
		appove.SetBytes(ongappove)
	}
	fmt.Printf("Ont:    %s\nOng:    %s\nOngAppove:    %s\n", ont.String(), ong.String(), appove.String())
	return nil
}
