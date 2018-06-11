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
package signature

import (
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"time"
)

var encryptNames = []string{
	"",
	"SHA224withECDSA",
	"SHA256withECDSA",
	"SHA384withECDSA",
	"SHA512withECDSA",
	"SHA3-224withECDSA",
	"SHA3-256withECDSA",
	"SHA3-384withECDSA",
	"SHA3-512withECDSA",
	"RIPEMD160withECDSA",
	"SM3withSM2",
	"SHA512withEdDSA",
}

func TestSignAndVerify(t *testing.T) {
	for _, encryptName := range encryptNames {
		acc := account.NewAccount(encryptName)
		data := []byte{1, 2, 3}
		sig, err := Sign(acc, data)
		if err != nil {
			t.Errorf("%s sign failed, err is %s", encryptName, err)
		}

		err = Verify(acc.PublicKey, data, sig)
		if err != nil {
			t.Errorf("%s verify failed, err is %s", encryptName, err)
		}
	}
}

func TestVerifyMultiSignature(t *testing.T) {
	data := []byte{1, 2, 3}
	accs := make([]*account.Account, 0)
	pubkeys := make([]keypair.PublicKey, 0)
	N := 4
	// test different encrypt scheme sign and verify
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < N; i++ {
		encryptSchemeIndex := rand.Intn(len(encryptNames))
		accs = append(accs, account.NewAccount(encryptNames[encryptSchemeIndex]))
	}
	sigs := make([][]byte, 0)

	for _, acc := range accs {
		sig, _ := Sign(acc, data)
		sigs = append(sigs, sig)
		pubkeys = append(pubkeys, acc.PublicKey)
	}

	err := VerifyMultiSignature(data, pubkeys, N, sigs)
	assert.Nil(t, err)

	pubkeys[0], pubkeys[1] = pubkeys[1], pubkeys[0]
	err = VerifyMultiSignature(data, pubkeys, N, sigs)
	assert.Nil(t, err)

}
