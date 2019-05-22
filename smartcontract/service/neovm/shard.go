package neovm

import (
	vm "github.com/ontio/ontology/vm/neovm"
)

// ShardGetShardId push shardId to vm stack
func ShardGetShardId(service *NeoVmService, engine *vm.ExecutionEngine) error {
	vm.PushData(engine, service.ShardID.ToUint64())
	return nil
}
