package oep4

import (
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type TransferEvent struct {
	Asset  common.Address
	From   common.Address
	To     common.Address
	Amount *big.Int
}

func (this *TransferEvent) ToNotify() []interface{} {
	return []interface{}{this.Asset.ToHexString(), this.From.ToBase58(), this.To.ToBase58(), this.Amount.String()}
}

type ApproveEvent struct {
	Asset     common.Address
	Owner     common.Address
	Spender   common.Address
	Allowance *big.Int
}

func (this *ApproveEvent) ToNotify() []interface{} {
	return []interface{}{this.Asset.ToHexString(), this.Owner.ToBase58(), this.Spender.ToBase58(), this.Allowance.String()}
}

type DepositToShardEvent struct {
	*TransferEvent
	DepositId uint64
	ToShard   types.ShardID
}

func (this *DepositToShardEvent) ToNotify() []interface{} {
	transferEvent := this.TransferEvent.ToNotify()
	return append(transferEvent, this.DepositId, this.ToShard.ToUint64())
}

type WithdrawFromShardEvent struct {
	*TransferEvent
	WithdrawId types.ShardID
	FromShard  types.ShardID
}

func (this *WithdrawFromShardEvent) ToNotify() []interface{} {
	transferEvent := this.TransferEvent.ToNotify()
	return append(transferEvent, this.WithdrawId, this.FromShard.ToUint64())
}

func NotifyEvent(native *native.NativeService, notify []interface{}) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: utils.ShardAssetAddress,
			States:          notify,
		})
}
