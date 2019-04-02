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
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBigIntFromBytes(t *testing.T) {
	cases := []string{
		"7491324e0ed37bcf702ad3540b7a1dc724d31d5cdd9ff0803a2172d5da5c00ff",
		"80",
		"00",
		"0000000000000080",
	}
	for _, cs := range cases {
		buf, _ := hex.DecodeString(cs)
		v := BigIntFromNeoBytes(buf)
		buf2 := BigIntToNeoBytes(v)
		v2 := BigIntFromNeoBytes(buf2)

		assert.Equal(t, v.Cmp(v2), 0, fmt.Sprintf("message:%d, %d, %x", v, v2, buf))
		assert.Equal(t, bytes.Equal(simplifyNeoBytes(buf), buf2), true)
	}
}

func TestBigInt(t *testing.T) {
	cases := [][2]string{
		{"-1", "FF"},
		{"1", "01"},
		//{"0", "00"},
		{"120", "78"},
		{"128", "8000"},
		{"255", "FF00"},
		{"1024", "0004"},
		{"-9223372036854775808", "0000000000000080"},
		{"9223372036854775807", "FFFFFFFFFFFFFF7F"},
		{"90123123981293054321", "71E975A9C4A7B5E204"},
	}
	for _, cs := range cases {
		v, b := big.NewInt(0).SetString(cs[0], 10)
		assert.True(t, b)
		buf := BigIntToNeoBytes(v)
		orig, _ := hex.DecodeString(cs[1])

		assert.Equal(t, string(buf), string(orig))
	}
}

func TestRandBigIntFromBytes(t *testing.T) {
	const N = 1000000
	for i := 0; i < N; i++ {
		length := (rand.Uint32() % 100) + 1
		buf := make([]byte, length)
		_, _ = crand.Read(buf)

		v := BigIntFromNeoBytes(buf)
		buf2 := BigIntToNeoBytes(v)
		v2 := BigIntFromNeoBytes(buf2)

		assert.Equal(t, v.Cmp(v2), 0, fmt.Sprintf("message:%d, %d, %x", v, v2, buf))
		assert.Equal(t, bytes.Equal(simplifyNeoBytes(buf), buf2), true, fmt.Sprintf("buff: %x, %x", buf, buf2))
	}
}

func trimFF(buf []byte) []byte {
	if len(buf) <= 1 {
		return buf
	}
	i := len(buf)
	for ; i > 0 && buf[i-1] == 255; i-- {
	}
	if i == len(buf) {
		return buf
	}

	buf = buf[:i+1]
	if i > 0 && buf[i-1] >= 128 {
		buf = buf[:i]
	}

	return buf
}

func trim00(buf []byte) []byte {
	if len(buf) <= 1 {
		return buf
	}
	i := len(buf)
	for ; i > 0 && buf[i-1] == 0; i-- {
	}
	if i == len(buf) {
		return buf
	}

	buf = buf[:i+1]
	if i > 0 && buf[i-1] < 128 {
		buf = buf[:i]
	}

	return buf
}

func simplifyNeoBytes(buf []byte) []byte {
	if len(buf) <= 1 {
		if bytes.Equal(buf, []byte{0}) {
			return nil
		}
		return buf
	}
	i := len(buf)
	if buf[i-1] == 255 {
		return trimFF(buf)
	} else if buf[i-1] == 0 {
		buf = trim00(buf)
		// treat bytes(0) to nil
		if bytes.Equal(buf, []byte{0}) {
			return nil
		}
		return buf
	}
	return buf
}

func TestSimplifyNeoBytes(t *testing.T) {
	assert.Equal(t, simplifyNeoBytes([]byte{255}), []byte{255})
	assert.Equal(t, simplifyNeoBytes([]byte{1, 2, 255}), []byte{1, 2, 255})
	assert.Equal(t, simplifyNeoBytes([]byte{1, 2, 255, 255}), []byte{1, 2, 255})
	assert.Equal(t, simplifyNeoBytes([]byte{1, 2, 255, 255, 255}), []byte{1, 2, 255})
	assert.Equal(t, simplifyNeoBytes([]byte{1, 128, 255, 255, 255}), []byte{1, 128})

	assert.Equal(t, simplifyNeoBytes([]byte{0, 0, 0}), []byte(nil))
	assert.Equal(t, simplifyNeoBytes([]byte{0, 1, 0}), []byte{0, 1})
	assert.Equal(t, simplifyNeoBytes([]byte{0, 234, 0}), []byte{0, 234, 0})
	assert.Equal(t, simplifyNeoBytes([]byte{0, 234, 0, 0}), []byte{0, 234, 0})

}
