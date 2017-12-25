package merkle

import (
	"crypto/sha256"

	. "github.com/Ontology/common"
)

type TreeHasher struct {
}

func (self TreeHasher) hash_empty() Uint256 {
	return sha256.Sum256(nil)
}

func (self TreeHasher) hash_leaf(data []byte) Uint256 {
	tmp := append([]byte{0}, data...)
	return sha256.Sum256(tmp)
}

func (self TreeHasher) hash_children(left, right Uint256) Uint256 {
	data := append([]byte{1}, left[:]...)
	data = append(data, right[:]...)
	return sha256.Sum256(data)
}

func (self TreeHasher) HashFullTreeWithLeafHash(leaves []Uint256) Uint256 {
	length := uint32(len(leaves))
	root_hash, hashes := self._hash_full(leaves, 0, length)

	if uint(len(hashes)) != countBit(length) {
		panic("assert failed in hash full tree")
	}

	// assert len(hashes) == countBit(len(leaves))
	// assert self._hash_fold(hashes) == root_hash if hashes else root_hash == self.hash_empty()

	return root_hash
}
func (self TreeHasher) HashFullTree(leaves [][]byte) Uint256 {
	length := uint32(len(leaves))
	leafhashes := make([]Uint256, length, length)
	for i := range leaves {
		leafhashes[i] = self.hash_leaf(leaves[i])
	}
	root_hash, hashes := self._hash_full(leafhashes, 0, length)

	if uint(len(hashes)) != countBit(length) {
		panic("assert failed in hash full tree")
	}

	// assert len(hashes) == countBit(len(leaves))
	// assert self._hash_fold(hashes) == root_hash if hashes else root_hash == self.hash_empty()

	return root_hash
}

func (self TreeHasher) _hash_full(leaves []Uint256, l_idx, r_idx uint32) (root_hash Uint256, hashes []Uint256) {
	width := r_idx - l_idx
	if width == 0 {
		return self.hash_empty(), nil
	} else if width == 1 {
		leaf_hash := leaves[l_idx]
		return leaf_hash, []Uint256{leaf_hash}
	} else {
		var split_width uint32 = 1 << (countBit(width - 1) - 1)
		l_root, l_hashes := self._hash_full(leaves, l_idx, l_idx + split_width)
		if len(l_hashes) != 1 {
			panic("left tree always full")
		}
		r_root, r_hashes := self._hash_full(leaves, l_idx + split_width, r_idx)
		root_hash = self.hash_children(l_root, r_root)
		var hashes []Uint256
		if split_width * 2 == width {
			hashes = []Uint256{root_hash}
		} else {
			hashes = append(l_hashes, r_hashes[:]...)
		}
		return root_hash, hashes
	}
}

func (self TreeHasher) _hash_fold(hashes []Uint256) Uint256 {
	l := len(hashes)
	accum := hashes[l - 1]
	for i := l - 2; i >= 0; i-- {
		accum = self.hash_children(hashes[i], accum)
	}

	return accum
}
