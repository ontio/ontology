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
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"fmt"
)

type kv struct {
	k, v []byte
	t    bool
}

func TestSimpleProof(t *testing.T) {
	trie, vals := simpleTrie()
	root := trie.Hash()
	for _, kv := range vals {
		proves := trie.Prove(kv.k)
		val, err := VerifyProof(root, kv.k, proves)
		assert.Nil(t, err)
		assert.Equal(t, val, kv.v)
	}
	t.Log("Test Simple Trie Proof Successful")
}

func simpleTrie() (*Trie, map[string]*kv) {
	trie := new(Trie)
	vals := make(map[string]*kv)
	trie.Update([]byte("111"), []byte("111"))
	vals["111"] = &kv{[]byte("111"), []byte("111"), false}
	trie.Update([]byte("222"), []byte("222"))
	vals["222"] = &kv{[]byte("222"), []byte("222"), false}
	trie.Update([]byte("333"), []byte("333"))
	vals["333"] = &kv{[]byte("333"), []byte("333"), false}
	trie.Update([]byte("444"), []byte("444"))
	vals["444"] = &kv{[]byte("444"), []byte("444"), false}
	return trie, vals
}

func TestRandomProof(t *testing.T) {
	trie, vals := createRandomTrie(100)
	root := trie.Hash()
	for _, kv := range vals {
		proves := trie.Prove(kv.k)
		val, err := VerifyProof(root, kv.k, proves)
		assert.Nil(t, err)
		assert.Equal(t, val, kv.v)
	}
	t.Log("Test Random Trie Proof Successful")
}

func createRandomTrie(n int) (*Trie, map[string]*kv) {
	trie := new(Trie)
	vals := make(map[string]*kv)
	for i := 0; i < n; i++ {
		key := randBytes(32)
		value := randBytes(20)
		trie.TryUpdate(key, value)
		vals[string(key)] = &kv{key, []byte("123"), false}
	}
	return trie, vals
}

func randBytes(n int) []byte {
	r := make([]byte, n)
	rand.Read(r)
	return r
}
