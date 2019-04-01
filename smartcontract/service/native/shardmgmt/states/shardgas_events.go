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
	"fmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"io"

	"github.com/ontio/ontology/common"
)

const (
	EVENT_SHARD_GAS_DEPOSIT = iota + 128
	EVENT_SHARD_GAS_WITHDRAW_REQ
	EVENT_SHARD_GAS_WITHDRAW_DONE
	EVENT_SHARD_COMMIT_DPOS
)

type DepositGasEvent struct {
	*ImplSourceTargetShardID
	Height uint32
	User   common.Address
	Amount uint64
}

func (evt *DepositGasEvent) GetHeight() uint32 {
	return evt.Height
}

func (evt *DepositGasEvent) GetType() uint32 {
	return EVENT_SHARD_GAS_DEPOSIT
}

func (evt *DepositGasEvent) Serialize(w io.Writer) error {
	if err := evt.ImplSourceTargetShardID.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write impl failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(evt.Height)); err != nil {
		return fmt.Errorf("serialize: write height failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, evt.User); err != nil {
		return fmt.Errorf("serialize: write user failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, evt.Amount); err != nil {
		return fmt.Errorf("serialize: write amount failed, err: %s", err)
	}
	return nil
}

func (evt *DepositGasEvent) Deserialize(r io.Reader) error {
	var err error = nil
	evt.ImplSourceTargetShardID = &ImplSourceTargetShardID{}
	if err = evt.ImplSourceTargetShardID.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read impl failed, err: %s", err)
	}
	height, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read height failed, err: %s", err)
	}
	evt.Height = uint32(height)
	if evt.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user failed, err: %s", err)
	}
	if evt.Amount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read amount failed, err: %s", err)
	}
	return nil
}

func (this *DepositGasEvent) Serialization(sink *common.ZeroCopySink) {
	this.ImplSourceTargetShardID.Serialization(sink)
	sink.WriteUint32(this.Height)
	sink.WriteAddress(this.User)
	sink.WriteUint64(this.Amount)
}

func (this *DepositGasEvent) Deserialization(source *common.ZeroCopySource) error {
	this.ImplSourceTargetShardID = &ImplSourceTargetShardID{}
	if err := this.ImplSourceTargetShardID.Deserialization(source); err != nil {
		return fmt.Errorf("read impl err: %s", err)
	}
	var eof bool
	this.Height, eof = source.NextUint32()
	this.User, eof = source.NextAddress()
	this.Amount, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

type WithdrawGasReqEvent struct {
	*ImplSourceTargetShardID
	Height     uint32
	User       common.Address
	WithdrawId uint64
	Amount     uint64
}

func (evt *WithdrawGasReqEvent) GetHeight() uint32 {
	return evt.Height
}

func (evt *WithdrawGasReqEvent) GetType() uint32 {
	return EVENT_SHARD_GAS_WITHDRAW_REQ
}

func (evt *WithdrawGasReqEvent) Serialize(w io.Writer) error {
	if err := evt.ImplSourceTargetShardID.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write impl failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(evt.Height)); err != nil {
		return fmt.Errorf("serialize: write height failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, evt.User); err != nil {
		return fmt.Errorf("serialize: write user failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, evt.Amount); err != nil {
		return fmt.Errorf("serialize: write amount failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, evt.WithdrawId); err != nil {
		return fmt.Errorf("serialize: write withdraw id failed, err: %s", err)
	}
	return nil
}

func (evt *WithdrawGasReqEvent) Deserialize(r io.Reader) error {
	var err error = nil
	evt.ImplSourceTargetShardID = &ImplSourceTargetShardID{}
	if err = evt.ImplSourceTargetShardID.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read impl failed, err: %s", err)
	}
	height, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read height failed, err: %s", err)
	}
	evt.Height = uint32(height)
	if evt.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user failed, err: %s", err)
	}
	if evt.Amount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read amount failed, err: %s", err)
	}
	if evt.WithdrawId, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read withdraw id failed, err: %s", err)
	}
	return nil
}

func (this *WithdrawGasReqEvent) Serialization(sink *common.ZeroCopySink) {
	this.ImplSourceTargetShardID.Serialization(sink)
	sink.WriteUint32(this.Height)
	sink.WriteAddress(this.User)
	sink.WriteUint64(this.WithdrawId)
	sink.WriteUint64(this.Amount)
}

func (this *WithdrawGasReqEvent) Deserialization(source *common.ZeroCopySource) error {
	this.ImplSourceTargetShardID = &ImplSourceTargetShardID{}
	if err := this.ImplSourceTargetShardID.Deserialization(source); err != nil {
		return fmt.Errorf("read impl err: %s", err)
	}
	var eof bool
	this.Height, eof = source.NextUint32()
	this.User, eof = source.NextAddress()
	this.WithdrawId, eof = source.NextUint64()
	this.Amount, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

type WithdrawGasDoneEvent struct {
	*ImplSourceTargetShardID
	Height     uint32
	User       common.Address
	WithdrawId uint64
}

func (evt *WithdrawGasDoneEvent) GetHeight() uint32 {
	return evt.Height
}

func (evt *WithdrawGasDoneEvent) GetType() uint32 {
	return EVENT_SHARD_GAS_WITHDRAW_DONE
}

func (evt *WithdrawGasDoneEvent) Serialize(w io.Writer) error {
	if err := evt.ImplSourceTargetShardID.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write impl failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(evt.Height)); err != nil {
		return fmt.Errorf("serialize: write height failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, evt.User); err != nil {
		return fmt.Errorf("serialize: write user failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, evt.WithdrawId); err != nil {
		return fmt.Errorf("serialize: write withdraw id failed, err: %s", err)
	}
	return nil
}

func (evt *WithdrawGasDoneEvent) Deserialize(r io.Reader) error {
	var err error = nil
	evt.ImplSourceTargetShardID = &ImplSourceTargetShardID{}
	if err = evt.ImplSourceTargetShardID.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read impl failed, err: %s", err)
	}
	height, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read height failed, err: %s", err)
	}
	evt.Height = uint32(height)
	if evt.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user failed, err: %s", err)
	}
	if evt.WithdrawId, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read withdraw id failed, err: %s", err)
	}
	return nil
}

func (this *WithdrawGasDoneEvent) Serialization(sink *common.ZeroCopySink) {
	this.ImplSourceTargetShardID.Serialization(sink)
	sink.WriteUint32(this.Height)
	sink.WriteAddress(this.User)
	sink.WriteUint64(this.WithdrawId)
}

func (this *WithdrawGasDoneEvent) Deserialization(source *common.ZeroCopySource) error {
	this.ImplSourceTargetShardID = &ImplSourceTargetShardID{}
	if err := this.ImplSourceTargetShardID.Deserialization(source); err != nil {
		return fmt.Errorf("read impl err: %s", err)
	}
	var eof bool
	this.Height, eof = source.NextUint32()
	this.User, eof = source.NextAddress()
	this.WithdrawId, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

type ShardCommitDposEvent struct {
	*ImplSourceTargetShardID
	Height    uint32
	FeeAmount uint64
}

func (evt *ShardCommitDposEvent) GetHeight() uint32 {
	return evt.Height
}

func (evt *ShardCommitDposEvent) GetType() uint32 {
	return EVENT_SHARD_COMMIT_DPOS
}
func (evt *ShardCommitDposEvent) Serialize(w io.Writer) error {
	if err := evt.ImplSourceTargetShardID.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write impl failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, uint64(evt.Height)); err != nil {
		return fmt.Errorf("serialize: write height failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, evt.FeeAmount); err != nil {
		return fmt.Errorf("serialize: write fee amount failed, err: %s", err)
	}
	return nil
}

func (evt *ShardCommitDposEvent) Deserialize(r io.Reader) error {
	var err error = nil
	evt.ImplSourceTargetShardID = &ImplSourceTargetShardID{}
	if err = evt.ImplSourceTargetShardID.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read impl failed, err: %s", err)
	}
	height, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read height failed, err: %s", err)
	}
	evt.Height = uint32(height)
	if evt.FeeAmount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read fee amount failed, err: %s", err)
	}
	return nil
}

func (this *ShardCommitDposEvent) Serialization(sink *common.ZeroCopySink) {
	this.ImplSourceTargetShardID.Serialization(sink)
	sink.WriteUint32(this.Height)
	sink.WriteUint64(this.FeeAmount)
}

func (this *ShardCommitDposEvent) Deserialization(source *common.ZeroCopySource) error {
	this.ImplSourceTargetShardID = &ImplSourceTargetShardID{}
	if err := this.ImplSourceTargetShardID.Deserialization(source); err != nil {
		return fmt.Errorf("read impl err: %s", err)
	}
	var eof bool
	this.Height, eof = source.NextUint32()
	this.FeeAmount, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
