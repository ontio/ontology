package shardgas_states

import (
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

const (
	EVENT_SHARD_GAS_DEPOSIT = iota + 100
	EVENT_SHARD_GAS_WITHDRAW_REQ
	EVENT_SHARD_GAS_WITHDRAW_DONE
)

const (
	CAP_PENDING_WITHDRAW        = 10
	WITHDRAW_GAS_DELAY_DURATION = 50000
)

type GasWithdrawInfo struct {
	Height uint64 `json:"height"`
	Amount uint64 `json:"amount"`
}

type UserGasInfo struct {
	Balance         uint64             `json:"balance"`
	WithdrawBalance uint64             `json:"withdraw_balance"`
	PendingWithdraw []*GasWithdrawInfo `json:"pending_withdraw"`
}

func (this *UserGasInfo) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *UserGasInfo) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type DepositGasEvent struct {
	SourceShardID uint64         `json:"source_shard_id"`
	Height        uint64         `json:"height"`
	ShardID       uint64         `json:"shard_id"`
	User          common.Address `json:"user"`
	Amount        uint64         `json:"amount"`
}

func (evt *DepositGasEvent) GetSourceShardID() uint64 {
	return evt.SourceShardID
}

func (evt *DepositGasEvent) GetTargetShardID() uint64 {
	return evt.ShardID
}

func (evt *DepositGasEvent) GetHeight() uint64 {
	return evt.Height
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

type WithdrawGasReqEvent struct {
	SourceShardID uint64         `json:"source_shard_id"`
	Height        uint64         `json:"height"`
	ShardID       uint64         `json:"shard_id"`
	User          common.Address `json:"user"`
	Amount        uint64         `json:"amount"`
}

func (evt *WithdrawGasReqEvent) GetSourceShardID() uint64 {
	return evt.SourceShardID
}

func (evt *WithdrawGasReqEvent) GetTargetShardID() uint64 {
	return evt.ShardID
}

func (evt *WithdrawGasReqEvent) GetHeight() uint64 {
	return evt.Height
}

func (evt *WithdrawGasReqEvent) GetType() uint32 {
	return EVENT_SHARD_GAS_WITHDRAW_REQ
}

func (evt *WithdrawGasReqEvent) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *WithdrawGasReqEvent) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}

type WithdrawGasDoneEvent struct {
	SourceShardID uint64         `json:"source_shard_id"`
	Height        uint64         `json:"height"`
	ShardID     uint64         `json:"shard_id"`
	User        common.Address `json:"user"`
	Amount      uint64         `json:"amount"`
}

func (evt *WithdrawGasDoneEvent) GetSourceShardID() uint64 {
	return evt.SourceShardID
}

func (evt *WithdrawGasDoneEvent) GetTargetShardID() uint64 {
	return evt.ShardID
}

func (evt *WithdrawGasDoneEvent) GetHeight() uint64 {
	return evt.Height
}

func (evt *WithdrawGasDoneEvent) GetType() uint32 {
	return EVENT_SHARD_GAS_WITHDRAW_DONE
}

func (evt *WithdrawGasDoneEvent) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *WithdrawGasDoneEvent) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}
