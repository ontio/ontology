package shardsysmsg

import (
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/common/log"
	"bytes"
	"fmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

const (
	INIT_NAME               = "init"
	PROCESS_CROSS_SHARD_MSG = "processShardMsg"
)

func InitShardSystemMessageContract() {
	native.Contracts[utils.ShardSysMsgContractAddress] = RegisterShardSysMsgContract
}

func RegisterShardSysMsgContract(native *native.NativeService) {
	native.Register(INIT_NAME, ShardSysMsgInit)
	native.Register(PROCESS_CROSS_SHARD_MSG, ProcessCrossShardMsg)
}

func ShardSysMsgInit(native *native.NativeService) ([]byte, error) {
	return utils.BYTE_FALSE, nil
}

func ProcessCrossShardMsg(native *native.NativeService) ([]byte, error) {

	param := new(CrossShardMsgParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("cross-shard msg, invalid input: %s", err)
	}

	for _, evt := range param.Events {
		log.Infof("processing cross shard msg %d(%d)", evt.EventType, evt.FromHeight)
		if evt.Version != shardmgmt.VERSION_CONTRACT_SHARD_MGMT {
			continue
		}

		shardEvt, err := shardstates.DecodeShardEvent(evt.EventType, evt.Payload)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("processing shard event %d", evt.EventType)
		}

		switch evt.EventType {
		case shardstates.EVENT_SHARD_GAS_DEPOSIT:
			if err := processShardGasDeposit(shardEvt.(*shardstates.DepositGasEvent)); err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("process gas deposit: %s", err)
			}
		case shardstates.EVENT_SHARD_GAS_WITHDRAW_REQ:
		}
	}

	return utils.BYTE_TRUE, nil
}

func processShardGasDeposit(evt *shardstates.DepositGasEvent) error {
	return nil
}
