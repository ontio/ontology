package LevelDBStore

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	. "GoOnchain/core/ledger"
	"GoOnchain/core/contract/program"
	. "GoOnchain/core/asset"
	"GoOnchain/common/serialization"
	"bytes"
	"fmt"
	tx "GoOnchain/core/transaction"
	. "GoOnchain/common"
)

type LevelDBStore struct {
	db *leveldb.DB // LevelDB instance
	b  *leveldb.Batch
	it *Iterator
	header_index *[]Uint256
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

	var headerindex = make ([]Uint256,0)

	return &LevelDBStore{
		db: db,
		b: nil,
		it: nil,
		header_index: &headerindex,
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

func (bd *LevelDBStore) InitLedgerStore ( l * Ledger ) error {
	// TODO: InitLedgerStore
	return nil
}

func (bd *LevelDBStore) IsDoubleSpend( tx tx.Transaction ) bool {
	// TODO: IsDoubleSpend Check

	return false
}

func (bd *LevelDBStore) GetBlockHash(height uint32) Uint256 {
	// TODO: GetBlockHash
	x := new(Uint256)

	return *x
}

func (bd *LevelDBStore) GetContract(hash []byte) ([]byte, error) {
	prefix := []byte { byte(DATA_Contract) }
	bData,err_get := bd.Get( append(prefix,hash...) )
	if ( err_get != nil ) {
		//TODO: implement error process
		return nil, err_get
	}

	fmt.Println("GetContract Data: ", bData)

	return bData,nil
}

func (bd *LevelDBStore) SaveHeader(header *Header) error {
	w := bytes.NewBuffer(nil)
	header.Serialize(w)

	// generate key
	headerkey := bytes.NewBuffer(nil)
	// add header prefix.
	headerkey.WriteByte( byte(DATA_Header) )
	// contact asset id
	blockhash := header.Blockdata.Hash()
	blockhash.Serialize(headerkey)

	fmt.Printf( "header key: %x\n",  headerkey )
	fmt.Printf( "header data: %x\n",  w )

	// PUT VALUE
	err := bd.Put( headerkey.Bytes(), w.Bytes() )
	if ( err != nil ){
		return err
	}

	return nil
}

func (bd *LevelDBStore) GetHeader(hash Uint256) (*Header, error) {
	// TODO: GET HEADER
	var h * Header = new(Header)

	h.Blockdata = new(Blockdata)
	h.Blockdata.Program = new(program.Program)

	prefix := []byte{ byte(DATA_Header) }
	data,err_get := bd.Get( append(prefix,hash.ToArray()...) )
	fmt.Printf( "Get Header Data: %x\n",  data )
	if ( err_get != nil ) {
		//TODO: implement error process
		return nil, err_get
	}

	r := bytes.NewReader(data)

	// first 8 bytes is sys_fee
	sysfee,err := serialization.ReadUint64(r)
	fmt.Printf( "sysfee: %d\n",  sysfee )

	// Deserialize block data
	err = h.Deserialize( r )

	return h,err
}

func (bd *LevelDBStore) SaveAsset(asset *Asset) error {
	w := bytes.NewBuffer(nil)

	asset.Serialize(w)

	// generate key
	assetkey := bytes.NewBuffer(nil)
	// add asset prefix.
	assetkey.WriteByte( byte(ST_QuantityIssued) )
	// contact asset id
	asset.ID.Serialize(assetkey)

	fmt.Printf( "asset key: %x\n",  assetkey )

	// PUT VALUE
	err := bd.Put( assetkey.Bytes(), w.Bytes() )
	if ( err != nil ){
		return err
	}

	return nil
}

func (bd *LevelDBStore) GetAsset(hash Uint256) (*Asset, error) {
	fmt.Printf( "GetAsset Hash: %x\n",  hash )

	asset := new(Asset)

	prefix := []byte{ byte(ST_QuantityIssued) }
	data,err_get := bd.Get( append(prefix,hash.ToArray()...) )

	fmt.Printf( "GetAsset Data: %x\n",  data )
	if ( err_get != nil ) {
		//TODO: implement error process
		return nil, err_get
	}

	r := bytes.NewReader(data)
	asset.Deserialize(r)

	return asset,nil
}

/*
func (bd *LevelDBStore) GetNextBlockHash(hash []byte) common.Uint256 {
	h,_ := bd.GetHeader( hash )

	if ( h == nil ) {
		return nil
	}

	if ( h.Blockdata.Height + 1 >= uint32(len(*bd.header_index)) ) {
		return nil
	}

	return (*bd.header_index)[h.Blockdata.Height + 1];
}
*/

func (bd *LevelDBStore) GetTransaction(t *tx.Transaction,hash []byte) error {
	fmt.Printf( "GetTransaction Hash: %x\n",  hash )

	prefix := []byte{ byte(DATA_Transaction) }
	tHash,err_get := bd.Get( append(prefix,hash...) )
	fmt.Printf( "GetTransaction Data: %x\n",  tHash )
	if ( err_get != nil ) {
		//TODO: implement error process
		return err_get
	}

	r := bytes.NewReader(tHash)

	// get height
	height,err := serialization.ReadUint32(r)
	fmt.Printf( "tx height: %d\n",  height )

	// Deserialize Transaction
	t.Deserialize(r)

	return err
}

func (bd *LevelDBStore) SaveTransaction(tx *tx.Transaction,height uint32) error {
	w := bytes.NewBuffer(nil)

	// generate value
	serialization.WriteUint32(w, height)
	tx.Serialize(w)

	// generate key
	txhash := bytes.NewBuffer(nil)
	// add transaction header prefix.
	txhash.WriteByte( byte(DATA_Transaction) )
	// get transaction hash
	txHashValue := tx.Hash()
	txHashValue.Serialize(txhash)

	fmt.Printf( "transaction header + hash: %x\n",  txhash )
	fmt.Printf( "transaction tx data: %x\n",  w )

	// put value
	err := bd.Put( txhash.Bytes(), w.Bytes() )
	if ( err != nil ){
		return err
	}

	return nil
}


func (bd *LevelDBStore) GetBlock(hash Uint256) (*Block, error) {
	var b *Block = new (Block)

	b.Blockdata = new (Blockdata)
	b.Blockdata.Program = new (program.Program)

	prefix := []byte{ byte(DATA_Header) }
	bHash,err_get := bd.Get( append(prefix,hash.ToArray()...) )
	fmt.Printf( "GetBlock Data: %x\n",  bHash )
	if ( err_get != nil ) {
		//TODO: implement error process
		return nil, err_get
	}

	r := bytes.NewReader(bHash)

	// first 8 bytes is sys_fee
	sysfee,err := serialization.ReadUint64(r)
	fmt.Printf( "sysfee: %d\n",  sysfee )

	// Deserialize block data
	err = b.Deserialize( r )

	// Deserialize transaction
	for i:=0; i<len(b.Transcations); i++ {
		temp := b.Transcations[i].Hash()
		bd.GetTransaction(b.Transcations[i],temp.ToArray())
	}

	return b, err
}

func (bd *LevelDBStore) SaveBlock(b *Block) error {
	w := bytes.NewBuffer(nil)

	// generate value
	var sysfee uint64 = 0xFFFFFFFFFFFFFFFF
	serialization.WriteUint64(w, sysfee)
	b.Serialize(w)

	// generate key
	bhhash := bytes.NewBuffer(nil)
	// add block header prefix.
	bhhash.WriteByte( byte(DATA_Header) )
	// calc block hash
	blockhash := b.Hash()
	blockhash.Serialize(bhhash)

	fmt.Printf( "block header + hash: %x\n",  bhhash )

	// PUT VALUE
	err := bd.Put( bhhash.Bytes(), w.Bytes() )
	if ( err != nil ){
		return err
	}

	// save transcations to leveldb
	nLen := len(b.Transcations)
	for i:=0; i<nLen; i++{
		/*
		// for test
		if i==1 {
			b.Transcations[i].Hash = Uint256{0x00,0x01,0x02,0x03,0x00,0x01,0x02,0x03,0x00,0x01,0x02,0x03,0x00,0x01,0x02,0x03,0x00,0x01,0x02,0x03,0x00,0x01,0x02,0x03,0x00,0x01,0x02,0x03,0x00,0x01,0x02,0x03}
			fmt.Printf( "txhash: %x\n",  b.Transcations[i].Hash )
			bd.SaveTransaction(b.Transcations[i],b.Blockdata.Height)
		}
		*/

		// now support RegisterAsset and Miner tx ONLY.
		if ( b.Transcations[i].TxType == 0x40 || b.Transcations[i].TxType == 0x00 ) {
			bd.SaveTransaction(b.Transcations[i],b.Blockdata.Height)
		}
	}

	return nil
}