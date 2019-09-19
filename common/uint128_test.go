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
	"crypto/rand"
	"encoding/binary"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestU128LittleInt(t *testing.T) {
	for _, test := range []struct {
		input  int
		output string
	}{
		{
			0,
			"00000000000000000000000000000000",
		},
		{
			1,
			"00000000000000000000000000000001",
		},
		{
			16,
			"00000000000000000000000000000010",
		},
		{
			-1,
			"ffffffffffffffffffffffffffffffff",
		},
		{
			-2,
			"fffffffffffffffffffffffffffffffe",
		},
	} {
		u128 := U128FromInt64(int64(test.input))
		assert.Equal(t, u128.ToBEHex(), test.output)
	}
}

func TestU128BigInt(t *testing.T) {
	for _, test := range []struct {
		input  *big.Int
		output string
	}{
		{
			big.NewInt(0),
			"00000000000000000000000000000000",
		},
		{
			big.NewInt(1),
			"00000000000000000000000000000001",
		},
		{
			big.NewInt(16),
			"00000000000000000000000000000010",
		},
		{
			big.NewInt(-1),
			"ffffffffffffffffffffffffffffffff",
		},
		{
			big.NewInt(-2),
			"fffffffffffffffffffffffffffffffe",
		},
		{
			new(big.Int).Sub(bigPow(2, 128), big.NewInt(1)),
			"ffffffffffffffffffffffffffffffff",
		},
		{
			new(big.Int).Neg(bigPow(2, 128)),
			"00000000000000000000000000000000",
		},
		{
			big.NewInt(130),
			"00000000000000000000000000000082",
		},
		{
			big.NewInt(255),
			"000000000000000000000000000000ff",
		},
	} {
		u128 := U128FromBigInt(test.input)
		assert.Equal(t, u128.ToBEHex(), test.output)
	}
}

func TestU128Conv(t *testing.T) {
	var buf [16]byte
	const N = 1000000
	for i := 0; i < N; i++ {
		_, err := rand.Read(buf[:])
		assert.Nil(t, err)
		randInt := int64(binary.LittleEndian.Uint64(buf[:]))

		u1 := U128FromInt64(randInt)
		u2 := U128FromBigInt(big.NewInt(randInt))

		assert.Equal(t, u1, u2)
	}
}
