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
package sigsvr

import (
	"fmt"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/sigsvr/store"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/urfave/cli"
)

var ImportWalletCommand = cli.Command{
	Name:      "import",
	Usage:     "Import accounts from a wallet file",
	ArgsUsage: "",
	Action:    importWallet,
	Flags: []cli.Flag{
		utils.CliWalletDirFlag,
		utils.WalletFileFlag,
	},
	Description: "",
}

func importWallet(ctx *cli.Context) error {
	walletDirPath := ctx.String(utils.GetFlagName(utils.CliWalletDirFlag))
	walletFilePath := ctx.String(utils.GetFlagName(utils.WalletFileFlag))
	if walletDirPath == "" || walletFilePath == "" {
		fmt.Printf("walletdir or wallet flag cannot empty\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	if !common.FileExisted(walletFilePath) {
		fmt.Printf("wallet file:%s doesnot exist\n", walletFilePath)
		return nil
	}
	walletStore, err := store.NewWalletStore(walletDirPath)
	if err != nil {
		fmt.Printf("NewWalletStore dir path:%s error:%s", walletDirPath, err)
		return nil
	}
	wallet, err := account.Open(walletFilePath)
	if err != nil {
		fmt.Printf("Open wallet:%s error:%s\n", walletFilePath, err)
		return nil
	}
	walletData := wallet.GetWalletData()
	if *walletStore.WalletScrypt != *walletData.Scrypt {
		fmt.Printf("Import account failed, wallet scrypt:%+v != %+v\n", walletData.Scrypt, walletStore.WalletScrypt)
		return nil
	}
	for i := 0; i < len(walletData.Accounts); i++ {
		err = walletStore.AddAccountData(walletData.Accounts[i])
		if err != nil {
			fmt.Printf("Import account address:%s error:%s\n", walletData.Accounts[i].Address, err)
			return nil
		}
	}
	fmt.Printf("Import account success\n")
	fmt.Printf("Account number:%d\n", len(walletData.Accounts))
	return nil
}
