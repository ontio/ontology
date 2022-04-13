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
	"github.com/ethereum/go-ethereum/core/types"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"os"
)

const (
	DBDirBloom = "bloom"
)

//Bloom store save the data of block & transaction
type BloomStore struct {
	dbDir string                     //The path of store file
	store *leveldbstore.LevelDBStore //block store handler
}

//NewBloomStore return bloom store instance
func NewBloomStore(dataDir string) (*BloomStore, error) {
	dbDir := fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), DBDirBloom)
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	return &BloomStore{
		dbDir: dbDir,
		store: store,
	}, nil
}

func (this *BloomStore) SaveBloomData(height uint32, bloom types.Bloom) error {
	key := this.genBloomKey(height)
	return this.store.Put(key, bloom.Bytes())
}

func (this *BloomStore) GetBloomData(height uint32) (types.Bloom, error) {
	key := this.genBloomKey(height)
	value, err := this.store.Get(key)
	if err != nil && err != scom.ErrNotFound {
		return types.Bloom{}, err
	}
	if err == scom.ErrNotFound {
		return types.Bloom{}, nil
	}
	return types.BytesToBloom(value), nil
}

//Close BloomStore store
func (this *BloomStore) Close() error {
	return this.store.Close()
}

func (this *BloomStore) genBloomKey(height uint32) []byte {
	temp := make([]byte, 5)
	temp[0] = byte(scom.DATA_BLOOM)
	binary.LittleEndian.PutUint32(temp[1:], height)
	return temp
}
