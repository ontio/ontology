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
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"

	"github.com/ontio/ontology/common"
)

const (
	LEFT byte = iota
	RIGHT
	MAX_SIZE = 1024 * 1024
)

var debugCheck = false

type TreeHasher struct {
}

func (self TreeHasher) hash_empty() common.Uint256 {
	return sha256.Sum256(nil)
}

func (self TreeHasher) hash_leaf(data []byte) common.Uint256 {
	tmp := append([]byte{0}, data...)
	return sha256.Sum256(tmp)
}

func (self TreeHasher) hash_children(left, right common.Uint256) common.Uint256 {
	data := append([]byte{1}, left[:]...)
	data = append(data, right[:]...)
	return sha256.Sum256(data)
}

func (self TreeHasher) HashFullTreeWithLeafHash(leaves []common.Uint256) common.Uint256 {
	length := uint32(len(leaves))
	root_hash, hashes := self._hash_full(leaves, 0, length)

	if uint(len(hashes)) != countBit(length) {
		panic("hashes length mismatch")
	}

	if debugCheck {
		root2 := self.hash_empty()
		if len(hashes) != 0 {
			root2 = self._hash_fold(hashes)
		}

		if root_hash != root2 {
			panic("root hash mismatch")
		}
	}

	// assert len(hashes) == countBit(len(leaves))
	// assert self._hash_fold(hashes) == root_hash if hashes else root_hash == self.hash_empty()

	return root_hash
}

func (self TreeHasher) HashFullTree(leaves [][]byte) common.Uint256 {
	length := uint32(len(leaves))
	leafhashes := make([]common.Uint256, length, length)
	for i := range leaves {
		leafhashes[i] = self.hash_leaf(leaves[i])
	}

	return self.HashFullTreeWithLeafHash(leafhashes)
}

func (self TreeHasher) _hash_full(leaves []common.Uint256, l_idx, r_idx uint32) (root_hash common.Uint256, hashes []common.Uint256) {
	width := r_idx - l_idx
	if width == 0 {
		return self.hash_empty(), nil
	} else if width == 1 {
		leaf_hash := leaves[l_idx]
		return leaf_hash, []common.Uint256{leaf_hash}
	} else {
		var split_width uint32 = 1 << (highBit(width-1) - 1)
		l_root, l_hashes := self._hash_full(leaves, l_idx, l_idx+split_width)
		if len(l_hashes) != 1 {
			panic("left tree always full")
		}
		r_root, r_hashes := self._hash_full(leaves, l_idx+split_width, r_idx)
		root_hash = self.hash_children(l_root, r_root)
		var hashes []common.Uint256
		if split_width*2 == width {
			hashes = []common.Uint256{root_hash}
		} else {
			hashes = append(l_hashes, r_hashes[:]...)
		}
		return root_hash, hashes
	}
}

func (self TreeHasher) _hash_fold(hashes []common.Uint256) common.Uint256 {
	l := len(hashes)
	accum := hashes[l-1]
	for i := l - 2; i >= 0; i-- {
		accum = self.hash_children(hashes[i], accum)
	}

	return accum
}

func HashLeaf(data []byte) common.Uint256 {
	tmp := append([]byte{0}, data...)
	return sha256.Sum256(tmp)
}

func HashChildren(left, right common.Uint256) common.Uint256 {
	data := append([]byte{1}, left[:]...)
	data = append(data, right[:]...)
	return sha256.Sum256(data)
}

func MerkleLeafPath(data []byte, hashes []common.Uint256) ([]byte, error) {
	size := len(hashes)*(common.UINT256_SIZE+1) + len(data) + 8
	if size > MAX_SIZE {
		return nil, fmt.Errorf("data length over max value:%d", MAX_SIZE)
	}
	index := getIndex(HashLeaf(data), hashes)
	if index < 0 {
		return nil, fmt.Errorf("%s", "values doesn't exist!")
	}
	sink := common.NewZeroCopySink(make([]byte, 0, size))
	sink.WriteVarBytes(data)
	d := depth(len(hashes))
	merkleTree := MerkleHashes(hashes, d)
	for i := d; i > 0; i-- {
		subTree := merkleTree[i]
		subLen := len(subTree)
		nIndex := index / 2
		if index == subLen-1 && subLen%2 != 0 {
			index = nIndex
			continue
		}
		if index%2 != 0 {
			sink.WriteByte(LEFT)
			sink.WriteHash(subTree[index-1])
		} else {
			sink.WriteByte(RIGHT)
			sink.WriteHash(subTree[index+1])
		}
		index = nIndex
	}
	return sink.Bytes(), nil
}

func MerkleHashes(preLeaves []common.Uint256, depth int) [][]common.Uint256 {
	levels := make([][]common.Uint256, depth+1, depth+1)
	levels[depth] = preLeaves
	for i := depth; i > 0; i -= 1 {
		level := levels[i]
		levelLen := len(level)
		remainder := levelLen % 2
		nextLevel := make([]common.Uint256, levelLen/2+remainder)
		k := 0
		for j := 0; j < len(level)-1; j += 2 {
			left := level[j]
			right := level[j+1]

			nextLevel[k] = HashChildren(left, right)
			k += 1
		}
		if remainder != 0 {
			nextLevel[k] = level[len(level)-1]
		}
		levels[i-1] = nextLevel
	}
	return levels
}

func MerkleProve(path []byte, root common.Uint256) ([]byte, error) {
	source := common.NewZeroCopySource(path)
	value, _, irr, eof := source.NextVarBytes()
	if eof || irr {
		return nil, errors.New("read bytes error")
	}
	hash := HashLeaf(value)
	size := int((source.Size() - source.Pos()) / common.UINT256_SIZE)
	for i := 0; i < size; i++ {
		f, eof := source.NextByte()
		if eof {
			return nil, errors.New("read byte error")
		}
		v, eof := source.NextHash()
		if eof {
			return nil, errors.New("read hash error")
		}
		if f == LEFT {
			hash = HashChildren(v, hash)
		} else {
			hash = HashChildren(hash, v)
		}
	}

	if hash != root {
		return nil, fmt.Errorf("excepted root is not equal actual root, excepted:%x, actual:%x", hash, root)
	}
	return value, nil
}

func depth(n int) int {
	return int(math.Ceil(math.Log2(float64(n))))
}

func getIndex(leaf common.Uint256, hashes []common.Uint256) int {
	for i, v := range hashes {
		if bytes.Equal(v[:], leaf[:]) {
			return i
		}
	}
	return -1
}
