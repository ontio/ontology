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
	crand "crypto/rand"
	"testing"
	"bytes"
)

func TestSimpleProof(t *testing.T) {
	trie, vals := simpleTrie()
	root := trie.Hash()
	for _, kv := range vals {
		proof := trie.Prove(kv.k)
		if proof == nil {
			t.Fatalf("missing key %x while constructing proof", kv.k)
		}
		val, err := VerifyProof(root, kv.k, proof)
		if err != nil {
			t.Fatalf("VerifyProof error for key %x: %v\nraw proof: %x", kv.k, err, proof)
		}
		if !bytes.Equal(val, kv.v) {
			t.Fatalf("VerifyProof returned wrong value for key %x: got %x, want %x", kv.k, val, kv.v)
		}
	}
	t.Log("Test Simple Trie Proof Successful")
}

func TestRandomTrie(t *testing.T) {
	trie, vals := randomTrie(500)
	root := trie.Hash()
	for _, kv := range vals {
		proof := trie.Prove(kv.k)
		if proof == nil {
			t.Fatalf("missing key %x while constructing proof", kv.k)
		}
		val, err := VerifyProof(root, kv.k, proof)
		if err != nil {
			t.Fatalf("VerifyProof error for key %x: %v\nraw proof: %x", kv.k, err, proof)
		}
		if !bytes.Equal(val, kv.v) {
			t.Fatalf("VerifyProof returned wrong value for key %x: got %x, want %x", kv.k, val, kv.v)
		}
	}
	t.Log("Test Random Trie Proof Successful")
}

type kv struct {
	k, v []byte
	t    bool
}

func simpleTrie() (*Trie, map[string]*kv) {
	trie := new(Trie)
	vals := make(map[string]*kv)
	trie.Update([]byte("1"), []byte("1"))
	vals["1"] = &kv{[]byte("1"), []byte("1"), false}
	trie.Update([]byte("12"), []byte("12"))
	vals["12"] = &kv{[]byte("12"), []byte("12"), false}
	trie.Update([]byte("123"), []byte("123"))
	vals["123"] = &kv{[]byte("123"), []byte("123"), false}
	trie.Update([]byte("1234"), []byte("1234"))
	vals["1234"] = &kv{[]byte("1234"), []byte("1234"), false}
	return trie, vals
}

func randomTrie(n int) (*Trie, map[string]*kv) {
	trie := new(Trie)
	vals := make(map[string]*kv)
	for i := 0; i < 100; i++ {
		value1 := &kv{[]byte{byte(i)}, []byte{byte(i)}, false}
		value2 := &kv{[]byte{byte(i + 10)}, []byte{byte(i)}, false}
		trie.Update(value1.k, value1.v)
		trie.Update(value2.k, value2.v)
		vals[string(value1.k)] = value1
		vals[string(value2.k)] = value2
	}

	for i := 0; i < n; i++ {
		value := &kv{randBytes(32), randBytes(20), false}
		trie.Update(value.k, value.v)
		vals[string(value.k)] = value
	}
	return trie, vals
}

func randBytes(n int) []byte {
	r := make([]byte, n)
	crand.Read(r)
	return r
}