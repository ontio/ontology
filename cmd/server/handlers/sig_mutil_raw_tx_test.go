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
	"github.com/ontio/ontology/account"
	clisvrcom "github.com/ontio/ontology/cmd/server/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/core/types"
	"testing"
)

func TestSigMutilRawTransaction(t *testing.T) {
	acc1 := account.NewAccount("")
	acc2 := account.NewAccount("")
	pubKeys := []keypair.PublicKey{acc1.PublicKey, acc2.PublicKey}
	m := 2
	fromAddr, err := types.AddressFromMultiPubKeys(pubKeys, m)
	if err != nil {
		t.Errorf("TestSigMutilRawTransaction AddressFromMultiPubKeys error:%s", err)
		return
	}
	defAcc := clisvrcom.DefAccount
	tx, err := utils.TransferTx(0, 0, "ont", fromAddr.ToBase58(), defAcc.Address.ToBase58(), 10)
	if err != nil {
		t.Errorf("TransferTx error:%s", err)
		return
	}
	buf := bytes.NewBuffer(nil)
	err = tx.Serialize(buf)
	if err != nil {
		t.Errorf("tx.Serialize error:%s", err)
		return
	}

	rawReq := &SigMutilRawTransactionReq{
		RawTx:   hex.EncodeToString(buf.Bytes()),
		M:       m,
		PubKeys: []string{hex.EncodeToString(keypair.SerializePublicKey(acc1.PublicKey)), hex.EncodeToString(keypair.SerializePublicKey(acc2.PublicKey))},
	}
	data, err := json.Marshal(rawReq)
	if err != nil {
		t.Errorf("json.Marshal SigRawTransactionReq error:%s", err)
		return
	}
	req := &clisvrcom.CliRpcRequest{
		Qid:    "t",
		Method: "sigmutilrawtx",
		Params: data,
	}
	resp := &clisvrcom.CliRpcResponse{}
	clisvrcom.DefAccount = acc1
	SigMutilRawTransaction(req, resp)
	if resp.ErrorCode != clisvrcom.CLIERR_OK {
		t.Errorf("SigMutilRawTransaction failed,ErrorCode:%d ErrorString:%s", resp.ErrorCode, resp.ErrorInfo)
		return
	}

	clisvrcom.DefAccount = acc2
	SigMutilRawTransaction(req, resp)
	if resp.ErrorCode != clisvrcom.CLIERR_OK {
		t.Errorf("SigMutilRawTransaction failed,ErrorCode:%d ErrorString:%s", resp.ErrorCode, resp.ErrorInfo)
		return
	}
}
