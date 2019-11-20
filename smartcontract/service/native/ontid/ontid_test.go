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
	"github.com/ontio/ontology/smartcontract/service/native/testsuite"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func testcase(t *testing.T, f func(t *testing.T, n *native.NativeService)) {
	testsuite.InvokeNativeContract(t, utils.OntIDContractAddress,
		func(n *native.NativeService) ([]byte, error) {
			f(t, n)
			return nil, nil
		},
	)
}

func TestReg(t *testing.T) {
	testcase(t, CaseRegID)
}

func TestOwner(t *testing.T) {
	testcase(t, CaseOwner)
}

func TestRecovery(t *testing.T) {
	testcase(t, CaseRecovery)
}

// Register id with account acc
func regID(n *native.NativeService, id string, a *account.Account) error {
	// make arguments
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes([]byte(id))
	pk := keypair.SerializePublicKey(a.PubKey())
	sink.WriteVarBytes(pk)
	n.Input = sink.Bytes()
	// set signing address
	n.Tx.SignedAddr = []common.Address{a.Address}
	// call
	_, err := regIdWithPublicKey(n)
	return err
}

func CaseRegID(t *testing.T, n *native.NativeService) {
	id, err := account.GenerateID()
	if err != nil {
		t.Fatal(err)
	}
	a := account.NewAccount("")

	// 1. register invalid id, should fail
	if err := regID(n, "did:ont:abcd1234", a); err == nil {
		t.Error("invalid id registered")
	}

	// 2. register without valid signature, should fail
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	sink.WriteVarBytes(keypair.SerializePublicKey(a.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{}
	if _, err := regIdWithPublicKey(n); err == nil {
		t.Error("id registered without signature")
	}

	// 3. register with invalid key, should fail
	sink.Reset()
	sink.WriteString(id)
	sink.WriteVarBytes([]byte("invalid public key"))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a.Address}
	if _, err := regIdWithPublicKey(n); err == nil {
		t.Error("id registered with invalid key")
	}

	// 4. register id
	if err := regID(n, id, a); err != nil {
		t.Fatal(err)
	}

	// 5. register again, should fail
	if err := regID(n, id, a); err == nil {
		t.Error("id registered twice")
	}

	// 6. revoke id
	sink.Reset()
	sink.WriteString(id)
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	if _, err := revokeID(n); err != nil {
		t.Fatal(err)
	}

	// 7. register again, should fail
	if err := regID(n, id, a); err == nil {
		t.Error("revoked id should not be registered again")
	}
}

func CaseOwner(t *testing.T, n *native.NativeService) {
	// 1. register ID
	id, err := account.GenerateID()
	if err != nil {
		t.Fatal("generate ID error")
	}
	a0 := account.NewAccount("")
	if err := regID(n, id, a0); err != nil {
		t.Fatal("register ID error", err)
	}

	// 2. add new key
	a1 := account.NewAccount("")
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	sink.WriteVarBytes(keypair.SerializePublicKey(a1.PubKey()))
	sink.WriteVarBytes(keypair.SerializePublicKey(a0.PubKey()))
	n.Input = sink.Bytes()
	// tx signer remains the same
	if _, err = addKey(n); err != nil {
		t.Fatal("add key error", err)
	}

	// 3. verify new key
	sink.Reset()
	sink.WriteString(id)
	utils.EncodeVarUint(sink, 2)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a1.Address}
	res, err := verifySignature(n)
	if err != nil || !bytes.Equal(res, utils.BYTE_TRUE) {
		t.Fatal("verify signature failed")
	}

	// 4. remove key
	sink.Reset()
	sink.WriteString(id)
	sink.WriteVarBytes(keypair.SerializePublicKey(a0.PubKey()))
	sink.WriteVarBytes(keypair.SerializePublicKey(a1.PubKey()))
	n.Input = sink.Bytes()
	if _, err = removeKey(n); err != nil {
		t.Fatal("remove key error")
	}

	// 5. check removed key
	sink.Reset()
	sink.WriteString(id)
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	res, err = verifySignature(n)
	if err == nil && bytes.Equal(res, utils.BYTE_TRUE) {
		t.Fatal("signature passed unexpectedly")
	}
}

func CaseRecovery(t *testing.T, n *native.NativeService) {
	//1. generate and register id
	id0, err := account.GenerateID()
	if err != nil {
		t.Fatal("generate id0 error")
	}
	a0 := account.NewAccount("")
	if regID(n, id0, a0) != nil {
		t.Fatal("register id0 error")
	}
	id1, err := account.GenerateID()
	if err != nil {
		t.Fatal("generate id1 error")
	}
	a1 := account.NewAccount("")
	if regID(n, id1, a1) != nil {
		t.Fatal("register id1 error")
	}
	id2, err := account.GenerateID()
	if err != nil {
		t.Fatal("generate id2 error")
	}
	a2 := account.NewAccount("")
	if regID(n, id2, a2) != nil {
		t.Fatal("register id2 error")
	}
	id3, err := account.GenerateID()
	if err != nil {
		t.Fatal("generate id3 error")
	}
	a3 := account.NewAccount("")
	if regID(n, id3, a3) != nil {
		t.Fatal("register id3 error")
	}
	//2. set id1 and id2 as id0's recovery id
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes([]byte(id0))
	g := Group{
		Threshold: 1,
		Members:   []interface{}{[]byte(id1), []byte(id2)},
	}
	sink.WriteVarBytes(g.Serialize())
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	if _, err = setRecovery(n); err != nil {
		t.Fatal("add recovery error", err)
	}
	//3. update recovery
	g = Group{
		Threshold: 1,
		Members: []interface{}{
			[]byte(id1),
			&Group{
				Threshold: 1,
				Members: []interface{}{
					[]byte(id2),
					[]byte(id3),
				},
			},
		},
	}
	s := []Signer{Signer{[]byte(id1), 1}}
	sink.Reset()
	sink.WriteVarBytes([]byte(id0))
	sink.WriteVarBytes(g.Serialize())
	sink.WriteVarBytes(SerializeSigners(s))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a1.Address}
	if _, err = updateRecovery(n); err != nil {
		t.Fatal("update recovery error", err)
	}
	//4. add new key by recovery id
	a4 := account.NewAccount("")
	sink.Reset()
	sink.WriteVarBytes([]byte(id0))
	pk := keypair.SerializePublicKey(a4.PubKey())
	sink.WriteVarBytes(pk)
	s = []Signer{Signer{[]byte(id3), 1}}
	sink.WriteVarBytes(SerializeSigners(s))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a3.Address}
	if _, err = addKeyByRecovery(n); err != nil {
		t.Fatal("add key by recovery error", err)
	}
	//5. remove key
	sink.Reset()
	sink.WriteVarBytes([]byte(id0))
	utils.EncodeVarUint(sink, 1)
	s = []Signer{Signer{[]byte(id2), 1}}
	sink.WriteVarBytes(SerializeSigners(s))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a2.Address}
	if _, err = removeKeyByRecovery(n); err != nil {
		t.Fatal("remove key by recovery error", err)
	}
	//6. verify signature
	sink.Reset()
	sink.WriteVarBytes([]byte(id0))
	utils.EncodeVarUint(sink, 2)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a4.Address}
	res, err := verifySignature(n)
	if err != nil || !bytes.Equal(res, utils.BYTE_TRUE) {
		t.Fatal("verify signature failed")
	}
	//7. verify signature generated by removed key
	// this should fail
	sink.Reset()
	sink.WriteVarBytes([]byte(id0))
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	res, err = verifySignature(n)
	if err == nil && bytes.Equal(res, utils.BYTE_TRUE) {
		t.Fatal("signature passed unexpectedly")
	}
}
