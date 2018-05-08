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
	"bytes"
	"encoding/hex"
	"encoding/json"
	clisvrcom "github.com/ontio/ontology/cmd/server/common"
	cliutil "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common/log"
)

type SigTransferTransactionReq struct {
	GasPrice uint64 `json:"gas_price"`
	GasLimit uint64 `json:"gas_limit"`
	Asset    string `json:"asset"`
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   uint64 `json:"amount"`
}

type SinTransferTransactionRsp struct {
	SignedTx string `json:"signed_tx"`
}

func SigTransferTransaction(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	rawReq := &SigTransferTransactionReq{}
	err := json.Unmarshal(req.Params, rawReq)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	transferTx, err := cliutil.TransferTx(rawReq.GasPrice, rawReq.GasLimit, rawReq.Asset, rawReq.From, rawReq.To, rawReq.Amount)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	signer := clisvrcom.DefAccount
	if signer == nil {
		resp.ErrorCode = clisvrcom.CLIERR_ACCOUNT_UNLOCK
		return
	}
	err = cliutil.SignTransaction(signer, transferTx)
	if err != nil {
		log.Infof("Cli Qid:%s SigTransferTransaction SignTransaction error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	buf := bytes.NewBuffer(nil)
	err = transferTx.Serialize(buf)
	if err != nil {
		log.Infof("Cli Qid:%s SigTransferTransaction tx Serialize error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	resp.Result = &SinTransferTransactionRsp{
		SignedTx: hex.EncodeToString(buf.Bytes()),
	}
}
