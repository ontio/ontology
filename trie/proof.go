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
	"bytes"
	"github.com/Ontology/common/log"
	"fmt"
	"github.com/Ontology/rlp"
	"github.com/Ontology/common"
	"errors"
)

func (t *Trie) Prove(key []byte) []rlp.RawValue {
	key = keyBytesToHex(key)
	nodes := []node{}
	tn := t.root
	for len(key) > 0 && tn != nil {
		switch n := tn.(type) {
		case *shortNode:
			if len(key) < len(n.Key) || !bytes.Equal(n.Key, key[:len(n.Key)]) {
				tn = nil
			} else {
				tn = n.Val
				key = key[len(n.Key):]
			}
			nodes = append(nodes, n)
		case *fullNode:
			tn = n.Children[key[0]]
			key = key[1:]
			nodes = append(nodes, n)
		case hashNode:
			var err error
			tn, err = t.resolveHash(n, nil)
			if err != nil {
				log.Error(fmt.Sprintf("[Prove] hashNode resolveHash error: %v", err))
			}
		default:
			panic(fmt.Sprintf("[Prove] Invalid node type :%v", tn))
		}
	}
	h := newHasher()
	proof := make([]rlp.RawValue, 0, len(nodes))
	for i, n := range nodes {
		n, _, _ = h.hasChildren(n, nil)
		hn, _ := h.store(n, nil, false)
		if _, ok := hn.(hashNode); ok || i == 0 {
			enc, _ := rlp.EncodeToBytes(n)
			proof = append(proof, enc)
		}
	}
	return proof
}

func VerifyProof(rootHash common.Uint256, key []byte, proof []rlp.RawValue) ([]byte, error) {
	key = keyBytesToHex(key)
	root := rootHash.ToArray()
	for i, buf := range proof {
		if !bytes.Equal(ToHash256(buf), root) {
			return nil, fmt.Errorf("[VerifyProof] bad proof node %d: hash mismatch", i)
		}
		n, err := decodeNode(root, buf)
		if err != nil {
			return nil, fmt.Errorf("[VerifyProof] bad proof node %d: %v", i, err)
		}
		keyRest, cld := get(n, key)
		switch cld := cld.(type) {
		case nil:
			if i != len(proof) - 1 {
				return nil, fmt.Errorf("[VerifyProof] key mismatch at proof node %d", i)
			} else {
				return nil, nil
			}
		case hashNode:
			key = keyRest
			root = cld
		case valueNode:
			if i != len(proof) - 1 {
				return nil, errors.New("[VerifyProof] additional nodes at end of proof")
			}
			return cld, nil
		}
	}
	return nil, errors.New("unexpected end of proof")
}

func get(tn node, key []byte) ([]byte, node) {
	for {
		switch n := tn.(type) {
		case *shortNode:
			if len(key) < len(n.Key) || !bytes.Equal(n.Key, key[:len(n.Key)]) {
				return nil, nil
			}
			tn = n.Val
			key = key[len(n.Key):]
		case *fullNode:
			tn = n.Children[key[0]]
			key = key[1:]
		case hashNode:
			return key, n
		case nil:
			return key, nil
		case valueNode:
			return nil, n
		default:
			panic(fmt.Sprintf("[get] Invalid node: %v", tn))
		}
	}
}
