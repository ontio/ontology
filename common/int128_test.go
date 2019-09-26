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

func TestI128LittleInt(t *testing.T) {
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
		i128 := I128FromInt64(int64(test.input))
		assert.Equal(t, i128.ToBEHex(), test.output)
	}
}

func TestI128BigInt(t *testing.T) {
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
			new(big.Int).Sub(bigPow(2, 127), big.NewInt(1)),
			"7fffffffffffffffffffffffffffffff",
		},
		{
			new(big.Int).Neg(bigPow(2, 127)),
			"80000000000000000000000000000000",
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
		u128, err := I128FromBigInt(test.input)
		assert.Nil(t, err)
		assert.Equal(t, u128.ToBEHex(), test.output)
	}
}

func TestI128Conv(t *testing.T) {
	var buf [16]byte
	const N = 1000000
	for i := 0; i < N; i++ {
		_, err := rand.Read(buf[:])
		assert.Nil(t, err)
		randInt := int64(binary.LittleEndian.Uint64(buf[:]))

		u1 := I128FromInt64(randInt)
		u2, err := I128FromBigInt(big.NewInt(randInt))
		assert.Nil(t, err)

		assert.Equal(t, u1, u2)
	}
}

func TestI128ToNumString(t *testing.T) {
	numbers := []int64{1, 255, 256, 123456, -1, -127, -128, -255, -256}
	strings := []string{"1", "255", "256", "123456", "-1", "-127", "-128", "-255", "-256"}
	for i, num := range numbers {
		u1 := I128FromInt64(num)
		u2, err := I128FromBigInt(big.NewInt(num))

		assert.Nil(t, err)
		assert.Equal(t, u1, u2)
		assert.Equal(t, u2.ToNumString(), strings[i])
	}
}
