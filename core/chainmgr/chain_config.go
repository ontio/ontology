package chainmgr

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common/config"
)

func (self *ChainManager) buildShardConfig(shardID uint64) (*config.OntologyConfig, error) {
	if cfg := self.GetShardConfig(shardID); cfg != nil {
		return cfg, nil
	}

	shardState, err := self.getShardState(shardID)
	if err != nil {
		return nil, fmt.Errorf("get shardmgmt state: %s", err)
	}

	// check if shardID is in children
	buf := new(bytes.Buffer)
	if err := config.DefConfig.Serialize(buf); err != nil {
		return nil, fmt.Errorf("serialize parent config: %s", err)
	}

	shardConfig := &config.OntologyConfig{}
	if err := shardConfig.Deserialize(buf); err != nil {
		return nil, fmt.Errorf("init child config: %s", err)
	}

	// FIXME: solo only
	if shardConfig.Genesis.ConsensusType != config.CONSENSUS_TYPE_SOLO {
		return nil, fmt.Errorf("only solo suppported")
	}
	seedlist := make([]string, 0)
	bookkeepers := make([]string, 0)
	for peerPK, info := range shardState.Peers {
		seedlist = append(seedlist, info.PeerAddress)
		bookkeepers = append(bookkeepers, peerPK)
	}
	shardConfig.Genesis.SOLO.Bookkeepers = bookkeepers
	shardConfig.Genesis.SeedList = seedlist

	// TODO: init config for shard $shardID, including genesis config, data dir, net port, etc
	shardName := GetShardName(shardID)
	shardConfig.P2PNode.NodePort = 10000 + uint(shardID)
	shardConfig.P2PNode.NetworkName = shardName

	// init child shard config
	shardConfig.Shard = &config.ShardConfig{
		ShardID:              shardID,
		ShardPort:            uint(uint64(self.ShardPort) + shardID - self.ShardID),
		ParentShardID:        self.ShardID,
		ParentShardIPAddress: config.DEFAULT_PARENTSHARD_IPADDR,
		ParentShardPort:      self.ShardPort,
	}

	self.setShardConfig(shardID, shardConfig)

	return shardConfig, nil
}

func GetShardName(shardID uint64) string {
	return fmt.Sprintf("shard_%d", shardID)
}