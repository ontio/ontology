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

package merkle_pdp

import (
	"crypto/rand"
	"testing"
)

func TestMerkleProof(t *testing.T) {
	data := make([]byte, 256*1024)
	rand.Read(data)

	prf := MerkleProof(data, 1)
	if err := VerifyMerkleProof(prf, prf[len(prf)-1], 1); err != nil {
		t.Fatal(err.Error())
	}
}

func BenchmarkHash(b *testing.B) {
	data := make([]byte, 256*1024)
	rand.Read(data)

	for i := 0; i < b.N; i++ {
		MerkleProof(data, 10)
	}
}

func BenchmarkVerify(b *testing.B) {
	data := make([]byte, 256*1024)
	rand.Read(data)
	proof := MerkleProof(data, 10)

	for i := 0; i < b.N; i++ {
		if err := VerifyMerkleProof(proof, proof[len(proof)-1], 10); err != nil {
			b.Fatal(err.Error())
		}
	}
}
