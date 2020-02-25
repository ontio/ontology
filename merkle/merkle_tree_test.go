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

package merkle

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func TestMerkleLeaf3(t *testing.T) {
	hasher := TreeHasher{}
	leafs := []common.Uint256{hasher.hash_leaf([]byte{1}),
		hasher.hash_leaf([]byte{2}),
		hasher.hash_leaf([]byte{3})}
	store, _ := NewFileHashStore("merkletree.db", 0)
	tree := NewTree(0, nil, store)
	if tree.Root() != sha256.Sum256(nil) {
		t.Fatal("root error")
	}
	for i := range leafs {
		tree.Append([]byte{byte(i + 1)})
	}

	hashes := make([]common.Uint256, 5, 5)
	for i := 0; i < 4; i++ {
		hashes[i], _ = tree.hashStore.GetHash(uint32(i))
	}
	hashes[4] = tree.Root()

	cmp := []common.Uint256{
		leafs[0],
		leafs[1],
		hasher.hash_children(leafs[0], leafs[1]),
		leafs[2],
		hasher.hash_children(hasher.hash_children(leafs[0], leafs[1]),
			leafs[2]),
	}

	for i := 0; i < 5; i++ {
		if hashes[i] != cmp[i] {
			t.Fatal(fmt.Sprintf("error: %d, expected %x, found %x", i, cmp[i], hashes[i]))
		}
	}

}

func TestCompactMerkleTree_GetRootWithNewLeaves(t *testing.T) {
	N := 1000
	tree1 := NewTree(0, nil, nil)
	tree2 := NewTree(0, nil, nil)
	leaves := make([]common.Uint256, N)
	for i := 0; i < N; i++ {
		leaves[i][:][0] = byte(i)
		hash := leaves[i]
		assert.Equal(t, tree1.GetRootWithNewLeaf(hash), tree2.GetRootWithNewLeaves([]common.Uint256{hash}))
		tree1.AppendHash(hash)
		tree2.AppendHash(hash)
	}
}

func TestMerkle(t *testing.T) {
	hasher := TreeHasher{}
	leafs := []common.Uint256{hasher.hash_leaf([]byte{1}),
		hasher.hash_leaf([]byte{2}),
		hasher.hash_leaf([]byte{3}),
		hasher.hash_leaf([]byte{4})}

	store, _ := NewFileHashStore("merkletree.db", 0)
	tree := NewTree(0, nil, store)
	if tree.Root() != sha256.Sum256(nil) {
		t.Fatal("root error")
	}
	for i := range leafs {
		tree.Append([]byte{byte(i + 1)})
	}

	hashes := make([]common.Uint256, 6, 6)
	for i := 0; i < 6; i++ {
		hashes[i], _ = tree.hashStore.GetHash(uint32(i))
	}
	cmp := []common.Uint256{
		leafs[0],
		leafs[1],
		hasher.hash_children(leafs[0], leafs[1]),
		leafs[2],
		leafs[3],
		hasher.hash_children(leafs[2], leafs[3]),
		hasher.hash_children(hasher.hash_children(leafs[0], leafs[1]),
			hasher.hash_children(leafs[2], leafs[3])),
	}

	for i := 0; i < 6; i++ {
		if hashes[i] != cmp[i] {
			fmt.Println(hashes)
			fmt.Println(cmp)
			t.Fatal(fmt.Sprintf("error: %d, expected %x, found %x", i, cmp[i], hashes[i]))
		}
	}

}

func TestMerkleHashes(t *testing.T) {
	store, _ := NewFileHashStore("merkletree.db", 0)
	tree := NewTree(0, nil, store)
	for i := 0; i < 100; i++ {
		tree.Append([]byte{byte(i + 1)})
	}

	// 100 == 110 0100
	if len(tree.hashes) != 3 {
		t.Fatal(fmt.Sprintf("error tree hashes size"))
	}

}

// zero based return merkle root of D[0:n]
func TestMerkleRoot(t *testing.T) {
	n := 100
	roots := make([]common.Uint256, n, n)
	store, _ := NewFileHashStore("merkletree.db", 0)
	tree := NewTree(0, nil, store)
	for i := 0; i < n; i++ {
		tree.Append([]byte{byte(i + 1)})
		roots[i] = tree.Root()
	}

	cmp := make([]common.Uint256, n, n)
	for i := 0; i < n; i++ {
		cmp[i] = tree.merkleRoot(uint32(i) + 1)
		if cmp[i] != roots[i] {
			t.Error(fmt.Sprintf("error merkle root is not equal at %d", i))
		}
	}

}

func TestGetSubTreeSize(t *testing.T) {
	sizes := getSubTreeSize(7)
	fmt.Println("sub tree size", sizes)
}

// zero based return merkle root of D[0:n]
func TestMerkleIncludeProof(t *testing.T) {
	n := uint32(9)
	store, _ := NewFileHashStore("merkletree.db", 0)
	tree := NewTree(0, nil, store)
	for i := uint32(0); i < n; i++ {
		tree.Append([]byte{byte(i + 1)})
	}

	verify := NewMerkleVerifier()

	root := tree.Root()
	for i := uint32(0); i < n; i++ {
		proof, _ := tree.InclusionProof(i, n)
		leaf_hash := tree.hasher.hash_leaf([]byte{byte(i + 1)})
		res := verify.VerifyLeafHashInclusion(leaf_hash, i, proof, root, n)
		if res != nil {
			t.Fatal(res, i, proof)
		}
	}
}

func TestMerkleConsistencyProofLen(t *testing.T) {
	n := uint32(7)
	store, _ := NewFileHashStore("merkletree.db", 0)
	tree := NewTree(0, nil, store)
	for i := uint32(0); i < n; i++ {
		tree.Append([]byte{byte(i + 1)})
	}

	cmp := []int{3, 2, 4, 1, 4, 3, 0}
	for i := uint32(0); i < n; i++ {
		proof := tree.ConsistencyProof(i+1, n)
		if len(proof) != cmp[i] {
			t.Fatal("error: wrong proof length")
		}
	}

}

func TestMerkleConsistencyProof(t *testing.T) {
	n := uint32(140)
	roots := make([]common.Uint256, n, n)
	store, _ := NewFileHashStore("merkletree.db", 0)
	tree := NewTree(0, nil, store)
	for i := uint32(0); i < n; i++ {
		tree.Append([]byte{byte(i + 1)})
		roots[i] = tree.Root()
	}

	verify := NewMerkleVerifier()

	for i := uint32(0); i < n; i++ {
		proof := tree.ConsistencyProof(i+1, n)
		err := verify.VerifyConsistency(i+1, n, roots[i], roots[n-1], proof)
		if err != nil {
			t.Fatal("verify consistency error:", i, err)
		}

	}
}

//~70w
func BenchmarkMerkleInsert(b *testing.B) {
	store, _ := NewFileHashStore("merkletree.db", 0)
	tree := NewTree(0, nil, store)
	for i := 0; i < b.N; i++ {
		//use b.N for looping
		tree.Append([]byte(fmt.Sprintf("bench %d", i)))
	}
}

var treeTest *CompactMerkleTree
var storeTest HashStore
var N = 100 //00

func init() {
	storeTest, _ := NewFileHashStore("merkletree.db", 0)
	treeTest := NewTree(0, nil, storeTest)
	for i := 0; i < N; i++ {
		treeTest.Append([]byte(fmt.Sprintf("setup %d", i)))
	}

}

/*
// ~20w
func BenchmarkMerkleInclusionProof(b *testing.B) {
	for i := 0; i < b.N; i++ {
		treeTest.InclusionProof(uint32(i), uint32(N))
	}
}

// ~20w
func BenchmarkMerkleConsistencyProof(b *testing.B) {
	for i := 0; i < b.N; i++ {
		treeTest.ConsistencyProof(uint32(i+1), uint32(N))
	}
}
*/

//~70w
func BenchmarkMerkleInsert2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		treeTest.Append([]byte(fmt.Sprintf("bench %d", i)))
	}
}

func TestTreeHasher_HashFullTree(t *testing.T) {
	debugCheck = true
	leaves := make([][]byte, 0)
	for i := byte(0); i < 200; i++ {
		leaves = append(leaves, []byte{i})
		TreeHasher{}.HashFullTree(leaves)
	}
}

func TestTreeHasher(t *testing.T) {
	tree := NewTree(0, nil, nil)
	leaves := make([][]byte, 0)
	for i := uint32(0); i < 1000; i++ {
		leaf := []byte{byte(i + 1)}
		leaves = append(leaves, leaf)
		tree.Append(leaf)
		root := TreeHasher{}.HashFullTree(leaves)
		assert.Equal(t, root, tree.Root())
	}
}

func TestAudit(t *testing.T) {
	var hashes []common.Uint256
	n := 10
	tree := TreeHasher{}
	for i := 0; i < n; i++ {
		hashes = append(hashes, HashLeaf([]byte(fmt.Sprintf("%d", i))))
	}
	root := tree.HashFullTreeWithLeafHash(hashes)
	treeHashes := MerkleHashes(hashes, depth(len(hashes)))
	assert.Equal(t, root, treeHashes[0][0])
	for i := 0; i < n; i++ {
		auditPath, err := MerkleLeafPath([]byte(fmt.Sprintf("%d", i)), hashes)
		assert.NoError(t, err)
		value, err := MerkleProve(auditPath, root)
		assert.NoError(t, err)
		assert.Equal(t, []byte(fmt.Sprintf("%d", i)), value)
	}
}
