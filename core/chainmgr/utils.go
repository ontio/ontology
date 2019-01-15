package chainmgr

import "github.com/ontio/ontology/common/config"

func (self *ChainManager) GetShardConfig(shardID uint64) *config.OntologyConfig {
	if s := self.Shards[shardID]; s != nil {
		return s.Config
	}
	return nil
}

func (self *ChainManager) setShardConfig(shardID uint64, cfg *config.OntologyConfig) error {
	if info := self.Shards[shardID]; info != nil {
		info.Config = cfg
		return nil
	}

	self.Shards[shardID] = &ShardInfo{
		Config: cfg,
	}
	return nil
}

