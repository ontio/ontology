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
	"encoding/hex"
	"encoding/json"

	"github.com/ontio/ontology-crypto/keypair"
	clisvrcom "github.com/ontio/ontology/cmd/sigsvr/common"
	cliutil "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
)

type SigMutilRawTransactionReq struct {
	RawTx   string   `json:"raw_tx"`
	M       int      `json:"m"`
	PubKeys []string `json:"pub_keys"`
}

type SigMutilRawTransactionRsp struct {
	SignedTx string `json:"signed_tx"`
}

func SigMutilRawTransaction(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	rawReq := &SigMutilRawTransactionReq{}
	err := json.Unmarshal(req.Params, rawReq)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	numkeys := len(rawReq.PubKeys)
	if rawReq.M <= 0 || numkeys < rawReq.M || numkeys <= 1 || numkeys > constants.MULTI_SIG_MAX_PUBKEY_SIZE {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	rawTxData, err := hex.DecodeString(rawReq.RawTx)
	if err != nil {
		log.Infof("Cli Qid:%s SigMutilRawTransaction hex.DecodeString error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}

	tmpTx, err := types.TransactionFromRawBytes(rawTxData)
	if err != nil {
		log.Infof("Cli Qid:%s SigMutilRawTransaction TransactionFromRawBytes error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_TX
		return
	}
	mutTx, err := tmpTx.IntoMutable()
	if err != nil {
		log.Infof("Cli Qid:%s SigMutilRawTransaction IntoMutable error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_TX
		return
	}

	pubKeys := make([]keypair.PublicKey, 0, len(rawReq.PubKeys))
	for _, pkStr := range rawReq.PubKeys {
		pkData, err := hex.DecodeString(pkStr)
		if err != nil {
			log.Info("Cli Qid:%s SigMutilRawTransaction pk hex.DecodeString error:%s", req.Qid, err)
			resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
			return
		}
		pk, err := keypair.DeserializePublicKey(pkData)
		if err != nil {
			log.Info("Cli Qid:%s SigMutilRawTransaction keypair.DeserializePublicKey error:%s", req.Qid, err)
			resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
			return
		}
		pubKeys = append(pubKeys, pk)
	}

	signer, err := req.GetAccount()
	if err != nil {
		log.Infof("Cli Qid:%s SigMutilRawTransaction GetAccount:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_ACCOUNT_UNLOCK
		return
	}
	err = cliutil.MultiSigTransaction(mutTx, uint16(rawReq.M), pubKeys, signer)
	if err != nil {
		log.Infof("Cli Qid:%s SigMutilRawTransaction MultiSigTransaction error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	tmpTx, err = mutTx.IntoImmutable()
	if err != nil {
		log.Infof("Cli Qid:%s SigMutilRawTransaction tx Serialize error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	sink := common.ZeroCopySink{}
	tmpTx.Serialization(&sink)
	resp.Result = &SigRawTransactionRsp{
		SignedTx: hex.EncodeToString(sink.Bytes()),
	}
}
