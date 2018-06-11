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
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func genTestSig() (*Sig, error) {
	sig := new(Sig)
	sig.M = 1
	_, pub, err := keypair.GenerateKeyPair(keypair.PK_EDDSA, keypair.ED25519)
	sig.PubKeys = make([]keypair.PublicKey, 0)
	sig.PubKeys = append(sig.PubKeys, pub)
	sig.SigData = make([][]byte, 0)
	sig.SigData = append(sig.SigData, keypair.SerializePublicKey(pub))
	return sig, err
}

func genTestTx(txType TransactionType) (*Transaction, error) {
	tx := new(Transaction)
	tx.TxType = txType
	tx.Nonce = uint32(time.Now().Unix())
	sig, err := genTestSig()
	tx.Sigs = []*Sig{sig}
	tx.Payer = AddressFromPubKey(tx.Sigs[0].PubKeys[0])
	tx.Payload = &payload.InvokeCode{
		[]byte{0x00, 0x00, 0x01, 0x11, 0x22, 0x45, 0x55},
	}
	return tx, err
}
func TestSig_Serialize_Deserialize(t *testing.T) {
	sig, err := genTestSig()
	assert.Nil(t, err)

	bf := new(bytes.Buffer)
	err = sig.Serialize(bf)
	assert.Nil(t, err)

	deserializeSig := new(Sig)
	deserializeSig.Deserialize(bf)
	assert.Equal(t, deserializeSig, sig)
}

func TestTransaction_Serialize_Deserialize(t *testing.T) {
	invokeTx, err := genTestTx(Invoke)
	assert.Nil(t, err)
	bf := new(bytes.Buffer)
	err = invokeTx.Serialize(bf)
	assert.Nil(t, err)
	deserializeTx := new(Transaction)
	err = deserializeTx.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, invokeTx, deserializeTx)

	deployTx, err := genTestTx(Invoke)
	assert.Nil(t, err)
	bf.Reset()
	err = deployTx.Serialize(bf)
	assert.Nil(t, err)
	deserializeTx = new(Transaction)
	err = deserializeTx.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, deployTx, deserializeTx)

	bookkeeperTypeTx, err := genTestTx(Bookkeeper)
	assert.Nil(t, err)
	bf.Reset()
	err = bookkeeperTypeTx.Serialize(bf)
	assert.Nil(t, err)
	deserializeTx = new(Transaction)
	err = deserializeTx.Deserialize(bf)
	assert.NotNil(t, err)
}

func TestTransaction(t *testing.T) {
	testTx, err := genTestTx(Invoke)
	assert.Nil(t, err)
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
