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
	"testing"

	clisvrcom "github.com/ontio/ontology/cmd/sigsvr/common"
)

var testNeovmAbi = `{
  "hash": "e827bf96529b5780ad0702757b8bad315e2bb8ce",
  "entrypoint": "Main",
  "functions": [
    {
      "name": "Main",
      "parameters": [
        {
          "name": "operation",
          "type": "String"
        },
        {
          "name": "args",
          "type": "Array"
        }
      ],
      "returntype": "Any"
    },
    {
      "name": "Add",
      "parameters": [
        {
          "name": "a",
          "type": "Integer"
        },
        {
          "name": "b",
          "type": "Integer"
        }
      ],
      "returntype": "Integer"
    }
  ],
  "events": []
}`

func TestSigNeoVMInvokeAbiTx(t *testing.T) {
	defAcc, err := testWallet.GetDefaultAccount(pwd)
	if err != nil {
		t.Errorf("GetDefaultAccount error:%s", err)
		return
	}

	invokeReq := &SigNeoVMInvokeTxAbiReq{
		GasPrice: 0,
		GasLimit: 0,
		Address:  "e827bf96529b5780ad0702757b8bad315e2bb8ce",
		Method:   "Add",
		Params: []string{
			"12",
			"13",
		},
		ContractAbi: []byte(testNeovmAbi),
	}
	data, err := json.Marshal(invokeReq)
	if err != nil {
		t.Errorf("json.Marshal SigNeoVMInvokeTxReq error:%s", err)
		return
	}
	req := &clisvrcom.CliRpcRequest{
		Qid:     "t",
		Method:  "signeovminvokeabitx",
		Params:  data,
		Account: defAcc.Address.ToBase58(),
		Pwd:     string(pwd),
	}
	rsp := &clisvrcom.CliRpcResponse{}
	SigNeoVMInvokeAbiTx(req, rsp)
	if rsp.ErrorCode != 0 {
		t.Errorf("SigNeoVMInvokeAbiTx failed. ErrorCode:%d ErrorInfo:%s", rsp.ErrorCode, rsp.ErrorInfo)
		return
	}
}
