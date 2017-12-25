package trie

import (
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"fmt"
)

var secureKeyPrefix = []byte("secure-key-")

const secureKeyLength = 11 + 32

type SecureTrie struct {
	trie             Trie
	hashKeyBuf       [secureKeyLength]byte
	secKeyBuf        [200]byte
	secKeyCache      map[string][]byte
	secKeyCacheOwner *SecureTrie
}

func NewSecure(root common.Uint256, db Database) (*SecureTrie, error) {
	if db == nil {
		panic("NewSecure called with nil database")
	}
	trie, err := New(root, db)
	if err != nil {
		return nil, err
	}
	return &SecureTrie{trie: *trie}, nil
}

func (t *SecureTrie) Get(key []byte) []byte {
	res, err := t.TryGet(key)
	if err != nil {
		log.Error(fmt.Sprintf("Unhandled trie error: %v", err))
	}
	return res
}

func (t *SecureTrie) TryGet(key []byte) ([]byte, error) {
	return t.trie.TryGet(t.hashKey(key))
}

func (t *SecureTrie) Update(key, value []byte) {
	if err := t.TryUpdate(key, value); err != nil {
		log.Error(fmt.Sprintf("Unhandled trie error: %v", err))
	}
}

func (t *SecureTrie) TryUpdate(key, value []byte) error {
	hk := t.hashKey(key)
	err := t.trie.TryUpdate(hk, value)
	if err != nil {
		return err
	}
	t.getSecKeyCache()[string(hk)] = common.CopyBytes(key)
	return nil
}

func (t *SecureTrie) Delete(key []byte) {
	if err := t.TryDelete(key); err != nil {
		log.Error(fmt.Sprintf("Unhandled trie error: %v", err))
	}
}

func (t *SecureTrie) TryDelete(key []byte) error {
	hk := t.hashKey(key)
	delete(t.getSecKeyCache(), string(hk))
	return t.trie.TryDelete(hk)
}

func (t *SecureTrie) Commit() (common.Uint256, error) {
	return t.CommitTo(t.trie.db)
}

func (t *SecureTrie) CommitTo(db DatabaseWriter) (common.Uint256, error) {
	if len(t.getSecKeyCache()) > 0 {
		for hk, key := range t.secKeyCache {
			if err := db.Put(t.secKey([]byte(hk)), key); err != nil {
				return common.Uint256{}, err
			}
		}
		t.secKeyCache = make(map[string][]byte)
	}
	return t.trie.commitTo(db)
}

func (t *SecureTrie) secKey(key []byte) []byte {
	buf := append(t.secKeyBuf[:0], secureKeyPrefix...)
	buf = append(buf, key...)
	return buf
}

func (t *SecureTrie) hashKey(key []byte) []byte {
	h := newHasher()
	h.sha, _ = common.Uint256ParseFromBytes(ToHash256(key))
	return h.sha.ToArray()
}

func (t *SecureTrie) getSecKeyCache() map[string][]byte {
	if t != t.secKeyCacheOwner {
		t.secKeyCacheOwner = t
		t.secKeyCache = make(map[string][]byte)
	}
	return t.secKeyCache
}

func (t *SecureTrie) Hash() common.Uint256 {
	return t.trie.Hash()
}

func (t *SecureTrie) Copy() *SecureTrie {
	cpy := *t
	return &cpy
}