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
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/types"
)

//saving cross shard msg
type CrossShardStore struct {
	dbDir string                     //Store path
	store *leveldbstore.LevelDBStore //Store handler
}

//NewCrossShardStore return cross shard store instance
func NewCrossShardStore(dbDir string) (*CrossShardStore, error) {
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	return &CrossShardStore{
		dbDir: dbDir,
		store: store,
	}, nil
}

//NewBatch start event commit batch
func (this *CrossShardStore) NewBatch() {
	this.store.NewBatch()
}

func (this *CrossShardStore) SaveCrossShardMsgByShardID(shardID common.ShardID, crossShardTxInfos []*types.CrossShardTxInfos) error {
	key := this.getCrossShardMsgKeyByShard(shardID)

	value := common.NewZeroCopySink(1024)
	value.WriteUint32(uint32(len(crossShardTxInfos)))
	for _, crossShardTx := range crossShardTxInfos {
		err := crossShardTx.Serialization(value)
		if err != nil {
			return err
		}
	}
	this.store.BatchPut(key, value.Bytes())
	return nil
}

func (this *CrossShardStore) GetCrossShardMsgByShardID(shardID common.ShardID) ([]*types.CrossShardTxInfos, error) {
	key := this.getCrossShardMsgKeyByShard(shardID)
	data, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	source := common.NewZeroCopySource(data)
	txCnt, eof := source.NextUint32()
	if eof {
		return nil, io.ErrUnexpectedEOF
	}
	crossShardTxInfos := make([]*types.CrossShardTxInfos, 0)
	for i := uint32(0); i < txCnt; i++ {
		crossShardTxInfo := new(types.CrossShardTxInfos)
		err := crossShardTxInfo.Deserialization(source)
		if err != nil {
			return nil, fmt.Errorf("deserialize shard tx: %s", err)
		}
		crossShardTxInfos = append(crossShardTxInfos, crossShardTxInfo)
	}
	return crossShardTxInfos, nil
}

func (this *CrossShardStore) getCrossShardMsgKeyByShard(shardID common.ShardID) []byte {
	key := common.NewZeroCopySink(8)
	key.WriteByte(byte(scom.CROSS_SHARD_MSG))
	key.WriteShardID(shardID)
	return key.Bytes()
}

func (this *CrossShardStore) AddShardConsensusConfig(shardID common.ShardID, height uint32, value []byte) {
	key := this.genShardConsensusConfigKey(shardID, height)
	this.store.BatchPut(key, value)
}

func (this *CrossShardStore) GetShardConsensusConfig(shardID common.ShardID, height uint32) ([]byte, error) {
	key := this.genShardConsensusConfigKey(shardID, height)
	return this.store.Get(key)
}

func (this *CrossShardStore) genShardConsensusConfigKey(shardID common.ShardID, height uint32) []byte {
	key := common.NewZeroCopySink(16)
	key.WriteByte(byte(scom.SHARD_CONFIG_DATA))
	key.WriteShardID(shardID)
	key.WriteUint32(height)
	return key.Bytes()
}

func (this *CrossShardStore) AddShardConsensusHeight(shardID common.ShardID, value []byte) {
	key := this.genShardConsensusHeightKey(shardID)
	this.store.BatchPut(key, value)
}

func (this *CrossShardStore) GetShardConsensusHeight(shardID common.ShardID) ([]uint32, error) {
	key := this.genShardConsensusHeightKey(shardID)
	data, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	source := common.NewZeroCopySource(data)
	m, _, irregular, eof := source.NextVarUint()
	if eof {
		return nil, io.ErrUnexpectedEOF
	}
	if irregular {
		return nil, common.ErrIrregularData
	}
	heights := make([]uint32, 0)
	for i := 0; i < int(m); i++ {
		config_height, eof := source.NextUint32()
		if eof {
			return nil, io.ErrUnexpectedEOF
		}
		heights = append(heights, config_height)
	}
	return heights, nil
}

func (this *CrossShardStore) genShardConsensusHeightKey(shardID common.ShardID) []byte {
	key := common.NewZeroCopySink(8)
	key.WriteByte(byte(scom.CROSS_SHARD_HEIGHT))
	key.WriteShardID(shardID)
	return key.Bytes()
}

//CommitTo cross shard store batch to store
func (this *CrossShardStore) CommitTo() error {
	return this.store.BatchCommit()
}

//Close cross shard store
func (this *CrossShardStore) Close() error {
	return this.store.Close()
}

//ClearAll all data in cross shard store
func (this *CrossShardStore) ClearAll() error {
	this.NewBatch()
	iter := this.store.NewIterator(nil)
	for iter.Next() {
		this.store.BatchDelete(iter.Key())
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return err
	}
	return this.CommitTo()
}
