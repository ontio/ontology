/*
 * Copyright (C) 2018 The ontology Authors
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

package event

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
)

const (
	CONTRACT_STATE_FAIL    byte = 0
	CONTRACT_STATE_SUCCESS byte = 1
)

// NotifyEventInfo describe smart contract event notify info struct
type NotifyEventInfo struct {
	ContractAddress common.Address
	States          interface{}

	IsEvm bool
}

type ExecuteNotify struct {
	TxHash      common.Uint256
	State       byte
	GasConsumed uint64
	Notify      []*NotifyEventInfo

	GasStepUsed     uint64
	TxIndex         uint32
	CreatedContract common.Address
}

func ExecuteNotifyFromEthReceipt(receipt *types.Receipt) *ExecuteNotify {
	notify := &ExecuteNotify{
		TxHash:          common.Uint256(receipt.TxHash),
		State:           byte(receipt.Status),
		GasConsumed:     receipt.GasUsed * receipt.GasPrice,
		GasStepUsed:     receipt.GasUsed,
		TxIndex:         receipt.TxIndex,
		CreatedContract: common.Address(receipt.ContractAddress),
	}

	for _, log := range receipt.Logs {
		notify.Notify = append(notify.Notify, NotifyEventInfoFromEvmLog(log))
	}

	return notify
}

func NotifyEventInfoFromEvmLog(log *types.StorageLog) *NotifyEventInfo {
	raw := common.SerializeToBytes(log)

	return &NotifyEventInfo{
		ContractAddress: common.Address(log.Address),
		States:          hexutil.Bytes(raw),
		IsEvm:           true,
	}
}
