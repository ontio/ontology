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
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	// function name
	SHARD_HOTEL_INIT_NAME      = "shardHotelInit"
	SHARD_RESERVE_NAME         = "shardReserveRoom"
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
	native.Register(SHARD_RESERVE_NAME, ShardReserveRoom)
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

	for i := uint64(0); i < param.Count; i++ {
		setRoomOwner(ctx, i, common.ADDRESS_EMPTY)
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

	if user, err := getRoomUser(ctx, param.RoomNo); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserve: %s", err)
	} else if user != common.ADDRESS_EMPTY {
		if user == param.User {
			return utils.BYTE_TRUE, nil
		}
		return utils.BYTE_FALSE, fmt.Errorf("room reserved by %s", hex.EncodeToString(user[:]))
	}

	setRoomOwner(ctx, param.RoomNo, param.User)
	log.Errorf("user %s reserved room %d OK", param.User.ToBase58(), param.RoomNo)

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

	setRoomOwner(ctx, param.RoomNo, common.ADDRESS_EMPTY)

	log.Errorf("user %v checkout room %d OK", hex.EncodeToString(param.User[:]), param.RoomNo)

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

	log.Infof(">>>> double reserve %s", string(cp.Input))

	if user, err := getRoomUser(ctx, param.RoomNo1); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserver1: %s", err)
	} else if user != common.ADDRESS_EMPTY {
		if user == param.User {
			return utils.BYTE_TRUE, nil
		}
		return utils.BYTE_FALSE, fmt.Errorf("hotel reserver1: room reserved by %s", hex.EncodeToString(user[:]))
	}

	setRoomOwner(ctx, param.RoomNo1, param.User)

	if err := appcallReserveRoom(ctx, param.User, param.Shard2, param.RoomNo2, param.Transactional); err != nil {
		log.Errorf(">>>> hotel contract reserve remote room: %s", err)
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

	setRoomOwner(ctx, param.RoomNo1, common.ADDRESS_EMPTY)

	if err := appcallCheckoutRoom(ctx, param.User, param.Shard2, param.RoomNo2, param.Transactional); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("hotel checkout1: to remote shard: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func setRoomOwner(ctx *native.NativeService, roomNo uint64, owner common.Address) {
	roomBytes := utils.GetUint64Bytes(roomNo)
	contract := ctx.ContextRef.CurrentContext().ContractAddress
	ctx.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_ROOM), roomBytes), states.GenRawStorageItem(owner[:]))
}

func getRoomUser(ctx *native.NativeService, roomNo uint64) (common.Address, error) {
	roomBytes := utils.GetUint64Bytes(roomNo)
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
	return userAddr, nil
}

func appcallReserveRoom(ctx *native.NativeService, user common.Address, toShard types.ShardID, roomNo uint64,
	transactional bool) error {
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

	log.Infof(">>>> to reserve room: shard %d, transactional: %v, req: %s", toShard, transactional,
		string(buf2.Bytes()))

	return appcallSendReq(ctx, toShard, SHARD_RESERVE_NAME, buf2.Bytes(), transactional)
}

func appcallCheckoutRoom(ctx *native.NativeService, user common.Address, toShard types.ShardID, roomNo uint64,
	transactional bool) error {
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

	return appcallSendReq(ctx, toShard, SHARD_CHECKOUT_NAME, buf2.Bytes(), transactional)
}

func appcallSendReq(native *native.NativeService, toShard types.ShardID, method string, payload []byte,
	transactional bool) error {
	paramBytes := new(bytes.Buffer)
	params := shardsysmsg.NotifyReqParam{
		ToShard:    toShard,
		ToContract: utils.ShardHotelAddress,
		Method:     method,
		Args:       payload,
	}
	if err := params.Serialize(paramBytes); err != nil {
		return fmt.Errorf("hotel remote req, marshal param: %s", err)
	}

	cmnBytes := new(bytes.Buffer)
	cmnParam := shardmgmt.CommonParam{paramBytes.Bytes()}
	if err := cmnParam.Serialize(cmnBytes); err != nil {
		return fmt.Errorf("hotel remote req, marshal cmn param: %s", err)
	}

	if transactional {
		if _, err := native.NativeCall(utils.ShardSysMsgContractAddress, shardsysmsg.REMOTE_INVOKE, cmnBytes.Bytes());
			err != nil {
			return fmt.Errorf("hotel remote req, appcallSend Tx: %s", err)
		}
	} else {
		if _, err := native.NativeCall(utils.ShardSysMsgContractAddress, shardsysmsg.REMOTE_NOTIFY, cmnBytes.Bytes());
			err != nil {
			return fmt.Errorf("hotel remote req, appcallSend Notify: %s", err)
		}
	}
	return nil
}
