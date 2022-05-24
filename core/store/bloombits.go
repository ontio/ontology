/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */
package store

import (
	"encoding/binary"
	"io"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/bitutil"
	"github.com/ethereum/go-ethereum/core/bloombits"
	"github.com/ethereum/go-ethereum/core/types"
	common2 "github.com/ontio/ontology/common"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
)

const (
	// bloomServiceThreads is the number of goroutines used globally by an Ethereum
	// instance to service bloombits lookups for all running filters.
	BloomServiceThreads = 16

	// bloomFilterThreads is the number of goroutines used locally per filter to
	// multiplex requests onto the global servicing goroutines.
	BloomFilterThreads = 3

	// bloomRetrievalBatch is the maximum number of bloom bit retrievals to service
	// in a single batch.
	BloomRetrievalBatch = 16

	// bloomRetrievalWait is the maximum time to wait for enough bloom bit requests
	// to accumulate request an entire batch (avoiding hysteresis).
	BloomRetrievalWait = time.Duration(0)

	// BloomBitsBlocks is the number of blocks a single bloom bit section vector
	// contains on the server side.
	BloomBitsBlocks uint32 = 4096
)

const (
	// bloomThrottling is the time to wait between processing two consecutive index
	// sections. It's useful during chain upgrades to prevent disk overload.
	bloomThrottling = 100 * time.Millisecond

	bloomIdxDir = "bloomIdx"
)

var (
	bloomBitsPrefix = []byte("B") // bloomBitsPrefix + bit (uint16 big endian) + section (uint32 big endian) + hash -> bloom bits
)

// bloomIndexer implements a core.ChainIndexer, building up a rotated bloom bits index
// for the Ethereum header bloom filters, permitting blazing fast filtering.
type bloomIndexer struct {
	store   *leveldbstore.LevelDBStore // database instance to write index data and metadata into
	gen     *bloombits.Generator       // generator to rotate the bloom bits crating the bloom index
	section uint32                     // Section is the section number being processed currently
	head    common.Hash                // Head is the hash of the last header processed
}

func initBloomIndexer(store *leveldbstore.LevelDBStore) bloomIndexer {
	return bloomIndexer{
		store: store,
	}
}

// Reset implements core.ChainIndexerBackend, starting a new bloombits index
// section.
func (b *bloomIndexer) Reset(section uint32) {
	gen, err := bloombits.NewGenerator(uint(BloomBitsBlocks))
	if err != nil {
		panic(err) // never fired since BloomBitsBlocks is multiple of 8
	}
	b.gen, b.section, b.head = gen, section, common.Hash{}
}

// Process implements core.ChainIndexerBackend, adding a new header's bloom into
// the index.
func (b *bloomIndexer) Process(hash common.Hash, height uint32, bloom types.Bloom, start uint32) {
	// the initial height is 1 but it on ethereum is 0. so subtract 1
	b.gen.AddBloom(uint(height-b.section*BloomBitsBlocks-start), bloom)
	b.head = hash
}

// Commit implements core.ChainIndexerBackend, finalizing the bloom section and
// writing it out into the database.
func (b *bloomIndexer) Commit() error {
	b.NewBatch()
	for i := 0; i < types.BloomBitLength; i++ {
		bits, err := b.gen.Bitset(uint(i))
		if err != nil {
			return err
		}
		value := bitutil.CompressBytes(bits)
		b.store.BatchPut(bloomBitsKey(uint(i), b.section, b.head), value)
	}
	err := b.CommitTo()
	return err
}

//NewBatch start a commit batch
func (this *bloomIndexer) NewBatch() {
	this.store.NewBatch()
}

//CommitTo commit the batch to store
func (this *bloomIndexer) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *bloomIndexer) PutFilterStart(height uint32) error {
	key := genFilterStartKey()
	sink := common2.NewZeroCopySink(nil)
	sink.WriteUint32(height)
	return this.store.Put(key, sink.Bytes())
}

func (this *bloomIndexer) GetFilterStart() (uint32, error) {
	key := genFilterStartKey()
	data, err := this.store.Get(key)
	if err != nil {
		return 0, err
	}
	height, eof := common2.NewZeroCopySource(data).NextUint32()
	if eof {
		return 0, io.ErrUnexpectedEOF
	}
	return height, nil
}

func genFilterStartKey() []byte {
	return []byte{byte(scom.ST_ETH_FILTER_START)}
}

// setValidSections writes the number of valid sections to the index database
func (this *bloomIndexer) SetValidSections(sections uint32) error {
	// Set the current number of valid sections in the database
	var data [4]byte
	binary.BigEndian.PutUint32(data[:], sections)
	return this.store.Put([]byte{byte(scom.ST_ETH_SECTIONS_COUNT)}, data[:])
}

func (this *bloomIndexer) GetValidSections() uint32 {
	data, _ := this.store.Get([]byte{byte(scom.ST_ETH_SECTIONS_COUNT)})
	if len(data) == 4 {
		return binary.BigEndian.Uint32(data)
	}
	return 0
}

// bloomBitsKey = bloomBitsPrefix + bit (uint16 big endian) + section (uint32 big endian) + hash
func bloomBitsKey(bit uint, section uint32, hash common.Hash) []byte {
	key := append(append(bloomBitsPrefix, make([]byte, 6)...), hash.Bytes()...)

	binary.BigEndian.PutUint16(key[1:], uint16(bit))
	binary.BigEndian.PutUint32(key[3:], section)

	return key
}

// ReadBloomBits retrieves the compressed bloom bit vector belonging to the given
// section and bit index from the.
func ReadBloomBits(db *leveldbstore.LevelDBStore, bit uint, section uint32, head common.Hash) ([]byte, error) {
	return db.Get(bloomBitsKey(bit, section, head))
}
