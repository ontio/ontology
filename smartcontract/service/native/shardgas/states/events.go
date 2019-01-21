package shardgas_states

import (
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

const (
	EVENT_SHARD_GAS_DEPOSIT = iota + 100
	EVENT_SHARD_GAS_WITHDRAW
)

type DepositGasEvent struct {
	ShardID uint64         `json:"shard_id"`
	User    common.Address `json:"user"`
	Amount  uint64         `json:"amount"`
}

func (evt *DepositGasEvent) GetType() uint32 {
	return EVENT_SHARD_GAS_DEPOSIT
}

func (evt *DepositGasEvent) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *DepositGasEvent) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}

type WithdrawGasEvent struct {
	ShardID uint64         `json:"shard_id"`
	User    common.Address `json:"user"`
	Amount  uint64         `json:"amount"`
}

func (evt *WithdrawGasEvent) GetType() uint32 {
	return EVENT_SHARD_GAS_WITHDRAW
}

func (evt *WithdrawGasEvent) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *WithdrawGasEvent) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}
