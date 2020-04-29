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

func TestRecovery(t *testing.T) {
	testcase(t, CaseRecovery)
}

func CaseRecovery(t *testing.T, n *native.NativeService) {
	id0, _ := account.GenerateID()
	a0 := account.NewAccount("")
	id1, _ := account.GenerateID()
	a1 := account.NewAccount("")
	id2, _ := account.GenerateID()
	a2 := account.NewAccount("")
	id3, _ := account.GenerateID()
	a3 := account.NewAccount("")

	if regID(n, id0, a0) != nil {
		t.Fatal("register id0 error")
	}

	g := &Group{
		Threshold: 2,
		Members:   []interface{}{[]byte(id1), []byte(id2)},
	}

	// 1. set unregistered id as recovery, should fail
	if err := setRec(n, id0, g, a0.Address); err == nil {
		t.Error("unregistered id is setted as recovery")
	}

	// 2. register id
	if regID(n, id1, a1) != nil {
		t.Fatal("register id1 error")
	}
	if regID(n, id2, a2) != nil {
		t.Fatal("register id2 error")
	}
	if regID(n, id3, a3) != nil {
		t.Fatal("register id3 error")
	}

	// 3. set recovery without valid signature, should fail
	if err := setRec(n, id0, g, common.ADDRESS_EMPTY); err == nil {
		t.Error("recovery setted without valid signature")
	}

	// 4. set id1 and id2 as id0's recovery id
	if err := setRec(n, id0, g, a0.Address); err != nil {
		t.Fatal(err)
	}

	// 5. set recovery again, should fail
	g = &Group{
		Threshold: 2,
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
	if err := setRec(n, id0, g, a0.Address); err == nil {
		t.Error("should not set recovery twice")
	}

	// 6. update recovery by invalid recovery, should fail
	s := []Signer{
		{[]byte(id0), 1},
		{[]byte(id1), 1},
	}
	addr := []common.Address{a0.Address, a1.Address}
	if err := updateRec(n, id0, g, s, addr); err == nil {
		t.Error("recovery updated by invalid recovery")
	}

	// 7. update without enough signature, should fail
	s[0].Id = []byte(id2)
	addr[0] = a2.Address
	if err := updateRec(n, id0, g, s, addr[1:]); err == nil {
		t.Error("recovery updated without enough signature")
	}

	// 8. update without enough signers, should fail
	if err := updateRec(n, id0, g, s[1:], addr); err == nil {
		t.Error("recovery updated without enough signers")
	}

	// 9. update recovery
	if err := updateRec(n, id0, g, s, addr); err != nil {
		t.Fatal(err)
	}

	// 10. add key without enough signer, should fail
	a4 := account.NewAccount("")
	pk := keypair.SerializePublicKey(a4.PubKey())
	if err := addKeyByRec(n, id0, pk, s[1:], addr); err == nil {
		t.Error("add key by recovery without enough signer")
	}

	// 11. add key without enough signature, should fail
	if err := addKeyByRec(n, id0, pk, s, addr[1:]); err == nil {
		t.Error("add key by recovery without enough signature")
	}

	// 12. add key by recovery
	if err := addKeyByRec(n, id0, pk, s, addr); err != nil {
		t.Fatal(err)
	}

	// 13. verify added key
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id0)
	utils.EncodeVarUint(sink, 2)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a4.Address}
	res, err := verifySignature(n)
	if err != nil || !bytes.Equal(res, utils.BYTE_TRUE) {
		t.Fatal("verifying added key failed")
	}

	// 14. remove key without enough signer, should fail
	if err := rmKeyByRec(n, id0, 2, s[1:], addr); err == nil {
		t.Error("key removed by recovery without enought signer")
	}

	// 15. remove key without enough signature, should fail
	if err := rmKeyByRec(n, id0, 2, s, addr[1:]); err == nil {
		t.Error("key removed by recovery without enought signature")
	}

	// 16. remove key by recovery
	if err := rmKeyByRec(n, id0, 2, s, addr); err != nil {
		t.Error(err)
	}

	// 17. check removed key
	sink.Reset()
	sink.WriteString(id0)
	utils.EncodeVarUint(sink, 2)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a4.Address}
	res, err = verifySignature(n)
	if err == nil && bytes.Equal(res, utils.BYTE_TRUE) {
		t.Error("removed key passed verification")
	}

	// 18. add the removed key again, should fail
	if err := addKeyByRec(n, id0, pk, s, addr); err == nil {
		t.Error("removed key should not be added again by recovery")
	}
}

func setRec(n *native.NativeService, id string, g *Group, addr common.Address) error {
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	sink.WriteVarBytes(g.Serialize())
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{addr}
	_, err := setRecovery(n)
	return err
}

func updateRec(n *native.NativeService, id string, g *Group, s []Signer, addr []common.Address) error {
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	sink.WriteVarBytes(g.Serialize())
	sink.WriteVarBytes(SerializeSigners(s))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = addr
	_, err := updateRecovery(n)
	return err
}

func addKeyByRec(n *native.NativeService, id string, pk []byte, s []Signer, addr []common.Address) error {
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	sink.WriteVarBytes(pk)
	sink.WriteVarBytes(SerializeSigners(s))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = addr
	_, err := addKeyByRecovery(n)
	return err
}

func rmKeyByRec(n *native.NativeService, id string, index uint64, s []Signer, addr []common.Address) error {
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	utils.EncodeVarUint(sink, index)
	sink.WriteVarBytes(SerializeSigners(s))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = addr
	_, err := removeKeyByRecovery(n)
	return err
}
