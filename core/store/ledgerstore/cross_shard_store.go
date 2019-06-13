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
	"os"

	"github.com/ontio/ontology/common"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/types"
)

var (
	//Storage save path.
	DBDirCrossShard = "crossshard"
)

//saving cross shard msg
type CrossShardStore struct {
	dbDir string                     //Store path
	store *leveldbstore.LevelDBStore //Store handler
}

//NewCrossShardStore return cross shard store instance
func NewCrossShardStore(dataDir string) (*CrossShardStore, error) {
	dbDir := fmt.Sprintf("%s%s%s", dataDir, string(os.PathSeparator), DBDirCrossShard)
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, fmt.Errorf("NewCrossShardStore error %s", err)
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

func (this *CrossShardStore) SaveCrossShardMsgByHash(msgHash common.Uint256, crossShardMsg *types.CrossShardMsg) error {
	this.NewBatch()
	key := this.getCrossShardMsgKeyByHash(msgHash)

	value := common.NewZeroCopySink(1024)

	crossShardMsg.Serialization(value)
	this.store.BatchPut(key, value.Bytes())
	err := this.CommitTo()
	if err != nil {
		return fmt.Errorf("crossShardStore.CommitTo msgHash:%s, error %s", msgHash.ToHexString(), err)
	}
	return nil
}

func (this *CrossShardStore) GetCrossShardMsgByHash(msgHash common.Uint256) (*types.CrossShardMsg, error) {
	key := this.getCrossShardMsgKeyByHash(msgHash)
	data, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	source := common.NewZeroCopySource(data)
	crossShardMsg := &types.CrossShardMsg{}
	crossShardMsg.Deserialization(source)
	return crossShardMsg, nil
}

func (this *CrossShardStore) getCrossShardMsgKeyByHash(msgHash common.Uint256) []byte {
	key := common.NewZeroCopySink(8)
	key.WriteByte(byte(scom.CROSS_SHARD_MSG))
	key.WriteBytes(msgHash[:])
	return key.Bytes()
}

func (this *CrossShardStore) SaveAllShardIDs(shardIDs []common.ShardID) error {
	this.NewBatch()
	key := this.getCrossShardIDKey()
	value := common.NewZeroCopySink(1024)
	value.WriteUint32(uint32(len(shardIDs)))
	for _, shardID := range shardIDs {
		value.WriteShardID(shardID)
	}
	this.store.BatchPut(key, value.Bytes())
	err := this.CommitTo()
	if err != nil {
		return fmt.Errorf("crossShardStore SaveAllShardIDs error %s", err)
	}
	return nil
}
func (this *CrossShardStore) GetAllShardIDs() ([]common.ShardID, error) {
	key := this.getCrossShardIDKey()
	data, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	source := common.NewZeroCopySource(data)
	shardIdCnt, eof := source.NextUint32()
	if eof {
		return nil, io.ErrUnexpectedEOF
	}
	shardIds := make([]common.ShardID, 0)
	for i := uint32(0); i < shardIdCnt; i++ {
		shardId, err := source.NextShardID()
		if err != nil {
			return nil, io.ErrUnexpectedEOF
		}
		shardIds = append(shardIds, shardId)
	}
	return shardIds, nil
}
func (this *CrossShardStore) getCrossShardIDKey() []byte {
	key := common.NewZeroCopySink(8)
	key.WriteByte(byte(scom.CROSS_ALL_SHARDS))
	return key.Bytes()
}

func (this *CrossShardStore) SaveCrossShardHash(shardID common.ShardID, msgHash common.Uint256) error {
	this.NewBatch()
	key := this.getCrossShardKeyByHash(shardID)
	value := common.NewZeroCopySink(64)
	value.WriteBytes(msgHash[:])
	this.store.BatchPut(key, value.Bytes())
	err := this.CommitTo()
	if err != nil {
		return fmt.Errorf("crossShardStore.CommitTo shardID:%v,msgHash:%s, error %s", shardID, msgHash.ToHexString(), err)
	}
	return nil
}

func (this *CrossShardStore) GetCrossShardHash(shardID common.ShardID) (common.Uint256, error) {
	key := this.getCrossShardKeyByHash(shardID)
	buf, err := this.store.Get(key)
	if err != nil {
		return common.Uint256{}, err
	}
	source := common.NewZeroCopySource(buf)
	msgHash, eof := source.NextHash()
	if eof {
		return common.Uint256{}, io.ErrUnexpectedEOF
	}
	return msgHash, nil
}

func (this *CrossShardStore) getCrossShardKeyByHash(shardID common.ShardID) []byte {
	key := common.NewZeroCopySink(8)
	key.WriteByte(byte(scom.CROSS_SHARD_HASH))
	key.WriteShardID(shardID)
	return key.Bytes()
}

func (this *CrossShardStore) AddShardConsensusConfig(shardID common.ShardID, height uint32, value []byte) error {
	this.NewBatch()
	key := this.genShardConsensusConfigKey(shardID, height)
	this.store.BatchPut(key, value)
	err := this.CommitTo()
	if err != nil {
		return fmt.Errorf("crossShardStore.CommitTo shardID:%v,height:%d error %s", shardID, height, err)
	}
	return nil
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

func (this *CrossShardStore) AddShardConsensusHeight(shardID common.ShardID, data []uint32) error {
	this.NewBatch()
	key := this.genShardConsensusHeightKey(shardID)
	value := common.NewZeroCopySink(16)
	value.WriteUint32(uint32(len(data)))
	for _, height := range data {
		value.WriteUint32(height)
	}
	this.store.BatchPut(key, value.Bytes())
	err := this.CommitTo()
	if err != nil {
		return fmt.Errorf("crossShardStore.CommitTo shardID:%v error %s", shardID, err)
	}
	return nil
}

func (this *CrossShardStore) GetShardConsensusHeight(shardID common.ShardID) ([]uint32, error) {
	key := this.genShardConsensusHeightKey(shardID)
	data, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	source := common.NewZeroCopySource(data)
	m, eof := source.NextUint32()
	if eof {
		return nil, io.ErrUnexpectedEOF
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

func (this *CrossShardStore) SaveShardMsgHash(shardID common.ShardID, msgHash common.Uint256) error {
	this.NewBatch()
	key := this.getCrossShardKeyByHash(shardID)
	value := common.NewZeroCopySink(64)
	value.WriteBytes(msgHash[:])
	this.store.BatchPut(key, value.Bytes())
	err := this.CommitTo()
	if err != nil {
		return fmt.Errorf("crossShardStore.CommitTo shardID:%v,msgHash:%s, error %s", shardID, msgHash.ToHexString(), err)
	}
	return nil
}

func (self *CrossShardStore) GetShardMsgHash(shardID common.ShardID) (common.Uint256, error) {
	keys := common.NewZeroCopySink(16)
	keys.WriteByte(byte(scom.XSHARD_KEY_MSG_HASH))
	keys.WriteShardID(shardID)
	buf, err := self.store.Get(keys.Bytes())
	if err != nil {
		return common.Uint256{}, err
	}
	source := common.NewZeroCopySource(buf)
	msgHash, eof := source.NextHash()
	if eof {
		return common.Uint256{}, io.ErrUnexpectedEOF
	}
	return msgHash, nil
}

func (this *CrossShardStore) getShardMsgKeyByShard(shardID common.ShardID) []byte {
	key := common.NewZeroCopySink(8)
	key.WriteByte(byte(scom.XSHARD_KEY_MSG_HASH))
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
