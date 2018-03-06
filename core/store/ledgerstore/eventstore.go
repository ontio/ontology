package ledgerstore

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/core/store/common"
	"github.com/Ontology/core/store/leveldbstore"
	"github.com/Ontology/smartcontract/event"
)

type EventStore struct {
	dbDir string
	store *leveldbstore.LevelDBStore
}

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

func (this *EventStore) NewBatch() error {
	return this.store.NewBatch()
}

func (this *EventStore) SaveEventNotifyByTx(txHash *common.Uint256, notifies []*event.NotifyEventInfo) error {
	result, err := json.Marshal(notifies)
	if err != nil {
		return fmt.Errorf("json.Marshal error %s", err)
	}
	key, err := this.getEventNotifyByTxKey(txHash)
	if err != nil {
		return err
	}
	return this.store.BatchPut(key, result)
}

func (this *EventStore) SaveEventNotifyByBlock(height uint32, txHashs []*common.Uint256) error {
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
		_, err = txHash.Serialize(values)
		if err != nil {
			return err
		}
	}
	return this.store.BatchPut(key, values.Bytes())
}

func (this *EventStore) GetEventNotifyByTx(txHash *common.Uint256) ([]*event.NotifyEventInfo, error) {
	key, err := this.getEventNotifyByTxKey(txHash)
	if err != nil {
		return nil, err
	}
	data, err := this.store.Get(key)
	if err != nil {
		if IsLevelDBNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	var notifies []*event.NotifyEventInfo
	if err = json.Unmarshal(data, &notifies); err != nil {
		return nil, fmt.Errorf("json.Unmarshal error %s", err)
	}
	return notifies, nil
}

func (this *EventStore) GetEventNotifyByBlock(height uint32) ([]*common.Uint256, error) {
	key, err := this.getEventNotifyByBlockKey(height)
	if err != nil {
		return nil, err
	}
	data, err := this.store.Get(key)
	if err != nil {
		if IsLevelDBNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	reader := bytes.NewBuffer(data)
	size, err := serialization.ReadUint32(reader)
	if err != nil {
		return nil, fmt.Errorf("ReadUint32 error %s", err)
	}
	txHashs := make([]*common.Uint256, 0, size)
	for i := uint32(0); i < size; i++ {
		var txHash common.Uint256
		err = txHash.Deserialize(reader)
		if err != nil {
			return nil, fmt.Errorf("txHash.Deserialize error %s", err)
		}
		txHashs = append(txHashs, &txHash)
	}
	return txHashs, nil
}

func (this *EventStore) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *EventStore) Close() error{
	return this.store.Close()
}

func (this *EventStore) ClearAll() error {
	err := this.NewBatch()
	if err != nil {
		return err
	}
	iter := this.store.NewIterator(nil)
	for iter.Next() {
		err = this.store.BatchDelete(iter.Key())
		if err != nil {
			return fmt.Errorf("BatchDelete error %s", err)
		}
	}
	iter.Release()
	return this.CommitTo()
}

func (this *EventStore) getEventNotifyByBlockKey(height uint32) ([]byte, error) {
	key := make([]byte, 5, 5)
	key[0] = byte(EVENT_Notify)
	binary.LittleEndian.PutUint32(key[1:], height)
	return key, nil
}

func (this *EventStore) getEventNotifyByTxKey(txHash *common.Uint256) ([]byte, error) {
	data := txHash.ToArray()
	key := make([]byte, 1+len(data))
	key[0] = byte(EVENT_Notify)
	copy(key[1:], data)
	return key, nil
}
