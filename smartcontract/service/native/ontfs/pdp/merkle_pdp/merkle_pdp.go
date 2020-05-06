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
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/crypto/sha3"
)

const MerkleRootLen = 32

func merkleHash(depth, indexA, indexB uint64, nodeA, nodeB []byte, res []byte) []byte {
	if depth > 256 {
		panic("depth should be less than 256")
	}
	d := []byte{byte(depth)}
	ia := make([]byte, 8)
	ib := make([]byte, 8)
	binary.LittleEndian.PutUint64(ia, indexA)
	binary.LittleEndian.PutUint64(ib, indexB)

	h := sha3.New256()
	h.Write(d)
	h.Write(ia)
	h.Write(nodeA)
	h.Write(ib)
	h.Write(nodeB)
	return h.Sum(res)
}

func MerkleProof(block []byte, c uint64) ([][]byte, error) {
	data := make([]byte, 64)
	if len(block) < 64 {
		copy(data, block[:])
	} else {
		copy(data, block[0:64])
	}
	size := uint64(len(data))
	if size%32 != 0 {
		return nil, fmt.Errorf("data length should be multiple of 32")
	}

	res := [][]byte{data[c*32 : (c+1)*32]}
	nodes := data
	depth := uint64(0)
	for ; size > 32; size = size / 2 {
		if size%64 != 0 {
			return nil, fmt.Errorf("data length should be power of 2")
		}
		tmp := make([]byte, 0, size/2)
		for pos := uint64(0); pos < size/32; pos += 2 {
			if pos == c {
				res = append(res, nodes[(pos+1)*32:(pos+2)*32])
			} else if pos+1 == c {
				res = append(res, nodes[pos*32:(pos+1)*32])
			}
			tmp = merkleHash(depth, pos, pos+1, nodes[pos*32:(pos+1)*32], nodes[(pos+1)*32:(pos+2)*32], tmp)
		}

		depth++
		nodes = tmp
		c = c / 2
	}
	res = append(res, nodes)

	return res, nil
}

func VerifyMerkleProof(proof [][]byte, uniqueId []byte, i uint64) error {
	root := proof[len(proof)-1]
	if !bytes.Equal(root, uniqueId) {
		return fmt.Errorf("[VerifyMerkleProof] rootHash not match")
	}
	n := proof[0]
	siblings := proof[1 : len(proof)-1]
	depth := uint64(0)
	c := i
	for _, s := range siblings {
		if c%2 == 0 {
			n = merkleHash(depth, c, c+1, n, s, nil)
		} else {
			n = merkleHash(depth, c-1, c, s, n, nil)
		}
		depth++
		c = c / 2
	}
	if !bytes.Equal(root, n) {
		return fmt.Errorf("[VerifyMerkleProof] verify failed")
	}
	return nil
}

func CalcRootHash(data []byte) ([]byte, error) {
	proof, err := MerkleProof(data, 1)
	if err != nil {
		return nil, err
	}
	return proof[len(proof)-1], nil
}
