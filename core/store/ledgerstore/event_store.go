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
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/payload"
	scom "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/events/message"
	"github.com/ontio/ontology/smartcontract/event"
)

//Saving event notifies gen by smart contract execution
type EventStore struct {
	dbDir string                     //Store path
	store *leveldbstore.LevelDBStore //Store handler
}

//NewEventStore return event store instance
func NewEventStore(dbDir string) (*EventStore, error) {
	store, err := leveldbstore.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	return &EventStore{
		dbDir: dbDir,
		store: store,
	}, nil
}

//NewBatch start event commit batch
func (this *EventStore) NewBatch() {
	this.store.NewBatch()
}

//SaveEventNotifyByTx persist event notify by transaction hash
func (this *EventStore) SaveEventNotifyByTx(txHash common.Uint256, notify *event.ExecuteNotify) error {
	result, err := json.Marshal(notify)
	if err != nil {
		return fmt.Errorf("json.Marshal error %s", err)
	}
	key := this.getEventNotifyByTxKey(txHash)
	this.store.BatchPut(key, result)
	return nil
}

//SaveEventNotifyByBlock persist transaction hash which have event notify to store
func (this *EventStore) SaveEventNotifyByBlock(height uint32, txHashs []common.Uint256) error {
	key, err := this.getEventNotifyByBlockKey(height)
	if err != nil {
		return err
	}

	values := bytes.NewBuffer(nil)
	err = serialization.WriteUint32(values, uint32(len(txHashs)))
	if err != nil {
		return err
	}
	for _, txHash := range txHashs {
		err = txHash.Serialize(values)
		if err != nil {
			return err
		}
	}
	this.store.BatchPut(key, values.Bytes())

	return nil
}

//GetEventNotifyByTx return event notify by trasanction hash
func (this *EventStore) GetEventNotifyByTx(txHash common.Uint256) (*event.ExecuteNotify, error) {
	key := this.getEventNotifyByTxKey(txHash)
	data, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	var notify event.ExecuteNotify
	if err = json.Unmarshal(data, &notify); err != nil {
		return nil, fmt.Errorf("json.Unmarshal error %s", err)
	}
	return &notify, nil
}

//GetEventNotifyByBlock return all event notify of transaction in block
func (this *EventStore) GetEventNotifyByBlock(height uint32) ([]*event.ExecuteNotify, error) {
	key, err := this.getEventNotifyByBlockKey(height)
	if err != nil {
		return nil, err
	}
	data, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewBuffer(data)
	size, err := serialization.ReadUint32(reader)
	if err != nil {
		return nil, fmt.Errorf("ReadUint32 error %s", err)
	}
	evtNotifies := make([]*event.ExecuteNotify, 0)
	for i := uint32(0); i < size; i++ {
		var txHash common.Uint256
		err = txHash.Deserialize(reader)
		if err != nil {
			return nil, fmt.Errorf("txHash.Deserialize error %s", err)
		}
		evtNotify, err := this.GetEventNotifyByTx(txHash)
		if err != nil {
			log.Errorf("getEventNotifyByTx Height:%d by txhash:%s error:%s", height, txHash.ToHexString(), err)
			continue
		}
		evtNotifies = append(evtNotifies, evtNotify)
	}
	return evtNotifies, nil
}

//CommitTo event store batch to store
func (this *EventStore) CommitTo() error {
	return this.store.BatchCommit()
}

//Close event store
func (this *EventStore) Close() error {
	return this.store.Close()
}

//ClearAll all data in event store
func (this *EventStore) ClearAll() error {
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

//SaveCurrentBlock persist current block height and block hash to event store
func (this *EventStore) SaveCurrentBlock(height uint32, blockHash common.Uint256) error {
	key := this.getCurrentBlockKey()
	sink := common.NewZeroCopySink(0)
	sink.WriteHash(blockHash)
	sink.WriteUint32(height)
	this.store.BatchPut(key, sink.Bytes())

	return nil
}

//GetCurrentBlock return current block hash, and block height
func (this *EventStore) GetCurrentBlock() (common.Uint256, uint32, error) {
	key := this.getCurrentBlockKey()
	data, err := this.store.Get(key)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	reader := bytes.NewReader(data)
	blockHash := common.Uint256{}
	err = blockHash.Deserialize(reader)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	height, err := serialization.ReadUint32(reader)
	if err != nil {
		return common.Uint256{}, 0, err
	}
	return blockHash, height, nil
}

func (this *EventStore) getCurrentBlockKey() []byte {
	return []byte{byte(scom.SYS_CURRENT_BLOCK)}
}

func (this *EventStore) getEventNotifyByBlockKey(height uint32) ([]byte, error) {
	key := make([]byte, 5, 5)
	key[0] = byte(scom.EVENT_NOTIFY)
	binary.LittleEndian.PutUint32(key[1:], height)
	return key, nil
}

func (this *EventStore) getEventNotifyByTxKey(txHash common.Uint256) []byte {
	data := txHash.ToArray()
	key := make([]byte, 1+len(data))
	key[0] = byte(scom.EVENT_NOTIFY)
	copy(key[1:], data)
	return key
}

func (this *EventStore) SaveContractMetaDataEvent(height uint32, metaData *payload.MetaDataCode) error {
	heightsList, err := this.GetContractMetaHeights(metaData.Contract)
	if err != nil && err != scom.ErrNotFound {
		return fmt.Errorf("SaveContractMetaDataEvent: get contract meta heights failed, err: %s", err)
	}
	heightsNum := len(heightsList)
	if heightsNum > 0 {
		if height < heightsList[heightsNum-1] {
			return fmt.Errorf("SaveContractMetaDataEvent: save height unmatch")
		} else if height > heightsList[heightsNum-1] {
			heightsList = append(heightsList, height)
		}
	} else {
		heightsList = []uint32{height}
	}
	this.SaveContractMetaHeights(metaData.Contract, heightsList)
	key := getContractMetaDataKey(height, metaData.Contract)
	value := common.NewZeroCopySink(64)
	metaData.Serialization(value)
	this.store.BatchPut(key, value.Bytes())
	return nil
}

func (this *EventStore) GetContractMetaDataEvent(height uint32, contractAddr common.Address) (*payload.MetaDataCode, error) {
	heightsList, err := this.GetContractMetaHeights(contractAddr)
	if err != nil {
		return nil, fmt.Errorf("GetContractMetaDataEvent: get contract meta heights failed, err: %s", err)
	}
	heightNum := len(heightsList)
	if heightNum == 0 {
		return nil, fmt.Errorf("GetContractMetaDataEvent: heights list empty")
	}
	destHeight := height
	if heightsList[heightNum-1] < height {
		destHeight = heightsList[heightNum-1]
	} else if heightsList[0] > height {
		return nil, fmt.Errorf("GetContractMetaDataEvent: height is too low")
	} else {
		for i, h := range heightsList {
			if h > height {
				destHeight = heightsList[i-1]
				break
			} else if h == height {
				destHeight = heightsList[i]
				break
			}
		}
	}

	key := getContractMetaDataKey(destHeight, contractAddr)
	data, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	source := common.NewZeroCopySource(data)
	metaData := &payload.MetaDataCode{}
	err = metaData.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return metaData, nil
}

func getContractMetaDataKey(height uint32, contractAddr common.Address) []byte {
	key := common.NewZeroCopySink(10)
	key.WriteByte(byte(scom.CROSS_SHARD_CONTRACT_META))
	key.WriteUint32(height)
	key.WriteBytes(contractAddr[:])
	return key.Bytes()
}

func (this *EventStore) SaveContractMetaHeights(contractAddr common.Address, data []uint32) {
	key := getContractMetaDataHeightsKey(contractAddr)
	value := common.NewZeroCopySink(16)
	value.WriteUint32(uint32(len(data)))
	for _, height := range data {
		value.WriteUint32(height)
	}
	this.store.BatchPut(key, value.Bytes())
}

func (this *EventStore) GetContractMetaHeights(contractAddr common.Address) ([]uint32, error) {
	key := getContractMetaDataHeightsKey(contractAddr)
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
		configheight, eof := source.NextUint32()
		if eof {
			return nil, io.ErrUnexpectedEOF
		}
		heights = append(heights, configheight)
	}
	return heights, nil

}

func getContractMetaDataHeightsKey(contractAddr common.Address) []byte {
	key := common.NewZeroCopySink(10)
	key.WriteByte(byte(scom.CROSS_SHARD_CONTRACT_META_HEIGHT))
	key.WriteBytes(contractAddr[:])
	return key.Bytes()
}

func (this *EventStore) SaveContractEvent(evt *message.ContractLifetimeEvent) error {
	oldEvt, err := this.GetContractEvent(evt.Contract.Address())
	if err != nil && err != scom.ErrNotFound {
		return fmt.Errorf("read old contract evt failed, err: %s", err)
	}
	// update event because contract destroyed or migrated
	if oldEvt != nil {
		evt.DeployHeight = oldEvt.DeployHeight
		evt.Contract = oldEvt.Contract
	}
	key := getContractEventKey(evt.Contract.Address())
	value := common.NewZeroCopySink(0)
	evt.Serialization(value)
	this.store.BatchPut(key, value.Bytes())
	return nil
}

func (this *EventStore) GetContractEvent(addr common.Address) (*message.ContractLifetimeEvent, error) {
	key := getContractEventKey(addr)
	data, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	source := common.NewZeroCopySource(data)
	evt := &message.ContractLifetimeEvent{}
	err = evt.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return evt, nil
}

func getContractEventKey(contractAddr common.Address) []byte {
	key := common.NewZeroCopySink(10)
	key.WriteByte(byte(scom.CROSS_SHARD_CONTRACT_EVENT))
	key.WriteBytes(contractAddr[:])
	return key.Bytes()
}

func (this *EventStore) AddShardConsensusHeight(shardID common.ShardID, data []uint32) {
	key := genShardConsensusHeightKey(shardID)
	value := common.NewZeroCopySink(16)
	value.WriteUint32(uint32(len(data)))
	for _, height := range data {
		value.WriteUint32(height)
	}
	this.store.BatchPut(key, value.Bytes())
}

func (this *EventStore) GetShardConsensusHeight(shardID common.ShardID) ([]uint32, error) {
	key := genShardConsensusHeightKey(shardID)
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
		configheight, eof := source.NextUint32()
		if eof {
			return nil, io.ErrUnexpectedEOF
		}
		heights = append(heights, configheight)
	}
	return heights, nil
}

func genShardConsensusHeightKey(shardID common.ShardID) []byte {
	key := common.NewZeroCopySink(9)
	key.WriteByte(byte(scom.CROSS_SHARD_HEIGHT))
	key.WriteShardID(shardID)
	return key.Bytes()
}

func (this *EventStore) AddShardConsensusConfig(shardID common.ShardID, height uint32, value []byte) error {
	key := genShardConsensusConfigKey(shardID, height)
	this.store.BatchPut(key, value)
	return nil
}

func (this *EventStore) GetShardConsensusConfig(shardID common.ShardID, height uint32) ([]byte, error) {
	key := genShardConsensusConfigKey(shardID, height)
	return this.store.Get(key)
}

func genShardConsensusConfigKey(shardID common.ShardID, height uint32) []byte {
	key := common.NewZeroCopySink(16)
	key.WriteByte(byte(scom.SHARD_CONFIG_DATA))
	key.WriteShardID(shardID)
	key.WriteUint32(height)
	return key.Bytes()
}
