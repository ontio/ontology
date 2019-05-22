package neovm

import (
	"fmt"
	"github.com/ontio/ontology/common"
	vm "github.com/ontio/ontology/vm/neovm"
)

// GetExecutingAddress push transaction's hash to vm stack
func ShardGetShardId(service *NeoVmService, engine *vm.ExecutionEngine) error {
	shardId, _ := vm.PopInteropInterface(engine)
	id, ok := shardId.(*common.ShardID)
	if !ok {
		return fmt.Errorf("get shardId failed")
	}
	vm.PushData(engine, id)
	return nil
}
