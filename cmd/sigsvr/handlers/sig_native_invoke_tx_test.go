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

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/cmd/abi"
	clisvrcom "github.com/ontio/ontology/cmd/sigsvr/common"
	nutils "github.com/ontio/ontology/smartcontract/service/native/utils"
)

func TestSigNativeInvokeTx(t *testing.T) {
	defAcc, err := testWallet.GetDefaultAccount(pwd)
	if err != nil {
		t.Errorf("GetDefaultAccount error:%s", err)
		return
	}
	acc1, err := clisvrcom.DefWalletStore.NewAccountData(keypair.PK_ECDSA, keypair.P256, signature.SHA256withECDSA, pwd)
	if err != nil {
		t.Errorf("wallet.NewAccount error:%s", err)
		return
	}
	clisvrcom.DefWalletStore.AddAccountData(acc1)
	invokeReq := &SigNativeInvokeTxReq{
		GasPrice: 0,
		GasLimit: 40000,
		Address:  nutils.OntContractAddress.ToHexString(),
		Method:   "transfer",
		Version:  0,
		Params: []interface{}{
			[]interface{}{
				[]interface{}{
					defAcc.Address.ToBase58(),
					acc1.Address,
					"10000000000",
				},
			},
		},
	}
	data, err := json.Marshal(invokeReq)
	if err != nil {
		t.Errorf("json.Marshal SigNativeInvokeTxReq error:%s", err)
		return
	}
	req := &clisvrcom.CliRpcRequest{
		Qid:     "t",
		Method:  "signativeinvoketx",
		Params:  data,
		Account: acc1.Address,
		Pwd:     string(pwd),
	}
	rsp := &clisvrcom.CliRpcResponse{}
	abiPath := "../../abi/native_abi_script"
	abi.DefAbiMgr.Init(abiPath)
	SigNativeInvokeTx(req, rsp)
	if rsp.ErrorCode != 0 {
		t.Errorf("SigNativeInvokeTx failed. ErrorCode:%d ErrorInfo:%s", rsp.ErrorCode, rsp.ErrorInfo)
		return
	}
}
