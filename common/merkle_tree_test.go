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
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
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

const N = 120000

func BenchmarkComputeMerkleRoot(b *testing.B) {
	data := make([]Uint256, N)
	for i := range data {
		data[i] = Uint256(sha256.Sum256([]byte(fmt.Sprint(i))))
	}

	for i := 0; i < b.N; i++ {
		ComputeMerkleRoot(data)
	}
}

func BenchmarkComputeMerkleRootOld(b *testing.B) {
	data := make([]Uint256, N)
	for i := range data {
		data[i] = Uint256(sha256.Sum256([]byte(fmt.Sprint(i))))
	}

	for i := 0; i < b.N; i++ {
		computeMerkleRootOld(data)
	}
}

func TestComputeMerkleRoot(t *testing.T) {
	for n := 0; n < 100; n++ {
		data := make([]Uint256, n)
		for i := range data {
			data[i] = Uint256(sha256.Sum256([]byte(fmt.Sprint(i))))
		}

		h1 := computeMerkleRootOld(data)

		h2 := ComputeMerkleRoot(data)
		assert.Equal(t, h1, h2)
	}
}

func doubleSha256(s []Uint256) Uint256 {
	b := new(bytes.Buffer)
	for _, d := range s {
		d.Serialize(b)
	}
	temp := sha256.Sum256(b.Bytes())
	f := sha256.Sum256(temp[:])

	return Uint256(f)
}

type merkleTree struct {
	Depth uint
	Root  *merkleTreeNode
}

type merkleTreeNode struct {
	Hash  Uint256
	Left  *merkleTreeNode
	Right *merkleTreeNode
}

func (t *merkleTreeNode) IsLeaf() bool {
	return t.Left == nil && t.Right == nil
}

//use []Uint256 to create a new merkleTree
func newMerkleTree(hashes []Uint256) (*merkleTree, error) {
	if len(hashes) == 0 {
		return nil, errors.New("NewMerkleTree input no item error.")
	}
	var height uint

	height = 1
	nodes := generateLeaves(hashes)
	for len(nodes) > 1 {
		nodes = levelUp(nodes)
		height += 1
	}
	mt := &merkleTree{
		Root:  nodes[0],
		Depth: height,
	}
	return mt, nil

}

//Generate the leaves nodes
func generateLeaves(hashes []Uint256) []*merkleTreeNode {
	var leaves []*merkleTreeNode
	for _, d := range hashes {
		node := &merkleTreeNode{
			Hash: d,
		}
		leaves = append(leaves, node)
	}
	return leaves
}

//calc the next level's hash use double sha256
func levelUp(nodes []*merkleTreeNode) []*merkleTreeNode {
	var nextLevel []*merkleTreeNode
	for i := 0; i < len(nodes)/2; i++ {
		var data []Uint256
		data = append(data, nodes[i*2].Hash)
		data = append(data, nodes[i*2+1].Hash)
		hash := doubleSha256(data)
		node := &merkleTreeNode{
			Hash:  hash,
			Left:  nodes[i*2],
			Right: nodes[i*2+1],
		}
		nextLevel = append(nextLevel, node)
	}
	if len(nodes)%2 == 1 {
		var data []Uint256
		data = append(data, nodes[len(nodes)-1].Hash)
		data = append(data, nodes[len(nodes)-1].Hash)
		hash := doubleSha256(data)
		node := &merkleTreeNode{
			Hash:  hash,
			Left:  nodes[len(nodes)-1],
			Right: nodes[len(nodes)-1],
		}
		nextLevel = append(nextLevel, node)
	}
	return nextLevel
}

//input a []uint256, create a merkleTree & calc the root hash
func computeMerkleRootOld(hashes []Uint256) Uint256 {
	if len(hashes) == 0 {
		return Uint256{}
	}
	if len(hashes) == 1 {
		return hashes[0]
	}
	tree, _ := newMerkleTree(hashes)
	return tree.Root.Hash
}
