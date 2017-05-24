package LevelDBStore

import (
	. "DNA/core/store"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type LevelDBStore struct {
	db    *leveldb.DB // LevelDB instance
	batch *leveldb.Batch
}

func NewLevelDBStore(file string) (*LevelDBStore, error) {

	// default Options
	o := opt.Options{
		NoSync: false,
	}

	db, err := leveldb.OpenFile(file, &o)

	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(file, nil)
	}

	if err != nil {
		return nil, err
	}

	return &LevelDBStore{
		db:    db,
		batch: nil,
	}, nil
}

func (self *LevelDBStore) Put(key []byte, value []byte) error {
	return self.db.Put(key, value, nil)
}

func (self *LevelDBStore) Get(key []byte) ([]byte, error) {
	dat, err := self.db.Get(key, nil)
	return dat, err
}

func (self *LevelDBStore) Delete(key []byte) error {
	return self.db.Delete(key, nil)
}

func (self *LevelDBStore) NewBatch() error {
	self.batch = new(leveldb.Batch)
	return nil
}

func (self *LevelDBStore) BatchPut(key []byte, value []byte) error {
	self.batch.Put(key, value)
	return nil
}

func (self *LevelDBStore) BatchDelete(key []byte) error {
	self.batch.Delete(key)
	return nil
}

func (self *LevelDBStore) BatchCommit() error {
	err := self.db.Write(self.batch, nil)
	if err != nil {
		return err
	}
	return nil
}

func (self *LevelDBStore) Close() error {
	err := self.db.Close()
	return err
}

func (self *LevelDBStore) NewIterator(prefix []byte) IIterator {

	iter := self.db.NewIterator(util.BytesPrefix(prefix), nil)

	return &Iterator{
		iter: iter,
	}
}
