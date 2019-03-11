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
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const KEY_CURR = "current"

type Storage struct {
	*StorageBackend
	task     chan Task
	currHash common.Uint256
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

	task := make(chan Task)
	go backend.blockSaveLoop(task)

	return &Storage{backend, task, backend.CurrHash()}, nil
}

func (self *Storage) SaveBlock(block *types.Block) {
	if block.Header.PrevBlockHash != self.currHash {
		return
	}

	sink := common.NewZeroCopySink(nil)
	headerLen, err := block.SerializeExt(sink)
	if err != nil {
		log.Errorf("serialize block err: %v", err)
		return
	}
	raw := sink.Bytes()
	self.task <- &SaveTask{
		block: &RawBlock{Hash: block.Hash(), HeaderSize:headerLen, Height: block.Header.Height, Payload: raw},
	}
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
	hash     common.Uint256
	offset   uint64
	height   uint32
	headerSize uint32
	size     uint32
	checksum common.Uint256
}

type RawBlock struct {
	Hash       common.Uint256
	Height     uint32
	HeaderSize uint32
	Payload    []byte
}

func (self *RawBlock) Size() int {
	return len(self.Payload)
}

func (self *BlockMeta) Bytes() []byte {
	buf := make([]byte, 0, 32+8+4+4+32+4)
	sink := common.NewZeroCopySink(buf)

	sink.WriteHash(self.hash)
	sink.WriteUint64(self.offset)
	sink.WriteUint32(self.height)
	sink.WriteUint32(self.headerSize)
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

func (self *StorageBackend) GetBlockByHash(hash common.Uint256) (*RawBlock, error) {
	metaRaw, err := self.metaDB.Get(hash[:], nil)
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

	return &RawBlock{Hash: hash, HeaderSize:meta.headerSize, Height: meta.height, Payload: buf}, nil
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

func (self *StorageBackend) saveBlock(block *RawBlock) error {
	if self.currInfo.nextHeight != block.Height {
		return fmt.Errorf("need continue block")
	}
	self.currInfo.checksum.Write(block.Payload)

	meta := BlockMeta{
		hash:   block.Hash,
		height: block.Height,
		headerSize: uint32(block.HeaderSize),
		size:   uint32(block.Size()),
		offset: self.currInfo.blockOffset,
	}
	self.currInfo.checksum.Sum(meta.checksum[:0])
	_, err := self.blockDB.Write(block.Payload)
	checkerr(err)

	self.batch.Put(meta.hash[:], meta.Bytes())

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

func (self *StorageBackend) blockSaveLoop(task <-chan Task) {
	for {
		select {
		case t, ok := <-task:
			if ok == false {
				self.flush()
				return
			}
			switch task := t.(type) {
			case *SaveTask:
				self.saveBlock(task.block)
			case *FlushTask:
				self.flush()
				task.finished <- self.currInfo.nextHeight - 1
			}
		case <-time.After(MAX_TIME_OUT):
			self.flush()
		}
	}
}

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}
