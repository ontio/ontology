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
	"github.com/ontio/ontology/common"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/types"
	"os"
)

const (
	DBDirLayer2 = "layer2"
)

//Block store save the data of block & transaction
type Layer2Store struct {
	dbDir string                     //The path of store file
	store *leveldbstore.LevelDBStore //block store handler
}

//NewCrossChainStore return layer2 store instance
func NewLayer2Store(dataDir string) (*Layer2Store, error) {
	dbDir := fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), DBDirLayer2)
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, fmt.Errorf("Newlayer2Store error %s", err)
	}
	return &Layer2Store{
		dbDir: dbDir,
		store: store,
	}, nil
}

func (this *Layer2Store) SaveMsgToLayer2Store(layer2Msg *types.Layer2State) error {
	if layer2Msg == nil {
		return nil
	}
	key := this.genLayer2StateKey(layer2Msg.Height)
	sink := common.NewZeroCopySink(nil)
	layer2Msg.Serialization(sink)
	return this.store.Put(key, sink.Bytes())
}

func (this *Layer2Store) GetLayer2State(height uint32) (*types.Layer2State, error) {
	key := this.genLayer2StateKey(height)
	value, err := this.store.Get(key)
	if err != nil && err != scom.ErrNotFound {
		return nil, err
	}
	if err == scom.ErrNotFound {
		return nil, nil
	}
	source := common.NewZeroCopySource(value)
	msg := new(types.Layer2State)
	if err := msg.Deserialization(source); err != nil {
		return nil, err
	}
	return msg, nil
}

func (this *Layer2Store) genLayer2StateKey(height uint32) []byte {
	temp := make([]byte, 5)
	temp[0] = byte(scom.SYS_CROSS_CHAIN_MSG)
	binary.LittleEndian.PutUint32(temp[1:], height)
	return temp
}
