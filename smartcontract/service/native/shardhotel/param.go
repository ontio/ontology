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
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
)

type ShardHotelInitParam struct {
	Count int `json:"count"`
}

func (this *ShardHotelInitParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardHotelInitParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type ShardHotelReserveParam struct {
	User   common.Address `json:"user"`
	RoomNo int            `json:"room_no"`
}

func (this *ShardHotelReserveParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardHotelReserveParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type ShardHotelCheckoutParam struct {
	User   common.Address `json:"user"`
	RoomNo int            `json:"room_no"`
}

func (this *ShardHotelCheckoutParam) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardHotelCheckoutParam) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

type ShardHotelReserve2Param struct {
	User             common.Address `json:"user"`
	RoomNo1          int            `json:"room_no"`
	Shard2           types.ShardID  `json:"shard_2"`
	ContractAddress2 common.Address `json:"contract_address_2"`
	RoomNo2          int            `json:"room_no_2"`
}

func (this *ShardHotelReserve2Param) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ShardHotelReserve2Param) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}
