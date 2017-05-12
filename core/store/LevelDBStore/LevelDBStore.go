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
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"
)

type LevelDBStore struct {
	db *leveldb.DB // LevelDB instance
	b  *leveldb.Batch
	it *Iterator

	headerIndex map[uint32]Uint256
	blockCache  map[Uint256]*Block
	headerCache map[Uint256]*Header

	currentBlockHeight uint32
	storedHeaderCount  uint32

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
		db:                 db,
		b:                  nil,
		it:                 nil,
		headerIndex:        map[uint32]Uint256{},
		blockCache:         map[Uint256]*Block{},
		headerCache:        map[Uint256]*Header{},
		currentBlockHeight: 0,
		storedHeaderCount:  0,
		disposed:           false,
	}, nil
}

func (bd *LevelDBStore) InitLevelDBStoreWithGenesisBlock(genesisBlock *Block) (uint32, error) {

	hash := genesisBlock.Hash()
	bd.headerIndex[0] = hash
	log.Debug(fmt.Sprintf("listhash genesis: %x\n", hash))

	prefix := []byte{byte(CFG_Version)}
	version, err := bd.Get(prefix)
	if err != nil {
		version = []byte{0x00}
	}

	if version[0] == 0x01 {
		// Get Current Block
		currentBlockPrefix := []byte{byte(SYS_CurrentBlock)}
		data, err := bd.Get(currentBlockPrefix)
		if err != nil {
			return 0, err
		}

		r := bytes.NewReader(data)
		var blockHash Uint256
		blockHash.Deserialize(r)
		bd.currentBlockHeight, err = serialization.ReadUint32(r)
		current_Header_Height := bd.currentBlockHeight
		////////////////////////////////////////////////

		// Get Current Header
		var headerHash Uint256
		currentHeaderPrefix := []byte{byte(SYS_CurrentHeader)}
		data, err = bd.Get(currentHeaderPrefix)
		if err == nil {
			r = bytes.NewReader(data)
			headerHash.Deserialize(r)

			headerHeight, err_get := serialization.ReadUint32(r)
			if err_get != nil {
				return 0, err_get
			}

			current_Header_Height = headerHeight
		}

		log.Debug(fmt.Sprintf("blockHash: %x\n", blockHash.ToArray()))
		log.Debug(fmt.Sprintf("blockheight: %d\n", current_Header_Height))
		////////////////////////////////////////////////

		var listHash Uint256
		iter := bd.db.NewIterator(util.BytesPrefix([]byte{byte(IX_HeaderHashList)}), nil)
		for iter.Next() {
			rk := bytes.NewReader(iter.Key())
			// read prefix
			_, _ = serialization.ReadBytes(rk, 1)
			startNum, err := serialization.ReadUint32(rk)
			if err != nil {
				return 0, err
			}
			log.Debug(fmt.Sprintf("start index: %d\n", startNum))

			r = bytes.NewReader(iter.Value())
			listNum, err := serialization.ReadVarUint(r, 0)
			if err != nil {
				return 0, err
			}

			for i := 0; i < int(listNum); i++ {
				listHash.Deserialize(r)
				bd.headerIndex[startNum+uint32(i)] = listHash
				bd.storedHeaderCount++
				//log.Debug( fmt.Sprintf( "listHash %d: %x\n", startNum+uint32(i), listHash ) )
			}
		}

		if bd.storedHeaderCount == 0 {
			iter = bd.db.NewIterator(util.BytesPrefix([]byte{byte(DATA_BlockHash)}), nil)
			for iter.Next() {
				rk := bytes.NewReader(iter.Key())
				// read prefix
				_, _ = serialization.ReadBytes(rk, 1)
				listheight, err := serialization.ReadUint32(rk)
				if err != nil {
					return 0, err
				}
				//log.Debug(fmt.Sprintf( "DATA_BlockHash block height: %d\n", listheight ))

				r := bytes.NewReader(iter.Value())
				listHash.Deserialize(r)
				//log.Debug(fmt.Sprintf( "DATA_BlockHash block hash: %x\n", listHash ))

				bd.headerIndex[listheight] = listHash
			}
		} else if current_Header_Height >= bd.storedHeaderCount {
			hash = headerHash
			for {
				if hash == bd.headerIndex[bd.storedHeaderCount-1] {
					break
				}

				header, err := bd.GetHeader(hash)
				if err != nil {
					return 0, err
				}

				//log.Debug(fmt.Sprintf( "header height: %d\n", header.Blockdata.Height ))
				//log.Debug(fmt.Sprintf( "header hash: %x\n", hash ))

				bd.headerIndex[header.Blockdata.Height] = hash
				hash = header.Blockdata.PrevBlockHash
			}
		}

		return current_Header_Height, nil

	} else {
		// batch delete old data
		batch := new(leveldb.Batch)
		iter := bd.db.NewIterator(nil, nil)
		for iter.Next() {
			batch.Delete(iter.Key())
		}
		iter.Release()

		err := bd.db.Write(batch, nil)
		if err != nil {
			return 0, err
		}

		// persist genesis block
		bd.persist(genesisBlock)

		// put version to db
		err = bd.Put(prefix, []byte{0x01})
		if err != nil {
			return 0, err
		}

		return 0, nil
	}
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

func (bd *LevelDBStore) IsDoubleSpend(tx *tx.Transaction) bool {
	if len(tx.UTXOInputs) == 0 {
		return false
	}

	unspentPrefix := []byte{byte(IX_Unspent)}
	for i := 0; i < len(tx.UTXOInputs); i++ {
		txhash := tx.UTXOInputs[i].ReferTxID
		unspentValue, err_get := bd.Get(append(unspentPrefix, txhash.ToArray()...))
		if err_get != nil {
			return true
		}

		unspents, _ := GetUint16Array(unspentValue)
		findFlag := false
		for k := 0; k < len(unspents); k++ {
			if unspents[k] == tx.UTXOInputs[i].ReferTxOutputIndex {
				findFlag = true
				break
			}
		}

		if !findFlag {
			return true
		}
	}

	return false
}

func (bd *LevelDBStore) GetBlockHash(height uint32) (Uint256, error) {

	if height >= 0 {
		queryKey := bytes.NewBuffer(nil)
		queryKey.WriteByte(byte(DATA_BlockHash))
		err := serialization.WriteUint32(queryKey, height)

		if err == nil {
			blockHash, err_get := bd.Get(queryKey.Bytes())
			if err_get != nil {
				//TODO: implement error process
				return Uint256{}, err_get
			} else {
				blockHash256, err_parse := Uint256ParseFromBytes(blockHash)
				if err_parse == nil {
					return blockHash256, nil
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

	return bd.headerIndex[bd.currentBlockHeight]
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

func (bd *LevelDBStore) GetHeaderWithCache(hash Uint256) *Header {
	if _, ok := bd.headerCache[hash]; ok {
		return bd.headerCache[hash]
	}

	header, _ := bd.GetHeader(hash)

	return header
}

func (bd *LevelDBStore) containsBlock(hash Uint256) bool {
	header := bd.GetHeaderWithCache(hash)
	if header != nil {
		return header.Blockdata.Height <= bd.currentBlockHeight
	} else {
		return false
	}
}

func (bd *LevelDBStore) VerifyHeader(header *Header) bool {
	if bd.containsBlock(header.Blockdata.Hash()) {
		return true
	}

	prevHeader := bd.GetHeaderWithCache(header.Blockdata.PrevBlockHash)

	if prevHeader == nil {
		return false
	}

	if prevHeader.Blockdata.Height+1 != header.Blockdata.Height {
		return false
	}

	if prevHeader.Blockdata.Timestamp >= header.Blockdata.Timestamp {
		return false
	}

	flag, err := validation.VerifySignableData(header.Blockdata)
	if flag == false || err != nil {
		return false
	}

	return true
}

func (bd *LevelDBStore) AddHeaders(headers []Header, ledger *Ledger) error {
	bd.mu.Lock()
	defer bd.mu.Unlock()

	batch := new(leveldb.Batch)
	for i := 0; i < len(headers); i++ {
		if headers[i].Blockdata.Height >= (uint32(len(bd.headerIndex)) + 1) {
			break
		}

		if headers[i].Blockdata.Height < uint32(len(bd.headerIndex)) {
			continue
		}

		//header verify
		if !bd.VerifyHeader(&headers[i]) {
			log.Error("Verify header failed")
			break
		}

		bd.onAddHeader(&headers[i], batch)

		// add hash to header_cache
		bd.headerCache[headers[i].Blockdata.Hash()] = &headers[i]
	}

	err := bd.db.Write(batch, nil)
	if err != nil {
		return err
	}

	// clear header_cache
	for k, _ := range bd.headerCache {
		delete(bd.headerCache, k)
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

func (bd *LevelDBStore) SaveAsset(assetId Uint256, asset *Asset) error {
	w := bytes.NewBuffer(nil)

	asset.Serialize(w)

	// generate key
	assetKey := bytes.NewBuffer(nil)
	// add asset prefix.
	assetKey.WriteByte(byte(ST_Info))
	// contact asset id
	assetId.Serialize(assetKey)

	log.Debug(fmt.Sprintf("asset key: %x\n", assetKey))

	// PUT VALUE
	err := bd.Put(assetKey.Bytes(), w.Bytes())
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
	log.Debug()
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
	unspentPrefix := []byte{byte(IX_Unspent)}
	for i := 0; i < len(b.Transactions); i++ {
		txhash := b.Transactions[i].Hash()
		unspentValue, err_get := bd.Get(append(unspentPrefix, txhash.ToArray()...))

		if err_get != nil {
			unspentValue = []byte{}
		}

		unspents[txhash], err_get = GetUint16Array(unspentValue)
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
	blockHash := b.Hash()
	blockHash.Serialize(bhhash)
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
	hashWriter := bytes.NewBuffer(nil)
	hashValue := b.Blockdata.Hash()
	hashValue.Serialize(hashWriter)
	log.Debug(fmt.Sprintf("DATA_BlockHash table value: %x\n", hashValue))

	// BATCH PUT VALUE
	batch.Put(bhash.Bytes(), hashWriter.Bytes())

	//////////////////////////////////////////////////////////////
	// save transactions to leveldb
	nLen := len(b.Transactions)
	for i := 0; i < nLen; i++ {

		// now support RegisterAsset / IssueAsset / TransferAsset and Miner TX ONLY.
		if b.Transactions[i].TxType == tx.RegisterAsset ||
			b.Transactions[i].TxType == tx.IssueAsset ||
			b.Transactions[i].TxType == tx.TransferAsset ||
			b.Transactions[i].TxType == tx.Record ||
			b.Transactions[i].TxType == tx.BookKeeping {
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
			results := b.Transactions[i].GetMergedAssetIDValueFromOutputs()

			for assetId, value := range results {
				if _, ok := quantities[assetId]; !ok {
					quantities[assetId] += value
				} else {
					quantities[assetId] = value
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
				unspentValue, err_get := bd.Get(append(unspentPrefix, txhash.ToArray()...))

				if err_get != nil {
					return err_get
				}

				unspents[txhash], err_get = GetUint16Array(unspentValue)
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
		unspentKey := bytes.NewBuffer(nil)
		unspentKey.WriteByte(byte(IX_Unspent))
		txhash.Serialize(unspentKey)

		if len(value) == 0 {
			batch.Delete(unspentKey.Bytes())
		} else {
			unspentArray := ToByteArray(value)
			batch.Put(unspentKey.Bytes(), unspentArray)
		}
	}

	// batch put quantities
	for assetId, value := range quantities {
		quantityKey := bytes.NewBuffer(nil)
		quantityKey.WriteByte(byte(ST_QuantityIssued))
		assetId.Serialize(quantityKey)

		qt, err := bd.GetQuantityIssued(assetId)
		if err != nil {
			return err
		}

		qt = qt + value

		quantityArray := bytes.NewBuffer(nil)
		qt.Serialize(quantityArray)

		batch.Put(quantityKey.Bytes(), quantityArray.Bytes())
		log.Debug(fmt.Sprintf("quantityKey: %x\n", quantityKey.Bytes()))
		log.Debug(fmt.Sprintf("quantityArray: %x\n", quantityArray.Bytes()))
	}

	// save hash with height
	bd.currentBlockHeight = b.Blockdata.Height

	// generate key with SYS_CurrentHeader prefix
	currentBlockKey := bytes.NewBuffer(nil)
	currentBlockKey.WriteByte(byte(SYS_CurrentBlock))
	//fmt.Printf( "SYS_CurrentHeader key: %x\n",  currentBlockKey )

	currentBlock := bytes.NewBuffer(nil)
	blockHash.Serialize(currentBlock)
	serialization.WriteUint32(currentBlock, b.Blockdata.Height)

	// BATCH PUT VALUE
	batch.Put(currentBlockKey.Bytes(), currentBlock.Bytes())

	//bh := b.Blockdata.Hash()
	//bd.headerIndex[bd.currentBlockHeight] = &bh

	err = bd.db.Write(batch, nil)
	if err != nil {
		return err
	}

	return nil
}

func (bd *LevelDBStore) onAddHeader(header *Header, batch *leveldb.Batch) {
	log.Debug(fmt.Sprintf("onAddHeader(), Height=%d\n", header.Blockdata.Height))

	hash := header.Blockdata.Hash()
	bd.headerIndex[header.Blockdata.Height] = hash
	for header.Blockdata.Height-bd.storedHeaderCount >= 2000 {
		hashBuffer := new(bytes.Buffer)
		serialization.WriteVarUint(hashBuffer, uint64(2000))
		var hashArray []byte
		for i := 0; i < 2000; i++ {
			index := bd.storedHeaderCount + uint32(i)
			//fmt.Println("index:",index)
			thash := bd.headerIndex[index]
			thehash := thash.ToArray()
			hashArray = append(hashArray, thehash...)
			//fmt.Printf("%x\n",thehash)
		}
		hashBuffer.Write(hashArray)
		//fmt.Printf( "%x\n", hashBuffer )

		// generate key with DATA_Header prefix
		hhlPrefix := bytes.NewBuffer(nil)
		// add block header prefix.
		hhlPrefix.WriteByte(byte(IX_HeaderHashList))
		serialization.WriteUint32(hhlPrefix, bd.storedHeaderCount)
		//fmt.Printf( "%x\n", hhlPrefix )

		batch.Put(hhlPrefix.Bytes(), hashBuffer.Bytes())
		bd.storedHeaderCount += 2000
	}

	//////////////////////////////////////////////////////////////
	// generate key with DATA_Header prefix
	headerKey := bytes.NewBuffer(nil)
	// add header prefix.
	headerKey.WriteByte(byte(DATA_Header))
	// contact block hash
	blockHash := header.Blockdata.Hash()
	blockHash.Serialize(headerKey)
	log.Debug(fmt.Sprintf("header key: %x\n", headerKey))
	//fmt.Println( "header key:",  headerKey.Bytes() )

	// generate value
	w := bytes.NewBuffer(nil)
	var sysfee uint64 = 0xFFFFFFFFFFFFFFFF
	serialization.WriteUint64(w, sysfee)
	header.Serialize(w)
	log.Debug(fmt.Sprintf("header data: %x\n", w))
	//fmt.Println( "header data:",  w.Bytes() )

	// PUT VALUE
	batch.Put(headerKey.Bytes(), w.Bytes())

	//////////////////////////////////////////////////////////////
	// generate key with SYS_CurrentHeader prefix
	currentHeaderKey := bytes.NewBuffer(nil)
	currentHeaderKey.WriteByte(byte(SYS_CurrentHeader))
	//fmt.Printf( "SYS_CurrentHeader key: %x\n",  currentHeaderKey )

	currentHeader := bytes.NewBuffer(nil)
	blockHash.Serialize(currentHeader)
	serialization.WriteUint32(currentHeader, header.Blockdata.Height)
	//fmt.Printf( "SYS_CurrentHeader data: %x\n",  currentHeader )

	// PUT VALUE
	batch.Put(currentHeaderKey.Bytes(), currentHeader.Bytes())
}

func (bd *LevelDBStore) persistBlocks(ledger *Ledger) {
	bd.mu.Lock()
	defer bd.mu.Unlock()
	for !bd.disposed {
		if uint32(len(bd.headerIndex)) < bd.currentBlockHeight+1 {
			log.Warn("[persistBlocks]: warn, headerIndex.count < currentBlockHeight + 1")
			break
		}

		hash := bd.headerIndex[bd.currentBlockHeight+1]

		block, ok := bd.blockCache[hash]
		if !ok {
			log.Warn("[persistBlocks]: warn, blockCache not contain key hash.")
			break
		}
		bd.persist(block)

		// PersistCompleted event
		ledger.Blockchain.BCEvents.Notify(events.EventBlockPersistCompleted, block)

		delete(bd.blockCache, hash)
	}

}

func (bd *LevelDBStore) SaveBlock(b *Block, ledger *Ledger) error {
	log.Debug("SaveBlock()")

	bd.mu.Lock()
	defer bd.mu.Unlock()

	if bd.blockCache[b.Hash()] == nil {
		bd.blockCache[b.Hash()] = b
	}

	log.Info("len(bd.headerIndex) is ", len(bd.headerIndex), " ,b.Blockdata.Height is ", b.Blockdata.Height)
	if b.Blockdata.Height >= (uint32(len(bd.headerIndex)) + 1) {
		//return false,NewDetailErr(errors.New(fmt.Sprintf("WARNING: [SaveBlock] block height - headerIndex.count >= 1, block height:%d, headerIndex.count:%d",b.Blockdata.Height, uint32(len(bd.headerIndex)) )),ErrDuplicatedBlock,"")
		return errors.New(fmt.Sprintf("WARNING: [SaveBlock] block height - headerIndex.count >= 1, block height:%d, headerIndex.count:%d", b.Blockdata.Height, uint32(len(bd.headerIndex))))
	}

	if b.Blockdata.Height == uint32(len(bd.headerIndex)) {
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
		return errors.New("[SaveBlock] block height != headerIndex")
	}

	if b.Blockdata.Height < uint32(len(bd.headerIndex)) {
		go bd.persistBlocks(ledger)
	} else {
		return errors.New("[SaveBlock] block height < headerIndex")
	}

	return nil
}

func (db *LevelDBStore) BlockInCache(hash Uint256) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if _, ok := db.blockCache[hash]; ok {
		return true
	}
	return false
}

func (bd *LevelDBStore) GetQuantityIssued(assetId Uint256) (Fixed64, error) {
	log.Debug(fmt.Sprintf("GetQuantityIssued Hash: %x\n", assetId))

	prefix := []byte{byte(ST_QuantityIssued)}
	data, err_get := bd.Get(append(prefix, assetId.ToArray()...))
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
		Tx, err := bd.GetTransaction(txid)
		if err != nil {
			return nil, err
		}

		return Tx.Outputs[index], nil
	}

	return nil, errors.New("[GetUnspent] NOT ContainsUnspent.")
}

func (bd *LevelDBStore) ContainsUnspent(txid Uint256, index uint16) (bool, error) {
	unspentPrefix := []byte{byte(IX_Unspent)}
	unspentValue, err_get := bd.Get(append(unspentPrefix, txid.ToArray()...))

	if err_get != nil {
		return false, err_get
	}

	unspentArray, err_get := GetUint16Array(unspentValue)
	if err_get != nil {
		return false, err_get
	}

	for i := 0; i < len(unspentArray); i++ {
		if unspentArray[i] == index {
			return true, nil
		}
	}

	return false, nil
}

func (bd *LevelDBStore) GetCurrentHeaderHash() Uint256 {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	return bd.headerIndex[uint32(len(bd.headerIndex)-1)]
}

func (bd *LevelDBStore) GetHeaderHashByHeight(height uint32) Uint256 {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	return bd.headerIndex[height]
}

func (bd *LevelDBStore) GetHeaderHeight() uint32 {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	return uint32(len(bd.headerIndex) - 1)
}

func (bd *LevelDBStore) GetHeight() uint32 {
	bd.mu.RLock()
	defer bd.mu.RUnlock()

	return bd.currentBlockHeight
}
