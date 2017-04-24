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
	"DNA/events"
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
	block_cache  map[Uint256]*Block

	current_block_height uint32
	stored_header_count  uint32

	mu sync.RWMutex

	disposed bool
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
		db:                   db,
		b:                    nil,
		it:                   nil,
		header_index:         map[uint32]Uint256{},
		block_cache:          map[Uint256]*Block{},
		current_block_height: 0,
		disposed:             false,
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
	bd.mu.RLock()
	defer bd.mu.RUnlock()

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

func (bd *LevelDBStore) AddHeaders(headers []Header, ledger *Ledger) error {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	batch := new(leveldb.Batch)
	for i := 0; i < len(headers); i++ {
		if headers[i].Blockdata.Height >= (uint32(len(bd.header_index)) + 1) {
			break
		}

		if headers[i].Blockdata.Height < uint32(len(bd.header_index)) {
			continue
		}

		//header verify
		err := validation.VerifyHeader(&headers[i], ledger)
		if err != nil {
			break
		}

		bd.onAddHeader(&headers[i], batch)
	}

	err := bd.db.Write(batch, nil)
	if err != nil {
		return err
	}

	return nil
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
	assetkey.WriteByte(byte(ST_Info))
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

	prefix := []byte{byte(ST_Info)}
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

func (bd *LevelDBStore) GetTransaction(hash Uint256) (*tx.Transaction, error) {
	log.Trace()
	log.Debug(fmt.Sprintf("GetTransaction Hash: %x\n", hash))

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
		//log.Warn("Get TX from DB error")
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
	_, err := serialization.ReadUint64(r)
	//sysfee, err := serialization.ReadUint64(r)
	if err != nil {
		return nil, err
	}
	//log.Debug(fmt.Sprintf("sysfee: %d\n", sysfee))

	// Deserialize block data
	err = b.FromTrimmedData(r)
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
	unspents := make(map[Uint256][]uint16)
	quantities := make(map[Uint256]Fixed64)
	///////////////////////////////////////////////////////////////
	// Get Unspents for every tx
	unspentprefix := []byte{byte(IX_Unspent)}
	for i := 0; i < len(b.Transactions); i++ {
		txhash := b.Transactions[i].Hash()
		unspentvalue, err_get := bd.Get(append(unspentprefix, txhash.ToArray()...))

		if err_get != nil {
			unspentvalue = []byte{}
		}

		unspents[txhash], err_get = GetUint16Array(unspentvalue)
		if err_get != nil {
			return err_get
		}
	}

	///////////////////////////////////////////////////////////////
	// batch write begin
	batch := new(leveldb.Batch)

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
	b.Trim(w)

	// BATCH PUT VALUE
	batch.Put(bhhash.Bytes(), w.Bytes())

	//////////////////////////////////////////////////////////////
	// generate key with DATA_BlockHash prefix
	bhash := bytes.NewBuffer(nil)
	bhash.WriteByte(byte(DATA_BlockHash))
	err := serialization.WriteUint32(bhash, b.Blockdata.Height)
	if err != nil {
		return err
	}
	log.Debug(fmt.Sprintf("DATA_BlockHash table key: %x\n", bhash))

	// generate value
	hashwriter := bytes.NewBuffer(nil)
	hashvalue := b.Blockdata.Hash()
	hashvalue.Serialize(hashwriter)
	log.Debug(fmt.Sprintf("DATA_BlockHash table value: %x\n", hashvalue))

	// BATCH PUT VALUE
	batch.Put(bhash.Bytes(), hashwriter.Bytes())

	//////////////////////////////////////////////////////////////
	// save transactions to leveldb
	nLen := len(b.Transactions)
	for i := 0; i < nLen; i++ {

		// now support RegisterAsset / IssueAsset / TransferAsset and Miner TX ONLY.
		if b.Transactions[i].TxType == tx.RegisterAsset || b.Transactions[i].TxType == tx.IssueAsset || b.Transactions[i].TxType == tx.TransferAsset ||
			b.Transactions[i].TxType == tx.Record || b.Transactions[i].TxType == tx.BookKeeping {
			err = bd.SaveTransaction(b.Transactions[i], b.Blockdata.Height)
			if err != nil {
				return err
			}
		}
		if b.Transactions[i].TxType == tx.RegisterAsset {
			ar := b.Transactions[i].Payload.(*payload.RegisterAsset)
			err = bd.SaveAsset(b.Transactions[i].Hash(), ar.Asset)
			if err != nil {
				return err
			}
		}

		if b.Transactions[i].TxType == tx.IssueAsset {
			results, err := b.Transactions[i].GetTransactionResults()
			if err != nil {
				return err
			}

			for _, result := range results {
				if _, ok := quantities[result.AssetId]; !ok {
					quantities[result.AssetId] -= result.Amount
				} else {
					quantities[result.AssetId] = -result.Amount
				}
			}
		}

		// init unspent in tx
		for index := 0; index < len(b.Transactions[i].Outputs); index++ {
			txhash := b.Transactions[i].Hash()
			unspents[txhash] = append(unspents[txhash], uint16(index))
		}

		// delete unspent when spent in input
		for index := 0; index < len(b.Transactions[i].UTXOInputs); index++ {
			txhash := b.Transactions[i].UTXOInputs[index].ReferTxID

			// if get unspent by utxo
			if _, ok := unspents[txhash]; !ok {
				unspentvalue, err_get := bd.Get(append(unspentprefix, txhash.ToArray()...))

				if err_get != nil {
					return err_get
				}

				unspents[txhash], err_get = GetUint16Array(unspentvalue)
				if err_get != nil {
					return err_get
				}
			}

			// find Transactions[i].UTXOInputs[index].ReferTxOutputIndex and delete it
			for k := 0; k < len(unspents[txhash]); k++ {
				if unspents[txhash][k] == uint16(b.Transactions[i].UTXOInputs[index].ReferTxOutputIndex) {
					unspents[txhash] = append(unspents[txhash], unspents[txhash][:k]...)
					unspents[txhash] = append(unspents[txhash], unspents[txhash][k+1:]...)
					break
				}
			}
		}

	}

	// batch put the unspents
	for txhash, value := range unspents {
		unspentkey := bytes.NewBuffer(nil)
		unspentkey.WriteByte(byte(IX_Unspent))
		txhash.Serialize(unspentkey)

		if len(value) == 0 {
			batch.Delete(unspentkey.Bytes())
		} else {
			unspentarray := ToByteArray(value)
			batch.Put(unspentkey.Bytes(), unspentarray)
		}
	}

	// batch put quantities
	for assetid, value := range quantities {
		quantitykey := bytes.NewBuffer(nil)
		quantitykey.WriteByte(byte(ST_QuantityIssued))
		assetid.Serialize(quantitykey)

		qt, err := bd.GetQuantityIssued(assetid)
		if err != nil {
			return err
		}

		qt = qt + value

		quantityarray := bytes.NewBuffer(nil)
		qt.Serialize(quantityarray)

		batch.Put(quantitykey.Bytes(), quantityarray.Bytes())
		log.Debug(fmt.Sprintf("quantitykey: %x\n", quantitykey.Bytes()))
		log.Debug(fmt.Sprintf("quantityarray: %x\n", quantityarray.Bytes()))
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

	// BATCH PUT VALUE
	batch.Put(currentblockkey.Bytes(), currentblock.Bytes())

	//bh := b.Blockdata.Hash()
	//bd.header_index[bd.current_block_height] = &bh

	err = bd.db.Write(batch, nil)
	if err != nil {
		return err
	}

	return nil
}

func (bd *LevelDBStore) onAddHeader(header *Header, batch *leveldb.Batch) {
	log.Debug(fmt.Sprintf("onAddHeader(), Height=%d\n", header.Blockdata.Height))

	hash := header.Blockdata.Hash()
	bd.header_index[header.Blockdata.Height] = hash
	for header.Blockdata.Height-bd.stored_header_count >= 2000 {
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
	blockhash := header.Blockdata.Hash()
	blockhash.Serialize(headerkey)
	log.Debug(fmt.Sprintf("header key: %x\n", headerkey))
	//fmt.Println( "header key:",  headerkey.Bytes() )

	// generate value
	w := bytes.NewBuffer(nil)
	var sysfee uint64 = 0xFFFFFFFFFFFFFFFF
	serialization.WriteUint64(w, sysfee)
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
	serialization.WriteUint32(currentheader, header.Blockdata.Height)
	//fmt.Printf( "SYS_CurrentHeader data: %x\n",  currentheader )

	// PUT VALUE
	batch.Put(currentheaderkey.Bytes(), currentheader.Bytes())
}

func (bd *LevelDBStore) persistBlocks(ledger *Ledger) {

	bd.mu.Lock()
	defer bd.mu.Unlock()

	for !bd.disposed {
		if uint32(len(bd.header_index)) < bd.current_block_height+1 {
			log.Warn("[persistBlocks]: warn, header_index.count < current_block_height + 1")
			break
		}

		hash := bd.header_index[bd.current_block_height+1]

		block, ok := bd.block_cache[hash]
		if !ok {
			log.Warn("[persistBlocks]: warn, block_cache not contain key hash.")
			break
		}
		bd.persist(block)

		// PersistCompleted event
		ledger.Blockchain.BCEvents.Notify(events.EventBlockPersistCompleted, block)

		delete(bd.block_cache, hash)
	}

}

func (bd *LevelDBStore) SaveBlock(b *Block, ledger *Ledger) error {
	log.Debug("SaveBlock()")

	bd.mu.Lock()
	defer bd.mu.Unlock()

	if bd.block_cache[b.Hash()] == nil {
		bd.block_cache[b.Hash()] = b
	}

	log.Info("len(bd.header_index) is ", len(bd.header_index), " ,b.Blockdata.Height is ", b.Blockdata.Height)
	if b.Blockdata.Height >= (uint32(len(bd.header_index)) + 1) {
		//return false,NewDetailErr(errors.New(fmt.Sprintf("WARNING: [SaveBlock] block height - header_index.count >= 1, block height:%d, header_index.count:%d",b.Blockdata.Height, uint32(len(bd.header_index)) )),ErrDuplicatedBlock,"")
		return errors.New(fmt.Sprintf("WARNING: [SaveBlock] block height - header_index.count >= 1, block height:%d, header_index.count:%d", b.Blockdata.Height, uint32(len(bd.header_index))))
	}

	if b.Blockdata.Height == uint32(len(bd.header_index)) {
		//Block verify
		err := validation.VerifyBlock(b, ledger, false)
		if err != nil {
			log.Debug("VerifyBlock() error!")
			return err
		}

		batch := new(leveldb.Batch)
		h := new(Header)
		h.Blockdata = b.Blockdata
		bd.onAddHeader(h, batch)
		//log.Debug("batch dump: ", batch.Dump())
		err = bd.db.Write(batch, nil)
		if err != nil {
			return err
		}
	} else {
		return errors.New("[SaveBlock] block height != header_index")
	}

	if b.Blockdata.Height < uint32(len(bd.header_index)) {
		go bd.persistBlocks(ledger)
	} else {
		return errors.New("[SaveBlock] block height < header_index")
	}

	return nil
}

func (bd *LevelDBStore) GetQuantityIssued(assetid Uint256) (Fixed64, error) {
	log.Debug(fmt.Sprintf("GetQuantityIssued Hash: %x\n", assetid))

	prefix := []byte{byte(ST_QuantityIssued)}
	data, err_get := bd.Get(append(prefix, assetid.ToArray()...))
	log.Debug(fmt.Sprintf("GetQuantityIssued Data: %x\n", data))

	var quantity Fixed64
	if err_get != nil {
		quantity = Fixed64(0)
	} else {
		r := bytes.NewReader(data)
		quantity.Deserialize(r)
	}

	return quantity, nil
}

func (bd *LevelDBStore) GetUnspent(txid Uint256, index uint16) (*tx.TxOutput, error) {
	//fmt.Println( "GetUnspent()" )

	if ok, _ := bd.ContainsUnspent(txid, index); ok {
		tx, err := bd.GetTransaction(txid)
		if err != nil {
			return nil, err
		}

		return tx.Outputs[index], nil
	}

	return nil, errors.New("[GetUnspent] NOT ContainsUnspent.")
}

func (bd *LevelDBStore) ContainsUnspent(txid Uint256, index uint16) (bool, error) {
	unspentprefix := []byte{byte(IX_Unspent)}
	unspentvalue, err_get := bd.Get(append(unspentprefix, txid.ToArray()...))

	if err_get != nil {
		return false, err_get
	}

	unspentarray, err_get := GetUint16Array(unspentvalue)
	if err_get != nil {
		return false, err_get
	}

	for i := 0; i < len(unspentarray); i++ {
		if unspentarray[i] == index {
			return true, nil
		}
	}

	return false, nil
}

func (bd *LevelDBStore) GetCurrentHeaderHash() Uint256 {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	return bd.header_index[uint32(len(bd.header_index)-1)]
}

func (bd *LevelDBStore) GetHeaderHeight() uint32 {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	return uint32(len(bd.header_index) - 1)
}

func (bd *LevelDBStore) GetHeight() uint32 {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	return bd.current_block_height
}
