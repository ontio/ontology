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
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/testsuite"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/stretchr/testify/assert"
)

func init() {
	Init()
}

func testcase(t *testing.T, f func(t *testing.T, n *native.NativeService)) {
	testsuite.InvokeNativeContract(t, utils.OntIDContractAddress,
		func(n *native.NativeService) ([]byte, error) {
			f(t, n)
			return nil, nil
		},
	)
}

func TestGetDocument(t *testing.T) {
	testcase(t, CaseGetDocument)
}

func TestReg(t *testing.T) {
	testcase(t, CaseRegID)
}

func TestOwner(t *testing.T) {
	testcase(t, CaseOwner)
}

func TestOwnerSize(t *testing.T) {
	testcase(t, CaseOwnerSize)
}

func TestNewApiFork(t *testing.T) {
	testcase(t, CaseBeforeNewApiFork)
	testcase(t, CaseAfterNewApiFork)
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

	// 5. get DDO
	sink.Reset()
	sink.WriteString(id)
	n.Input = sink.Bytes()
	_, err = GetDDO(n)
	if err != nil {
		t.Error(err)
	}

	// 6. register again, should fail
	if err := regID(n, id, a); err == nil {
		t.Error("id registered twice")
	}

	// 7. revoke with invalid key, should fail
	sink.Reset()
	sink.WriteString(id)
	utils.EncodeVarUint(sink, 2)
	n.Input = sink.Bytes()
	if _, err := revokeID(n); err == nil {
		t.Error("revoked by invalid key")
	}

	// 8. revoke without valid signature, should fail
	sink.Reset()
	sink.WriteString(id)
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{common.ADDRESS_EMPTY}
	if _, err := revokeID(n); err == nil {
		t.Error("revoked without valid signature")
	}

	// 9. revoke id
	sink.Reset()
	sink.WriteString(id)
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a.Address}
	if _, err := revokeID(n); err != nil {
		t.Fatal(err)
	}

	// 10. register again, should fail
	if err := regID(n, id, a); err == nil {
		t.Error("revoked id should not be registered again")
	}

	// 11. get DDO of the revoked id
	sink.Reset()
	sink.WriteString(id)
	n.Input = sink.Bytes()
	_, err = GetDDO(n)
	if err == nil {
		t.Error("get DDO of the revoked id should fail")
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

	// 2. add key without valid signature, should fail
	a1 := account.NewAccount("")
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id)
	sink.WriteVarBytes(keypair.SerializePublicKey(a1.PubKey()))
	sink.WriteVarBytes(keypair.SerializePublicKey(a0.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{common.ADDRESS_EMPTY}
	if _, err = addKey(n); err == nil {
		t.Error("key added without valid signature")
	}

	// 3. add key by invalid owner, should fail
	a2 := account.NewAccount("")
	sink.Reset()
	sink.WriteString(id)
	sink.WriteVarBytes(keypair.SerializePublicKey(a1.PubKey()))
	sink.WriteVarBytes(keypair.SerializePublicKey(a2.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a2.Address}
	if _, err = addKey(n); err == nil {
		t.Error("key added by invalid owner")
	}

	// 4. add invalid key, should fail
	sink.Reset()
	sink.WriteString(id)
	sink.WriteVarBytes([]byte("test invalid key"))
	sink.WriteVarBytes(keypair.SerializePublicKey(a0.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	if _, err = addKey(n); err == nil {
		t.Error("invalid key added")
	}

	// 5. add key
	sink.Reset()
	sink.WriteString(id)
	sink.WriteVarBytes(keypair.SerializePublicKey(a1.PubKey()))
	sink.WriteVarBytes(keypair.SerializePublicKey(a0.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	if _, err = addKey(n); err != nil {
		t.Fatal(err)
	}

	// 6. verify new key
	sink.Reset()
	sink.WriteString(id)
	utils.EncodeVarUint(sink, 2)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a1.Address}
	res, err := verifySignature(n)
	if err != nil || !bytes.Equal(res, utils.BYTE_TRUE) {
		t.Fatal("verify the added key failed")
	}

	// 7. add key again, should fail
	sink.Reset()
	sink.WriteString(id)
	sink.WriteVarBytes(keypair.SerializePublicKey(a1.PubKey()))
	sink.WriteVarBytes(keypair.SerializePublicKey(a0.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	if _, err = addKey(n); err == nil {
		t.Fatal("should not add the same key twice")
	}

	// 8. remove key without valid signature, should fail
	sink.Reset()
	sink.WriteString(id)
	sink.WriteVarBytes(keypair.SerializePublicKey(a0.PubKey()))
	sink.WriteVarBytes(keypair.SerializePublicKey(a1.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a2.Address}
	if _, err = removeKey(n); err == nil {
		t.Error("key removed without valid signature")
	}

	// 9. remove key by invalid owner, should fail
	sink.Reset()
	sink.WriteString(id)
	sink.WriteVarBytes(keypair.SerializePublicKey(a0.PubKey()))
	sink.WriteVarBytes(keypair.SerializePublicKey(a2.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a2.Address}
	if _, err = removeKey(n); err == nil {
		t.Error("key removed by invalid owner")
	}

	// 10. remove invalid key, should fail
	sink.Reset()
	sink.WriteString(id)
	sink.WriteVarBytes(keypair.SerializePublicKey(a2.PubKey()))
	sink.WriteVarBytes(keypair.SerializePublicKey(a1.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a1.Address}
	if _, err = removeKey(n); err == nil {
		t.Error("invalid key removed")
	}

	// 11. remove key
	sink.Reset()
	sink.WriteString(id)
	sink.WriteVarBytes(keypair.SerializePublicKey(a0.PubKey()))
	sink.WriteVarBytes(keypair.SerializePublicKey(a1.PubKey()))
	n.Input = sink.Bytes()
	if _, err = removeKey(n); err != nil {
		t.Fatal(err)
	}

	// 12. check removed key
	sink.Reset()
	sink.WriteString(id)
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	res, err = verifySignature(n)
	if err == nil && bytes.Equal(res, utils.BYTE_TRUE) {
		t.Fatal("removed key passed verification")
	}

	// 13. add removed key again, should fail
	sink.Reset()
	sink.WriteString(id)
	sink.WriteVarBytes(keypair.SerializePublicKey(a0.PubKey()))
	sink.WriteVarBytes(keypair.SerializePublicKey(a1.PubKey()))
	n.Input = sink.Bytes()
	res, err = verifySignature(n)
	if err == nil && bytes.Equal(res, utils.BYTE_TRUE) {
		t.Error("the removed key should not be added again")
	}

	// 14. query removed key
	sink.Reset()
	sink.WriteString(id)
	sink.WriteInt32(1)
	n.Input = sink.Bytes()
	_, err = GetPublicKeyByID(n)
	if err == nil {
		t.Error("query removed key should fail")
	}
}

func CaseOwnerSize(t *testing.T, n *native.NativeService) {
	id, _ := account.GenerateID()
	a := account.NewAccount("")
	err := regID(n, id, a)
	if err != nil {
		t.Fatal(err)
	}

	enc, err := encodeID([]byte(id))
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, OWNER_TOTAL_SIZE)
	_, err = insertPk(n, enc, buf, []byte("controller"), true, false)
	if err == nil {
		t.Fatal("total size of the owner's key should be limited")
	}
}

func CaseGetDocument(t *testing.T, n *native.NativeService) {
	n.Height = config.GetNewOntIdHeight()
	// 1. register ID
	id0, _ := account.GenerateID()
	a0 := account.NewAccount("")
	id1, _ := account.GenerateID()
	a1 := account.NewAccount("")
	id2, _ := account.GenerateID()
	a2 := account.NewAccount("")
	if err := regID(n, id0, a0); err != nil {
		t.Fatal("register ID error", err)
	}
	if err := regID(n, id1, a1); err != nil {
		t.Fatal("register ID error", err)
	}
	if err := regID(n, id2, a2); err != nil {
		t.Fatal("register ID error", err)
	}
	if err := regID(n, "did:ont:TVuSUjQQYsbK9WBnikeEgVy6GuWQ9qz1iW", a2); err != nil {
		t.Fatal("register ID error", err)
	}
	if err := regID(n, "did:ont:TYtNor9XRNYXe2XrM4abzaRbd37WgGKDC1", a2); err != nil {
		t.Fatal("register ID error", err)
	}
	if err := regID(n, "did:ont:TUgmqNEqDJSpN5AgcV2mQ4HtG4qVNVWb89", a2); err != nil {
		t.Fatal("register ID error", err)
	}

	// 2. add key
	sink := common.NewZeroCopySink(nil)
	sink.WriteString(id0)
	sink.WriteVarBytes(keypair.SerializePublicKey(a1.PubKey()))
	sink.WriteVarBytes(keypair.SerializePublicKey(a0.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	if _, err := addKey(n); err != nil {
		t.Fatal(err)
	}

	// 3. set recovery
	r, _ := hex.DecodeString("01022a6469643a6f6e743a54567553556a51515973624b3957426e696b6545675679364775575139717a3169575a01022a6469643a6f6e743a5459744e6f723958524e5958653258724d3461627a6152626433375767474b4443312a6469643a6f6e743a5455676d714e4571444a53704e3541676356326d51344874473471564e565762383901020101")
	sink.Reset()
	sink.WriteString(id0)
	sink.WriteVarBytes(r)
	utils.EncodeVarUint(sink, 1)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	_, err := setRecovery(n)
	if err != nil {
		t.Fatal(err)
	}

	// 4. add context
	var contexts = [][]byte{[]byte("https://www.w3.org/ns0/did/v1"), []byte("https://ontid.ont.io0/did/v1"), []byte("https://ontid.ont.io0/did/v1")}
	context := &Context{
		OntId:    []byte(id0),
		Contexts: contexts,
		Index:    1,
	}
	sink.Reset()
	context.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	_, err = addContext(n)
	if err != nil {
		t.Fatal()
	}

	// 5. add attribute
	attr := attribute{
		key:       []byte("test key"),
		valueType: []byte("test type"),
		value:     []byte("test value"),
	}
	sink.Reset()
	sink.WriteString(id0)
	utils.EncodeVarUint(sink, 1)
	attr.Serialization(sink)
	sink.WriteVarBytes(keypair.SerializePublicKey(a0.PubKey()))
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	if _, err := addAttributes(n); err != nil {
		t.Fatal(err)
	}

	// 6. add auth key
	newPublicKey := &NewPublicKey{
		key:        keypair.SerializePublicKey(a2.PublicKey),
		controller: []byte(id2),
	}
	authKeyParam := &AddNewAuthKeyParam{
		OntId:        []byte(id0),
		NewPublicKey: newPublicKey,
		SignIndex:    1,
	}

	sink.Reset()
	authKeyParam.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	if _, err := addNewAuthKey(n); err != nil {
		t.Fatal(err)
	}

	// 7. set auth key
	setAuthKeyParam := &SetAuthKeyParam{
		OntId:     []byte(id0),
		Index:     1,
		SignIndex: 1,
	}

	sink.Reset()
	setAuthKeyParam.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	if _, err := setAuthKey(n); err != nil {
		t.Fatal(err)
	}

	// 8. add service
	service := &ServiceParam{
		OntId:          []byte(id0),
		ServiceId:      []byte("someService"),
		Type:           []byte("sss"),
		ServiceEndpint: []byte("http;;s;s;s;;s"),
		Index:          1,
	}

	sink.Reset()
	service.Serialization(sink)
	n.Input = sink.Bytes()
	n.Tx.SignedAddr = []common.Address{a0.Address}
	_, err = addService(n)
	if err != nil {
		t.Fatal(err)
	}

	// 9. get document
	res, err := GetDocumentJson(n)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(res))
}

func CaseBeforeNewApiFork(t *testing.T, n *native.NativeService) {
	// 1. register ID
	id0, _ := account.GenerateID()
	a0 := account.NewAccount("")
	id1, _ := account.GenerateID()
	a1 := account.NewAccount("")
	id2, _ := account.GenerateID()
	a2 := account.NewAccount("")
	if err := regID(n, id0, a0); err != nil {
		t.Fatal("register ID error", err)
	}
	if err := regID(n, id1, a1); err != nil {
		t.Fatal("register ID error", err)
	}
	if err := regID(n, id2, a2); err != nil {
		t.Fatal("register ID error", err)
	}

	// 2. add context before fork height
	var contexts = [][]byte{[]byte("https://www.w3.org/ns0/did/v1"), []byte("https://ontid.ont.io0/did/v1"), []byte("https://ontid.ont.io0/did/v1")}
	context := &Context{
		OntId:    []byte(id0),
		Contexts: contexts,
		Index:    1,
	}
	sink := common.NewZeroCopySink(nil)
	context.Serialization(sink)
	n.Tx.SignedAddr = []common.Address{a0.Address}
	_, err := n.NativeCall(utils.OntIDContractAddress, "addContext", sink.Bytes())
	assert.Contains(t, err.Error(), "doesn't support this function")

	// 3. add auth key before fork height
	newPublicKey := &NewPublicKey{
		key:        keypair.SerializePublicKey(a2.PublicKey),
		controller: []byte(id2),
	}
	authKeyParam := &AddNewAuthKeyParam{
		OntId:        []byte(id0),
		NewPublicKey: newPublicKey,
		SignIndex:    1,
	}

	sink.Reset()
	authKeyParam.Serialization(sink)
	n.Tx.SignedAddr = []common.Address{a0.Address}
	_, err = n.NativeCall(utils.OntIDContractAddress, "addNewAuthKey", sink.Bytes())
	assert.Contains(t, err.Error(), "doesn't support this function")

	// 4. set auth key before fork height
	setAuthKeyParam := &SetAuthKeyParam{
		OntId:     []byte(id0),
		Index:     1,
		SignIndex: 1,
	}

	sink.Reset()
	setAuthKeyParam.Serialization(sink)
	n.Tx.SignedAddr = []common.Address{a0.Address}
	_, err = n.NativeCall(utils.OntIDContractAddress, "setAuthKey", sink.Bytes())
	assert.Contains(t, err.Error(), "doesn't support this function")

	// 5. add service before fork height
	service := &ServiceParam{
		OntId:          []byte(id0),
		ServiceId:      []byte("someService"),
		Type:           []byte("sss"),
		ServiceEndpint: []byte("http;;s;s;s;;s"),
		Index:          1,
	}

	sink.Reset()
	service.Serialization(sink)
	n.Tx.SignedAddr = []common.Address{a0.Address}
	_, err = n.NativeCall(utils.OntIDContractAddress, "addService", sink.Bytes())
	assert.Contains(t, err.Error(), "doesn't support this function")

	// 6. get document before fork height
	_, err = n.NativeCall(utils.OntIDContractAddress, "getDocumentJson", sink.Bytes())
	assert.Contains(t, err.Error(), "doesn't support this function")
}

func CaseAfterNewApiFork(t *testing.T, n *native.NativeService) {
	n.Height = config.GetNewOntIdHeight()
	// 1. register ID
	id0, _ := account.GenerateID()
	a0 := account.NewAccount("")
	id1, _ := account.GenerateID()
	a1 := account.NewAccount("")
	id2, _ := account.GenerateID()
	a2 := account.NewAccount("")
	if err := regID(n, id0, a0); err != nil {
		t.Fatal("register ID error", err)
	}
	if err := regID(n, id1, a1); err != nil {
		t.Fatal("register ID error", err)
	}
	if err := regID(n, id2, a2); err != nil {
		t.Fatal("register ID error", err)
	}

	// 2. add context after fork height
	var contexts = [][]byte{[]byte("https://www.w3.org/ns0/did/v1"), []byte("https://ontid.ont.io0/did/v1"), []byte("https://ontid.ont.io0/did/v1")}
	context := &Context{
		OntId:    []byte(id0),
		Contexts: contexts,
		Index:    1,
	}
	sink := common.NewZeroCopySink(nil)
	context.Serialization(sink)
	n.Tx.SignedAddr = []common.Address{a0.Address}
	_, err := n.NativeCall(utils.OntIDContractAddress, "addContext", sink.Bytes())
	assert.NoError(t, err)

	// 3. add auth key after fork height
	newPublicKey := &NewPublicKey{
		key:        keypair.SerializePublicKey(a2.PublicKey),
		controller: []byte(id2),
	}
	authKeyParam := &AddNewAuthKeyParam{
		OntId:        []byte(id0),
		NewPublicKey: newPublicKey,
		SignIndex:    1,
	}

	sink.Reset()
	authKeyParam.Serialization(sink)
	n.Tx.SignedAddr = []common.Address{a0.Address}
	_, err = n.NativeCall(utils.OntIDContractAddress, "addNewAuthKey", sink.Bytes())
	assert.NoError(t, err)

	// 4. set auth key after fork height
	setAuthKeyParam := &SetAuthKeyParam{
		OntId:     []byte(id0),
		Index:     1,
		SignIndex: 1,
	}

	sink.Reset()
	setAuthKeyParam.Serialization(sink)
	n.Tx.SignedAddr = []common.Address{a0.Address}
	_, err = n.NativeCall(utils.OntIDContractAddress, "setAuthKey", sink.Bytes())
	assert.NoError(t, err)

	// 5. add service after fork height
	service := &ServiceParam{
		OntId:          []byte(id0),
		ServiceId:      []byte("someService"),
		Type:           []byte("sss"),
		ServiceEndpint: []byte("http;;s;s;s;;s"),
		Index:          1,
	}

	sink.Reset()
	service.Serialization(sink)
	n.Tx.SignedAddr = []common.Address{a0.Address}
	_, err = n.NativeCall(utils.OntIDContractAddress, "addService", sink.Bytes())
	assert.NoError(t, err)

	// 6. get document after fork height
	res, err := n.NativeCall(utils.OntIDContractAddress, "getDocumentJson", sink.Bytes())
	assert.NoError(t, err)
	fmt.Println(string(res))
}
