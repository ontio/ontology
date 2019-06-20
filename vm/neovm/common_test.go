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
package neovm

import (
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ripemd160"
	"io"
	"testing"
)

func TestHash(t *testing.T) {
	engine := NewExecutionEngine(0)
	engine.OpCode = HASH160

	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	hd160 := Hash(data, engine)

	temp := sha256.Sum256(data)
	md := ripemd160.New()
	io.WriteString(md, string(temp[:]))
	assert.Equal(t, hd160, md.Sum(nil))

	temp1 := sha256.Sum256(data)
	data1 := sha256.Sum256(temp1[:])

	engine.OpCode = HASH256
	hd256 := Hash(data, engine)

	assert.Equal(t, data1[:], hd256)
}
