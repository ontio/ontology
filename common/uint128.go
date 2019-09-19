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
	"encoding/binary"
	"fmt"
	"math/big"
)

const U128_SIZE = 16

type U128 [U128_SIZE]byte // little endian u128

var U128_EMPTY = U128{}
var U128_MAX = func() U128 {
	var u128 U128
	for i := 0; i < U128_SIZE; i++ {
		u128[i] = 255
	}
	return u128
}()

var pow128 = bigPow(2, 128)
var maxBigU128 = new(big.Int).Sub(bigPow(2, 128), big.NewInt(1))
var maxBigS128 = new(big.Int).Sub(bigPow(2, 127), big.NewInt(1))

// returns a ** b
func bigPow(a, b int64) *big.Int {
	r := big.NewInt(a)
	return r.Exp(r, big.NewInt(b), nil)
}

func U128FromUint64(val uint64) U128 {
	var u128 U128
	binary.LittleEndian.PutUint64(u128[:], val)
	return u128
}

func U128FromInt64(val int64) U128 {
	var u128 U128
	if val < 0 {
		u128 = U128_MAX
	}
	binary.LittleEndian.PutUint64(u128[:], uint64(val))

	return u128
}

// val should in u128 range
func U128FromBigInt(val *big.Int) U128 {
	var u128 U128
	if val.Sign() < 0 {
		val = new(big.Int).Add(val, pow128)
	}
	buf := val.Bytes()
	buf = ToArrayReverse(buf)
	copy(u128[:], buf)

	return u128
}

// to big endian hex string
func (self *U128) ToBEHex() string {
	return fmt.Sprintf("%x", ToArrayReverse(self[:]))
}

// to little endian hex string
func (self *U128) ToLEHex() string {
	return fmt.Sprintf("%x", self[:])
}
