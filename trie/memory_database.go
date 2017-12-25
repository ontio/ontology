package trie

import (
	"sync"
	"fmt"
)

type MemDatabase struct {
	db map[string][]byte
	lock sync.RWMutex
}

func NewMemDatabase() *MemDatabase {
	return &MemDatabase {
		db: make(map[string][]byte),
	}
}

func (db *MemDatabase) Put(key []byte, value []byte) error{
	db.lock.Lock()
	defer db.lock.Unlock()
	db.db[string(key)] = CopyBytes(value)
	return nil
}

func (db *MemDatabase) Has(key []byte) (bool, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	_, ok := db.db[string(key)]
	return ok, nil
}

func (db *MemDatabase) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	if entry, ok := db.db[string(key)]; ok {
		return CopyBytes(entry), nil
	}
	return nil, nil
}

func (db *MemDatabase) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	delete(db.db, string(key))
	return nil
}

func (db *MemDatabase) ViewDB() {
	for k, v := range db.db {
		fmt.Printf("DB Cache Key:%v value:%v \n", k, v)
	}
}

func CopyBytes(b []byte) (copiedBytes []byte) {
	if b == nil {
		return nil
	}
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)

	return
}


