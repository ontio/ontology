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
package account

import (
	"github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewAccount(t *testing.T) {
	defer func() {
		os.RemoveAll("Log/")
	}()

	names := []string{
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
	accounts := make([]*Account, len(names))
	for k, v := range names {
		accounts[k] = NewAccount(v)
		assert.NotNil(t, accounts[k])
		assert.NotNil(t, accounts[k].PrivateKey)
		assert.NotNil(t, accounts[k].PublicKey)
		assert.NotNil(t, accounts[k].Address)
		assert.NotNil(t, accounts[k].PrivKey())
		assert.NotNil(t, accounts[k].PubKey())
		assert.NotNil(t, accounts[k].Scheme())
		assert.Equal(t, accounts[k].Address, types.AddressFromPubKey(accounts[k].PublicKey))
	}
}
