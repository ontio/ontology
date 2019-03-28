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

package shardgas

import (
	"fmt"
	"github.com/ontio/ontology/core/types"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type DepositGasParam struct {
	User    common.Address
	ShardId types.ShardID
	Amount  uint64
}

func (this *DepositGasParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write addr failed, err: %s", err)
	}
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.Amount); err != nil {
		return fmt.Errorf("serialize: write amount failed, err: %s", err)
	}
	return nil
}

func (this *DepositGasParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user failed, err: %s", err)
	}
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.Amount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read amount failed, err: %s", err)
	}
	return nil
}

type UserWithdrawGasParam struct {
	User   common.Address
	Amount uint64
}

func (this *UserWithdrawGasParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write addr failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.Amount); err != nil {
		return fmt.Errorf("serialize: write amount failed, err: %s", err)
	}
	return nil
}

func (this *UserWithdrawGasParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user failed, err: %s", err)
	}
	if this.Amount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read amount failed, err: %s", err)
	}
	return nil
}

type UserRetryWithdrawParam struct {
	User       common.Address
	WithdrawId uint64
}

func (this *UserRetryWithdrawParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write addr failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.WithdrawId); err != nil {
		return fmt.Errorf("serialize: write withdraw id failed, err: %s", err)
	}
	return nil
}

func (this *UserRetryWithdrawParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user failed, err: %s", err)
	}
	if this.WithdrawId, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read withdraw id failed, err: %s", err)
	}
	return nil
}

type UserWithdrawSuccessParam struct {
	User       common.Address
	WithdrawId uint64
}

func (this *UserWithdrawSuccessParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write addr failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.WithdrawId); err != nil {
		return fmt.Errorf("serialize: write withdraw id failed, err: %s", err)
	}
	return nil
}

func (this *UserWithdrawSuccessParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user failed, err: %s", err)
	}
	if this.WithdrawId, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read withdraw id failed, err: %s", err)
	}
	return nil
}

type PeerWithdrawGasParam struct {
	Signer     common.Address
	PeerPubKey string
	User       common.Address
	ShardId    types.ShardID
	Amount     uint64
	WithdrawId uint64
}

func (this *PeerWithdrawGasParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Signer); err != nil {
		return fmt.Errorf("serialize: write signer failed, err: %s", err)
	}
	if err := serialization.WriteString(w, this.PeerPubKey); err != nil {
		return fmt.Errorf("serialize: write peer pub key failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write user failed, err: %s", err)
	}
	if err := utils.SerializeShardId(w, this.ShardId); err != nil {
		return fmt.Errorf("serialize: write shard id failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.Amount); err != nil {
		return fmt.Errorf("serialize: write amount failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.WithdrawId); err != nil {
		return fmt.Errorf("serialize: write withdraw id failed, err: %s", err)
	}
	return nil
}

func (this *PeerWithdrawGasParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.Signer, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read signer failed, err: %s", err)
	}
	if this.PeerPubKey, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read signer failed, err: %s", err)
	}
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user failed, err: %s", err)
	}
	if this.ShardId, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard id failed, err: %s", err)
	}
	if this.Amount, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read amount failed, err: %s", err)
	}
	if this.WithdrawId, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read withdraw id failed, err: %s", err)
	}
	return nil
}

type CommitDposParam struct {
	Signer     common.Address
	PeerPubKey string
	*shardmgmt.CommitDposParam
}

func (this *CommitDposParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Signer); err != nil {
		return fmt.Errorf("serialize: write signer failed, err: %s", err)
	}
	if err := serialization.WriteString(w, this.PeerPubKey); err != nil {
		return fmt.Errorf("serialize: write peer pub key failed, err: %s", err)
	}
	if err := this.CommitDposParam.Serialize(w); err != nil {
		return fmt.Errorf("serialize: write commit dpos param failed, err: %s", err)
	}
	return nil
}

func (this *CommitDposParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.Signer, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read signer failed, err: %s", err)
	}
	if this.PeerPubKey, err = serialization.ReadString(r); err != nil {
		return fmt.Errorf("deserialize: read peer pub key failed, err: %s", err)
	}
	commitDpos := &shardmgmt.CommitDposParam{}
	if err := commitDpos.Deserialize(r); err != nil {
		return fmt.Errorf("deserialize: read commit dpos param failed, err: %s", err)
	}
	this.CommitDposParam = commitDpos
	return nil
}

type GetWithdrawByIdParam struct {
	User       common.Address
	WithdrawId uint64
}

func (this *GetWithdrawByIdParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write addr failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.WithdrawId); err != nil {
		return fmt.Errorf("serialize: write withdraw id failed, err: %s", err)
	}
	return nil
}

func (this *GetWithdrawByIdParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user failed, err: %s", err)
	}
	if this.WithdrawId, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read withdraw id failed, err: %s", err)
	}
	return nil
}
