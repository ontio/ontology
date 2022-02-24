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

	"github.com/laizy/bigint"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/storage"
)

func InitOng() {
	native.Contracts[utils.OngContractAddress] = RegisterOngContract
}

func RegisterOngContract(native *native.NativeService) {
	native.Register(ont.INIT_NAME, OngInit)
	native.Register(ont.NAME_NAME, OngName)
	native.Register(ont.SYMBOL_NAME, OngSymbol)
	native.Register(ont.TRANSFER_NAME, OngTransfer)
	native.Register(ont.APPROVE_NAME, OngApprove)
	native.Register(ont.TRANSFERFROM_NAME, OngTransferFrom)
	native.Register(ont.DECIMALS_NAME, OngDecimals)
	native.Register(ont.TOTAL_SUPPLY_NAME, OngTotalSupply)
	native.Register(ont.BALANCEOF_NAME, OngBalanceOf)
	native.Register(ont.ALLOWANCE_NAME, OngAllowance)
	native.Register(ont.TOTAL_ALLOWANCE_NAME, OngTotalAllowance)

	if native.Height >= config.GetAddDecimalsHeight() || native.PreExec {
		native.Register(ont.BALANCEOF_V2_NAME, OngBalanceOfV2)
		native.Register(ont.ALLOWANCE_V2_NAME, OngAllowanceV2)
		native.Register(ont.TOTAL_ALLOWANCE_V2_NAME, OngTotalAllowanceV2)
	}

	if native.Height >= config.GetAddDecimalsHeight() {
		native.Register(ont.TRANSFER_V2_NAME, OngTransferV2)
		native.Register(ont.APPROVE_V2_NAME, OngApproveV2)
		native.Register(ont.TRANSFERFROM_V2_NAME, OngTransferFromV2)
		native.Register(ont.DECIMALS_V2_NAME, OngDecimalsV2)
		native.Register(ont.TOTAL_SUPPLY_V2_NAME, OngTotalSupplyV2)
	}
}

func OngInit(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	amount, err := utils.GetNativeTokenBalance(native.CacheDB, ont.GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	if !amount.IsZero() {
		return utils.BYTE_FALSE, errors.NewErr("Init ong has been completed!")
	}
	addr := common.Address{}
	if config.DefConfig.P2PNode.NetworkId == config.NETWORK_ID_SOLO_NET {
		bookkeepers, _ := config.DefConfig.GetBookkeepers()
		addr = types.AddressFromPubKey(bookkeepers[0])
	} else {
		addr = utils.OntContractAddress
	}

	item := utils.GenUInt64StorageItem(constants.ONG_TOTAL_SUPPLY)
	native.CacheDB.Put(ont.GenTotalSupplyKey(contract), item.ToArray())
	native.CacheDB.Put(append(contract[:], addr[:]...), item.ToArray())
	state := &ont.TransferState{To: utils.OntContractAddress, Value: constants.ONG_TOTAL_SUPPLY}
	ont.AddTransferNotifications(native, contract, state.ToV2())
	return utils.BYTE_TRUE, nil
}

func OngTransfer(native *native.NativeService) ([]byte, error) {
	var transfers ont.TransferStates
	source := common.NewZeroCopySource(native.Input)
	if err := transfers.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OngTransfer] TransferStates deserialize error!")
	}

	return doTransfer(native, transfers.ToV2())
}

func OngTransferV2(native *native.NativeService) ([]byte, error) {
	var transfers ont.TransferStatesV2
	source := common.NewZeroCopySource(native.Input)
	if err := transfers.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OngTransfer] TransferStates deserialize error!")
	}

	return doTransfer(native, &transfers)
}

func doTransfer(native *native.NativeService, transfers *ont.TransferStatesV2) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	for _, v := range transfers.States {
		if v.Value.IsZero() {
			continue
		}
		if constants.ONG_TOTAL_SUPPLY_V2.LessThan(v.Value.Balance) {
			return utils.BYTE_FALSE, fmt.Errorf("transfer ong amount:%s over totalSupply:%d", v.Value, constants.ONG_TOTAL_SUPPLY)
		}
		if _, _, err := ont.Transfer(native, contract, v.From, v.To, v.Value); err != nil {
			return utils.BYTE_FALSE, err
		}
		ont.AddTransferNotifications(native, contract, v)
	}
	return utils.BYTE_TRUE, nil
}

func doApprove(native *native.NativeService, state *ont.TransferStateV2) ([]byte, error) {
	if constants.ONG_TOTAL_SUPPLY_V2.LessThan(state.Value.Balance) {
		return utils.BYTE_FALSE, fmt.Errorf("approve ong amount:%s over totalSupply:%s", state.Value, constants.ONG_TOTAL_SUPPLY_V2)
	}
	if !native.ContextRef.CheckWitness(state.From) {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	native.CacheDB.Put(ont.GenApproveKey(contract, state.From, state.To), state.Value.MustToStorageItemBytes())
	return utils.BYTE_TRUE, nil

}

func OngApprove(native *native.NativeService) ([]byte, error) {
	var state ont.TransferState
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OngApprove] state deserialize error!")
	}

	return doApprove(native, state.ToV2())
}

func OngApproveV2(native *native.NativeService) ([]byte, error) {
	var state ont.TransferStateV2
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OngApprove] state deserialize error!")
	}

	return doApprove(native, &state)
}

func doTransferFrom(native *native.NativeService, state *ont.TransferFromStateV2) ([]byte, error) {
	if state.Value.IsZero() {
		return utils.BYTE_FALSE, nil
	}
	if constants.ONG_TOTAL_SUPPLY_V2.LessThan(state.Value.Balance) {
		return utils.BYTE_FALSE, fmt.Errorf("approve ong amount:%s over totalSupply:%d", state.Value, constants.ONG_TOTAL_SUPPLY)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	if _, _, err := ont.TransferedFrom(native, contract, state); err != nil {
		return utils.BYTE_FALSE, err
	}
	ont.AddTransferNotifications(native, contract, &state.TransferStateV2)
	return utils.BYTE_TRUE, nil

}

func OngTransferFromV2(native *native.NativeService) ([]byte, error) {
	var state ont.TransferFromStateV2
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntTransferFrom] State deserialize error!")
	}

	return doTransferFrom(native, &state)
}

func OngTransferFrom(native *native.NativeService) ([]byte, error) {
	var state ont.TransferFrom
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntTransferFrom] State deserialize error!")
	}

	return doTransferFrom(native, state.ToV2())
}

func OngName(native *native.NativeService) ([]byte, error) {
	return []byte(constants.ONG_NAME), nil
}

func OngDecimals(native *native.NativeService) ([]byte, error) {
	return big.NewInt(int64(constants.ONG_DECIMALS)).Bytes(), nil
}

func OngDecimalsV2(native *native.NativeService) ([]byte, error) {
	return big.NewInt(int64(constants.ONG_DECIMALS_V2)).Bytes(), nil
}

func OngSymbol(native *native.NativeService) ([]byte, error) {
	return []byte(constants.ONG_SYMBOL), nil
}

func OngTotalSupply(native *native.NativeService) ([]byte, error) {
	return common.BigIntToNeoBytes(big.NewInt(constants.ONG_TOTAL_SUPPLY)), nil
}

func OngTotalSupplyV2(native *native.NativeService) ([]byte, error) {
	return common.BigIntToNeoBytes(constants.ONG_TOTAL_SUPPLY_V2.BigInt()), nil
}

func OngBalanceOf(native *native.NativeService) ([]byte, error) {
	return ont.GetBalanceValue(native, ont.TRANSFER_FLAG, false)
}

func OngBalanceOfV2(native *native.NativeService) ([]byte, error) {
	return ont.GetBalanceValue(native, ont.TRANSFER_FLAG, true)
}

func OngAllowance(native *native.NativeService) ([]byte, error) {
	return ont.GetBalanceValue(native, ont.APPROVE_FLAG, false)
}

func OngAllowanceV2(native *native.NativeService) ([]byte, error) {
	return ont.GetBalanceValue(native, ont.APPROVE_FLAG, true)
}

func OngTotalAllowance(native *native.NativeService) ([]byte, error) {
	return ont.TotalAllowance(native)
}
func OngTotalAllowanceV2(native *native.NativeService) ([]byte, error) {
	return ont.TotalAllowanceV2(native)
}

type OngBalanceHandle struct{}

func (self OngBalanceHandle) SubBalance(cache *storage.CacheDB, addr common.Address, val *big.Int) error {
	balance, err := getNativeTokenBalance(cache, addr)
	if err != nil {
		return err
	}

	newBalance, err := balance.Sub(states.NativeTokenBalance{Balance: bigint.New(val)})
	if err != nil {
		return err
	}

	return self.SetBalance(cache, addr, newBalance.ToBigInt())
}

func (self OngBalanceHandle) AddBalance(cache *storage.CacheDB, addr common.Address, val *big.Int) error {
	balance, err := self.GetBalance(cache, addr)
	if err != nil {
		return err
	}

	balance.Add(balance, val)
	return self.SetBalance(cache, addr, balance)
}

func (self OngBalanceHandle) SetBalance(cache *storage.CacheDB, addr common.Address, val *big.Int) error {
	balanceKey := ont.GenBalanceKey(utils.OngContractAddress, addr)

	balance := states.NativeTokenBalance{Balance: bigint.New(val)}
	if balance.IsZero() {
		cache.Delete(balanceKey)
	} else {
		cache.Put(balanceKey, balance.MustToStorageItemBytes())
	}

	return nil
}

func getNativeTokenBalance(cache *storage.CacheDB, addr common.Address) (states.NativeTokenBalance, error) {
	balanceKey := ont.GenBalanceKey(utils.OngContractAddress, addr)
	return utils.GetNativeTokenBalance(cache, balanceKey)
}

func (self OngBalanceHandle) GetBalance(cache *storage.CacheDB, addr common.Address) (*big.Int, error) {
	amount, err := getNativeTokenBalance(cache, addr)
	if err != nil {
		return nil, err
	}

	return amount.ToBigInt(), nil
}
