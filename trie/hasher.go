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

package trie

import (
	"github.com/Ontology/common"
	"bytes"
	"github.com/Ontology/rlp"
	"sync"
	"crypto/sha256"
)

type hasher struct {
	tmp *bytes.Buffer
	sha common.Uint256
}

var hasherPool = sync.Pool{
	New: func() interface{} {
		return &hasher{tmp: new(bytes.Buffer), sha: common.Uint256{}}
	},
}

func (h *hasher) hash(n node, db DatabaseWriter, force bool) (node, node, error) {
	if hash, dirty := n.cache(); hash != nil {
		if !dirty {
			return hash, n, nil
		}
	}
	collapsed, cached, err := h.hasChildren(n, db)
	if err != nil {
		return nil, n, err
	}
	hashed, err := h.store(collapsed, db, force)
	if err != nil {
		return nil, n, err
	}
	cachedHash, _ := hashed.(hashNode)
	switch cn := cached.(type) {
	case *shortNode:
		cn.flags.hash = cachedHash
		cn.flags.dirty = false
	case *fullNode:
		cn.flags.hash = cachedHash
		cn.flags.dirty = false
	}
	return hashed, cached, nil
}

func (h *hasher) hasChildren(original node, db DatabaseWriter) (node, node, error) {
	var err error
	switch n := original.(type) {
	case *shortNode:
		collapsed, cached := n.copy(), n.copy()
		collapsed.Key = hexToCompact(n.Key)
		cached.Key = common.CopyBytes(n.Key)
		if _, ok := n.Val.(valueNode); !ok {
			collapsed.Val, cached.Val, err = h.hash(n.Val, db, false)
			if err != nil {
				return original, original, err
			}
		}
		if collapsed.Val == nil {
			collapsed.Val = valueNode(nil)
		}
		return collapsed, cached, nil
	case *fullNode:
		collapsed, cached := n.copy(), n.copy()
		for i := 0; i < 16; i++ {
			if n.Children[i] != nil {
				collapsed.Children[i], cached.Children[i], err = h.hash(n.Children[i], db, false)
				if err != nil {
					return original, original, err
				}
			} else {
				collapsed.Children[i] = valueNode(nil)
			}
		}
		cached.Children[16] = n.Children[16]
		if collapsed.Children[16] == nil {
			collapsed.Children[16] = valueNode(nil)
		}
		return collapsed, cached, nil
	default:
		return n, original, nil
	}
}

func (h *hasher) store(n node, db DatabaseWriter, force bool) (node, error) {
	if _, isHash := n.(hashNode); n == nil || isHash {
		return n, nil
	}
	h.tmp.Reset()
	if err := rlp.Encode(h.tmp, n); err != nil {
		panic("enocde error:" + err.Error())
	}
	if h.tmp.Len() < 32 && !force {
		return n, nil
	}
	hs, _ := n.cache()
	if hs == nil {
		u256 := ToHash256(h.tmp.Bytes())
		hs = hashNode(u256[:])
	}
	if db != nil {
		return hs, db.BatchPut(append(secureKeyPrefix, []byte(hs)...), h.tmp.Bytes())
	}
	return hs, nil
}

func newHasher() *hasher {
	h := hasherPool.Get().(*hasher)
	return h
}

func returnHasherToPool(h *hasher) {
	hasherPool.Put(h)
}

func ToHash256(bs []byte) []byte {
	temp := sha256.Sum256([]byte(bs))
	u256 := sha256.Sum256(temp[:])
	return u256[:]
}