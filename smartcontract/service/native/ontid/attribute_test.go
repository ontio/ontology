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
	"fmt"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func TestAttribute(t *testing.T) {
	testcase(t, CaseAttribute)
}

func CaseAttribute(t *testing.T, n *native.NativeService) {
	// 1. register id
	a := account.NewAccount("")
	id, err := account.GenerateID()
	if err != nil {
		t.Fatal("generate id error")
	}
	if err := regID(n, id, a); err != nil {
		t.Fatal(err)
	}

	attr := attribute{
		key:       []byte("test key"),
		valueType: []byte("test type"),
		value:     []byte("test value"),
	}

	// 2. add attribute by invalid owner
	a1 := account.NewAccount("")
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	utils.EncodeVarUint(sink, 1)
	attr.Serialization(sink)
	sink.WriteVarBytes(keypair.SerializePublicKey(a1.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a1.Address}
	if _, err := addAttributes(n); err == nil {
		t.Error("attribute added by invalid owner")
	}

	// 3. add invalid attribute, should fail
	sink.Reset()
	sink.WriteString(id)
	utils.EncodeVarUint(sink, 1)
	sink.WriteVarBytes([]byte("invalid attribute"))
	sink.WriteVarBytes(keypair.SerializePublicKey(a.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a.Address}
	if _, err := addAttributes(n); err == nil {
		t.Error("invalid attribute added")
	}

	// 4. add attribute
	sink.Reset()
	sink.WriteString(id)
	utils.EncodeVarUint(sink, 1)
	attr.Serialization(sink)
	sink.WriteVarBytes(keypair.SerializePublicKey(a.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a.Address}
	if _, err := addAttributes(n); err != nil {
		t.Fatal(err)
	}

	// 5. check attribute
	if err := checkAttribute(n, id, []attribute{attr}); err != nil {
		t.Fatal(err)
	}

	// 6. remove attribute by invalid owner
	sink.Reset()
	sink.WriteString(id)
	sink.WriteVarBytes(attr.key)
	sink.WriteVarBytes(keypair.SerializePublicKey(a1.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a1.Address}
	if _, err := removeAttribute(n); err == nil {
		t.Error("attribute removed by invalid owner")
	}

	// 7. remove nonexistent attribute
	sink.Reset()
	sink.WriteString(id)
	sink.WriteVarBytes([]byte("invalid attribute key"))
	sink.WriteVarBytes(keypair.SerializePublicKey(a.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a.Address}
	if _, err := removeAttribute(n); err == nil {
		t.Error("attribute removed by invalid owner")
	}

	// 8. remove attribute
	sink.Reset()
	sink.WriteString(id)
	sink.WriteVarBytes(attr.key)
	sink.WriteVarBytes(keypair.SerializePublicKey(a.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a.Address}
	if _, err := removeAttribute(n); err != nil {
		t.Fatal(err)
	}

	// 9. check attribute
	if err := checkAttribute(n, id, []attribute{}); err != nil {
		t.Error("check attribute error,", err)
	}

	// 10. attribute size limit
	attr = attribute{
		key:       make([]byte, MAX_KEY_SIZE+1),
		valueType: []byte("test type"),
		value:     []byte("test value"),
	}
	sink.Reset()
	sink.WriteString(id)
	utils.EncodeVarUint(sink, 1)
	attr.Serialization(sink)
	sink.WriteVarBytes(keypair.SerializePublicKey(a.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a.Address}
	if _, err := addAttributes(n); err == nil {
		t.Error("attribute key size limit error")
	}
	attr = attribute{
		key:       []byte("test key"),
		valueType: []byte("test type"),
		value:     make([]byte, MAX_VALUE_SIZE+1),
	}
	sink.Reset()
	sink.WriteString(id)
	utils.EncodeVarUint(sink, 1)
	attr.Serialization(sink)
	sink.WriteVarBytes(keypair.SerializePublicKey(a.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a.Address}
	if _, err := addAttributes(n); err == nil {
		t.Error("attribute value size limit error")
	}
	attr = attribute{
		key:       []byte("test key"),
		valueType: make([]byte, MAX_TYPE_SIZE+1),
		value:     []byte("test value"),
	}
	sink.Reset()
	sink.WriteString(id)
	utils.EncodeVarUint(sink, 1)
	attr.Serialization(sink)
	sink.WriteVarBytes(keypair.SerializePublicKey(a.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a.Address}
	if _, err := addAttributes(n); err == nil {
		t.Error("attribute type size limit error")
	}
}

func checkAttribute(n *native.NativeService, id string, attributes []attribute) error {
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	n.Input = sink.Bytes()
	res, err := GetAttributes(n)
	if err != nil {
		return err
	}

	total := 0
	for _, a := range attributes {
		sink.Reset()
		a.Serialization(sink)
		b := sink.Bytes()
		if bytes.Index(res, b) == -1 {
			return fmt.Errorf("attribute %s not found", string(a.key))
		}
		total += len(b)
	}

	if len(res) != total {
		return fmt.Errorf("unmatched attribute number")
	}

	return nil
}
