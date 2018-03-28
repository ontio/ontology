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
)

	func doubleSha256(s []Uint256) Uint256 {
		b := new(bytes.Buffer)
		for _, d := range s {
			d.Serialize(b)
		}
		temp := sha256.Sum256(b.Bytes())
		f := sha256.Sum256(temp[:])
		return Uint256(f)
	}

type MerkleTree struct {
	Depth uint
	Root  *MerkleTreeNode
}

type MerkleTreeNode struct {
	Hash  Uint256
	Left  *MerkleTreeNode
	Right *MerkleTreeNode
}

func (t *MerkleTreeNode) IsLeaf() bool {
	return t.Left == nil && t.Right == nil
}

//use []Uint256 to create a new MerkleTree
func NewMerkleTree(hashes []Uint256) (*MerkleTree, error) {
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
	mt := &MerkleTree{
		Root:  nodes[0],
		Depth: height,
	}
	return mt, nil

}

//Generate the leaves nodes
func generateLeaves(hashes []Uint256) []*MerkleTreeNode {
	var leaves []*MerkleTreeNode
	for _, d := range hashes {
		node := &MerkleTreeNode{
			Hash: d,
		}
		leaves = append(leaves, node)
	}
	return leaves
}

//calc the next level's hash use double sha256
func levelUp(nodes []*MerkleTreeNode) []*MerkleTreeNode {
	var nextLevel []*MerkleTreeNode
	for i := 0; i < len(nodes)/2; i++ {
		var data []Uint256
		data = append(data, nodes[i*2].Hash)
		data = append(data, nodes[i*2+1].Hash)
		hash := doubleSha256(data)
		node := &MerkleTreeNode{
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
		node := &MerkleTreeNode{
			Hash:  hash,
			Left:  nodes[len(nodes)-1],
			Right: nodes[len(nodes)-1],
		}
		nextLevel = append(nextLevel, node)
	}
	return nextLevel
}

//input a []uint256, create a MerkleTree & calc the root hash
func ComputeRoot(hashes []Uint256) (Uint256, error) {
	if len(hashes) == 0 {
		return Uint256{}, errors.New("NewMerkleTree input no item error.")
	}
	if len(hashes) == 1 {
		return hashes[0], nil
	}
	tree, _ := NewMerkleTree(hashes)
	return tree.Root.Hash, nil
}
