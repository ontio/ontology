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

package shardhotel

import (
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type ShardHotelInitParam struct {
	Count uint64
}

func (this *ShardHotelInitParam) Serialize(w io.Writer) error {
	return utils.WriteVarUint(w, this.Count)
}

func (this *ShardHotelInitParam) Deserialize(r io.Reader) error {
	count, err := utils.ReadVarUint(r)
	if err != nil {
		return err
	}
	this.Count = count
	return nil
}

type ShardHotelReserveParam struct {
	User   common.Address
	RoomNo uint64
}

func (this *ShardHotelReserveParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write user failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.RoomNo); err != nil {
		return fmt.Errorf("serialize: write roomNo failed, err: %s", err)
	}
	return nil
}

func (this *ShardHotelReserveParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user failed, err: %s", err)
	}
	if this.RoomNo, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read roomNo failed, err: %s", err)
	}
	return nil
}

type ShardHotelCheckoutParam struct {
	User   common.Address
	RoomNo uint64
}

func (this *ShardHotelCheckoutParam) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write user failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.RoomNo); err != nil {
		return fmt.Errorf("serialize: write roomNo failed, err: %s", err)
	}
	return nil
}

func (this *ShardHotelCheckoutParam) Deserialize(r io.Reader) error {
	var err error = nil
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user failed, err: %s", err)
	}
	if this.RoomNo, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read roomNo failed, err: %s", err)
	}
	return nil
}

type ShardHotelReserve2Param struct {
	User             common.Address
	RoomNo1          uint64
	Shard2           common.ShardID
	ContractAddress2 common.Address
	RoomNo2          uint64
	Transactional    bool
}

func (this *ShardHotelReserve2Param) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.User); err != nil {
		return fmt.Errorf("serialize: write user failed, err: %s", err)
	}
	if err := utils.WriteVarUint(w, this.RoomNo1); err != nil {
		return fmt.Errorf("serialize: write roomNo1 failed, err: %s", err)
	}
	if err := utils.SerializeShardId(w, this.Shard2); err != nil {
		return fmt.Errorf("serialize: write shard2 failed, err: %s", err)
	}
	if err := utils.WriteAddress(w, this.ContractAddress2); err != nil {
		return fmt.Errorf("serialize: write contract addr2 failed, err: %s", err)
	}
	if err := serialization.WriteBool(w, this.Transactional); err != nil {
		return fmt.Errorf("serialize: write transactional failed, err: %s", err)
	}
	return nil
}

func (this *ShardHotelReserve2Param) Deserialize(r io.Reader) error {
	var err error = nil
	if this.User, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read user failed, err: %s", err)
	}
	if this.RoomNo1, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("deserialize: read roomNo1 failed, err: %s", err)
	}
	if this.Shard2, err = utils.DeserializeShardId(r); err != nil {
		return fmt.Errorf("deserialize: read shard2 failed, err: %s", err)
	}
	if this.ContractAddress2, err = utils.ReadAddress(r); err != nil {
		return fmt.Errorf("deserialize: read contract addr2 failed, err: %s", err)
	}
	if this.Transactional, err = serialization.ReadBool(r); err != nil {
		return fmt.Errorf("deserialize: read transactional failed, err: %s", err)
	}
	return nil
}
