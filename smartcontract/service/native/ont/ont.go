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
	"bytes"
	"fmt"
	"math/big"

	"github.com/laizy/bigint"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/common/log"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	TRANSFER_FLAG byte = 1
	APPROVE_FLAG  byte = 2
)

func InitOnt() {
	native.Contracts[utils.OntContractAddress] = RegisterOntContract
}

func RegisterOntContract(native *native.NativeService) {
	native.Register(INIT_NAME, OntInit)
	native.Register(NAME_NAME, OntName)
	native.Register(SYMBOL_NAME, OntSymbol)
	native.Register(TRANSFER_NAME, OntTransfer)
	native.Register(APPROVE_NAME, OntApprove)
	native.Register(TRANSFERFROM_NAME, OntTransferFrom)
	native.Register(DECIMALS_NAME, OntDecimals)
	native.Register(TOTAL_SUPPLY_NAME, OntTotalSupply)
	native.Register(BALANCEOF_NAME, OntBalanceOf)
	native.Register(ALLOWANCE_NAME, OntAllowance)
	native.Register(TOTAL_ALLOWANCE_NAME, TotalAllowance)

	if native.Height >= config.GetAddDecimalsHeight() {
		native.Register(TRANSFER_V2_NAME, OntTransferV2)
		native.Register(APPROVE_V2_NAME, OntApproveV2)
		native.Register(TRANSFERFROM_V2_NAME, OntTransferFromV2)
		native.Register(DECIMALS_V2_NAME, OntDecimalsV2)
		native.Register(TOTAL_SUPPLY_V2_NAME, OntTotalSupplyV2)
		native.Register(BALANCEOF_V2_NAME, OntBalanceOfV2)
		native.Register(ALLOWANCE_V2_NAME, OntAllowanceV2)
		native.Register(TOTAL_ALLOWANCE_V2_NAME, TotalAllowanceV2)
	}

	native.Register(UNBOUND_ONG_TO_GOVERNANCE, UnboundOngToGovernance)
}

func OntInit(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	amount, err := utils.GetNativeTokenBalance(native.CacheDB, GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	if amount.IsZero() == false {
		return utils.BYTE_FALSE, errors.NewErr("Init ont has been completed!")
	}

	distribute := make(map[common.Address]uint64)
	source := common.NewZeroCopySource(native.Input)
	buf, _, irregular, eof := source.NextVarBytes()
	if eof {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarBytes, contract params deserialize error!")
	}
	if irregular {
		return utils.BYTE_FALSE, common.ErrIrregularData
	}
	input := common.NewZeroCopySource(buf)
	num, err := utils.DecodeVarUint(input)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("read number error:%v", err)
	}
	sum := uint64(0)
	overflow := false
	for i := uint64(0); i < num; i++ {
		addr, err := utils.DecodeAddress(input)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("read address error:%v", err)
		}
		value, err := utils.DecodeVarUint(input)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("read value error:%v", err)
		}
		sum, overflow = common.SafeAdd(sum, value)
		if overflow {
			return utils.BYTE_FALSE, errors.NewErr("wrong config. overflow detected")
		}
		distribute[addr] += value
	}
	if sum != constants.ONT_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("wrong config. total supply %d != %d", sum, constants.ONT_TOTAL_SUPPLY)
	}

	for addr, val := range distribute {
		balanceKey := GenBalanceKey(contract, addr)
		balance := cstates.NativeTokenBalanceFromInteger(val)
		native.CacheDB.Put(balanceKey, balance.MustToStorageItemBytes())
		AddTransferNotifications(native, contract, &TransferStateV2{To: addr, Value: balance})
	}

	native.CacheDB.Put(GenTotalSupplyKey(contract), cstates.NativeTokenBalanceFromInteger(constants.ONT_TOTAL_SUPPLY).MustToStorageItemBytes())

	return utils.BYTE_TRUE, nil
}

func doTransfer(native *native.NativeService, transfers *TransferStatesV2) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	for _, v := range transfers.States {
		if v.Value.IsZero() {
			continue
		}
		if bigint.New(constants.ONT_TOTAL_SUPPLY_V2).LessThan(v.Value.Balance) {
			return utils.BYTE_FALSE, fmt.Errorf("transfer ont amount:%d over totalSupply:%d", v.Value, constants.ONT_TOTAL_SUPPLY)
		}
		fromBalance, toBalance, err := Transfer(native, contract, v.From, v.To, v.Value)
		if err != nil {
			return utils.BYTE_FALSE, err
		}

		if err := grantOng(native, contract, v.From, fromBalance.MustToInteger64()); err != nil {
			return utils.BYTE_FALSE, err
		}

		if err := grantOng(native, contract, v.To, toBalance.MustToInteger64()); err != nil {
			return utils.BYTE_FALSE, err
		}

		AddTransferNotifications(native, contract, v)
	}

	return utils.BYTE_TRUE, nil
}

func OntTransfer(native *native.NativeService) ([]byte, error) {
	var transfers TransferStates
	source := common.NewZeroCopySource(native.Input)
	if err := transfers.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[Transfer] TransferStates deserialize error!")
	}
	return doTransfer(native, transfers.ToV2())
}

func OntTransferV2(native *native.NativeService) ([]byte, error) {
	var transfers TransferStatesV2
	source := common.NewZeroCopySource(native.Input)
	if err := transfers.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[Transfer] TransferStates deserialize error!")
	}
	return doTransfer(native, &transfers)
}

func OntTransferFrom(native *native.NativeService) ([]byte, error) {
	var state TransferFrom
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntTransferFrom] State deserialize error!")
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if state.Value > constants.ONT_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("transferFrom ont amount:%d over totalSupply:%d", state.Value, constants.ONT_TOTAL_SUPPLY)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	fromBalance, toBalance, err := TransferedFrom(native, contract, state.ToV2())
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if err := grantOng(native, contract, state.From, fromBalance.MustToInteger64()); err != nil {
		return utils.BYTE_FALSE, err
	}
	if err := grantOng(native, contract, state.To, toBalance.MustToInteger64()); err != nil {
		return utils.BYTE_FALSE, err
	}
	AddTransferNotifications(native, contract, state.TransferState.ToV2())
	return utils.BYTE_TRUE, nil
}

func OntTransferFromV2(native *native.NativeService) ([]byte, error) {
	var state TransferFromStateV2
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntTransferFrom] State deserialize error!")
	}
	if state.Value.IsZero() {
		return utils.BYTE_FALSE, nil
	}
	if bigint.New(constants.ONT_TOTAL_SUPPLY_V2).LessThan(state.Value.Balance) {
		return utils.BYTE_FALSE, fmt.Errorf("transferFrom ont amount:%s over totalSupply:%d", state.Value, constants.ONT_TOTAL_SUPPLY)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	fromBalance, toBalance, err := TransferedFrom(native, contract, &state)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if err := grantOng(native, contract, state.From, fromBalance.MustToInteger64()); err != nil {
		return utils.BYTE_FALSE, err
	}
	if err := grantOng(native, contract, state.To, toBalance.MustToInteger64()); err != nil {
		return utils.BYTE_FALSE, err
	}
	AddTransferNotifications(native, contract, &state.TransferStateV2)
	return utils.BYTE_TRUE, nil
}

func OntApprove(native *native.NativeService) ([]byte, error) {
	var state TransferState
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntApprove] state deserialize error!")
	}
	if state.Value > constants.ONT_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("approve ont amount:%d over totalSupply:%d", state.Value, constants.ONT_TOTAL_SUPPLY)
	}
	if !native.ContextRef.CheckWitness(state.From) {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	native.CacheDB.Put(GenApproveKey(contract, state.From, state.To), utils.GenUInt64StorageItem(state.Value).ToArray())
	return utils.BYTE_TRUE, nil
}

func OntApproveV2(native *native.NativeService) ([]byte, error) {
	var state TransferStateV2
	source := common.NewZeroCopySource(native.Input)
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntApprove] state deserialize error!")
	}
	if bigint.New(constants.ONT_TOTAL_SUPPLY_V2).LessThan(state.Value.Balance) {
		return utils.BYTE_FALSE, fmt.Errorf("approve ont amount:%s over totalSupply:%d", state.Value, constants.ONT_TOTAL_SUPPLY)
	}
	if !native.ContextRef.CheckWitness(state.From) {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	native.CacheDB.Put(GenApproveKey(contract, state.From, state.To), state.Value.MustToStorageItemBytes())
	return utils.BYTE_TRUE, nil
}

func OntName(native *native.NativeService) ([]byte, error) {
	return []byte(constants.ONT_NAME), nil
}

func OntDecimals(native *native.NativeService) ([]byte, error) {
	return common.BigIntToNeoBytes(big.NewInt(int64(constants.ONT_DECIMALS))), nil
}

func OntDecimalsV2(native *native.NativeService) ([]byte, error) {
	return common.BigIntToNeoBytes(big.NewInt(int64(constants.ONT_DECIMALS_V2))), nil
}

func OntSymbol(native *native.NativeService) ([]byte, error) {
	return []byte(constants.ONT_SYMBOL), nil
}

func OntTotalSupply(native *native.NativeService) ([]byte, error) {
	return common.BigIntToNeoBytes(big.NewInt(constants.ONT_TOTAL_SUPPLY)), nil
}

func OntTotalSupplyV2(native *native.NativeService) ([]byte, error) {
	return common.BigIntToNeoBytes(big.NewInt(constants.ONT_TOTAL_SUPPLY_V2)), nil
}

func OntBalanceOf(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, TRANSFER_FLAG, false)
}

func OntAllowance(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, APPROVE_FLAG, false)
}

func OntBalanceOfV2(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, TRANSFER_FLAG, true)
}

func OntAllowanceV2(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, APPROVE_FLAG, true)
}

func GetBalanceValue(native *native.NativeService, flag byte, scaleDecimal9 bool) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	from, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] get from address error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	var key []byte
	if flag == APPROVE_FLAG {
		to, err := utils.DecodeAddress(source)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] get from address error!")
		}
		key = GenApproveKey(contract, from, to)
	} else if flag == TRANSFER_FLAG {
		key = GenBalanceKey(contract, from)
	}
	balance, err := utils.GetNativeTokenBalance(native.CacheDB, key)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] address parse error!")
	}
	amount := balance.ToBigInt()
	if !scaleDecimal9 {
		amount = balance.ToInteger().BigInt()
	}
	return common.BigIntToNeoBytes(amount), nil
}

func getTotalAllowance(native *native.NativeService, from common.Address) (result cstates.NativeTokenBalance, err error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	iter := native.CacheDB.NewIterator(utils.ConcatKey(contract, from[:]))
	defer iter.Release()
	r := cstates.NativeTokenBalanceFromInteger(0)
	for has := iter.First(); has; has = iter.Next() {
		if bytes.Equal(iter.Key(), utils.ConcatKey(contract, from[:])) {
			continue
		}
		item := new(cstates.StorageItem)
		err = item.Deserialization(common.NewZeroCopySource(iter.Value()))
		if err != nil {
			return result, errors.NewDetailErr(err, errors.ErrNoCode, "[TotalAllowance] instance isn't StorageItem!")
		}
		balance, err := cstates.NativeTokenBalanceFromStorageItem(item)
		if err != nil {
			return result, errors.NewDetailErr(err, errors.ErrNoCode, "[TotalAllowance] get token allowance from storage value error!")
		}
		r = r.Add(balance)
	}

	return r, nil
}

func TotalAllowance(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	from, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[TotalAllowance] get from address error!")
	}
	r, err := getTotalAllowance(native, from)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return common.BigIntToNeoBytes(r.ToInteger().BigInt()), nil
}

func TotalAllowanceV2(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	from, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[TotalAllowance] get from address error!")
	}
	r, err := getTotalAllowance(native, from)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return common.BigIntToNeoBytes(r.ToBigInt()), nil
}

func UnboundOngToGovernance(native *native.NativeService) ([]byte, error) {
	err := unboundOngToGovernance(native)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("unboundOngToGovernance error: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func grantOng(native *native.NativeService, contract, address common.Address, balance uint64) error {
	startOffset, err := getUnboundOffset(native, contract, address)
	if err != nil {
		return err
	}
	if native.Time <= constants.GENESIS_BLOCK_TIMESTAMP {
		return nil
	}
	endOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP
	if endOffset < startOffset {
		if native.PreExec {
			return nil
		}
		errstr := fmt.Sprintf("grant Ong error: wrong timestamp endOffset: %d < startOffset: %d", endOffset, startOffset)
		log.Error(errstr)
		return errors.NewErr(errstr)
	} else if endOffset == startOffset {
		return nil
	}

	if balance != 0 {
		value := utils.CalcUnbindOng(balance, startOffset, endOffset)

		args, amount, err := getApproveArgs(native, contract, utils.OngContractAddress, address, value)
		if err != nil {
			return err
		}
		if _, err := native.NativeCall(utils.OngContractAddress, "approve", args); err != nil {
			return err
		}
		if endOffset > config.GetOntHolderUnboundDeadline() {
			if address != utils.GovernanceContractAddress {
				args, err := getTransferFromArgs(address, contract, address, amount)
				if err != nil {
					return err
				}
				if _, err := native.NativeCall(utils.OngContractAddress, "transferFrom", args); err != nil {
					return err
				}
			}
		}
	}

	native.CacheDB.Put(genAddressUnboundOffsetKey(contract, address), utils.GenUInt32StorageItem(endOffset).ToArray())
	return nil
}

func unboundOngToGovernance(native *native.NativeService) error {
	contract := utils.OntContractAddress
	address := utils.GovernanceContractAddress
	startOffset, err := getGovernanceUnboundOffset(native, contract)
	if err != nil {
		return err
	}
	if native.Time <= constants.GENESIS_BLOCK_TIMESTAMP {
		return nil
	}
	endOffset := native.Time - constants.GENESIS_BLOCK_TIMESTAMP
	if endOffset < startOffset {
		if native.PreExec {
			return nil
		}
		errstr := fmt.Sprintf("grant Ong error: wrong timestamp endOffset: %d < startOffset: %d", endOffset, startOffset)
		log.Error(errstr)
		return errors.NewErr(errstr)
	} else if endOffset == startOffset {
		return nil
	}

	value := utils.CalcGovernanceUnbindOng(startOffset, endOffset)

	args, err := getTransferArgs(contract, address, value)
	if err != nil {
		return err
	}

	if _, err := native.NativeCall(utils.OngContractAddress, "transfer", args); err != nil {
		return err
	}

	native.CacheDB.Put(genGovernanceUnboundOffsetKey(contract), utils.GenUInt32StorageItem(endOffset).ToArray())
	return nil
}

func getApproveArgs(native *native.NativeService, contract, ongContract, address common.Address, value uint64) ([]byte, uint64, error) {
	bf := common.NewZeroCopySink(nil)
	approve := TransferState{
		From:  contract,
		To:    address,
		Value: value,
	}

	stateValue, err := utils.GetStorageUInt64(native.CacheDB, GenApproveKey(ongContract, approve.From, approve.To))
	if err != nil {
		return nil, 0, err
	}

	approve.Value += stateValue
	approve.Serialization(bf)
	return bf.Bytes(), approve.Value, nil
}

func getTransferArgs(contract, address common.Address, value uint64) ([]byte, error) {
	bf := common.NewZeroCopySink(nil)
	state := TransferState{
		From:  contract,
		To:    address,
		Value: value,
	}
	transfers := TransferStates{[]TransferState{state}}

	transfers.Serialization(bf)
	return bf.Bytes(), nil
}

func getTransferFromArgs(sender, from, to common.Address, value uint64) ([]byte, error) {
	sink := common.NewZeroCopySink(nil)
	param := TransferFrom{
		Sender: sender,
		TransferState: TransferState{
			From:  from,
			To:    to,
			Value: value,
		},
	}

	param.Serialization(sink)
	return sink.Bytes(), nil
}
