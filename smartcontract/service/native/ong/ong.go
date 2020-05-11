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

package ong

import (
	"fmt"
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func InitOng() {
	native.Contracts[utils.OngContractAddress] = RegisterOngContract
}

func RegisterOngContract(native *native.NativeService) {
	native.Register(ont.INIT_NAME, OngInit)
	native.Register(ont.TRANSFER_NAME, OngTransfer)
	native.Register(ont.APPROVE_NAME, OngApprove)
	native.Register(ont.TRANSFERFROM_NAME, OngTransferFrom)
	native.Register(ont.NAME_NAME, OngName)
	native.Register(ont.SYMBOL_NAME, OngSymbol)
	native.Register(ont.DECIMALS_NAME, OngDecimals)
	native.Register(ont.TOTALSUPPLY_NAME, OngTotalSupply)
	native.Register(ont.BALANCEOF_NAME, OngBalanceOf)
	native.Register(ont.ALLOWANCE_NAME, OngAllowance)
	native.Register(ont.TOTAL_ALLOWANCE_NAME, OngTotalAllowance)
}

func OngInit(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	amount, err := utils.GetStorageUInt64(native, ont.GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	if amount > 0 {
		return utils.BYTE_FALSE, errors.NewErr("Init ong has been completed!")
	}

	item := utils.GenUInt64StorageItem(constants.ONG_TOTAL_SUPPLY)
	native.CacheDB.Put(ont.GenTotalSupplyKey(contract), item.ToArray())
	native.CacheDB.Put(append(contract[:], utils.OntContractAddress[:]...), item.ToArray())
	ont.AddNotifications(native, contract, &ont.State{To: utils.OntContractAddress, Value: constants.ONG_TOTAL_SUPPLY})
	return utils.BYTE_TRUE, nil
}

func OngTransfer(native *native.NativeService) ([]byte, error) {
	var transfers ont.Transfers
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
		if _, _, err := ont.Transfer(native, contract, &v); err != nil {
			return utils.BYTE_FALSE, err
		}
		ont.AddNotifications(native, contract, &v)
	}
	return utils.BYTE_TRUE, nil
}

func OngApprove(native *native.NativeService) ([]byte, error) {
	var state ont.State
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OngApprove] state deserialize error!")
	}
	if state.Value > constants.ONG_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("approve ong amount:%d over totalSupply:%d", state.Value, constants.ONG_TOTAL_SUPPLY)
	}
	if native.ContextRef.CheckWitness(state.From) == false {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	native.CacheDB.Put(ont.GenApproveKey(contract, state.From, state.To), utils.GenUInt64StorageItem(state.Value).ToArray())
	return utils.BYTE_TRUE, nil
}

func OngTransferFrom(native *native.NativeService) ([]byte, error) {
	var state ont.TransferFrom
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
	if _, _, err := ont.TransferedFrom(native, contract, &state); err != nil {
		return utils.BYTE_FALSE, err
	}
	ont.AddNotifications(native, contract, &ont.State{From: state.From, To: state.To, Value: state.Value})
	return utils.BYTE_TRUE, nil
}

func OngName(native *native.NativeService) ([]byte, error) {
	return []byte(constants.ONG_NAME), nil
}

func OngDecimals(native *native.NativeService) ([]byte, error) {
	return big.NewInt(int64(constants.ONG_DECIMALS)).Bytes(), nil
}

func OngSymbol(native *native.NativeService) ([]byte, error) {
	return []byte(constants.ONG_SYMBOL), nil
}

func OngTotalSupply(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	amount, err := utils.GetStorageUInt64(native, ont.GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntTotalSupply] get totalSupply error!")
	}
	return common.BigIntToNeoBytes(big.NewInt(int64(amount))), nil
}

func OngBalanceOf(native *native.NativeService) ([]byte, error) {
	return ont.GetBalanceValue(native, ont.TRANSFER_FLAG)
}

func OngAllowance(native *native.NativeService) ([]byte, error) {
	return ont.GetBalanceValue(native, ont.APPROVE_FLAG)
}

func OngTotalAllowance(native *native.NativeService) ([]byte, error) {
	return ont.TotalAllowance(native)
}
