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
package handlers

import (
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	clisvrcom "github.com/ontio/ontology/cmd/sigsvr/common"
	"github.com/ontio/ontology/common/log"
)

type CreateAccountReq struct {
}

type CreateAccountRsp struct {
	Account string `json:"account"`
}

func CreateAccount(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	pwd := req.Pwd
	if pwd == "" {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		resp.ErrorInfo = "pwd cannot empty"
		return
	}
	walletPath := req.Wallet
	var err error
	var wallet account.Client
	if walletPath != "" {
		wallet, err = account.Open(walletPath)
		if err != nil {
			resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
			resp.ErrorInfo = "create wallet failed"
			log.Errorf("CreateAccount Qid:%s create wallet:%s error:%s", req.Qid, walletPath, err)
			return
		}
	} else {
		wallet = clisvrcom.DefWallet
	}
	if wallet == nil {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		resp.ErrorInfo = "no wallet to create account"
		return
	}

	acc, err := wallet.NewAccount("", keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, []byte(pwd))
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		resp.ErrorInfo = "create wallet failed"
		log.Errorf("CreateAccount Qid:%s create account  error:%s", req.Qid, err)
		return
	}

	resp.Result = &CreateAccountRsp{
		Account: acc.Address.ToBase58(),
	}
}
