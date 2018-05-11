package common

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
import (
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHash(t *testing.T) {

	var data []Uint256
	a1 := Uint256(sha256.Sum256([]byte("a")))
	a2 := Uint256(sha256.Sum256([]byte("b")))
	a3 := Uint256(sha256.Sum256([]byte("c")))
	a4 := Uint256(sha256.Sum256([]byte("d")))
	a5 := Uint256(sha256.Sum256([]byte("e")))
	data = append(data, a1)
	data = append(data, a2)
	data = append(data, a3)
	data = append(data, a4)
	data = append(data, a5)
	hash := ComputeMerkleRoot(data)
	assert.NotEqual(t, hash, UINT256_EMPTY)

}
