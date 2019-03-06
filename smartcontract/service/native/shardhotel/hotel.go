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
	"bytes"
	"errors"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	// function name
	SHARD_HOTEL_INIT_NAME      = "shardHotelInit"
	SHARD_RESERVE_ROOM_NAME    = "shardReserveRoom"
	SHARD_CHECKOUT_NAME        = "shardCheckout"
	SHARD_DOUBLE_RESERVE_NAME  = "shardDoubleReserve"
	SHARD_DOUBLE_CHECKOUT_NAME = "shardDoubleCheckout"

	// key prefix
	KEY_ROOM = "room"
)

var ErrNotFound = errors.New("Room Not Found")

func InitShardHotel() {
	native.Contracts[utils.ShardHotelAddress] = RegisterShardHotelContract
}

func RegisterShardHotelContract(native *native.NativeService) {
	native.Register(SHARD_HOTEL_INIT_NAME, ShardHotelInit)
	native.Register(SHARD_RESERVE_ROOM_NAME, ShardReserveRoom)
	native.Register(SHARD_CHECKOUT_NAME, ShardCheckout)
	native.Register(SHARD_DOUBLE_RESERVE_NAME, ShardDoubleReserve)
	native.Register(SHARD_DOUBLE_CHECKOUT_NAME, ShardDoubleCheckout)
}

func ShardHotelInit(ctx *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("init shard hotel, invalid cmd param: %s", err)
	}

	param := new(ShardHotelInitParam)
	if err := param.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("init shard hotel, invalid param: %s, %s", err, string(ctx.Input))
	}

	// check if admin
	adminAddress, err := global_params.GetStorageRole(ctx,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	if err := utils.ValidateOwner(ctx, adminAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("init shard hotel, checkWitness error: %v", err)
	}

	for i := 0; i < param.Count; i++ {
		if err := setRoomOwner(ctx, i, common.ADDRESS_EMPTY); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("init shard hotel, init room %d failed: %s", i, err)
		}
	}

	return utils.BYTE_TRUE, nil
}

func ShardReserveRoom(ctx *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserve, invalid cmd param: %s", err)
	}

	param := new(ShardHotelReserveParam)
	if err := param.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserve, invalid param: %s", err)
	}

	if err := utils.ValidateOwner(ctx, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserve, invalid owner: %s", err)
	}

	if user, err := getRoomUser(ctx, param.RoomNo); err != nil && err != ErrNotFound {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserve: %s", err)
	} else if user != common.ADDRESS_EMPTY {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserve: room %d reserved", param.RoomNo)
	}

	if err := setRoomOwner(ctx, param.RoomNo, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserve: reserve room %d: %s", param.RoomNo, err)
	}

	return utils.BYTE_TRUE, nil
}

func ShardCheckout(ctx *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel checkout, invalid cmd param: %s", err)
	}

	param := new(ShardHotelCheckoutParam)
	if err := param.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel checkout, invalid param: %s", err)
	}

	if err := utils.ValidateOwner(ctx, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel checkout, invalid owner: %s", err)
	}

	if user, err := getRoomUser(ctx, param.RoomNo); err != nil && err != ErrNotFound {
		return utils.BYTE_FALSE, fmt.Errorf("hotel checkout: %s", err)
	} else if user != param.User {
		return utils.BYTE_FALSE, fmt.Errorf("hotel checkout: invalid user")
	}

	if err := setRoomOwner(ctx, param.RoomNo, common.ADDRESS_EMPTY); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel checkout: room %d: %s", param.RoomNo, err)
	}

	return utils.BYTE_TRUE, nil
}

func ShardDoubleReserve(ctx *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserver1, invalid cmd param: %s", err)
	}

	param := new(ShardHotelReserve2Param)
	if err := param.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserver1, invalid param: %s", err)
	}

	if err := utils.ValidateOwner(ctx, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserver1, invalid owner: %s", err)
	}

	if user, err := getRoomUser(ctx, param.RoomNo1); err != nil && err != ErrNotFound {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserver1: %s", err)
	} else if user != param.User {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserver1: invalid user")
	}

	if err := setRoomOwner(ctx, param.RoomNo1, common.ADDRESS_EMPTY); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserver1: room %d: %s", param.RoomNo1, err)
	}

	if err := appcallReserveRoom(ctx, param.Shard2, param.User, param.ContractAddress2, param.RoomNo2); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserver1: to remote shard: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func ShardDoubleCheckout(ctx *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(ctx.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel checkout1, invalid cmd param: %s", err)
	}

	param := new(ShardHotelReserve2Param)
	if err := param.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel checkout1, invalid param: %s", err)
	}

	if err := utils.ValidateOwner(ctx, param.User); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel checkout1, invalid owner: %s", err)
	}

	if user, err := getRoomUser(ctx, param.RoomNo1); err != nil && err != ErrNotFound {
		return utils.BYTE_FALSE, fmt.Errorf("hotel checkout1: %s", err)
	} else if user != param.User {
		return utils.BYTE_FALSE, fmt.Errorf("hotel checkout1: invalid user")
	}

	if err := setRoomOwner(ctx, param.RoomNo1, common.ADDRESS_EMPTY); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel checkout1: room %d: %s", param.RoomNo1, err)
	}

	if err := appcallCheckoutRoom(ctx, param.Shard2, param.User, param.ContractAddress2, param.RoomNo2); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel checkout1: to remote shard: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func setRoomOwner(ctx *native.NativeService, roomNo int, owner common.Address) error {
	roomBytes := shardutil.GetUint32Bytes(uint32(roomNo))
	contract := ctx.ContextRef.CurrentContext().ContractAddress
	ctx.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_ROOM), roomBytes), states.GenRawStorageItem(owner[:]))
	return nil
}

func getRoomUser(ctx *native.NativeService, roomNo int) (common.Address, error) {
	roomBytes := shardutil.GetUint32Bytes(uint32(roomNo))
	contract := ctx.ContextRef.CurrentContext().ContractAddress

	userBytes, err := ctx.CacheDB.Get(utils.ConcatKey(contract, []byte(KEY_ROOM), roomBytes))
	if err != nil {
		return common.ADDRESS_EMPTY, fmt.Errorf("get from db: %s", err)
	}
	if userBytes == nil {
		return common.ADDRESS_EMPTY, ErrNotFound
	}

	user, err := states.GetValueFromRawStorageItem(userBytes)
	if err != nil {
		return common.ADDRESS_EMPTY, fmt.Errorf("get from storage item: %s, %x", err, userBytes)
	}

	var userAddr common.Address
	copy(userAddr[:], user)
	return userAddr, ErrNotFound
}

func appcallReserveRoom(ctx *native.NativeService, toShard types.ShardID, user, contract common.Address, roomNo int) error {
	buf := new(bytes.Buffer)
	param := &ShardHotelReserveParam{
		User:   user,
		RoomNo: roomNo,
	}
	if err := param.Serialize(buf); err != nil {
		return err
	}

	buf2 := new(bytes.Buffer)
	cp := &shardmgmt.CommonParam{buf.Bytes()}
	if err := cp.Serialize(buf2); err != nil {
		return err
	}

	return appcallSendReq(ctx, toShard, buf2.Bytes())
}

func appcallCheckoutRoom(ctx *native.NativeService, toShard types.ShardID, user, contract common.Address, roomNo int) error {
	buf := new(bytes.Buffer)
	param := &ShardHotelCheckoutParam{
		User:   user,
		RoomNo: roomNo,
	}
	if err := param.Serialize(buf); err != nil {
		return err
	}

	buf2 := new(bytes.Buffer)
	cp := &shardmgmt.CommonParam{buf.Bytes()}
	if err := cp.Serialize(buf2); err != nil {
		return err
	}

	return appcallSendReq(ctx, toShard, buf2.Bytes())
}

func appcallSendReq(native *native.NativeService, toShard types.ShardID, payload []byte) error {
	paramBytes := new(bytes.Buffer)
	params := shardsysmsg.NotifyReqParam{
		ToShard: toShard,
		Args:    payload,
	}
	if err := params.Serialize(paramBytes); err != nil {
		return fmt.Errorf("hotel remote req, marshal param: %s", err)
	}

	cmnBytes := new(bytes.Buffer)
	cmnParam := shardmgmt.CommonParam{paramBytes.Bytes()}
	if err := cmnParam.Serialize(cmnBytes); err != nil {
		return fmt.Errorf("hotel remote req, marshal cmn param: %s", err)
	}

	if _, err := native.NativeCall(utils.ShardSysMsgContractAddress, shardsysmsg.REMOTE_NOTIFY, cmnBytes.Bytes()); err != nil {
		return fmt.Errorf("hotel remote req, appcallSendReq: %s", err)
	}
	return nil
}
