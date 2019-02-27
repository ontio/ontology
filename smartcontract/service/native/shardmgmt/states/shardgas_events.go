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

package shardstates

import (
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

const (
	EVENT_SHARD_GAS_DEPOSIT = iota + 128
	EVENT_SHARD_GAS_WITHDRAW_REQ
	EVENT_SHARD_GAS_WITHDRAW_DONE
)

const (
	CAP_PENDING_WITHDRAW   = 10
	DEFAULE_WITHDRAW_DELAY = 50000
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
	ImplSourceTargetShardID
	Height uint64         `json:"height"`
	User   common.Address `json:"user"`
	Amount uint64         `json:"amount"`
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
	ImplSourceTargetShardID
	Height uint64         `json:"height"`
	User   common.Address `json:"user"`
	Amount uint64         `json:"amount"`
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
	ImplSourceTargetShardID
	Height uint64         `json:"height"`
	User   common.Address `json:"user"`
	Amount uint64         `json:"amount"`
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
