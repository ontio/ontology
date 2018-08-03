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
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
)

var DefWallet account.Client

type CliRpcRequest struct {
	Qid     string          `json:"qid"`
	Params  json.RawMessage `json:"params"`
	Wallet  string          `json:"wallet"`
	Account string          `json:"account"`
	Pwd     string          `json:"pwd"`
	Method  string          `json:"method"`
}

func (this *CliRpcRequest) GetAccount() (*account.Account, error) {
	var wallet account.Client
	var acc *account.Account
	var err error
	if this.Wallet != "" {
		if !common.FileExisted(this.Wallet) {
			return nil, fmt.Errorf("wallet doesnot exist")
		}
		wallet, err = account.Open(this.Wallet)
		if err != nil {
			return nil, err
		}
	} else {
		wallet = DefWallet
	}
	if wallet == nil {
		return nil, fmt.Errorf("no wallet to sig")
	}
	pwd := []byte(this.Pwd)
	if this.Pwd == "" {
		return nil, fmt.Errorf("pwd cannot empty")
	}
	if this.Account == "" {
		return nil, fmt.Errorf("account cannot empty")
	}
	acc, err = wallet.GetAccountByAddress(this.Account, pwd)
	if err != nil {
		return nil, err
	}
	if acc != nil {
		return acc, nil
	}
	acc, err = wallet.GetAccountByLabel(this.Account, pwd)
	if err != nil {
		return nil, err
	}
	if acc != nil {
		return acc, nil
	}
	return nil, fmt.Errorf("cannot find account by %s", this.Account)
}

type CliRpcResponse struct {
	Qid       string      `json:"qid"`
	Method    string      `json:"method"`
	Result    interface{} `json:"result"`
	ErrorCode int         `json:"error_code"`
	ErrorInfo string      `json:"error_info"`
}
