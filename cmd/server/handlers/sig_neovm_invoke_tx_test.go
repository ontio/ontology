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
	clisvrcom "github.com/ontio/ontology/cmd/server/common"
	"github.com/ontio/ontology/cmd/utils"
	"testing"
)

func TestSigNeoVMInvokeTx(t *testing.T) {
	invokeReq := &SigNeoVMInvokeTxReq{
		GasPrice: 0,
		GasLimit: 0,
		Address:  "TA6PCtD9qEvN6Rk7i4EGY6u1cWQBWZTr5A",
		Version:  0,
		Params: []interface{}{
			&utils.NeoVMInvokeParam{
				Type:  "string",
				Value: "foo",
			},
			&utils.NeoVMInvokeParam{
				Type: "array",
				Value: []interface{}{
					&utils.NeoVMInvokeParam{
						Type:  "int",
						Value: "0",
					},
					&utils.NeoVMInvokeParam{
						Type:  "bool",
						Value: "true",
					},
				},
			},
		},
	}
	data, err := json.Marshal(invokeReq)
	if err != nil {
		t.Errorf("json.Marshal SigNeoVMInvokeTxReq error:%s", err)
		return
	}
	req := &clisvrcom.CliRpcRequest{
		Qid:    "t",
		Method: "signeovminvoketx",
		Params: data,
	}
	rsp := &clisvrcom.CliRpcResponse{}
	SigNeoVMInvokeTx(req, rsp)
	if rsp.ErrorCode != 0 {
		t.Errorf("SigNeoVMInvokeTx failed. ErrorCode:%d ErrorInfo:%s", rsp.ErrorCode, rsp.ErrorInfo)
		return
	}
}
