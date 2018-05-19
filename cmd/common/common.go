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
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/password"
	"github.com/urfave/cli"
	"strconv"
)

func GetPasswd(ctx *cli.Context) ([]byte, error) {
	var passwd []byte
	var err error
	if ctx.IsSet(utils.GetFlagName(utils.AccountPassFlag)) {
		passwd = []byte(ctx.String(utils.GetFlagName(utils.AccountPassFlag)))
	} else {
		passwd, err = password.GetAccountPassword()
		if err != nil {
			return nil, fmt.Errorf("Input password error:%s", err)
		}
	}
	return passwd, nil
}

func OpenWallet(ctx *cli.Context) (account.Client, error) {
	walletFile := ctx.String(utils.GetFlagName(utils.WalletFileFlag))
	if walletFile == "" {
		walletFile = config.DEFAULT_WALLET_FILE_NAME
	}
	if !common.FileExisted(walletFile) {
		return nil, fmt.Errorf("cannot find wallet file:%s", walletFile)
	}
	wallet, err := account.Open(walletFile)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func GetAccountMulti(wallet account.Client, passwd []byte, accAddr string) (*account.Account, error) {
	//Address maybe address in base58, label or index
	if accAddr == "" {
		fmt.Printf("Using default account:%s\n", accAddr)
		return wallet.GetDefaultAccount(passwd)
	}
	acc, err := wallet.GetAccountByAddress(accAddr, passwd)
	if err != nil {
		return nil, fmt.Errorf("getAccountByAddress:%s error:%s", accAddr, err)
	}
	if acc != nil {
		return acc, nil
	}
	acc, err = wallet.GetAccountByLabel(accAddr, passwd)
	if err != nil {
		return nil, fmt.Errorf("getAccountByLabel:%s error:%s", accAddr, err)
	}
	if acc != nil {
		return acc, nil
	}
	index, err := strconv.ParseInt(accAddr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("cannot get account by:%s", accAddr)
	}
	acc, err = wallet.GetAccountByIndex(int(index), passwd)
	if err != nil {
		return nil, fmt.Errorf("getAccountByIndex:%d error:%s", index, err)
	}
	if acc != nil {
		return acc, nil
	}
	return nil, fmt.Errorf("cannot get account by:%s", accAddr)
}

func GetAccountMetadataMulti(wallet account.Client, accAddr string) *account.AccountMetadata {
	//Address maybe address in base58, label or index
	if accAddr == "" {
		fmt.Printf("Using default account:%s\n", accAddr)
		return wallet.GetDefaultAccountMetadata()
	}
	acc := wallet.GetAccountMetadataByAddress(accAddr)
	if acc != nil {
		return acc
	}
	acc = wallet.GetAccountMetadataByLabel(accAddr)
	if acc != nil {
		return acc
	}
	index, err := strconv.ParseInt(accAddr, 10, 32)
	if err != nil {
		return nil
	}
	return wallet.GetAccountMetadataByIndex(int(index))
}

func GetAccount(ctx *cli.Context, address ...string) (*account.Account, error) {
	wallet, err := OpenWallet(ctx)
	if err != nil {
		return nil, err
	}
	passwd, err := GetPasswd(ctx)
	if err != nil {
		return nil, err
	}
	defer ClearPasswd(passwd)
	accAddr := ""
	if len(address) > 0 {
		accAddr = address[0]
	} else {
		accAddr = ctx.String(utils.GetFlagName(utils.AccountAddressFlag))
	}
	return GetAccountMulti(wallet, passwd, accAddr)
}

func IsBase58Address(address string) bool {
	if address == "" {
		return false
	}
	_, err := common.AddressFromBase58(address)
	return err == nil
}

//ParseAddress return base58 address from base58, label or index
func ParseAddress(address string, ctx *cli.Context) (string, error) {
	if IsBase58Address(address) {
		return address, nil
	}
	wallet, err := OpenWallet(ctx)
	if err != nil {
		return "", err
	}
	acc := wallet.GetAccountMetadataByLabel(address)
	if acc != nil {
		return acc.Address, nil
	}
	index, err := strconv.ParseInt(address, 10, 32)
	if err != nil {
		return "", fmt.Errorf("cannot get account by:%s", address)
	}
	acc = wallet.GetAccountMetadataByIndex(int(index))
	if acc != nil {
		return acc.Address, nil
	}
	return "", fmt.Errorf("cannot get account by:%s", address)
}

func ClearPasswd(passwd []byte) {
	size := len(passwd)
	for i := 0; i < size; i++ {
		passwd[i] = 0
	}
}
