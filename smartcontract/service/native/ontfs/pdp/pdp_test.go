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

package pdp

import (
	"crypto/rand"
	"testing"

	"github.com/ontio/ontology/smartcontract/service/native/ontfs/pdp/types"
)

func TestPdpVerify(t *testing.T) {
	data := make([]byte, 256*1024)
	var blocks []types.Block
	for i := 0; i < 1024; i++ {
		rand.Read(data)
		blocks = append(blocks, data)
	}

	pdp := NewPdp(MerklePdp)
	fileUniqueId, err := pdp.GenUniqueIdWithFileBlocks(blocks)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("fileUniqueId: %v", fileUniqueId)
	}

	var nodeId [20]byte
	rand.Read(nodeId[:])

	blockHash := make([]byte, 32)
	rand.Read(blockHash[:])

	challenge, err := pdp.GenChallenge(nodeId, blockHash, 1024)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("challenge: %v", challenge)
	}

	proof, err := pdp.GenProofWithBlocks(blocks, fileUniqueId, challenge)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("proof: %v", proof)
		t.Logf("proofLen: %v", len(proof))
	}

	err = VerifyProofWithUniqueId(fileUniqueId, proof, challenge)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func BenchmarkGenUniqueIdWithFileBlocks(b *testing.B) {
	data := make([]byte, 256*1024)
	var blocks []types.Block
	for i := 0; i < 1024; i++ {
		rand.Read(data)
		blocks = append(blocks, data)
	}

	pdp := NewPdp(MerklePdp)
	for i := 0; i < b.N; i++ {
		_, err := pdp.GenUniqueIdWithFileBlocks(blocks)
		if err != nil {
			b.Fatal(err.Error())
		}
	}
}

func BenchmarkGenProofWithBlocks(b *testing.B) {
	data := make([]byte, 256*1024)
	var blocks []types.Block
	for i := 0; i < 1024; i++ {
		rand.Read(data)
		blocks = append(blocks, data)
	}

	pdp := NewPdp(MerklePdp)
	fileUniqueId, err := pdp.GenUniqueIdWithFileBlocks(blocks)
	if err != nil {
		b.Fatal(err.Error())
	}

	var nodeId [20]byte
	rand.Read(nodeId[:])

	blockHash := make([]byte, 32)
	rand.Read(blockHash[:])

	challenge, err := pdp.GenChallenge(nodeId, blockHash, 1024)
	if err != nil {
		b.Fatal(err.Error())
	}
	for i := 0; i < b.N; i++ {
		_, err := pdp.GenProofWithBlocks(blocks, fileUniqueId, challenge)
		if err != nil {
			b.Fatal(err.Error())
		}
	}
}

func BenchmarkVerifyProofWithUniqueId(b *testing.B) {
	data := make([]byte, 256*1024)
	var blocks []types.Block
	for i := 0; i < 1024; i++ {
		rand.Read(data)
		blocks = append(blocks, data)
	}

	pdp := NewPdp(MerklePdp)
	fileUniqueId, err := pdp.GenUniqueIdWithFileBlocks(blocks)
	if err != nil {
		b.Fatal(err.Error())
	}

	var nodeId [20]byte
	rand.Read(nodeId[:])

	blockHash := make([]byte, 32)
	rand.Read(blockHash[:])

	challenge, err := pdp.GenChallenge(nodeId, blockHash, 1024)
	if err != nil {
		b.Fatal(err.Error())
	}

	proof, err := pdp.GenProofWithBlocks(blocks, fileUniqueId, challenge)
	if err != nil {
		b.Fatal(err.Error())
	}
	for i := 0; i < b.N; i++ {
		if err = VerifyProofWithUniqueId(fileUniqueId, proof, challenge); err != nil {
			b.Fatal(err.Error())
		}
	}
}
