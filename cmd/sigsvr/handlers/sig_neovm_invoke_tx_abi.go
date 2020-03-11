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

	clisvrcom "github.com/ontio/ontology/cmd/sigsvr/common"
	cliutil "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	httpcom "github.com/ontio/ontology/http/base/common"
)

type SigNeoVMInvokeTxAbiReq struct {
	GasPrice    uint64          `json:"gas_price"`
	GasLimit    uint64          `json:"gas_limit"`
	Address     string          `json:"address"`
	Method      string          `json:"method"`
	Params      []string        `json:"params"`
	Payer       string          `json:"payer"`
	ContractAbi json.RawMessage `json:"contract_abi"`
}

type SigNeoVMInvokeTxAbiRsp struct {
	SignedTx string `json:"signed_tx"`
}

func SigNeoVMInvokeAbiTx(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	rawReq := &SigNeoVMInvokeTxAbiReq{}
	err := json.Unmarshal(req.Params, rawReq)
	if err != nil {
		log.Infof("SigNeoVMInvokeAbiTx json.Unmarshal SigNeoVMInvokeTxAbiReq:%s error:%s", req.Params, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	contractAbi, err := cliutil.NewNeovmContractAbi(rawReq.ContractAbi)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_ABI_UNMATCH
		resp.ErrorInfo = err.Error()
		return
	}
	funcAbi := contractAbi.GetFunc(rawReq.Method)
	if funcAbi == nil {
		resp.ErrorCode = clisvrcom.CLIERR_ABI_NOT_FOUND
		return
	}
	invokParams, err := cliutil.ParseNeovmFunc(rawReq.Params, funcAbi)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_ABI_UNMATCH
		resp.ErrorInfo = err.Error()
		return
	}
	contAddr, err := common.AddressFromHexString(rawReq.Address)
	if err != nil {
		log.Infof("Cli Qid:%s SigNeoVMInvokeAbiTx AddressParseFromBytes:%s error:%s", req.Qid, rawReq.Address, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	mutable, err := httpcom.NewNeovmInvokeTransaction(rawReq.GasPrice, rawReq.GasLimit, contAddr, invokParams)
	if err != nil {
		log.Infof("Cli Qid:%s SigNeoVMInvokeAbiTx InvokeNeoVMContractTx error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		return
	}
	if rawReq.Payer != "" {
		payerAddress, err := common.AddressFromBase58(rawReq.Payer)
		if err != nil {
			log.Infof("Cli Qid:%s SigNeoVMInvokeAbiTx AddressFromBase58 error:%s", req.Qid, err)
			resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
			return
		}
		mutable.Payer = payerAddress
	}
	signer, err := req.GetAccount()
	if err != nil {
		log.Infof("Cli Qid:%s SigNeoVMInvokeAbiTx GetAccount:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_ACCOUNT_UNLOCK
		return
	}
	err = cliutil.SignTransaction(signer, mutable)
	if err != nil {
		log.Infof("Cli Qid:%s SigNeoVMInvokeAbiTx SignTransaction error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}

	tx, err := mutable.IntoImmutable()
	if err != nil {
		log.Infof("Cli Qid:%s SigNeoVMInvokeAbiTx tx Serialize error:%s", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}
	resp.Result = &SigNeoVMInvokeTxAbiRsp{
		SignedTx: hex.EncodeToString(common.SerializeToBytes(tx)),
	}
}
