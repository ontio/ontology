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
	"os"

	"github.com/ontio/ontology/common"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/types"
)

const (
	DBDirCrossChain = "crosschain"
)

//Block store save the data of block & transaction
type CrossChainStore struct {
	dbDir string                     //The path of store file
	store *leveldbstore.LevelDBStore //block store handler
}

//NewCrossChainStore return cross chain store instance
func NewCrossChainStore(dataDir string) (*CrossChainStore, error) {
	dbDir := fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), DBDirCrossChain)
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, fmt.Errorf("NewCrossShardStore error %s", err)
	}
	return &CrossChainStore{
		dbDir: dbDir,
		store: store,
	}, nil
}

func (this *CrossChainStore) SaveMsgToCrossChainStore(crossChainMsg *types.CrossChainMsg) error {
	if crossChainMsg == nil {
		return nil
	}
	key := this.genCrossChainMsgKey(crossChainMsg.Height)
	sink := common.NewZeroCopySink(nil)
	crossChainMsg.Serialization(sink)
	return this.store.Put(key, sink.Bytes())
}

func (this *CrossChainStore) GetCrossChainMsg(height uint32) (*types.CrossChainMsg, error) {
	key := this.genCrossChainMsgKey(height)
	value, err := this.store.Get(key)
	if err != nil && err != scom.ErrNotFound {
		return nil, err
	}
	if err == scom.ErrNotFound {
		return nil, nil
	}
	source := common.NewZeroCopySource(value)
	msg := new(types.CrossChainMsg)
	if err := msg.Deserialization(source); err != nil {
		return nil, err
	}
	return msg, nil
}

func (this *CrossChainStore) genCrossChainMsgKey(height uint32) []byte {
	temp := make([]byte, 5)
	temp[0] = byte(scom.SYS_CROSS_CHAIN_MSG)
	binary.LittleEndian.PutUint32(temp[1:], height)
	return temp
}
