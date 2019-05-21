/*
 * Copyright (C) 2019 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */
package oep4

import (
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	EVENT_TRANSFER        = "transfer"
	EVENT_APPROVE         = "approve"
	EVENT_MINT            = "mint"
	EVENT_BURN            = "burn"
	EVENT_XSHARD_TRANSFER = "xshardTransfer"
	EVENT_XSHARD_RECEIVE  = "xshardReceive"
)

type TransferEvent struct {
	AssetId AssetId
	From    common.Address
	To      common.Address
	Amount  *big.Int
}

func (this *TransferEvent) toNotify() []interface{} {
	return []interface{}{uint64(this.AssetId), this.From.ToBase58(), this.To.ToBase58(), this.Amount.String()}
}

func (this *TransferEvent) ToNotify() []interface{} {
	transferEvent := this.toNotify()
	return append([]interface{}{EVENT_TRANSFER}, transferEvent...)
}

type ApproveEvent struct {
	AssetId   AssetId
	Owner     common.Address
	Spender   common.Address
	Allowance *big.Int
}

func (this *ApproveEvent) ToNotify() []interface{} {
	return []interface{}{EVENT_APPROVE, uint64(this.AssetId), this.Owner.ToBase58(), this.Spender.ToBase58(),
		this.Allowance.String()}
}

type XShardTransferEvent struct {
	*TransferEvent
	TransferId *big.Int
	ToShard    common.ShardID
}

func (this *XShardTransferEvent) ToNotify() []interface{} {
	transferEvent := this.TransferEvent.toNotify()
	evts := append(transferEvent, this.TransferId.String(), this.ToShard.ToUint64())
	return append([]interface{}{EVENT_XSHARD_TRANSFER}, evts...)
}

type XShardReceiveEvent struct {
	*TransferEvent
	TransferId *big.Int
	FromShard  common.ShardID
}

func (this *XShardReceiveEvent) ToNotify() []interface{} {
	transferEvent := this.TransferEvent.toNotify()
	evts := append(transferEvent, this.TransferId.String(), this.FromShard.ToUint64())
	return append([]interface{}{EVENT_XSHARD_RECEIVE}, evts...)
}

type MintEvent struct {
	User    common.Address
	AssetId AssetId
	Amount  *big.Int
}

func (this *MintEvent) ToNotify() []interface{} {
	return []interface{}{EVENT_MINT, this.User.ToBase58(), uint64(this.AssetId), this.Amount.String()}
}

type BurnEvent struct {
	User    common.Address
	AssetId AssetId
	Amount  *big.Int
}

func (this *BurnEvent) ToNotify() []interface{} {
	return []interface{}{EVENT_BURN, this.User.ToBase58(), uint64(this.AssetId), this.Amount.String()}
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
