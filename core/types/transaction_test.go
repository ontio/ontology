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
	"testing"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/payload"
	stypes "github.com/ontio/ontology/smartcontract/types"
)

func genTx() *Transaction {
	data := &payload.InvokeCode{
		Code: stypes.VmCode{
			VmType: stypes.Native,
			Code:   []byte("test1234"),
		},
	}

	_, pk, err := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	if err != nil {
		return nil
	}
	sigs := []*Sig{{
		PubKeys: []keypair.PublicKey{pk},
		M:       1,
		SigData: [][]byte{[]byte("test")},
	}}

	tx := &Transaction{
		Version:    1,
		TxType:     Invoke,
		Nonce:      100,
		GasPrice:   1000,
		GasLimit:   100000,
		Payer:      common.Address{},
		Payload:    data,
		Attributes: make([]*TxAttribute, 0),
		Sigs:       sigs,
	}

	return tx
}

func TestTransaction_Deserialize(t *testing.T) {
	tx := genTx()

	buf := bytes.NewBuffer([]byte{})
	tx.Serialize(buf)

	tx2 := new(Transaction)
	tx2.Deserialize(buf)

	h1 := tx.Hash()
	h2 := tx2.Hash()

	if bytes.Compare(h1.ToArray(), h2.ToArray()) != 0 {
		t.Fail()
	}
}

func TestTransaction_ParallelDeserialize(t *testing.T) {
	tx := genTx()
	h1 := tx.Hash()

	txBuf := bytes.NewBuffer([]byte{})
	tx.Serialize(txBuf)

	data := bytes.NewBuffer([]byte{})
	cnt := 60000
	for i := 0; i < cnt; i++ {
		serialization.WriteVarBytes(data, txBuf.Bytes())
	}

	input1 := bytes.NewBuffer(data.Bytes())
	input2 := bytes.NewBuffer(data.Bytes())

	start := time.Now()
	for i := 0; i < cnt; i++ {
		txdata, err := serialization.ReadVarBytes(input1)
		if err != nil {
			t.Errorf(err.Error())
		}
		tx2 := new(Transaction)
		if err := tx2.Deserialize(bytes.NewBuffer(txdata)); err != nil {
			t.Errorf(err.Error())
		}
	}
	t.Logf("deserialize %d tx, time: %d", cnt, time.Since(start).Nanoseconds())

	start2 := time.Now()
	_, hashes, err := parallelDeserializeTxs(input2, cnt)
	if err != nil {
		t.Errorf(err.Error())
	}
	t.Logf("parallel deserialize %d tx, time: %d", cnt, time.Since(start2).Nanoseconds())

	if len(hashes) != cnt {
		t.Errorf("mismatch deserialize tx count: %d vs %d", len(hashes), cnt)
	}
	for _, h2 := range hashes {
		if bytes.Compare(h1.ToArray(), h2.ToArray()) != 0 {
			t.Fail()
		}
	}
}

func TestTransaction_DeserializeBadTx1(t *testing.T) {
	tx := genTx()

	txBuf := bytes.NewBuffer([]byte{})
	tx.Serialize(txBuf)

	// bad txn in the middle
	data := bytes.NewBuffer([]byte{})
	cnt := 100
	for i := 0; i < cnt; i++ {
		serialization.WriteVarBytes(data, txBuf.Bytes())
		serialization.WriteVarBytes(data, []byte("bad transaction"))
	}

	_, _, err := parallelDeserializeTxs(data, cnt)
	if err == nil {
		t.Errorf("bad txn deserialized successfully")
	}
	t.Logf("deserialize bad txn get err: %s", err)
}

func TestTransaction_DeserializeBadTx2(t *testing.T) {
	tx := genTx()

	txBuf := bytes.NewBuffer([]byte{})
	tx.Serialize(txBuf)

	// first txn is bad
	data := bytes.NewBuffer([]byte{})
	serialization.WriteVarBytes(data, []byte("bad transaction"))
	cnt := 100
	for i := 0; i < cnt; i++ {
		serialization.WriteVarBytes(data, txBuf.Bytes())
	}

	_, _, err := parallelDeserializeTxs(data, cnt)
	if err == nil {
		t.Errorf("bad txn deserialized successfully")
	}
	t.Logf("deserialize bad txn get err: %s", err)
}

func TestTransaction_DeserializeBadTxCount(t *testing.T) {
	tx := genTx()

	txBuf := bytes.NewBuffer([]byte{})
	tx.Serialize(txBuf)

	data := bytes.NewBuffer([]byte{})
	cnt := 100
	for i := 0; i < cnt; i++ {
		serialization.WriteVarBytes(data, txBuf.Bytes())
	}

	_, _, err := parallelDeserializeTxs(data, 1000)
	if err == nil {
		t.Errorf("bad txn deserialized successfully")
	}
	t.Logf("deserialize bad txn count get err: %s", err)
}
