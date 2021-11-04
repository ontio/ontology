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

package ont

import (
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

// storage key
const (
	UNBOUND_TIME_OFFSET_KEY = "unboundTimeOffset"
	TOTAL_SUPPLY_KEY        = "totalSupply"
)

// methods
const (
	INIT_NAME            = "init"
	TRANSFER_NAME        = "transfer"
	APPROVE_NAME         = "approve"
	TRANSFERFROM_NAME    = "transferFrom"
	NAME_NAME            = "name"
	SYMBOL_NAME          = "symbol"
	DECIMALS_NAME        = "decimals"
	TOTAL_SUPPLY_NAME    = "totalSupply"
	BALANCEOF_NAME       = "balanceOf"
	ALLOWANCE_NAME       = "allowance"
	TOTAL_ALLOWANCE_NAME = "totalAllowance"

	TRANSFER_V2_NAME        = "transferV2"
	APPROVE_V2_NAME         = "approveV2"
	TRANSFERFROM_V2_NAME    = "transferFromV2"
	DECIMALS_V2_NAME        = "decimalsV2"
	TOTAL_SUPPLY_V2_NAME    = "totalSupplyV2"
	BALANCEOF_V2_NAME       = "balanceOfV2"
	ALLOWANCE_V2_NAME       = "allowanceV2"
	TOTAL_ALLOWANCE_V2_NAME = "totalAllowanceV2"

	UNBOUND_ONG_TO_GOVERNANCE = "unboundOngToGovernance"
)

func AddTransferNotifications(native *native.NativeService, contract common.Address, state *TransferStateV2) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	states := []interface{}{TRANSFER_NAME, state.From.ToBase58(), state.To.ToBase58(), state.Value.MustToInteger64()}
	if state.Value.IsFloat() {
		states = append(states, state.Value.FloatPart())
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          states,
		})
}

func GetToUInt64StorageItem(toBalance, value uint64) *cstates.StorageItem {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(toBalance + value)
	return &cstates.StorageItem{Value: sink.Bytes()}
}

func GenTotalSupplyKey(contract common.Address) []byte {
	return append(contract[:], TOTAL_SUPPLY_KEY...)
}

func GenBalanceKey(contract, addr common.Address) []byte {
	return append(contract[:], addr[:]...)
}

func Transfer(native *native.NativeService, contract common.Address, from, to common.Address,
	value cstates.NativeTokenBalance) (oldFrom, oldTo cstates.NativeTokenBalance, err error) {
	if !native.ContextRef.CheckWitness(from) {
		return oldFrom, oldTo, errors.NewErr("authentication failed!")
	}

	oldFrom, err = reduceFromBalance(native, GenBalanceKey(contract, from), value)
	if err != nil {
		return
	}

	oldTo, err = increaseToBalance(native, GenBalanceKey(contract, to), value)
	return
}

func GenApproveKey(contract, from, to common.Address) []byte {
	temp := append(contract[:], from[:]...)
	return append(temp, to[:]...)
}

func TransferedFrom(native *native.NativeService, currentContract common.Address, state *TransferFromStateV2) (oldFrom cstates.NativeTokenBalance, oldTo cstates.NativeTokenBalance, err error) {
	if native.Time <= config.GetOntHolderUnboundDeadline()+constants.GENESIS_BLOCK_TIMESTAMP {
		if !native.ContextRef.CheckWitness(state.Sender) {
			err = errors.NewErr("authentication failed!")
			return
		}
	} else {
		if state.Sender != state.To && !native.ContextRef.CheckWitness(state.Sender) {
			err = errors.NewErr("authentication failed!")
			return
		}
	}

	if err = fromApprove(native, genTransferFromKey(currentContract, state.From, state.Sender), state.Value); err != nil {
		return
	}

	oldFrom, err = reduceFromBalance(native, GenBalanceKey(currentContract, state.From), state.Value)
	if err != nil {
		return
	}

	oldTo, err = increaseToBalance(native, GenBalanceKey(currentContract, state.To), state.Value)
	if err != nil {
		return
	}
	return oldFrom, oldTo, nil
}

func getUnboundOffset(native *native.NativeService, contract, address common.Address) (uint32, error) {
	offset, err := utils.GetStorageUInt32(native.CacheDB, genAddressUnboundOffsetKey(contract, address))
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func getGovernanceUnboundOffset(native *native.NativeService, contract common.Address) (uint32, error) {
	offset, err := utils.GetStorageUInt32(native.CacheDB, genGovernanceUnboundOffsetKey(contract))
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func genTransferFromKey(contract common.Address, from, sender common.Address) []byte {
	temp := append(contract[:], from[:]...)
	return append(temp, sender[:]...)
}

func fromApprove(native *native.NativeService, fromApproveKey []byte, value cstates.NativeTokenBalance) error {
	approveValue, err := utils.GetNativeTokenBalance(native.CacheDB, fromApproveKey)
	if err != nil {
		return err
	}
	newApprove, err := approveValue.Sub(value)
	if err != nil {
		return fmt.Errorf("[TransferFrom] approve balance insufficient: %v", err)
	}
	if newApprove.IsZero() {
		native.CacheDB.Delete(fromApproveKey)
	} else {
		native.CacheDB.Put(fromApproveKey, newApprove.MustToStorageItemBytes())
	}
	return nil
}

func reduceFromBalance(native *native.NativeService, fromKey []byte, value cstates.NativeTokenBalance) (cstates.NativeTokenBalance, error) {
	fromBalance, err := utils.GetNativeTokenBalance(native.CacheDB, fromKey)
	if err != nil {
		return cstates.NativeTokenBalance{}, err
	}
	newFromBalance, err := fromBalance.Sub(value)
	if err != nil {
		addr, _ := common.AddressParseFromBytes(fromKey[20:])
		return cstates.NativeTokenBalance{}, fmt.Errorf("[Transfer] balance insufficient. contract:%s, account:%s, fromBalance:%s,value:%s, err: %v",
			native.ContextRef.CurrentContext().ContractAddress.ToHexString(), addr.ToBase58(), fromBalance.String(), value.String(), err)
	}

	if newFromBalance.IsZero() {
		native.CacheDB.Delete(fromKey)
	} else {
		native.CacheDB.Put(fromKey, newFromBalance.MustToStorageItemBytes())
	}

	return fromBalance, nil
}

func increaseToBalance(native *native.NativeService, toKey []byte, value cstates.NativeTokenBalance) (cstates.NativeTokenBalance, error) {
	toBalance, err := utils.GetNativeTokenBalance(native.CacheDB, toKey)
	if err != nil {
		return cstates.NativeTokenBalance{}, err
	}
	native.CacheDB.Put(toKey, toBalance.Add(value).MustToStorageItemBytes())
	return toBalance, nil
}

func genAddressUnboundOffsetKey(contract, address common.Address) []byte {
	temp := append(contract[:], UNBOUND_TIME_OFFSET_KEY...)
	return append(temp, address[:]...)
}

func genGovernanceUnboundOffsetKey(contract common.Address) []byte {
	temp := append(contract[:], UNBOUND_TIME_OFFSET_KEY...)
	return temp
}
