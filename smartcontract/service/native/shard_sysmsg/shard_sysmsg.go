package shardsysmsg

import (
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	INIT_NAME                = "init"
	PROCESS_PARENT_SHARD_MSG = "processParentShardMsg"
	PROCESS_SIB_SHARD_MSG    = "processSibShardMsg"
)

func InitShardSystemMessageContract() {
	native.Contracts[utils.ShardSysMsgContractAddress] = RegisterShardSysMsgContract
}

func RegisterShardSysMsgContract(native *native.NativeService) {
	native.Register(INIT_NAME, ShardSysMsgInit)
	native.Register(PROCESS_PARENT_SHARD_MSG, ProcessParentShardMsg)
	native.Register(PROCESS_SIB_SHARD_MSG, ProcessSibShardMsg)
}

func ShardSysMsgInit(native *native.NativeService) ([]byte, error) {
	return utils.BYTE_FALSE, nil
}

func ProcessParentShardMsg(native *native.NativeService) ([]byte, error) {
	return utils.BYTE_FALSE, nil
}

func ProcessSibShardMsg(native *native.NativeService) ([]byte, error) {
	return utils.BYTE_FALSE, nil
}
