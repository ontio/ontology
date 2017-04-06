package LevelDBStore

import (
	. "DNA/common"
	"DNA/common/log"
	"DNA/common/serialization"
	. "DNA/core/asset"
	"DNA/core/contract/program"
	. "DNA/core/ledger"
	tx "DNA/core/transaction"
	"DNA/core/transaction/payload"
	"DNA/core/validation"
	. "DNA/errors"
	"bytes"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"sync"
)

type LevelDBStore struct {
	db *leveldb.DB // LevelDB instance
	b  *leveldb.Batch
	it *Iterator

	header_index map[uint32]Uint256
	//header_cache map[Uint256]*Blockdata
	block_cache map[Uint256]*Block

	current_block_height uint32
	stored_header_count  uint32

	mutex sync.Mutex
}

func init() {
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
		db:           db,
		b:            nil,
		it:           nil,
		header_index: map[uint32]Uint256{},
		//header_cache: map[Uint256]*Blockdata{},
		block_cache:          map[Uint256]*Block{},
		current_block_height: 0,
	}, nil
}

func (bd *LevelDBStore) InitLevelDBStoreWithGenesisBlock(genesisblock *Block) {
	hash := genesisblock.Hash()
	bd.header_index[0] = hash
	bd.persist(genesisblock)
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
		b:  nil,
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
		iter: iter,
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

func (bd *LevelDBStore) InitLedgerStore(l *Ledger) error {
	// TODO: InitLedgerStore
	return nil
}

func (bd *LevelDBStore) IsDoubleSpend(tx tx.Transaction) bool {
	// TODO: IsDoubleSpend Check

	return false
}

func (bd *LevelDBStore) GetBlockHash(height uint32) (Uint256, error) {

	if height >= 0 {
		querykey := bytes.NewBuffer(nil)
		querykey.WriteByte(byte(DATA_BlockHash))
		err := serialization.WriteUint32(querykey, height)

		if err == nil {
			blockhash, err_get := bd.Get(querykey.Bytes())
			if err_get != nil {
				//TODO: implement error process
				return Uint256{}, err_get
			} else {
				blockhash256, err_parse := Uint256ParseFromBytes(blockhash)
				if err_parse == nil {
					return blockhash256, nil
				} else {
					return Uint256{}, err_parse
				}

			}
		} else {
			return Uint256{}, err
		}
	} else {
		return Uint256{}, NewDetailErr(errors.New("[LevelDB]: GetBlockHash error param height < 0."), ErrNoCode, "")
	}
}

func (bd *LevelDBStore) GetCurrentBlockHash() Uint256 {
	bd.mutex.Lock()
	defer bd.mutex.Unlock()
	
	return bd.header_index[bd.current_block_height]
}

func (bd *LevelDBStore) GetContract(hash []byte) ([]byte, error) {
	prefix := []byte{byte(DATA_Contract)}
	bData, err_get := bd.Get(append(prefix, hash...))
	if err_get != nil {
		//TODO: implement error process
		return nil, err_get
	}

	log.Debug("GetContract Data: ", bData)

	return bData, nil
}

func (bd *LevelDBStore) SaveHeader(header *Header, ledger *Ledger) error {
	bd.mutex.Lock()
	defer bd.mutex.Unlock()

	if header.Blockdata.Height-uint32(len(bd.header_index)) >= 1 {
		return errors.New("[SaveHeader] block height - header_index >= 1")
	}

	if header.Blockdata.Height < uint32(len(bd.header_index)) {
		return errors.New("[SaveHeader] block height < header_index")
	}

	//header verify
	err := validation.VerifyBlockData(header.Blockdata, ledger)
	if err != nil {
		return err
	}

	batch := new(leveldb.Batch)
	bd.onAddHeader(header.Blockdata, batch)
	err = bd.db.Write(batch, nil)
	if err != nil {
		return err
	}

	return nil

	/*
		//////////////////////////////////////////////////////////////
		// generate key with DATA_Header prefix
		headerkey := bytes.NewBuffer(nil)
		// add header prefix.
		headerkey.WriteByte( byte(DATA_Header) )
		// contact block hash
		blockhash := header.Blockdata.Hash()
		blockhash.Serialize(headerkey)

		fmt.Printf( "header key: %x\n",  headerkey )

		// generate value
		w := bytes.NewBuffer(nil)
		header.Serialize(w)
		fmt.Printf( "header data: %x\n",  w )

		// PUT VALUE
		err := bd.Put( headerkey.Bytes(), w.Bytes() )
		if ( err != nil ){
			return err
		}

		//////////////////////////////////////////////////////////////
		// generate key with DATA_BlockHash prefix
		bhash := bytes.NewBuffer(nil)
		bhash.WriteByte( byte(DATA_BlockHash) )
		err = serialization.WriteUint32( bhash, header.Blockdata.Height )
		fmt.Printf( "DATA_BlockHash table key: %x\n",  bhash )

		// generate value
		hashwriter := bytes.NewBuffer(nil)
		hashvalue := header.Blockdata.Hash()
		hashvalue.Serialize(hashwriter)
		fmt.Printf( "DATA_BlockHash table value: %x\n",  hashvalue )

		// PUT VALUE
		err = bd.Put( bhash.Bytes(), hashwriter.Bytes() )
		if ( err != nil ){
			return err
		}

		return nil
	*/
}

func (bd *LevelDBStore) GetHeader(hash Uint256) (*Header, error) {
	// TODO: GET HEADER
	var h *Header = new(Header)

	h.Blockdata = new(Blockdata)
	h.Blockdata.Program = new(program.Program)

	prefix := []byte{byte(DATA_Header)}
	log.Debug("GetHeader Data:", hash.ToArray())
	data, err_get := bd.Get(append(prefix, hash.ToArray()...))
	//log.Debug( "Get Header Data: %x\n",  data )
	if err_get != nil {
		//TODO: implement error process
		return nil, err_get
	}

	r := bytes.NewReader(data)

	// first 8 bytes is sys_fee
	sysfee, err := serialization.ReadUint64(r)
	if err != nil {
		return nil, err
	}
	log.Debug(fmt.Sprintf("sysfee: %d\n", sysfee))

	// Deserialize block data
	err = h.Deserialize(r)
	if err != nil {
		return nil, err
	}

	return h, err
}

func (bd *LevelDBStore) SaveAsset(assetid Uint256, asset *Asset) error {
	w := bytes.NewBuffer(nil)

	asset.Serialize(w)

	// generate key
	assetkey := bytes.NewBuffer(nil)
	// add asset prefix.
	assetkey.WriteByte(byte(ST_QuantityIssued))
	// contact asset id
	assetid.Serialize(assetkey)

	log.Debug(fmt.Sprintf("asset key: %x\n", assetkey))

	// PUT VALUE
	err := bd.Put(assetkey.Bytes(), w.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (bd *LevelDBStore) GetAsset(hash Uint256) (*Asset, error) {
	log.Debug(fmt.Sprintf("GetAsset Hash: %x\n", hash))

	asset := new(Asset)

	prefix := []byte{byte(ST_QuantityIssued)}
	data, err_get := bd.Get(append(prefix, hash.ToArray()...))

	log.Debug(fmt.Sprintf("GetAsset Data: %x\n", data))
	if err_get != nil {
		//TODO: implement error process
		return nil, err_get
	}

	r := bytes.NewReader(data)
	asset.Deserialize(r)

	return asset, nil
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

func (bd *LevelDBStore) GetTransaction(hash Uint256) (*tx.Transaction, error) {
	//Trace()
	//log.Debug(fmt.Sprintf("GetTransaction Hash: %x\n", hash))
	t := new(tx.Transaction)
	err := bd.getTx(t, hash)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (bd *LevelDBStore) getTx(tx *tx.Transaction, hash Uint256) error {
	prefix := []byte{byte(DATA_Transaction)}
	tHash, err_get := bd.Get(append(prefix, hash.ToArray()...))
	//log.Debug(fmt.Sprintf("getTx Data: %x\n", tHash))
	if err_get != nil {
		//TODO: implement error process
		log.Warn("Get TX from DB error")
		return err_get
	}

	r := bytes.NewReader(tHash)

	// get height
	_, err := serialization.ReadUint32(r)
	//height, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	//log.Debug(fmt.Sprintf("tx height: %d\n", height))

	// Deserialize Transaction
	err = tx.Deserialize(r)

	return err
}

func (bd *LevelDBStore) SaveTransaction(tx *tx.Transaction, height uint32) error {

	//////////////////////////////////////////////////////////////
	// generate key with DATA_Transaction prefix
	txhash := bytes.NewBuffer(nil)
	// add transaction header prefix.
	txhash.WriteByte(byte(DATA_Transaction))
	// get transaction hash
	txHashValue := tx.Hash()
	txHashValue.Serialize(txhash)
	log.Debug(fmt.Sprintf("transaction header + hash: %x\n", txhash))

	// generate value
	w := bytes.NewBuffer(nil)
	serialization.WriteUint32(w, height)
	tx.Serialize(w)
	log.Debug(fmt.Sprintf("transaction tx data: %x\n", w))

	// put value
	err := bd.Put(txhash.Bytes(), w.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (bd *LevelDBStore) GetBlock(hash Uint256) (*Block, error) {
	var b *Block = new(Block)

	b.Blockdata = new(Blockdata)
	b.Blockdata.Program = new(program.Program)

	prefix := []byte{byte(DATA_Header)}
	bHash, err_get := bd.Get(append(prefix, hash.ToArray()...))
	if err_get != nil {
		//TODO: implement error process
		return nil, err_get
	}
	//log.Debug(fmt.Sprintf("GetBlock Data: %x\n", bHash))

	r := bytes.NewReader(bHash)

	// first 8 bytes is sys_fee
	_,err := serialization.ReadUint64(r)
	//sysfee, err := serialization.ReadUint64(r)
	if err != nil {
		return nil, err
	}
	//log.Debug(fmt.Sprintf("sysfee: %d\n", sysfee))

	// Deserialize block data
	err = b.Deserialize(r)
	if err != nil {
		return nil, err
	}

	// Deserialize transaction
	for i := 0; i < len(b.Transactions); i++ {
		err = bd.getTx(b.Transactions[i], b.Transactions[i].Hash())
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

func (bd *LevelDBStore) persist(b *Block) error {

	//////////////////////////////////////////////////////////////
	// generate key with DATA_Header prefix
	bhhash := bytes.NewBuffer(nil)
	// add block header prefix.
	bhhash.WriteByte(byte(DATA_Header))
	// calc block hash
	blockhash := b.Hash()
	blockhash.Serialize(bhhash)
	log.Debug(fmt.Sprintf("block header + hash: %x\n", bhhash))

	// generate value
	w := bytes.NewBuffer(nil)
	var sysfee uint64 = 0xFFFFFFFFFFFFFFFF
	serialization.WriteUint64(w, sysfee)
	b.Serialize(w)

	// PUT VALUE
	err := bd.Put(bhhash.Bytes(), w.Bytes())
	if err != nil {
		return err
	}

	//////////////////////////////////////////////////////////////
	// generate key with DATA_BlockHash prefix
	bhash := bytes.NewBuffer(nil)
	bhash.WriteByte(byte(DATA_BlockHash))
	err = serialization.WriteUint32(bhash, b.Blockdata.Height)
	if err != nil {
		return err
	}
	log.Debug(fmt.Sprintf("DATA_BlockHash table key: %x\n", bhash))

	// generate value
	hashwriter := bytes.NewBuffer(nil)
	hashvalue := b.Blockdata.Hash()
	hashvalue.Serialize(hashwriter)
	log.Debug(fmt.Sprintf("DATA_BlockHash table value: %x\n", hashvalue))

	// PUT VALUE
	err = bd.Put(bhash.Bytes(), hashwriter.Bytes())
	if err != nil {
		return err
	}

	//////////////////////////////////////////////////////////////
	// save transcations to leveldb
	nLen := len(b.Transactions)
	for i := 0; i < nLen; i++ {
		/*
			// for test
			if i==1 {
				b.Transactions[i].Hash = Uint256{0x00,0x01,0x02,0x03,0x00,0x01,0x02,0x03,0x00,0x01,0x02,0x03,0x00,0x01,0x02,0x03,0x00,0x01,0x02,0x03,0x00,0x01,0x02,0x03,0x00,0x01,0x02,0x03,0x00,0x01,0x02,0x03}
				fmt.Printf( "txhash: %x\n",  b.Transactions[i].Hash )
				bd.SaveTransaction(b.Transactions[i],b.Blockdata.Height)
			}
		*/

		// now support RegisterAsset and Miner tx ONLY.
		if b.Transcations[i].TxType == 0x40 || b.Transcations[i].TxType == 0x00 {
			err = bd.SaveTransaction(b.Transcations[i], b.Blockdata.Height)
			if err != nil {
				return err
			}
		}
		if b.Transcations[i].TxType == 0x40 {
			ar := b.Transcations[i].Payload.(*payload.RegisterAsset)
			err = bd.SaveAsset(b.Transcations[i].Hash(),ar.Asset)
			if err != nil {
				return err
			}

		}
	}

	// save hash with height
	bd.current_block_height = b.Blockdata.Height

	// generate key with SYS_CurrentHeader prefix
	currentblockkey := bytes.NewBuffer(nil)
	currentblockkey.WriteByte(byte(SYS_CurrentBlock))
	//fmt.Printf( "SYS_CurrentHeader key: %x\n",  currentblockkey )

	currentblock := bytes.NewBuffer(nil)
	blockhash.Serialize(currentblock)
	serialization.WriteUint32(currentblock, b.Blockdata.Height)

	// PUT VALUE
	err = bd.Put(currentblockkey.Bytes(), currentblock.Bytes())
	if err != nil {
		return err
	}

	//bh := b.Blockdata.Hash()
	//bd.header_index[bd.current_block_height] = &bh

	return nil
}

func (bd *LevelDBStore) onAddHeader(header *Blockdata, batch *leveldb.Batch) {
	log.Debug(fmt.Sprintf("onAddHeader(), Height=%d\n", header.Height))

	hash := header.Hash()
	bd.header_index[header.Height] = hash
	for header.Height-bd.stored_header_count >= 2000 {
		hashbuffer := new(bytes.Buffer)
		serialization.WriteVarUint(hashbuffer, uint64(2000))
		var hasharray []byte
		for i := 0; i < 2000; i++ {
			index := bd.stored_header_count + uint32(i)
			//fmt.Println("index:",index)
			thash := bd.header_index[index]
			thehash := thash.ToArray()
			hasharray = append(hasharray, thehash...)
			//fmt.Printf("%x\n",thehash)
		}
		hashbuffer.Write(hasharray)
		//fmt.Printf( "%x\n", hashbuffer )

		// generate key with DATA_Header prefix
		hhlprefix := bytes.NewBuffer(nil)
		// add block header prefix.
		hhlprefix.WriteByte(byte(IX_HeaderHashList))
		serialization.WriteUint32(hhlprefix, bd.stored_header_count)
		//fmt.Printf( "%x\n", hhlprefix )

		batch.Put(hhlprefix.Bytes(), hashbuffer.Bytes())
		bd.stored_header_count += 2000
	}

	//////////////////////////////////////////////////////////////
	// generate key with DATA_Header prefix
	headerkey := bytes.NewBuffer(nil)
	// add header prefix.
	headerkey.WriteByte(byte(DATA_Header))
	// contact block hash
	blockhash := header.Hash()
	blockhash.Serialize(headerkey)
	log.Debug(fmt.Sprintf("header key: %x\n", headerkey))
	//fmt.Println( "header key:",  headerkey.Bytes() )

	// generate value
	w := bytes.NewBuffer(nil)
	header.Serialize(w)
	log.Debug(fmt.Sprintf("header data: %x\n", w))
	//fmt.Println( "header data:",  w.Bytes() )

	// PUT VALUE
	batch.Put(headerkey.Bytes(), w.Bytes())

	//////////////////////////////////////////////////////////////
	// generate key with SYS_CurrentHeader prefix
	currentheaderkey := bytes.NewBuffer(nil)
	currentheaderkey.WriteByte(byte(SYS_CurrentHeader))
	//fmt.Printf( "SYS_CurrentHeader key: %x\n",  currentheaderkey )

	currentheader := bytes.NewBuffer(nil)
	blockhash.Serialize(currentheader)
	serialization.WriteUint32(currentheader, header.Height)
	//fmt.Printf( "SYS_CurrentHeader data: %x\n",  currentheader )

	// PUT VALUE
	batch.Put(currentheaderkey.Bytes(), currentheader.Bytes())
}

func (bd *LevelDBStore) persistBlocks() {
	log.Debug("persistBlocks()")

	if uint32(len(bd.header_index)) <= bd.current_block_height+1 {
		fmt.Printf("[persistBlocks]: error, header_index.count < current_block_height + 1")
		return
	}

	hash := bd.header_index[bd.current_block_height+1]

	block, ok := bd.block_cache[hash]
	if ok {
		bd.persist(block)
		delete(bd.block_cache, hash)

		//TODO: PersistCompleted
	}
}

func (bd *LevelDBStore) SaveBlock(b *Block, ledger *Ledger) error {
	log.Debug("SaveBlock()")

	bd.mutex.Lock()
	defer bd.mutex.Unlock()

	if bd.block_cache[b.Hash()] == nil {
		bd.block_cache[b.Hash()] = b
	}

	if b.Blockdata.Height-uint32(len(bd.header_index)) >= 1 {
		//return false,NewDetailErr(errors.New(fmt.Sprintf("WARNING: [SaveBlock] block height - header_index.count >= 1, block height:%d, header_index.count:%d",b.Blockdata.Height, uint32(len(bd.header_index)) )),ErrDuplicatedBlock,"")
		return errors.New(fmt.Sprintf("WARNING: [SaveBlock] block height - header_index.count >= 1, block height:%d, header_index.count:%d", b.Blockdata.Height, uint32(len(bd.header_index))))
	}

	if b.Blockdata.Height == uint32(len(bd.header_index)) {
		//Block verify
		err := validation.VerifyBlock(b, ledger, true)
		if err != nil {
			log.Debug("VerifyBlock() error!")
			return err
		}

		batch := new(leveldb.Batch)
		bd.onAddHeader(b.Blockdata, batch)
		log.Debug("batch dump: ", batch.Dump())
		err = bd.db.Write(batch, nil)
		if err != nil {
			return err
		}
	} else {
		return errors.New("[SaveBlock] block height != header_index")
	}

	if b.Blockdata.Height < uint32(len(bd.header_index)) {
		// TODO: block event set
		//new_block_event.Set();
		bd.persistBlocks()
	} else {
		return errors.New("[SaveBlock] block height < header_index")
	}

	return nil
}

func (bd *LevelDBStore) GetQuantityIssued (AssetId Uint256) (*Fixed64, error) {

	return nil,nil
}