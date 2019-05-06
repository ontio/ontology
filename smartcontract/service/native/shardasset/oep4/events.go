package oep4

import (
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type TransferEvent struct {
	AssetId AssetId
	From    common.Address
	To      common.Address
	Amount  *big.Int
}

func (this *TransferEvent) ToNotify() []interface{} {
	return []interface{}{uint64(this.AssetId), this.From.ToBase58(), this.To.ToBase58(), this.Amount.String()}
}

type ApproveEvent struct {
	AssetId   AssetId
	Owner     common.Address
	Spender   common.Address
	Allowance *big.Int
}

func (this *ApproveEvent) ToNotify() []interface{} {
	return []interface{}{uint64(this.AssetId), this.Owner.ToBase58(), this.Spender.ToBase58(), this.Allowance.String()}
}

type XShardTransferEvent struct {
	*TransferEvent
	TransferId *big.Int
	ToShard    common.ShardID
}

func (this *XShardTransferEvent) ToNotify() []interface{} {
	transferEvent := this.TransferEvent.ToNotify()
	return append(transferEvent, this.TransferId.String(), this.ToShard.ToUint64())
}

type XShardReceiveEvent struct {
	*TransferEvent
	TransferId *big.Int
	FromShard  common.ShardID
}

func (this *XShardReceiveEvent) ToNotify() []interface{} {
	transferEvent := this.TransferEvent.ToNotify()
	return append(transferEvent, this.TransferId.String(), this.FromShard.ToUint64())
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
