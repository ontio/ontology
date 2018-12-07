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
package ongx
import (
	"fmt"
	"math/big"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/vm/neovm/types"
)
func InitOngx() {
	native.Contracts[utils.OngContractAddress] = RegisterOngContract
}
func RegisterOngContract(native *native.NativeService) {
	native.Register(TRANSFER_NAME, OngxTransfer)
	native.Register(APPROVE_NAME, OngxApprove)
	native.Register(TRANSFERFROM_NAME, OngxTransferFrom)
	native.Register(NAME_NAME, OngxName)
	native.Register(SYMBOL_NAME, OngxSymbol)
	native.Register(DECIMALS_NAME, OngxDecimals)
	native.Register(TOTALSUPPLY_NAME, OngxTotalSupply)
	native.Register(BALANCEOF_NAME, OngxBalanceOf)
	native.Register(ALLOWANCE_NAME, OngxAllowance)
	native.Register(ONG_SWAP, OngSwap)
	native.Register(ONGX_SWAP, OngxSwap)
	native.Register(SET_SYNC_ADDR_NAME, OngxSetSyncAddr)
}
func OngxTransfer(native *native.NativeService) ([]byte, error) {
	var transfers Transfers
	source := common.NewZeroCopySource(native.Input)
	if err := transfers.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OngTransfer] Transfers deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	for _, v := range transfers.States {
		if v.Value == 0 {
			continue
		}
		if v.Value > constants.ONG_TOTAL_SUPPLY {
			return utils.BYTE_FALSE, fmt.Errorf("transfer ong amount:%d over totalSupply:%d", v.Value, constants.ONG_TOTAL_SUPPLY)
		}
		if _, _, err := Transfer(native, contract, &v); err != nil {
			return utils.BYTE_FALSE, err
		}
		AddNotifications(native, contract, &v)
	}
	return utils.BYTE_TRUE, nil
}
func OngxApprove(native *native.NativeService) ([]byte, error) {
	var state State
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OngApprove] state deserialize error!")
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if state.Value > constants.ONG_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("approve ong amount:%d over totalSupply:%d", state.Value, constants.ONG_TOTAL_SUPPLY)
	}
	if native.ContextRef.CheckWitness(state.From) == false {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	native.CacheDB.Put(GenApproveKey(contract, state.From, state.To), utils.GenUInt64StorageItem(state.Value).ToArray())
	return utils.BYTE_TRUE, nil
}
func OngxTransferFrom(native *native.NativeService) ([]byte, error) {
	var state TransferFrom
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntTransferFrom] State deserialize error!")
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if state.Value > constants.ONG_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("approve ong amount:%d over totalSupply:%d", state.Value, constants.ONG_TOTAL_SUPPLY)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	if _, _, err := TransferedFrom(native, contract, &state); err != nil {
		return utils.BYTE_FALSE, err
	}
	AddNotifications(native, contract, &State{From: state.From, To: state.To, Value: state.Value})
	return utils.BYTE_TRUE, nil
}
func OngxName(native *native.NativeService) ([]byte, error) {
	return []byte(constants.ONG_NAME), nil
}
func OngxDecimals(native *native.NativeService) ([]byte, error) {
	return big.NewInt(int64(constants.ONG_DECIMALS)).Bytes(), nil
}
func OngxSymbol(native *native.NativeService) ([]byte, error) {
	return []byte(constants.ONG_SYMBOL), nil
}
func OngxTotalSupply(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	amount, err := utils.GetStorageUInt64(native, GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntTotalSupply] get totalSupply error!")
	}
	return types.BigIntToBytes(big.NewInt(int64(amount))), nil
}
func OngxBalanceOf(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, TRANSFER_FLAG)
}
func OngxAllowance(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, APPROVE_FLAG)
}
func OngxSetSyncAddr(native *native.NativeService) ([]byte, error) {
	context := native.ContextRef.CurrentContext().ContractAddress[:]
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}
	//check witness
	err = utils.ValidateOwner(native, adminAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ongxSetSyncAddr, checkWitness error: %v", err)
	}
	native.CacheDB.Put(append(context, SYNC_ADDRESS...), states.GenRawStorageItem(native.Input))
	return utils.BYTE_TRUE, nil
}
func OngSwap(native *native.NativeService) ([]byte, error) {
	context := native.ContextRef.CurrentContext().ContractAddress
	key := append(context[:], SYNC_ADDRESS...)
	result, err := native.CacheDB.Get(key)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngSwap] get address from cache error:%s", err)
	}
	if result == nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngSwap] sync address is nil")
	}
	syncAddrBytes, err := states.GetValueFromRawStorageItem(result)
	if err != nil {
		return nil, fmt.Errorf("[OngSwap], deserialize from raw storage item err:%v", err)
	}
	syncAddr := new(SyncAddress)
	if err := syncAddr.Deserialize(common.NewZeroCopySource(syncAddrBytes)); err != nil {
		return nil, fmt.Errorf("deserialize, deserialize syncAddr error: %v", err)
	}
	if !native.ContextRef.CheckWitness(syncAddr.SyncAddress) {
		return utils.BYTE_FALSE, errors.NewErr("[OngSwap] authentication failed!")
	}
	source := common.NewZeroCopySource(native.Input)
	var ongSwapParam OngSwapParam
	if err := ongSwapParam.Deserialize(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngSwap] error:%s", err)
	}
	totalSupplyKey := GenTotalSupplyKey(context)
	amount, err := utils.GetStorageUInt64(native, totalSupplyKey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngSwap] error:%s", err)
	}
	for _, v := range ongSwapParam.Swap {
		key := append(context[:], v.Addr[:]...)
		balance, err := utils.GetStorageUInt64(native, key)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("[OngSwap] error:%s", err)
		}
		native.CacheDB.Put(key, GetToUInt64StorageItem(balance, v.Value).ToArray())
		amount += v.Value
		AddNotifications(native, context, &State{To: v.Addr, Value: v.Value})
	}
	native.CacheDB.Put(totalSupplyKey, utils.GenUInt64StorageItem(amount).ToArray())
	return utils.BYTE_TRUE, nil
}
func OngxSwap(native *native.NativeService) ([]byte, error) {
	context := native.ContextRef.CurrentContext().ContractAddress
	source := common.NewZeroCopySource(native.Input)
	var swap Swap
	if err := swap.Deserialize(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngxSwap] error:%s", err)
	}
	if !native.ContextRef.CheckWitness(swap.Addr) {
		return utils.BYTE_FALSE, errors.NewErr("[OngxSwap] authentication failed!")
	}
	key := append(context[:], swap.Addr[:]...)
	balance, err := utils.GetStorageUInt64(native, key)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngxSwap] error:%s", err)
	}
	if swap.Value > balance {
		return utils.BYTE_FALSE, fmt.Errorf("[OngxSwap] swap ongx balance insufficient! have %d, want %d", balance, swap.Value)
	} else if swap.Value == balance {
		native.CacheDB.Delete(key)
	} else {
		native.CacheDB.Put(key, utils.GenUInt64StorageItem(balance-swap.Value).ToArray())
	}
	AddOngxSwapNotifications(native, context, &State{From: swap.Addr, Value: swap.Value})
	totalSupplyKey := GenTotalSupplyKey(context)
	amount, err := utils.GetStorageUInt64(native, totalSupplyKey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngxSwap] error:%s", err)
	}
	native.CacheDB.Put(totalSupplyKey, utils.GenUInt64StorageItem(amount-swap.Value).ToArray())
	return utils.BYTE_TRUE, nil
}