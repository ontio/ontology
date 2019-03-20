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
	"github.com/ontio/ontology/consensus/vbft/config"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

const (
	EVENT_SHARD_GAS_DEPOSIT = iota + 128
	EVENT_SHARD_GAS_WITHDRAW_REQ
	EVENT_SHARD_GAS_WITHDRAW_DONE
	EVENT_SHARD_COMMIT_DPOS
)

type DepositGasEvent struct {
	ImplSourceTargetShardID
	Height     uint32         `json:"height"`
	User       common.Address `json:"user"`
	Amount     uint64         `json:"amount"`
	WithdrawId uint64         `json:"withdraw_id"`
}

func (evt *DepositGasEvent) GetHeight() uint32 {
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
	Height     uint32         `json:"height"`
	User       common.Address `json:"user"`
	WithdrawId uint64         `json:"withdraw_id"`
	Amount     uint64         `json:"amount"`
}

func (evt *WithdrawGasReqEvent) GetHeight() uint32 {
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
	Height     uint32         `json:"height"`
	User       common.Address `json:"user"`
	WithdrawId uint64         `json:"withdraw_id"`
}

func (evt *WithdrawGasDoneEvent) GetHeight() uint32 {
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

type ShardCommitDposEvent struct {
	ImplSourceTargetShardID
	Height    uint32               `json:"height"`
	FeeAmount uint64               `json:"fee_amount"`
	View      uint64               `json:"view"`
	NewConfig *vconfig.ChainConfig `json:"new_config"`
}

func (evt *ShardCommitDposEvent) GetHeight() uint32 {
	return evt.Height
}

func (evt *ShardCommitDposEvent) GetType() uint32 {
	return EVENT_SHARD_COMMIT_DPOS
}
func (evt *ShardCommitDposEvent) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, evt)
}

func (evt *ShardCommitDposEvent) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, evt)
}
