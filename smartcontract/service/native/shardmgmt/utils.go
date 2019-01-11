package shardmgmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
)

func serJson(w io.Writer, v interface{}) error {
	buf, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("json marshal failed: %s", err)
	}

	if err := serialization.WriteVarBytes(w, buf); err != nil {
		return fmt.Errorf("json serialize write failed: %s", err)
	}
	return nil
}

func desJson(r io.Reader, v interface{}) error {
	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("json deserialize read failed: %s", err)
	}
	if err := json.Unmarshal(buf, v); err != nil {
		return fmt.Errorf("json unmarshal failed: %s", err)
	}
	return nil
}

func GetUint64Bytes(num uint64) ([]byte, error) {
	bf := new(bytes.Buffer)
	if err := serialization.WriteUint64(bf, num); err != nil {
		return nil, fmt.Errorf("serialization.WriteUint64, serialize uint64 error: %v", err)
	}
	return bf.Bytes(), nil
}

func GetBytesUint64(b []byte) (uint64, error) {
	num, err := serialization.ReadUint64(bytes.NewBuffer(b))
	if err != nil {
		return 0, fmt.Errorf("serialization.ReadUint64, deserialize uint64 error: %v", err)
	}
	return num, nil
}

func addNotification(native *native.NativeService, contract common.Address, info shardEvent) error {
	infoBuf := new(bytes.Buffer)
	if err := serJson(infoBuf, info); err != nil {
		return fmt.Errorf("addNotification, ser info: %s", err)
	}
	eventState := &shardEventState{
		Version:   VERSION_CONTRACT_SHARD_MGMT,
		EventType: info.getType(),
		Info:      infoBuf.Bytes(),
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{eventState},
		})
	return nil
}
