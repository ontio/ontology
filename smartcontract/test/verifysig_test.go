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

package test

import (
	"fmt"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract"
	svm "github.com/ontio/ontology/smartcontract/service/neovm"
	vtypes "github.com/ontio/ontology/vm/neovm/types"
)

func TestVerifySig(t *testing.T) {
	/**
	# the code of source python.
	Cversion = '2.0.0'
	from ontology.interop.Ontology.Runtime import VerifyMutiSig

	def Main(data, pks_list, m, sig_list):
	    return VerifyMutiSig(data, pks_list, m, sig_list)
	**/
	code := `52c56b05322e302e306a00527ac46c59c56b6a00527ac46a51527ac46a52527ac46a53527ac46a54527ac46a54c36a53c36a52c36a51c3681e4f6e746f6c6f67792e52756e74696d652e5665726966794d7574695369676c75660111c56b6a00527ac46a51527ac46a51c300947600a0640c00c16a52527ac4620e007562030000c56a52527ac46a52c3c0517d9c7c75641c00006a53527ac46a52c300c36a54527ac4516a55527ac4625c006a52c3c0527d9c7c756421006a52c300c36a53527ac46a52c351c36a54527ac4516a55527ac4616232006a52c3c0537d9c7c756424006a52c300c36a53527ac46a52c351c36a54527ac46a52c352c36a55527ac462050000f100c176c96a56527ac46a53c36a57527ac46a57c36a54c37d9f7c756419006a56c36a57c3c86a57c36a55c3936a57527ac462e0ff6a56c36c7566`

	data_pre := []byte{1, 2, 3}
	data, _ := vtypes.VmValueFromBytes(data_pre)
	pubkeys := vtypes.NewArrayValue()
	sigs := vtypes.NewArrayValue()
	accs := make([]*account.Account, 0)
	const N = 4
	for i := 0; i < N; i++ {
		accs = append(accs, account.NewAccount(""))
	}

	for _, acc := range accs {
		sig, _ := signature.Sign(acc, data_pre)
		key0, _ := vtypes.VmValueFromBytes(sig)
		_ = sigs.Append(key0)

		pk := keypair.SerializePublicKey(acc.PublicKey)
		key1, _ := vtypes.VmValueFromBytes(pk)
		_ = pubkeys.Append(key1)
	}

	hex, err := common.HexToBytes(code)

	if err != nil {
		t.Fatal("hex to byte error:", err)
	}

	config := &smartcontract.Config{
		Time:   10,
		Height: 10,
		Tx:     nil,
	}
	sc := smartcontract.SmartContract{
		Config: config,
		Gas:    100000,
	}
	engine, err := sc.NewExecuteEngine(hex, types.InvokeNeo)

	if err != nil {
		t.Fatal("hex to byte error:", err)
	}

	var service *svm.NeoVmService
	service = engine.(*svm.NeoVmService)
	e := service.Engine
	_ = e.EvalStack.Push(vtypes.VmValueFromArrayVal(sigs))
	_ = e.EvalStack.PushInt64(N)
	_ = e.EvalStack.Push(vtypes.VmValueFromArrayVal(pubkeys))
	_ = e.EvalStack.Push(data)

	_, err = engine.Invoke()
	if err != nil {
		t.Fatal("multisignature inovke err:", err)
	}

	arr, err := e.EvalStack.PopAsBool()
	if err != nil {
		t.Fatal("multisignature PopBoolean err:", err)
	}

	if !arr {
		t.Fatal("multisignature failed")
	}

	fmt.Printf("multisignature passed\n")

}
