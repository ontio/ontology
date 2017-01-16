package store

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"GoOnchain/core/ledger"
	"GoOnchain/core/contract/program"
	"bytes"
	"fmt"
)

type LevelDBStore struct {
	db *leveldb.DB // LevelDB instance
	b  *leveldb.Batch
	it *Iterator
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
		db: db,
		b: nil,
		it: nil,
	}, nil
}

func NewDBByOptions(file string, o *opt.Options) (*LevelDBStore, error) {

	db, err := leveldb.OpenFile(file, o)

	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(file, nil)
	}

	if err != nil {
		return nil, err
	}

	return &LevelDBStore{
		db: db,
		b: nil,
		it: nil,
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

func (self *LevelDBStore) Close() error {

	err := self.db.Close()

	return err
}

func (self *LevelDBStore) NewIterator(options *opt.ReadOptions) *Iterator {

	iter := self.db.NewIterator(nil, options)

	return &Iterator{
		iter : iter,
	}
}

func (self *LevelDBStore) BatchPut(key []byte, value []byte) error {
	self.b.Put(key, value)

	return nil
}

func (self *LevelDBStore) BatchDelete(key []byte) error {

	self.b.Delete(key)

	return nil
}

func (self *LevelDBStore) BatchWrite() error {

	return self.db.Write(self.b, nil)
}

func (bd *LevelDBStore) GetBlock(hash []byte) (*ledger.Block, error) {
	var b *ledger.Block = new (ledger.Block)
	b.Blockdata = new (ledger.Blockdata)
	b.Blockdata.Program = new (program.Program)

	bHash,err_get := bd.Get( hash )
	fmt.Println("GetBlock Data: ", bHash)
	if ( err_get != nil ) {
		//TODO: implement error process
		return nil, err_get
	}

	// first 8 bytes is sys_fee
	r := bytes.NewReader(bHash[8:])

	err := b.Deserialize( r )

	return b,err
}

func (bd *LevelDBStore) SaveBlock(b *ledger.Block) error {
	w := bytes.NewBuffer(nil)
	b.Serialize(w)
	nLen := len(b.Transcations)

	for i:=0; i<nLen; i++{
		txhash,_ := b.Transcations[i].GetOutputHashes()
		txhash[0].Serialize(w)
	}

	// GET KEY
	bhhash := bytes.NewBuffer(nil)
	// block prefix
	bhhash.WriteByte( byte(DATA_Header) )
	// block hash
	blockhash := b.GetHash()
	blockhash.Serialize(bhhash)

	// PUT VALUE
	bd.Put( bhhash, w )

	//fmt.Println("SaveBlock Data: ", w.Bytes())
	//fmt.Println("SaveBlock Data: ", w.Bytes())

	return nil
}