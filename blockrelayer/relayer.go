package blockrelayer

import (
	"bytes"
	"crypto/sha256"
	"encoding"
	"encoding/binary"
	"fmt"
	"github.com/ontio/ontology/common/log"
	"hash"
	"io"
	"math/rand"
	"os"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const KEY_CURR = "current"

var DefStorage *Storage

type Storage struct {
	backend          *StorageBackend
	task             chan Task
	currHash         common.Uint256
	currHeight       uint32
	lock             *sync.Mutex
	headers          map[common.Uint256]*types.RawTrustedHeader
	headerIndex      map[uint32]common.Uint256
	currHeaderHeight uint32
}

type Task interface {
	ImplementTask()
}

type implTask struct{}

func (self implTask) ImplementTask() {}

type FlushTask struct {
	implTask
	finished chan<- uint32
}

type SaveTask struct {
	implTask
	block *RawBlock
}

func Open(pt string) (*Storage, error) {
	backend, err := open(pt)
	if err != nil {
		return nil, err
	}

	task := make(chan Task, 1000)
	headers := make(map[common.Uint256]*types.RawTrustedHeader)
	headerIndex := make(map[uint32]common.Uint256)
	lock := new(sync.Mutex)
	store := &Storage{backend, task, backend.CurrHash(), backend.CurrHeight(),
		lock, headers, headerIndex, backend.CurrHeight()}
	go store.blockSaveLoop(task)
	return store, nil
}

func (self *Storage) SaveBlock(block *types.Block) error {

	sink := common.NewZeroCopySink(nil)
	headerLen, unsignedLen, err := block.SerializeExt(sink)
	if err != nil {
		log.Errorf("serialize block err: %v", err)
		return err
	}
	raw := sink.Bytes()
	self.task <- &SaveTask{
		block: &RawBlock{Hash: block.Hash(), HeaderSize: headerLen, unSignedHeaderSize: unsignedLen, Height: block.Header.Height, Payload: raw},
	}

	return nil
}

func (self *Storage) blockSaveLoop(task <-chan Task) {
	for {
		select {
		case t, ok := <-task:
			if ok == false {
				self.backend.flush()
				return
			}
			switch task := t.(type) {
			case *SaveTask:
				err := self.backend.saveBlock(task.block)
				if err != nil {
					log.Warnf("[replayer] saveBlock warning:%v", err)
					continue
				}
				self.currHeight = task.block.Height

			case *FlushTask:
				self.backend.flush()
				task.finished <- self.backend.currInfo.nextHeight - 1
			}
		case <-time.After(MAX_TIME_OUT):
			self.backend.flush()
			nextHeight := self.backend.currInfo.nextHeight

			self.lock.Lock()
			for k, v := range self.headers {
				if v.Height+100 < nextHeight {
					delete(self.headers, k)
				}
			}
			for height, _ := range self.headerIndex {
				if height+100 < nextHeight {
					delete(self.headerIndex, height)
				}
			}
			self.lock.Unlock()
		}
	}
}

func (self *Storage) AddHeader(headers []*types.RawHeader) error {
	sort.Slice(headers, func(i, j int) bool {
		return headers[i].Height < headers[j].Height
	})
	if self.CurrHeaderHash() != headers[0].PrevBlockHash {
		return fmt.Errorf("[relayer] AddHeader check hash failed")
	}
	self.lock.Lock()
	defer self.lock.Unlock()
	for _, header := range headers {
		self.headers[header.Hash()] = header.GetTrustedHeader()
		self.headerIndex[header.Height] = header.Hash()
	}
	self.currHeaderHeight = headers[len(headers)-1].Height
	return nil
}

func (self *Storage) GetHeaderByHash(hash common.Uint256) (*types.RawTrustedHeader, error) {
	self.lock.Lock()
	header, ok := self.headers[hash]
	self.lock.Unlock()
	if ok {
		return header, nil
	} else {
		header, err := self.backend.getHeader(hash[:])
		if err != nil {
			return nil, err
		}
		return header, nil
	}
}

func (self *Storage) CurrHeaderHeight() uint32 {
	if self.currHeaderHeight == 0 {
		return self.backend.CurrHeight()
	}
	return self.currHeaderHeight
}

func (self *Storage) CurrHeaderHash() common.Uint256 {
	self.lock.Lock()
	headerHash, ok := self.headerIndex[self.currHeaderHeight]
	self.lock.Unlock()
	if ok {
		return headerHash
	}
	return self.backend.CurrHash()
}

func (self *Storage) Flush() uint32 {
	finished := make(chan uint32)
	self.task <- &FlushTask{finished: finished}

	return <-finished
}

func (self *Storage) SaveBlockTest(height uint32) {
	raw := make([]byte, 20000)
	rand.Read(raw)
	var blockHash common.Uint256
	binary.LittleEndian.PutUint32(blockHash[:], height)

	self.task <- &SaveTask{
		block: &RawBlock{Hash: blockHash, HeaderSize: uint32(len(raw) / 3), Height: height, Payload: raw},
	}
}

func (self *Storage) CurrentHeight() uint32 {
	return self.currHeight
}

func (self *Storage) GetBlockByHash(hash common.Uint256) (*RawBlock, error) {
	return self.backend.GetBlockByHash(hash)
}

type CurrInfo struct {
	blockOffset uint64
	nextHeight  uint32
	currHash    common.Uint256
	checksum    hash.Hash // sha256
}

func NewCurrInfo() CurrInfo {
	return CurrInfo{checksum: sha256.New()}
}

func (self *CurrInfo) Bytes() []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(self.blockOffset)
	sink.WriteUint32(self.nextHeight)
	sink.WriteHash(self.currHash)
	cs, _ := self.checksum.(encoding.BinaryMarshaler).MarshalBinary()
	sink.WriteVarBytes(cs)
	return sink.Bytes()
}

func CurrInfoFromBytes(buf []byte) (info CurrInfo, err error) {
	var eof bool
	source := common.NewZeroCopySource(buf)
	info.blockOffset, eof = source.NextUint64()
	info.nextHeight, eof = source.NextUint32()
	info.currHash, eof = source.NextHash()
	cs, _, irr, eof := source.NextVarBytes()
	if irr {
		err = common.ErrIrregularData
		return
	}
	if eof {
		err = io.ErrUnexpectedEOF
		return
	}

	info.checksum = sha256.New()
	err = info.checksum.(encoding.BinaryUnmarshaler).UnmarshalBinary(cs)
	return
}

type BlockMeta struct {
	hash               common.Uint256
	offset             uint64
	height             uint32
	headerSize         uint32
	unSignedHeaderSize uint32
	size               uint32
	checksum           common.Uint256
}

type RawBlockMeta struct {
	rawMeta []byte
}

func NewRawBlockMeta(raw []byte) RawBlockMeta {
	if len(raw) != 32+8+4+4+4+32+4 {
		panic("wrong meta block len")
	}
	return RawBlockMeta{rawMeta: raw}
}

func (self *RawBlockMeta) Hash() common.Uint256 {
	var hs common.Uint256
	copy(hs[:], self.rawMeta)

	return hs
}

type RawBlock struct {
	Hash               common.Uint256
	Height             uint32
	HeaderSize         uint32
	unSignedHeaderSize uint32
	Payload            []byte
}

func (self *RawBlock) Size() int {
	return len(self.Payload)
}

func (self *BlockMeta) Bytes() []byte {
	buf := make([]byte, 0, 32+8+4+4+4+32+4)
	sink := common.NewZeroCopySink(buf)
	sink.WriteHash(self.hash)
	sink.WriteUint64(self.offset)
	sink.WriteUint32(self.height)
	sink.WriteUint32(self.headerSize)
	sink.WriteUint32(self.unSignedHeaderSize)
	sink.WriteUint32(self.size)
	sink.WriteHash(self.checksum)

	return sink.Bytes()
}

func BlockMetaFromBytes(raw []byte) (meta BlockMeta, err error) {
	var eof bool
	source := common.NewZeroCopySource(raw)
	meta.hash, eof = source.NextHash()
	meta.offset, eof = source.NextUint64()
	meta.height, eof = source.NextUint32()
	meta.headerSize, eof = source.NextUint32()
	meta.unSignedHeaderSize, eof = source.NextUint32()
	meta.size, eof = source.NextUint32()
	meta.checksum, eof = source.NextHash()
	if eof {
		err = io.ErrUnexpectedEOF
	}

	return
}

const MAX_PENDING_BLOCKS = 50
const MAX_PENDING_SIZE = 20 * 1024 * 1024
const MAX_TIME_OUT = 30 * time.Second

type StorageBackend struct {
	metaDB  *leveldb.DB
	blockDB *os.File

	currInfo      CurrInfo
	batch         *leveldb.Batch
	pendingBlocks int
	pendingSize   int
	checkedHeight uint32
}

func OpenLevelDB(file string) (*leveldb.DB, error) {
	openFileCache := opt.DefaultOpenFilesCacheCapacity

	// default Options
	o := opt.Options{
		NoSync:                 false,
		OpenFilesCacheCapacity: openFileCache,
		Filter:                 filter.NewBloomFilter(10),
	}

	db, err := leveldb.OpenFile(file, &o)

	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(file, nil)
	}

	if err != nil {
		return nil, err
	}

	return db, nil
}

func open(pt string) (*StorageBackend, error) {
	metaDB, err := OpenLevelDB(path.Join(pt, "metadb"))
	if err != nil {
		return nil, err
	}
	name := path.Join(pt, "blocks.bin")
	blockDB, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	info := NewCurrInfo()
	raw, err := metaDB.Get([]byte(KEY_CURR), nil)
	if err == nil {
		info, err = CurrInfoFromBytes(raw)
		if err != nil {
			return nil, err
		}
	} else if err != errors.ErrNotFound {
		return nil, err
	}

	stat, err := blockDB.Stat()
	if err != nil {
		return nil, fmt.Errorf("get block db stat err:%v", err)
	}

	if stat.Size() < int64(info.blockOffset) {
		return nil, errors.New("the length of blocks.bin is less than the record of metadb")
	} else if stat.Size() > int64(info.blockOffset) {
		err := blockDB.Truncate(int64(info.blockOffset))
		checkerr(err)
	}

	store := &StorageBackend{
		metaDB:   metaDB,
		blockDB:  blockDB,
		currInfo: info,
		batch:    new(leveldb.Batch),
	}

	valid, err := store.checkDataConsistence()
	if err != nil {
		return nil, err
	}

	if valid == false {
		//todo : add recover
		return nil, errors.New("db inconsistant")
	}

	return store, nil
}

func (self *StorageBackend) checkDataConsistence() (bool, error) {
	checksum := sha256.New()
	reader := &io.LimitedReader{R: self.blockDB, N: int64(self.currInfo.blockOffset)}
	_, err := io.Copy(checksum, reader)
	if err != nil {
		return false, err
	}

	return bytes.Equal(checksum.Sum(nil), self.currInfo.checksum.Sum(nil)), nil
}

func (self *StorageBackend) GetBlockByHeight(height uint32) (*RawBlock, error) {
	var metaKey [4]byte
	binary.BigEndian.PutUint32(metaKey[:], height)
	return self.getBlock(metaKey[:])
}

func (self *Storage) GetBlockHash(height uint32) (common.Uint256, error) {
	self.lock.Lock()
	hash, ok := self.headerIndex[height]
	self.lock.Unlock()
	if ok {
		return hash, nil
	}
	var metaKey [4]byte
	binary.BigEndian.PutUint32(metaKey[:], height)
	raw, err := self.backend.metaDB.Get(metaKey[:], nil)
	if err != nil {
		return common.UINT256_EMPTY, err
	}

	rawMeta := NewRawBlockMeta(raw)
	return rawMeta.Hash(), nil
}

func (self *StorageBackend) getBlock(metaKey []byte) (*RawBlock, error) {
	metaRaw, err := self.metaDB.Get(metaKey, nil)
	if err != nil {
		return nil, err
	}

	meta, err := BlockMetaFromBytes(metaRaw)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, meta.size)
	_, err = self.blockDB.ReadAt(buf, int64(meta.offset))
	if err != nil {
		return nil, err
	}
	if meta.height < self.checkedHeight {
		if checkBlockHashConsistence(buf[0:meta.unSignedHeaderSize], meta) {
			self.checkedHeight = meta.height
		} else {
			log.Error("[relayer] getBlock checkBlockHashConsistence failed")
			return nil, fmt.Errorf("[relayer] getBlock  checkBlockHashConsistence failed")
		}
	}
	return &RawBlock{Hash: meta.hash, HeaderSize: meta.headerSize, unSignedHeaderSize: meta.unSignedHeaderSize, Height: meta.height, Payload: buf}, nil
}

func checkBlockHashConsistence(buf []byte, meta BlockMeta) bool {
	temp := sha256.Sum256(buf)
	hash := common.Uint256(sha256.Sum256(temp[:]))
	return meta.hash == hash
}
func (self *StorageBackend) getHeader(metaKey []byte) (*types.RawTrustedHeader, error) {
	metaRaw, err := self.metaDB.Get(metaKey, nil)
	if err != nil {
		return nil, err
	}

	meta, err := BlockMetaFromBytes(metaRaw)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, meta.headerSize)
	_, err = self.blockDB.ReadAt(buf, int64(meta.offset))
	if err != nil {
		return nil, err
	}
	if checkBlockHashConsistence(buf[0:meta.unSignedHeaderSize], meta) {
		self.checkedHeight = meta.height
	} else {
		log.Error("[relayer] getHeader checkBlockHashConsistence failed")
		return nil, fmt.Errorf("[relayer] getHeader checkBlockHashConsistence failed")
	}
	header := &types.RawTrustedHeader{
		Height:  meta.height,
		Payload: buf,
	}

	return header, nil
}

func (self *StorageBackend) GetBlockByHash(hash common.Uint256) (*RawBlock, error) {
	return self.getBlock(hash[:])
}

func (self *StorageBackend) flush() {
	err := self.blockDB.Sync()
	checkerr(err)
	wo := opt.WriteOptions{Sync: true}
	self.batch.Put([]byte(KEY_CURR), self.currInfo.Bytes())
	err = self.metaDB.Write(self.batch, &wo)
	checkerr(err)
	self.batch = new(leveldb.Batch)
	self.pendingBlocks = 0
	self.pendingSize = 0
}

func (self *StorageBackend) CurrHash() common.Uint256 {
	return self.currInfo.currHash
}

func (self *StorageBackend) NextHeight() uint32 {
	return self.currInfo.nextHeight
}
func (self *StorageBackend) CurrHeight() uint32 {
	if self.currInfo.nextHeight > 0 {
		return self.currInfo.nextHeight - 1
	} else {
		return 0
	}
}
func (self *StorageBackend) saveBlock(block *RawBlock) error {
	if self.currInfo.nextHeight != block.Height {
		return fmt.Errorf("need continue block")
	}
	self.currInfo.checksum.Write(block.Payload)

	meta := BlockMeta{
		hash:               block.Hash,
		height:             block.Height,
		headerSize:         uint32(block.HeaderSize),
		unSignedHeaderSize: block.unSignedHeaderSize,
		size:               uint32(block.Size()),
		offset:             self.currInfo.blockOffset,
	}
	self.currInfo.checksum.Sum(meta.checksum[:0])
	_, err := self.blockDB.Write(block.Payload)
	checkerr(err)

	self.batch.Put(meta.hash[:], meta.Bytes())
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], meta.height)
	self.batch.Put(b[:], meta.Bytes())

	self.currInfo.blockOffset += uint64(block.Size())
	self.currInfo.nextHeight += 1
	self.currInfo.currHash = block.Hash
	self.pendingBlocks += 1
	self.pendingSize += block.Size()
	if self.pendingBlocks >= MAX_PENDING_BLOCKS || self.pendingSize >= MAX_PENDING_SIZE {
		self.flush()
	}

	return nil
}

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}
