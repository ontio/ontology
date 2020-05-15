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

	"github.com/ontio/ontology/smartcontract/service/native/ontfs/pdp/types"
)

const (
	PROOF_LEN = 192
	MAX_DEPTH = 65536
)

//BLOCK_LEN should equal to hash len
func merkleHash(depth, indexA, indexB uint64, nodeA, nodeB []byte, res []byte) ([]byte, error) {
	if depth > MAX_DEPTH {
		return nil, fmt.Errorf("merkleHash depth should be less than 65536")
	}
	id := make([]byte, 8)
	ia := make([]byte, 8)
	ib := make([]byte, 8)
	binary.LittleEndian.PutUint64(id, depth)
	binary.LittleEndian.PutUint64(ia, indexA)
	binary.LittleEndian.PutUint64(ib, indexB)

	h := sha3.New256()
	h.Write(id)
	h.Write(ia)
	h.Write(nodeA)
	h.Write(ib)
	h.Write(nodeB)
	return h.Sum(res[:0]), nil
}

func MerkleProof(blocks []types.Block, c uint64) ([][]byte, error) {
	blocksLen := uint64(len(blocks))
	if blocksLen == 0 {
		return nil, fmt.Errorf("MerkleProof blocksLen error")
	}
	if c >= blocksLen {
		return nil, fmt.Errorf("MerkleProof challenge pos error")
	}
	layerHashes := make([][]byte, blocksLen)
	for i := uint64(0); i < blocksLen; i++ {
		h := sha3.Sum256(blocks[i])

		id := make([]byte, 8)
		binary.LittleEndian.PutUint64(id, i)

		var layer []byte
		layer = append(layer, h[:]...)
		layer = append(layer, id[:]...)
		layerHashes[i] = layer[:]
	}

	var depth uint64 = 0
	res := [][]byte{layerHashes[c]}

	for {
		layerHashesLen := uint64(len(layerHashes))
		if layerHashesLen == 1 {
			break
		}
		n := layerHashesLen / 2

		for i := uint64(0); i < n; i++ {
			if 2*i == c {
				res = append(res, layerHashes[2*i+1])
			} else if 2*i+1 == c {
				res = append(res, layerHashes[2*i])
			}
			tmp, err := merkleHash(depth, 2*i, 2*i+1, layerHashes[2*i][:], layerHashes[2*i+1][:], nil)
			if err != nil {
				return nil, fmt.Errorf("MerkleProof error: %s", err.Error())
			}
			layerHashes[i] = tmp
		}
		if layerHashesLen == 2*n+1 {
			if 2*n == c {
				res = append(res, layerHashes[2*n])
			}
			tmp, err := merkleHash(depth, 2*n, 2*n+1, layerHashes[2*n][:], layerHashes[2*n][:], nil)
			if err != nil {
				return nil, fmt.Errorf("MerkleProof error: %s", err.Error())
			}
			layerHashes[n] = tmp
			layerHashes = layerHashes[:n+1]
		} else {
			layerHashes = layerHashes[:n]
		}
		depth++
		c = c / 2
	}
	res = append(res, layerHashes[0])
	return res, nil
}

func VerifyMerkleProof(proof [][]byte, uniqueId []byte, i uint64) error {
	var err error
	proofLength := uint64(len(proof))
	if proofLength == 0 {
		return fmt.Errorf("VerifyMerkleProof proof length error")
	}
	root := proof[proofLength-1]
	if !bytes.Equal(root, uniqueId) {
		return fmt.Errorf("VerifyMerkleProof root hash not equal")
	}

	n := proof[0]
	nLen := len(n)
	if nLen < 8 {
		return fmt.Errorf("VerifyMerkleProof proof length error")
	}
	challengeIndex := binary.LittleEndian.Uint64(n[nLen-8:])
	if challengeIndex != i {
		return fmt.Errorf("VerifyMerkleProof proof challenge index error: %d", challengeIndex)
	}

	siblings := proof[1 : proofLength-1]
	depth := uint64(0)
	c := i
	for _, s := range siblings {
		if c%2 == 0 {
			n, err = merkleHash(depth, c, c+1, n, s, nil)
			if err != nil {
				return fmt.Errorf("VerifyMerkleProof error: %s", err.Error())
			}
		} else {
			n, err = merkleHash(depth, c-1, c, s, n, nil)
			if err != nil {
				return fmt.Errorf("VerifyMerkleProof error: %s", err.Error())
			}
		}
		depth++
		c = c / 2
	}
	if !bytes.Equal(root, n) {
		return fmt.Errorf("proof verify failed")
	}
	return nil
}

func CalcRootHash(blocks []types.Block) ([]byte, error) {
	var err error
	blocksLen := uint64(len(blocks))
	if blocksLen == 0 {
		return nil, fmt.Errorf("CalcRootHash blocksLen error")
	}
	layerHashes := make([][]byte, blocksLen)
	for i := uint64(0); i < blocksLen; i++ {
		h := sha3.Sum256(blocks[i])

		id := make([]byte, 8)
		binary.LittleEndian.PutUint64(id, i)

		var layer []byte
		layer = append(layer, h[:]...)
		layer = append(layer, id[:]...)
		layerHashes[i] = layer[:]
	}

	depth := uint64(0)
	for {
		layerHashesLen := uint64(len(layerHashes))
		if layerHashesLen == 1 {
			break
		}

		n := layerHashesLen / 2
		for i := uint64(0); i < n; i++ {
			layerHashes[i], err = merkleHash(depth, 2*i, 2*i+1, layerHashes[2*i][:], layerHashes[2*i+1][:], nil)
			if err != nil {
				return nil, fmt.Errorf("CalcRootHash error: %s", err.Error())
			}
		}
		if layerHashesLen == 2*n+1 {
			layerHashes[n], err = merkleHash(depth, 2*n, 2*n+1, layerHashes[2*n][:], layerHashes[2*n][:], nil)
			if err != nil {
				return nil, fmt.Errorf("CalcRootHash error: %s", err.Error())
			}
			layerHashes = layerHashes[:n+1]
		} else {
			layerHashes = layerHashes[:n]
		}
		//fmt.Printf("depth %d hash %x\n", depth, nodes)
		depth++
	}
	return layerHashes[0], nil
}
