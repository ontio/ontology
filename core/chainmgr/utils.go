package chainmgr

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func (self *ChainManager) GetShardConfig(shardID uint64) *config.OntologyConfig {
	if s := self.shards[shardID]; s != nil {
		return s.Config
	}
	return nil
}

func (self *ChainManager) setShardConfig(shardID uint64, cfg *config.OntologyConfig) error {
	if info := self.shards[shardID]; info != nil {
		info.Config = cfg
		return nil
	}

	self.shards[shardID] = &ShardInfo{
		Config: cfg,
	}
	return nil
}

func (self *ChainManager) getShardMgmtGlobalState() (*shardstates.ShardMgmtGlobalState, error) {
	if self.ledger == nil {
		return nil, fmt.Errorf("uninitialized chain mgr")
	}

	data, err := self.ledger.GetStorageItem(utils.ShardMgmtContractAddress, []byte(shardmgmt.KEY_VERSION))
	if err != nil {
		return nil, fmt.Errorf("get shardmgmt version: %s", err)
	}
	if len(data) == 0 {
		return nil, nil
	}

	ver, err := serialization.ReadUint32(bytes.NewBuffer(data))
	if ver != shardmgmt.VERSION_CONTRACT_SHARD_MGMT {
		return nil, fmt.Errorf("uncompatible version: %d", ver)
	}

	data, err = self.ledger.GetStorageItem(utils.ShardMgmtContractAddress, []byte(shardmgmt.KEY_GLOBAL_STATE))
	if err != nil {
		return nil, fmt.Errorf("get shardmgmt global state: %s", err)
	}
	if len(data) == 0 {
		return nil, nil
	}

	globalState := &shardstates.ShardMgmtGlobalState{}
	if err := globalState.Deserialize(bytes.NewBuffer(data)); err != nil {
		return nil, fmt.Errorf("des shardmgmt global state: %s", err)
	}

	return globalState, nil
}

func (self *ChainManager) getShardState(shardID uint64) (*shardstates.ShardState, error) {
	if self.ledger == nil {
		return nil, fmt.Errorf("uninitialized chain mgr")
	}

	shardIDBytes, err := shardutil.GetUint64Bytes(shardID)
	if err != nil {
		return nil, fmt.Errorf("ser shardID failed: %s", err)
	}
	key := append([]byte(shardmgmt.KEY_SHARD_STATE), shardIDBytes...)
	data, err := self.ledger.GetStorageItem(utils.ShardMgmtContractAddress, key)
	if err != nil {
		return nil, fmt.Errorf("get shardmgmt global state: %s", err)
	}

	shardState := &shardstates.ShardState{}
	if err := shardState.Deserialize(bytes.NewBuffer(data)); err != nil {
		return nil, fmt.Errorf("des shardmgmt shard state: %s", err)
	}

	return shardState, nil
}
