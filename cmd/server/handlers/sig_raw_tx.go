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
	"github.com/ontio/ontology-crypto/keypair"
	clisvrcom "github.com/ontio/ontology/cmd/server/common"
	cliutil "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
)

type SigRawTransactionReq struct {
	RawTx string `json:"raw_tx"`
}

type SigRawTransactionRsp struct {
	SignedTx string `json:"signed_tx"`
}

func SigRawTransaction(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	rawReq := &SigRawTransactionReq{}
	err := json.Unmarshal(req.Params, rawReq)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	rawTxData, err := hex.DecodeString(rawReq.RawTx)
	if err != nil {
		log.Infof("Cli Qid:%s SigRawTransaction hex.DecodeString error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	rawTx := &types.Transaction{}
	err = rawTx.Deserialize(bytes.NewBuffer(rawTxData))
	if err != nil {
		log.Infof("Cli Qid:%s SigRawTransaction tx Deserialize error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_TX
		return
	}
	signer := clisvrcom.DefAccount
	var emptyAddress = common.Address{}
	if rawTx.Payer == emptyAddress {
		rawTx.Payer = signer.Address
	}

	txHash := rawTx.Hash()
	sigData, err := cliutil.Sign(txHash.ToArray(), signer)
	if err != nil {
		log.Infof("Cli Qid:%s SigRawTransaction Sign error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	if len(rawTx.Sigs) == 0 {
		rawTx.Sigs = make([]*types.Sig, 0)
	}
	rawTx.Sigs = append(rawTx.Sigs, &types.Sig{
		PubKeys: []keypair.PublicKey{signer.PublicKey},
		M:       1,
		SigData: [][]byte{sigData},
	})
	buf := bytes.NewBuffer(nil)
	err = rawTx.Serialize(buf)
	if err != nil {
		log.Infof("Cli Qid:%s SigRawTransaction tx Serialize error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	resp.Result = &SigRawTransactionRsp{
		SignedTx: hex.EncodeToString(buf.Bytes()),
	}
}
