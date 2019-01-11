package shardmgmt

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	SHARD_STATE_CREATED    = iota
	SHARD_STATE_CONFIGURED // all parameter configured
	SHARD_STATE_READY      // all peers joined
	SHARD_STATE_ACTIVE     // started
	SHARD_STATE_ARCHIVED
)

type shardMgmtGlobalState struct {
	NextShardID uint64 `json:"next_shard_id"`
}

func (this *shardMgmtGlobalState) Serialize(w io.Writer) error {
	return serJson(w, this)
}

func (this *shardMgmtGlobalState) Deserialize(r io.Reader) error {
	return desJson(r, this)
}

type shardConfig struct {
	Config               []byte         `json:"config"`
	StakeContractAddress common.Address `json:"stake_contract_address"`
}

func (this *shardConfig) Serialize(w io.Writer) error {
	return serJson(w, this)
}

func (this *shardConfig) Deserialize(r io.Reader) error {
	return desJson(r, this)
}

func getVersion(native *native.NativeService, contract common.Address) (uint32, error) {
	versionBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(KEY_VERSION)))
	if err != nil {
		return 0, fmt.Errorf("get version: %s", err)
	}

	if versionBytes == nil {
		return 0, nil
	}

	value, err := cstates.GetValueFromRawStorageItem(versionBytes)
	if err != nil {
		return 0, fmt.Errorf("get versoin, deserialized from raw storage item: %s", err)
	}

	ver, err := serialization.ReadUint32(bytes.NewBuffer(value))
	if err != nil {
		return 0, fmt.Errorf("serialization.ReadUint32, deserialize version: %s", err)
	}
	return ver, nil
}

func setVersion(native *native.NativeService, contract common.Address) error {
	buf := new(bytes.Buffer)
	if err := serialization.WriteUint32(buf, VERSION_CONTRACT_SHARD_MGMT); err != nil {
		return fmt.Errorf("failed to serialize version: %s", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_VERSION)), buf.Bytes())
	return nil
}

func getGlobalState(native *native.NativeService, contract common.Address) (*shardMgmtGlobalState, error) {
	stateBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(KEY_GLOBAL_STATE)))
	if err != nil {
		return nil, fmt.Errorf("get shardmgmt global state: %s", err)
	}

	value, err := cstates.GetValueFromRawStorageItem(stateBytes)
	if err != nil {
		return nil, fmt.Errorf("get shardmgmt global state, deserialize from raw storage: %s", err)
	}

	globalState := &shardMgmtGlobalState{}
	if err := globalState.Deserialize(bytes.NewBuffer(value)); err != nil {
		return nil, fmt.Errorf("get shardgmgmtm global state: deserialize state: %s", err)
	}

	return globalState, nil
}

func setGlobalState(native *native.NativeService, contract common.Address, state *shardMgmtGlobalState) error {
	if state == nil {
		return fmt.Errorf("setGlobalState, nil state")
	}

	buf := new(bytes.Buffer)
	if err := state.Serialize(buf); err != nil {
		return fmt.Errorf("serialize shardmgmt global state: %s", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_GLOBAL_STATE)), buf.Bytes())
	return nil
}

type shardState struct {
	ShardID        uint64         `json:"shard_id"`
	Creator        common.Address `json:"creator"`
	State          uint32         `json:"state"`
	Config         *shardConfig   `json:"config"`
	PeerPubkeyList []string       `json:"peer_pubkey_list"`
}

func (this *shardState) Serialize(w io.Writer) error {
	return serJson(w, this)
}

func (this *shardState) Deserialize(r io.Reader) error {
	return desJson(r, this)
}

func getShardState(native *native.NativeService, contract common.Address, shardID uint64) (*shardState, error) {
	shardIDBytes, err := GetUint64Bytes(shardID)
	if err != nil {
		return nil, fmt.Errorf("getShardState, serialize shardID: %s", err)
	}

	shardStateBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(KEY_SHARD_STATE), shardIDBytes))
	if err != nil {
		return nil, fmt.Errorf("getShardState: %s", err)
	}
	if shardIDBytes == nil {
		return nil, nil
	}

	value, err := cstates.GetValueFromRawStorageItem(shardStateBytes)
	if err != nil {
		return nil, fmt.Errorf("getShardState, deserialize from raw storage: %s", err)
	}

	state := &shardState{}
	if err := state.Deserialize(bytes.NewBuffer(value)); err != nil {
		return nil, fmt.Errorf("getShardState, deserialize shardState: %s", err)
	}

	return state, nil
}

func setShardState(native *native.NativeService, contract common.Address, state *shardState) error {
	if state == nil {
		return fmt.Errorf("setShardState, nil state")
	}

	shardIDBytes, err := GetUint64Bytes(state.ShardID)
	if err != nil {
		return fmt.Errorf("setShardState, serialize shardID: %s", err)
	}

	buf := new(bytes.Buffer)
	if err := state.Serialize(buf); err != nil {
		return fmt.Errorf("serialize shardstate: %s", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_SHARD_STATE), shardIDBytes), buf.Bytes())
	return nil
}
