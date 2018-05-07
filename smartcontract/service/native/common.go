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

package native

import (
	"fmt"
	"math/big"

	"bytes"
	"github.com/ontio/ontology/common"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native/states"
)

var (
	ADDRESS_HEIGHT    = []byte("addressHeight")
	TRANSFER_NAME     = "transfer"
	TOTAL_SUPPLY_NAME = []byte("totalSupply")
	SET_PARAM         = "SetGlobalParam"
)

func getAddressHeightKey(contract, address common.Address) []byte {
	temp := append(ADDRESS_HEIGHT, address[:]...)
	return append(contract[:], temp...)
}

func getHeightStorageItem(height uint32) *cstates.StorageItem {
	return &cstates.StorageItem{Value: big.NewInt(int64(height)).Bytes()}
}

func getAmountStorageItem(value *big.Int) *cstates.StorageItem {
	return &cstates.StorageItem{Value: value.Bytes()}
}

func getToAmountStorageItem(toBalance, value *big.Int) *cstates.StorageItem {
	return &cstates.StorageItem{Value: new(big.Int).Add(toBalance, value).Bytes()}
}

func getParamStorageItem(params *states.Params) *cstates.StorageItem {
	bf := new(bytes.Buffer)
	params.Serialize(bf)
	return &cstates.StorageItem{Value: bf.Bytes()}
}

func getAdminStorageItem(admin *states.Admin) *cstates.StorageItem {
	bf := new(bytes.Buffer)
	admin.Serialize(bf)
	return &cstates.StorageItem{Value: bf.Bytes()}
}

func getTotalSupplyKey(contract common.Address) []byte {
	return append(contract[:], TOTAL_SUPPLY_NAME...)
}

func getTransferKey(contract, from common.Address) []byte {
	return append(contract[:], from[:]...)
}

func getApproveKey(contract common.Address, state *states.State) []byte {
	temp := append(contract[:], state.From[:]...)
	return append(temp, state.To[:]...)
}

func getTransferFromKey(contract common.Address, state *states.TransferFrom) []byte {
	temp := append(contract[:], state.From[:]...)
	return append(temp, state.Sender[:]...)
}

func getParamKey(contract common.Address, valueType paramType) []byte {
	key := append(contract[:], "param"...)
	key = append(key[:], byte(valueType))
	return key
}

func getAdminKey(contract common.Address, isTransferAdmin bool) []byte {
	if isTransferAdmin {
		return append(contract[:], "transfer"...)
	} else {
		return append(contract[:], "admin"...)
	}
}

func isTransferValid(native *NativeService, state *states.State) error {
	if state.Value.Sign() < 0 {
		return errors.NewErr("Transfer amount invalid!")
	}

	if native.ContextRef.CheckWitness(state.From) == false {
		return errors.NewErr("[Sender] Authentication failed!")
	}
	return nil
}

func transfer(native *NativeService, contract common.Address, state *states.State) (*big.Int, *big.Int, error) {
	if err := isTransferValid(native, state); err != nil {
		return nil, nil, err
	}

	fromBalance, err := fromTransfer(native, getTransferKey(contract, state.From), state.Value)
	if err != nil {
		return nil, nil, err
	}

	toBalance, err := toTransfer(native, getTransferKey(contract, state.To), state.Value)
	if err != nil {
		return nil, nil, err
	}
	return fromBalance, toBalance, nil
}

func transferFrom(native *NativeService, currentContract common.Address, state *states.TransferFrom) error {
	if err := isTransferFromValid(native, state); err != nil {
		return err
	}

	if err := fromApprove(native, getTransferFromKey(currentContract, state), state.Value); err != nil {
		return err
	}

	if _, err := fromTransfer(native, getTransferKey(currentContract, state.From), state.Value); err != nil {
		return err
	}

	if _, err := toTransfer(native, getTransferKey(currentContract, state.To), state.Value); err != nil {
		return err
	}
	return nil
}

func isTransferFromValid(native *NativeService, state *states.TransferFrom) error {
	if state.Value.Sign() < 0 {
		return errors.NewErr("TransferFrom amount invalid!")
	}

	if native.ContextRef.CheckWitness(state.Sender) == false {
		return errors.NewErr("[Sender] Authentication failed!")
	}
	return nil
}

func isApproveValid(native *NativeService, state *states.State) error {
	if state.Value.Sign() < 0 {
		return errors.NewErr("Approve amount invalid!")
	}
	if native.ContextRef.CheckWitness(state.From) == false {
		return errors.NewErr("[Sender] Authentication failed!")
	}
	return nil
}

func fromApprove(native *NativeService, fromApproveKey []byte, value *big.Int) error {
	approveValue, err := getStorageBigInt(native, fromApproveKey)
	if err != nil {
		return err
	}
	approveBalance := new(big.Int).Sub(approveValue, value)
	sign := approveBalance.Sign()
	if sign < 0 {
		return fmt.Errorf("[TransferFrom] approve balance insufficient! have %d, got %d", approveValue.Int64(), value.Int64())
	} else if sign == 0 {
		native.CloneCache.Delete(scommon.ST_STORAGE, fromApproveKey)
	} else {
		native.CloneCache.Add(scommon.ST_STORAGE, fromApproveKey, getAmountStorageItem(approveBalance))
	}
	return nil
}

func fromTransfer(native *NativeService, fromKey []byte, value *big.Int) (*big.Int, error) {
	fromBalance, err := getStorageBigInt(native, fromKey)
	if err != nil {
		return nil, err
	}
	balance := new(big.Int).Sub(fromBalance, value)
	sign := balance.Sign()
	if sign < 0 {
		return nil, errors.NewErr("[Transfer] balance insufficient!")
	} else if sign == 0 {
		native.CloneCache.Delete(scommon.ST_STORAGE, fromKey)
	} else {
		native.CloneCache.Add(scommon.ST_STORAGE, fromKey, getAmountStorageItem(balance))
	}
	return fromBalance, nil
}

func toTransfer(native *NativeService, toKey []byte, value *big.Int) (*big.Int, error) {
	toBalance, err := getStorageBigInt(native, toKey)
	if err != nil {
		return nil, err
	}
	native.CloneCache.Add(scommon.ST_STORAGE, toKey, getToAmountStorageItem(toBalance, value))
	return toBalance, nil
}

func getStartHeight(native *NativeService, contract, from common.Address) (uint32, error) {
	startHeight, err := getStorageBigInt(native, getAddressHeightKey(contract, from))
	if err != nil {
		return 0, err
	}
	return uint32(startHeight.Int64()), nil
}

func getStorageBigInt(native *NativeService, key []byte) (*big.Int, error) {
	balance, err := native.CloneCache.Get(scommon.ST_STORAGE, key)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[getBalance] storage error!")
	}
	if balance == nil {
		return big.NewInt(0), nil
	}
	item, ok := balance.(*cstates.StorageItem)
	if !ok {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[getBalance] get amount error!")
	}
	return new(big.Int).SetBytes(item.Value), nil
}

func getStorageParam(native *NativeService, key []byte)(*states.Params, error){
	params , err:= native.CloneCache.Get(scommon.ST_STORAGE, key)
	tempParams := new(states.Params)
	*tempParams = make(map[string]string)
	if err != nil || params == nil{
		return tempParams, errors.NewDetailErr(err, errors.ErrNoCode, "[Get Param] storage error!")
	}
	item, ok := params.(*cstates.StorageItem)
	if !ok {
		return tempParams, errors.NewDetailErr(err, errors.ErrNoCode, "[Get Param] storage error!")
	}
	bf := bytes.NewBuffer(item.Value)
	tempParams.Deserialize(bf)
	return tempParams, nil
}

func getStorageAdmin(native *NativeService, key []byte) (*states.Admin, error) {
	admin, err := native.CloneCache.Get(scommon.ST_STORAGE, key)
	tempAdmin := new(states.Admin)
	if err != nil || admin == nil {
		return tempAdmin, errors.NewDetailErr(err, errors.ErrNoCode, "[Get Admin] storage error!")
	}
	item, ok := admin.(*cstates.StorageItem)
	if !ok {
		return tempAdmin, errors.NewDetailErr(err, errors.ErrNoCode, "[Get Admin] storage error!")
	}
	bf := bytes.NewBuffer(item.Value)
	tempAdmin.Deserialize(bf)
	return tempAdmin, nil
}

func addNotifications(native *NativeService, contract common.Address, state *states.State) {
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			TxHash:          native.Tx.Hash(),
			ContractAddress: contract,
			States:          []interface{}{TRANSFER_NAME, state.From.ToBase58(), state.To.ToBase58(), state.Value},
		})
}

func notifyParamSetSuccess(native *NativeService, contract common.Address, params states.Params) {
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			TxHash:          native.Tx.Hash(),
			ContractAddress: contract,
			States:          []interface{}{SET_PARAM, params},
		})
}
