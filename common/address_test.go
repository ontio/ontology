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

package common

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressFromBase58(t *testing.T) {
	var addr Address
	rand.Read(addr[:])

	base58 := addr.ToBase58()
	b1 := string(append([]byte{'X'}, []byte(base58)...))
	_, err := AddressFromBase58(b1)

	assert.NotNil(t, err)

	b2 := string([]byte(base58)[1:10])
	_, err = AddressFromBase58(b2)

	assert.NotNil(t, err)
}

func TestAddressParseFromBytes(t *testing.T) {
	var addr Address
	rand.Read(addr[:])

	addr2, _ := AddressParseFromBytes(addr[:])

	assert.Equal(t, addr, addr2)
}

func TestAddress_Serialize(t *testing.T) {
	var addr Address
	rand.Read(addr[:])

	buf := bytes.NewBuffer(nil)
	addr.Serialize(buf)

	var addr2 Address
	addr2.Deserialize(buf)
	assert.Equal(t, addr, addr2)
}
