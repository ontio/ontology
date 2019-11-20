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
package ontid

import (
	"bytes"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func TestCaseController(t *testing.T) {
	testcase(t, CaseController)
}

func TestGroupController(t *testing.T) {
	testcase(t, CaseGroupController)
}

// Test case: register an ID controlled by another ID
func CaseController(t *testing.T, n *native.NativeService) {
	// 1. register the controller
	// create account and id
	a0 := account.NewAccount("")
	id0, err := account.GenerateID()
	if err != nil {
		t.Fatal("generate id0 error")
	}
	if err := regID(n, id0, a0); err != nil {
		t.Fatal("register ID error", err)
	}

	// 2. register the controlled ID
	id1, err := account.GenerateID()
	if err != nil {
		t.Fatal("generate id1 error")
	}
	if err := regControlledID(n, id1, id0, a0); err != nil {
		t.Fatal("register by controller error", err)
	}

	// 3. add attribute
	attr := attribute{
		[]byte("test key"),
		[]byte("test value"),
		[]byte("test type"),
	}
	sink := common.NewZeroCopySink(nil)
	// id
	sink.WriteVarBytes([]byte(id1))
	// attribute
	utils.EncodeVarUint(sink, 1)
	attr.Serialization(sink)
	// signer
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	if _, err = addAttributesByController(n); err != nil {
		t.Fatal("add attribute error", err)
	}

	// 4. verify signature
	sink.Reset()
	// id
	sink.WriteVarBytes([]byte(id1))
	// signer
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	res, err := verifyController(n)
	if err != nil || !bytes.Equal(res, utils.BYTE_TRUE) {
		t.Fatal("verify signature failed")
	}

	// 5. add key
	a1 := account.NewAccount("")
	sink.Reset()
	// id
	sink.WriteVarBytes([]byte(id1))
	// key
	pk := keypair.SerializePublicKey(a1.PubKey())
	sink.WriteVarBytes(pk)
	// signer
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	if _, err = addKeyByController(n); err != nil {
		t.Fatal("add key error")
	}

	// 6. remove controller
	sink.Reset()
	// id
	sink.WriteVarBytes([]byte(id1))
	// signing key index
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	// set signing address to a1
	n.Tx.SignedAddr = []common.Address{a1.Address}
	if _, err = removeController(n); err != nil {
		t.Fatal("remove controller error")
	}
}

func CaseGroupController(t *testing.T, n *native.NativeService) {
	//1. create and register controllers
	id0, err := account.GenerateID()
	if err != nil {
		t.Fatal("create id0 error")
	}
	id1, err := account.GenerateID()
	if err != nil {
		t.Fatal("create id1 error")
	}
	id2, err := account.GenerateID()
	if err != nil {
		t.Fatal("create id2 error")
	}
	a0 := account.NewAccount("")
	if err := regID(n, id0, a0); err != nil {
		t.Fatal("register id0 error")
	}
	a1 := account.NewAccount("")
	if err := regID(n, id1, a1); err != nil {
		t.Fatal("register id1 error")
	}
	a2 := account.NewAccount("")
	if err := regID(n, id2, a2); err != nil {
		t.Fatal("register id2 error")
	}
	// controller group
	g := Group{
		Threshold: 1,
		Members: []interface{}{
			[]byte(id0),
			&Group{
				Threshold: 1,
				Members: []interface{}{
					[]byte(id1),
					[]byte(id2),
				},
			},
		},
	}

	//2. generate and register the controlled id
	id, err := account.GenerateID()
	if err != nil {
		t.Fatal("generate controlled id error")
	}
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes([]byte(id))
	sink.WriteVarBytes(g.Serialize())
	// signers
	signers := []Signer{
		Signer{[]byte(id0), 1},
		Signer{[]byte(id1), 1},
		Signer{[]byte(id2), 1},
	}
	sink.WriteVarBytes(SerializeSigners(signers))
	n.Input = sink.Bytes()
	// set signing address
	n.Tx.SignedAddr = []common.Address{a0.Address, a1.Address, a2.Address}
	if _, err = regIdWithController(n); err != nil {
		t.Fatal("register controlled id error")
	}

	//3. verify signature
	sink.Reset()
	sink.WriteVarBytes([]byte(id))
	signers = []Signer{
		Signer{[]byte(id2), 1},
	}
	sink.WriteVarBytes(SerializeSigners(signers))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a2.Address}
	res, err := verifyController(n)
	if err != nil || !bytes.Equal(res, utils.BYTE_TRUE) {
		t.Fatal("verify signature failed")
	}
}

// Register id0 which is controlled by id1
func regControlledID(n *native.NativeService, id0, id1 string, a *account.Account) error {
	// make arguments
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes([]byte(id0))
	sink.WriteVarBytes([]byte(id1))
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	// set signing address
	n.Tx.SignedAddr = []common.Address{a.Address}
	// call
	_, err := regIdWithController(n)
	return err
}
