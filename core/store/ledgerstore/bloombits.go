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
package ledgerstore

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common/bitutil"
	"github.com/ethereum/go-ethereum/core/types"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/bloombits"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
)

// BloomBitsBlocks is the number of blocks a single bloom bit section vector
// contains on the server side.
const BloomBitsBlocks = 4096

var (
	bloomBitsPrefix = []byte("B") // bloomBitsPrefix + bit (uint16 big endian) + section (uint32 big endian) + hash -> bloom bits
)

func PutBloomIndex(store *leveldbstore.LevelDBStore, blooms []types.Bloom, section uint32) {
	gen, err := bloombits.NewGenerator(uint(BloomBitsBlocks))
	if err != nil {
		panic(err) // never fired since BloomBitsBlocks is multiple of 8
	}
	for i, b := range blooms {
		err := gen.AddBloom(uint(i), b)
		if err != nil {
			panic(err) // never fired
		}
	}

	for i := 0; i < types.BloomBitLength; i++ {
		bits, err := gen.Bitset(uint(i))
		if err != nil {
			panic(err) // never fired since idx is always less than 8 and section should be right
		}
		value := bitutil.CompressBytes(bits)
		store.BatchPut(bloomBitsKey(uint(i), section), value)
	}
}

func PutFilterStart(db *leveldbstore.LevelDBStore, height uint32) error {
	key := genFilterStartKey()
	sink := common2.NewZeroCopySink(nil)
	sink.WriteUint32(height)
	return db.Put(key, sink.Bytes())
}

func GetOrSetFilterStart(db *leveldbstore.LevelDBStore, def uint32) (uint32, error) {
	start, err := GetFilterStart(db)
	if err != nil {
		if err != scom.ErrNotFound {
			return 0, fmt.Errorf("get filter start: %s", err.Error())
		}

		err = PutFilterStart(db, def)
		if err != nil {
			return 0, fmt.Errorf("put filter start: %s", err.Error())
		}
		start = def
	}

	return start, nil
}

func GetFilterStart(db *leveldbstore.LevelDBStore) (uint32, error) {
	key := genFilterStartKey()
	data, err := db.Get(key)
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

// bloomBitsKey = bloomBitsPrefix + bit (uint16 big endian) + section (uint32 big endian)
func bloomBitsKey(bit uint, section uint32) []byte {
	key := append(bloomBitsPrefix, make([]byte, 6)...)

	binary.BigEndian.PutUint16(key[1:], uint16(bit))
	binary.BigEndian.PutUint32(key[3:], section)

	return key
}

// ReadBloomBits retrieves the compressed bloom bit vector belonging to the given
// section and bit index from the.
func ReadBloomBits(db *leveldbstore.LevelDBStore, bit uint, section uint32) ([]byte, error) {
	return db.Get(bloomBitsKey(bit, section))
}
