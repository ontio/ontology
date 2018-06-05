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

package types

import (
	"bytes"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/smartcontract/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func genTestSig(m uint8) *Sig {
	sig := new(Sig)
	sig.M = m
	_, pub, _ := keypair.GenerateKeyPair(keypair.PK_EDDSA, keypair.ED25519)
	sig.PubKeys = make([]keypair.PublicKey, 0)
	sig.PubKeys = append(sig.PubKeys, pub)
	sig.SigData = make([][]byte, 0)
	sig.SigData = append(sig.SigData, keypair.SerializePublicKey(pub))
	return sig
}

func genTestTx(txType TransactionType) *Transaction {
	tx := new(Transaction)
	tx.TxType = txType
	tx.Nonce = uint32(time.Now().Unix())
	tx.Sigs = []*Sig{genTestSig(0)}
	tx.Payer = AddressFromPubKey(tx.Sigs[0].PubKeys[0])
	tx.Attributes = make([]*TxAttribute, 0)
	attr := NewTxAttribute(Nonce, []byte("test transaction attribute"))
	tx.Attributes = append(tx.Attributes, &attr)
	tx.Payload = &payload.InvokeCode{
		Code: types.VmCode{
			VmType: types.Native,
			Code:   []byte{0x00, 0x00, 0x01, 0x11, 0x22, 0x45, 0x55},
		},
	}
	return tx
}

func TestSig_Serialize_Deserialize(t *testing.T) {
	sig := genTestSig(0)
	bf := new(bytes.Buffer)
	err := sig.Serialize(bf)
	assert.Nil(t, err)

	deserializeSig := new(Sig)
	deserializeSig.Deserialize(bf)
	assert.Equal(t, deserializeSig, sig)
}

func TestTransaction_Serialize_Deserialize(t *testing.T) {
	invokeTx := genTestTx(Invoke)
	bf := new(bytes.Buffer)
	err := invokeTx.Serialize(bf)
	assert.Nil(t, err)
	deserializeTx := new(Transaction)
	err = deserializeTx.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, invokeTx, deserializeTx)

	deployTx := genTestTx(Invoke)
	bf.Reset()
	err = deployTx.Serialize(bf)
	assert.Nil(t, err)
	deserializeTx = new(Transaction)
	err = deserializeTx.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, deployTx, deserializeTx)

	otherTypeTx := genTestTx(Enrollment)
	bf.Reset()
	err = otherTypeTx.Serialize(bf)
	assert.Nil(t, err)
	deserializeTx = new(Transaction)
	err = deserializeTx.Deserialize(bf)
	assert.NotNil(t, err)
}

func TestTransaction(t *testing.T) {
	testTx := genTestTx(Invoke)
	sigAddress := testTx.GetSignatureAddresses()
	assert.NotNil(t, sigAddress)
	assert.Equal(t, sigAddress[0], AddressFromPubKey(testTx.Sigs[0].PubKeys[0]))

	txMessageLen := len(testTx.GetMessage())
	if txMessageLen <= 0 {
		t.Fatal("tx get message test failed")
	}

	txArrayLen := len(testTx.ToArray())
	if txArrayLen <= 0 {
		t.Fatal("tx to array test failed")
	}

	txHash := testTx.Hash()
	if len(txHash) <= 0 {
		t.Fatal("tx to hash test failed")
	}
	testHash := txHash
	testHash[0] = 0x01
	testTx.SetHash(testHash)
	assert.NotEqual(t, txHash, testTx.Hash())

	assert.Equal(t, testTx.Type(), common.TRANSACTION)
}
