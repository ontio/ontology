/*
 * Copyright (C) 2019 The ontology Authors
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

package chainmgr

import (
	"bytes"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/core/signature"
)

func TestShardAccountMarshal(t *testing.T) {
	acc := account.NewAccount("")
	if acc == nil {
		t.Fatalf("failed to new account")
	}

	accBytes, err := serializeShardAccount(acc)
	if err != nil {
		t.Fatalf("serialize shard account: %s", err)
	}

	acc2, err := deserializeShardAccount(accBytes)
	if err != nil {
		t.Fatalf("deserialize shard account: %s", err)
	}

	pk1Bytes := keypair.SerializePrivateKey(acc.PrivKey())
	pk2Bytes := keypair.SerializePrivateKey(acc2.PrivKey())
	if bytes.Compare(pk1Bytes, pk2Bytes) != 0 {
		t.Fatalf("different private key: %v vs %v", pk1Bytes, pk2Bytes)
	}

	if acc.SigScheme != acc2.SigScheme {
		t.Fatalf("differnt sig scheme: %v vs %v", acc.SigScheme, acc2.SigScheme)
	}

	text := []byte("abcdefg")
	sig1, err := signature.Sign(acc, text)
	if err != nil {
		t.Fatalf("sign sig with acc1 failed: %s", err)
	}
	if err := signature.Verify(acc2.PublicKey, text, sig1); err != nil {
		t.Fatalf("sign sig with acc2 failed: %s", err)
	}
}
