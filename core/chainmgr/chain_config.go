package chainmgr

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common/config"
)

func (self *ChainManager) buildShardConfig(shardID uint64) ([]byte, error) {

	// check if shardID is in children
	buf := new(bytes.Buffer)
	if err := config.DefConfig.Serialize(buf); err != nil {
		return nil, fmt.Errorf("serialize parent config: %s", err)
	}

	childConfig := &config.OntologyConfig{}
	if err := childConfig.Deserialize(buf); err != nil {
		return nil, fmt.Errorf("init child config: %s", err)
	}

	// TODO: init config for shard $shardID, including genesis config, data dir, net port, etc
	shardName := GetShardName(shardID)
	childConfig.P2PNode.NodePort = 10000 + uint(shardID)
	childConfig.P2PNode.NetworkName = shardName

	// init child shard config
	childConfig.Shard = &config.ShardConfig{
		ShardID:              shardID,
		ShardPort:            uint(uint64(self.ShardPort) + shardID - self.ShardID),
		ParentShardID:        self.ShardID,
		ParentShardIPAddress: config.DEFAULT_PARENTSHARD_IPADDR,
		ParentShardPort:      self.ShardPort,
	}

	buf.Reset()
	if err := childConfig.Serialize(buf); err != nil {
		return nil, fmt.Errorf("serialized child config: %s", err)
	}
	return buf.Bytes(), nil
}

func GetShardName(shardID uint64) string {
	return fmt.Sprintf("shard_%d", shardID)
}