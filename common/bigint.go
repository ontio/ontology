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

import "math/big"

func bytesReverse(u []byte) []byte {
	for i, j := 0, len(u)-1; i < j; i, j = i+1, j-1 {
		u[i], u[j] = u[j], u[i]
	}
	return u
}

var bigOne = big.NewInt(1)

// neo encoding: https://docs.microsoft.com/en-us/dotnet/api/system.numerics.biginteger.tobytearray?view=netframework-4.7.2
func BigIntToNeoBytes(data *big.Int) []byte {
	bs := data.Bytes()
	if len(bs) == 0 {
		return []byte{}
	}
	// golang big.Int use big-endian
	bytesReverse(bs)
	// bs now is little-endian
	if data.Sign() < 0 {
		for i, b := range bs {
			bs[i] = ^b
		}
		for i := 0; i < len(bs); i++ {
			if bs[i] == 255 {
				bs[i] = 0
			} else {
				bs[i] += 1
				break
			}
		}
		if bs[len(bs)-1] < 128 {
			bs = append(bs, 255)
		}
	} else {
		if bs[len(bs)-1] >= 128 {
			bs = append(bs, 0)
		}
	}
	return bs
}

func BigIntFromNeoBytes(ba []byte) *big.Int {
	res := big.NewInt(0)
	l := len(ba)
	if l == 0 {
		return res
	}

	bytes := make([]byte, 0, l)
	bytes = append(bytes, ba...)
	bytesReverse(bytes)

	if bytes[0]>>7 == 1 {
		for i, b := range bytes {
			bytes[i] = ^b
		}

		temp := big.NewInt(0)
		temp.SetBytes(bytes)
		temp.Add(temp, bigOne)
		bytes = temp.Bytes()
		res.SetBytes(bytes)
		return res.Neg(res)
	}

	res.SetBytes(bytes)
	return res
}
