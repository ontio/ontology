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
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

type CommonParam struct {
	Input []byte
}

func (this *CommonParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.Input); err != nil {
		return fmt.Errorf("CommonParam serialize write failed: %s", err)
	}
	return nil
}

func (this *CommonParam) Deserialize(r io.Reader) error {
	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("CommonParam deserialize read failed: %s", err)
	}
	this.Input = buf
	return nil
}

type DepositGasParam struct {
	UserAddress common.Address `json:"user_address"`
	ShardID     uint64         `json:"shard_id"`
	Amount      uint64         `json:"amount"`
}

func (this *DepositGasParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *DepositGasParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type WithdrawGasRequestParam struct {
	UserAddress common.Address `json:"user_address"`
	ShardID     uint64         `json:"shard_id"`
	Amount      uint64         `json:"amount"`
}

func (this *WithdrawGasRequestParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *WithdrawGasRequestParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type AcquireWithdrawGasParam struct {
	UserAddress common.Address `json:"user_address"`
	ShardID     uint64         `json:"shard_id"`
	Amount      uint64         `json:"amount"`
}

func (this *AcquireWithdrawGasParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *AcquireWithdrawGasParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}
