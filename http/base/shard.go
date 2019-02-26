package base

import (
	cfg "github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/chainmgr"
)

const shard_port_gap = 100

func GetShardRestPort() uint {
	return cfg.DefConfig.Restful.HttpRestPort + uint(chainmgr.GetShardID())*shard_port_gap
}

func GetShardRpcPort() uint {
	return cfg.DefConfig.Rpc.HttpJsonPort + uint(chainmgr.GetShardID())*shard_port_gap
}
