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
package tests

import (
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/core/signature"
	"github.com/stretchr/testify/assert"
)

func TestSign(t *testing.T) {
	acc := account.NewAccount("")
	data := []byte{1, 2, 3}
	sig, err := signature.Sign(acc, data)
	assert.Nil(t, err)

	err = signature.Verify(acc.PublicKey, data, sig)
	assert.Nil(t, err)
}

func TestVerifyMultiSignature(t *testing.T) {
	data := []byte{1, 2, 3}
	accs := make([]*account.Account, 0)
	pubkeys := make([]keypair.PublicKey, 0)
	N := 4
	for i := 0; i < N; i++ {
		accs = append(accs, account.NewAccount(""))
	}
	sigs := make([][]byte, 0)

	for _, acc := range accs {
		sig, _ := signature.Sign(acc, data)
		sigs = append(sigs, sig)
		pubkeys = append(pubkeys, acc.PublicKey)
	}

	err := signature.VerifyMultiSignature(data, pubkeys, N, sigs)
	assert.Nil(t, err)

	pubkeys[0], pubkeys[1] = pubkeys[1], pubkeys[0]
	err = signature.VerifyMultiSignature(data, pubkeys, N, sigs)
	assert.Nil(t, err)

}
