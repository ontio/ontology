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
	"github.com/ontio/ontology/common/log"
)

const (
	SHARD_STATE_CREATED    = iota
	SHARD_STATE_CONFIGURED // all parameter configured
	SHARD_STATE_READY      // all peers joined
	SHARD_STATE_ACTIVE     // started
	SHARD_STATE_ARCHIVED
)

type ShardMgmtGlobalState struct {
	NextShardID uint64 `json:"next_shard_id"`
}

func (this *ShardMgmtGlobalState) Serialize(w io.Writer) error {
	return serJson(w, this)
}

func (this *ShardMgmtGlobalState) Deserialize(r io.Reader) error {
	return desJson(r, this)
}

type ShardConfig struct {
	Config               []byte         `json:"config"`
	StakeContractAddress common.Address `json:"stake_contract_address"`
}

func (this *ShardConfig) Serialize(w io.Writer) error {
	return serJson(w, this)
}

func (this *ShardConfig) Deserialize(r io.Reader) error {
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

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_VERSION)), cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}

func getGlobalState(native *native.NativeService, contract common.Address) (*ShardMgmtGlobalState, error) {
	stateBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(KEY_GLOBAL_STATE)))
	if err != nil {
		return nil, fmt.Errorf("get shardmgmt global state: %s", err)
	}

	value, err := cstates.GetValueFromRawStorageItem(stateBytes)
	if err != nil {
		return nil, fmt.Errorf("get shardmgmt global state, deserialize from raw storage: %s", err)
	}

	globalState := &ShardMgmtGlobalState{}
	if err := globalState.Deserialize(bytes.NewBuffer(value)); err != nil {
		return nil, fmt.Errorf("get shardgmgmtm global state: deserialize state: %s", err)
	}

	return globalState, nil
}

func setGlobalState(native *native.NativeService, contract common.Address, state *ShardMgmtGlobalState) error {
	if state == nil {
		return fmt.Errorf("setGlobalState, nil state")
	}

	buf := new(bytes.Buffer)
	if err := state.Serialize(buf); err != nil {
		return fmt.Errorf("serialize shardmgmt global state: %s", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_GLOBAL_STATE)), cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}

type ShardState struct {
	ShardID        uint64         `json:"shard_id"`
	Creator        common.Address `json:"creator"`
	State          uint32         `json:"state"`
	Config         *ShardConfig   `json:"config"`
	PeerPubkeyList []string       `json:"peer_pubkey_list"`
}

func (this *ShardState) Serialize(w io.Writer) error {
	return serJson(w, this)
}

func (this *ShardState) Deserialize(r io.Reader) error {
	return desJson(r, this)
}

func getShardState(native *native.NativeService, contract common.Address, shardID uint64) (*ShardState, error) {
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

	state := &ShardState{}
	if err := state.Deserialize(bytes.NewBuffer(value)); err != nil {
		return nil, fmt.Errorf("getShardState, deserialize ShardState: %s", err)
	}

	return state, nil
}

func setShardState(native *native.NativeService, contract common.Address, state *ShardState) error {
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

	key := utils.ConcatKey(contract, []byte(KEY_SHARD_STATE), shardIDBytes)
	log.Infof("set shard %d , key %v, state: %s", state.ShardID, key, string(buf.Bytes()))
	native.CacheDB.Put(key, cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}
