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

package common

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/password"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/account"
)

func OpenWallet(ctx *cli.Context) (*account.ClientImpl, error){
	walletFile := ctx.GlobalString(utils.WalletFileFlag.Name)
	if walletFile == "" {
		walletFile = config.DEFAULT_WALLET_FILE_NAME
	}
	if !common.FileExisted(walletFile) {
		return nil, fmt.Errorf("Transfer failed, cannot found wallet:%s", walletFile)
	}
	var err error
	var passwd []byte
	if ctx.IsSet("password") {
		passwd = []byte(ctx.String("password"))
	} else {
		passwd, err = password.GetAccountPassword()
		if err != nil {
			return nil, fmt.Errorf("Input password error:%s", err)
		}
	}
	wallet := account.Open(walletFile, passwd)
	if wallet == nil {
		return  nil, fmt.Errorf("open wallet:%s failed", walletFile)
	}
	return wallet, nil
}

func CommonCommandErrorHandler(ctx *cli.Context, err error, isSubcommand bool)error {
	fmt.Printf("%s\n", err)
	if isSubcommand {
		cli.ShowSubcommandHelp(ctx)
	}else{
		cli.ShowCommandCompletions(ctx, ctx.Command.Name)
	}
	return nil
}