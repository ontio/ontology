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
	"encoding/json"
	"github.com/ontio/ontology/account"
	clisvrcom "github.com/ontio/ontology/cmd/server/common"
	"testing"
)

func TestSigTransferTransaction(t *testing.T) {
	acc := account.NewAccount("")
	defAcc := clisvrcom.DefAccount
	sigReq := &SigTransferTransactionReq{
		GasLimit: 0,
		GasPrice: 0,
		Asset:    "ont",
		From:     defAcc.Address.ToBase58(),
		To:       acc.Address.ToBase58(),
		Amount:   10,
	}
	data, err := json.Marshal(sigReq)
	if err != nil {
		t.Errorf("json.Marshal SigTransferTransactionReq error:%s", err)
	}
	req := &clisvrcom.CliRpcRequest{
		Qid:    "t",
		Method: "sigtransfertx",
		Params: data,
	}
	rsp := &clisvrcom.CliRpcResponse{}
	SigTransferTransaction(req, rsp)
	if rsp.ErrorCode != 0 {
		t.Errorf("SigTransferTransaction failed. ErrorCode:%d", rsp.ErrorCode)
		return
	}
}
