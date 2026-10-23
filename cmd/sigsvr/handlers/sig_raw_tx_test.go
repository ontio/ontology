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
	"os"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	clisvrcom "github.com/ontio/ontology/cmd/sigsvr/common"
	"github.com/ontio/ontology/cmd/sigsvr/store"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"
)

var (
	pwd                 = []byte("123456")
	testWalletPath      = "wallet.tmp.dat"
	testWalletStorePath = "wallet_data_tmp"
	testWallet          account.Client
)

func TestMain(m *testing.M) {
	log.InitLog(0, os.Stdout)
	var err error
	testWallet, err = account.Open(testWalletPath)
	if err != nil {
		log.Errorf("account.Open :%s error:%s", testWalletPath)
		return
	}

	_, err = testWallet.NewAccount("", keypair.PK_ECDSA, keypair.P256, signature.SHA256withECDSA, pwd)
	if err != nil {
		log.Errorf("wallet.NewAccount error:%s", err)
		return
	}

	clisvrcom.DefWalletStore, err = store.NewWalletStore(testWalletStorePath)
	if err != nil {
		log.Errorf("NewWalletStore error:%s", err)
		return
	}
	_, err = clisvrcom.DefWalletStore.AddAccountData(testWallet.GetWalletData().Accounts[0])
	if err != nil {
		log.Errorf("AddAccountData error:%s", err)
		return
	}
	m.Run()
	os.RemoveAll("./ActorLog")
	os.RemoveAll("./Log")
	os.RemoveAll(testWalletPath)
	os.RemoveAll(testWalletStorePath)
}

func TestSigRawTx(t *testing.T) {
	acc := account.NewAccount("")
	defAcc, err := testWallet.GetDefaultAccount(pwd)
	if err != nil {
		t.Errorf("GetDefaultAccount error:%s", err)
		return
	}
	mutable, err := utils.TransferTx(0, 0, "ont", defAcc.Address.ToBase58(), acc.Address.ToBase58(), 10)
	if err != nil {
		t.Errorf("TransferTx error:%s", err)
		return
	}
	tx, err := mutable.IntoImmutable()
	assert.Nil(t, err)
	sink := common.ZeroCopySink{}
	tx.Serialization(&sink)
	rawReq := &SigRawTransactionReq{
		RawTx: hex.EncodeToString(sink.Bytes()),
	}
	data, err := json.Marshal(rawReq)
	if err != nil {
		t.Errorf("json.Marshal SigRawTransactionReq error:%s", err)
		return
	}
	req := &clisvrcom.CliRpcRequest{
		Qid:     "t",
		Method:  "sigrawtx",
		Params:  data,
		Account: defAcc.Address.ToBase58(),
		Pwd:     string(pwd),
	}
	resp := &clisvrcom.CliRpcResponse{}
	SigRawTransaction(req, resp)
	if resp.ErrorCode != 0 {
		t.Errorf("SigRawTransaction failed. ErrorCode:%d", resp.ErrorCode)
		return
	}
}

func TestSigRawTxDeclaresSignerForDifferentPayer(t *testing.T) {
	payer := account.NewAccount("")
	signer, err := testWallet.GetDefaultAccount(pwd)
	if err != nil {
		t.Fatalf("GetDefaultAccount error:%s", err)
	}
	mutable, err := utils.TransferTx(0, 0, "ont", signer.Address.ToBase58(), payer.Address.ToBase58(), 10)
	if err != nil {
		t.Fatalf("TransferTx error:%s", err)
	}
	mutable.Payer = payer.Address
	unsigned, err := mutable.IntoImmutable()
	if err != nil {
		t.Fatalf("IntoImmutable error:%s", err)
	}

	data, err := json.Marshal(&SigRawTransactionReq{
		RawTx: hex.EncodeToString(common.SerializeToBytes(unsigned)),
	})
	if err != nil {
		t.Fatalf("json.Marshal error:%s", err)
	}
	resp := &clisvrcom.CliRpcResponse{}
	SigRawTransaction(&clisvrcom.CliRpcRequest{
		Qid:     "t",
		Method:  "sigrawtx",
		Params:  data,
		Account: signer.Address.ToBase58(),
		Pwd:     string(pwd),
	}, resp)
	if resp.ErrorCode != clisvrcom.CLIERR_OK {
		t.Fatalf("SigRawTransaction failed. ErrorCode:%d ErrorInfo:%s", resp.ErrorCode, resp.ErrorInfo)
	}

	result, ok := resp.Result.(*SigRawTransactionRsp)
	if !ok {
		t.Fatalf("unexpected response type %T", resp.Result)
	}
	raw, err := hex.DecodeString(result.SignedTx)
	if err != nil {
		t.Fatalf("hex.DecodeString error:%s", err)
	}
	signed, err := types.TransactionFromRawBytes(raw)
	if err != nil {
		t.Fatalf("TransactionFromRawBytes error:%s", err)
	}
	assert.Equal(t, payer.Address, signed.Payer)
	assert.Equal(t, []common.Address{signer.Address}, signed.ExtraSigners)
	assert.Len(t, signed.Sigs, 1)
}
