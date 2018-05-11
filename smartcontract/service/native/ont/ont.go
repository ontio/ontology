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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/genesis"
	scommon "github.com/ontio/ontology/core/store/common"
	ctypes "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	TRANSFER_FLAG byte = 1
	APPROVE_FLAG  byte = 2
)

var (
	ONT_NAME           = "ONT Token"
	ONT_SYMBOL         = "ONT"
	ONT_DECIMALS       = 1
	DECREMENT_INTERVAL = uint32(2000000)
	GENERATION_AMOUNT  = [17]uint32{80, 70, 60, 50, 40, 30, 20, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
	GL                 = uint32(len(GENERATION_AMOUNT))
	ONT_TOTAL_SUPPLY   = 1000000000
)

func InitOnt() {
	native.Contracts[genesis.OntContractAddress] = RegisterOntContract
}

func RegisterOntContract(native *native.NativeService) {
	native.Register("init", OntInit)
	native.Register("transfer", OntTransfer)
	native.Register("approve", OntApprove)
	native.Register("transferFrom", OntTransferFrom)
	native.Register("name", OntName)
	native.Register("symbol", OntSymbol)
	native.Register("decimals", OntDecimals)
	native.Register("totalSupply", OntTotalSupply)
	native.Register("balanceOf", OntBalanceOf)
	native.Register("allowance", OntAllowance)
}

func OntInit(native *native.NativeService) ([]byte, error) {
	booKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetBookkeepers error:%s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	amount, err := utils.GetStorageUInt64(native, GetTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	if amount > 0 {
		return utils.BYTE_FALSE, errors.NewErr("Init ont has been completed!")
	}

	assign := uint64(ONT_TOTAL_SUPPLY / len(booKeepers))
	item := utils.GetUInt64StorageItem(assign)
	for _, v := range booKeepers {
		address := ctypes.AddressFromPubKey(v)
		native.CloneCache.Add(scommon.ST_STORAGE, append(contract[:], address[:]...), item)
		native.CloneCache.Add(scommon.ST_STORAGE, GetTotalSupplyKey(contract), item)
		AddNotifications(native, contract, &State{To: address, Value: assign})
	}

	return utils.BYTE_TRUE, nil
}

func OntTransfer(native *native.NativeService) ([]byte, error) {
	transfers := new(Transfers)
	if err := transfers.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[Transfer] Transfers deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	for _, v := range transfers.States {
		if v.Value == 0 {
			continue
		}
		fromBalance, toBalance, err := Transfer(native, contract, v)
		if err != nil {
			return utils.BYTE_FALSE, err
		}

		fromStartHeight, err := getStartHeight(native, contract, v.From)
		if err != nil {
			return utils.BYTE_FALSE, err
		}

		toStartHeight, err := getStartHeight(native, contract, v.To)
		if err != nil {
			return utils.BYTE_FALSE, err
		}

		if err := grantOng(native, contract, v.From, fromBalance, fromStartHeight); err != nil {
			return utils.BYTE_FALSE, err
		}

		if err := grantOng(native, contract, v.To, toBalance, toStartHeight); err != nil {
			return utils.BYTE_FALSE, err
		}

		AddNotifications(native, contract, v)
	}
	return utils.BYTE_TRUE, nil
}

func OntTransferFrom(native *native.NativeService) ([]byte, error) {
	state := new(TransferFrom)
	if err := state.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntTransferFrom] State deserialize error!")
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	if err := TransferedFrom(native, contract, state); err != nil {
		return utils.BYTE_FALSE, err
	}
	AddNotifications(native, contract, &State{From: state.From, To: state.To, Value: state.Value})
	return utils.BYTE_TRUE, nil
}

func OntApprove(native *native.NativeService) ([]byte, error) {
	state := new(State)
	if err := state.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OngApprove] state deserialize error!")
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if err := IsApproveValid(native, state); err != nil {
		return utils.BYTE_FALSE, err
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	native.CloneCache.Add(scommon.ST_STORAGE, GetApproveKey(contract, state.From, state.To), utils.GetUInt64StorageItem(state.Value))
	return utils.BYTE_TRUE, nil
}

func OntName(native *native.NativeService) ([]byte, error) {
	return []byte(ONT_NAME), nil
}

func OntDecimals(native *native.NativeService) ([]byte, error) {
	return big.NewInt(int64(ONT_DECIMALS)).Bytes(), nil
}

func OntSymbol(native *native.NativeService) ([]byte, error) {
	return []byte(ONT_SYMBOL), nil
}

func OntTotalSupply(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	amount, err := utils.GetStorageUInt64(native, GetTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[OntTotalSupply] get totalSupply error!")
	}
	return big.NewInt(int64(amount)).Bytes(), nil
}

func OntBalanceOf(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, TRANSFER_FLAG)
}

func OntAllowance(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, APPROVE_FLAG)
}

func GetBalanceValue(native *native.NativeService, flag byte) ([]byte, error) {
	var key []byte
	fromAddr, err := serialization.ReadVarBytes(bytes.NewBuffer(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] get address error!")
	}
	from, err := common.AddressParseFromBytes(fromAddr)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] address parse error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	if flag == APPROVE_FLAG {
		toAddr, err := serialization.ReadVarBytes(bytes.NewBuffer(native.Input))
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] get address error!")
		}
		to, err := common.AddressParseFromBytes(toAddr)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] address parse error!")
		}
		key = GetApproveKey(contract, from, to)
	} else if flag == TRANSFER_FLAG {
		key = GetTransferKey(contract, from)
	}
	amount, err := utils.GetStorageUInt64(native, key)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewDetailErr(err, errors.ErrNoCode, "[GetBalanceValue] address parse error!")
	}
	return big.NewInt(int64(amount)).Bytes(), nil
}

func grantOng(native *native.NativeService, contract, address common.Address, balance uint64, startHeight uint32) error {
	var amount uint32 = 0
	ustart := startHeight / DECREMENT_INTERVAL
	if ustart < GL {
		istart := startHeight % DECREMENT_INTERVAL
		uend := native.Height / DECREMENT_INTERVAL
		iend := native.Height % DECREMENT_INTERVAL
		if uend >= GL {
			uend = GL
			iend = 0
		}
		if iend == 0 {
			uend--
			iend = DECREMENT_INTERVAL
		}
		for {
			if ustart >= uend {
				break
			}
			amount += (DECREMENT_INTERVAL - istart) * GENERATION_AMOUNT[ustart]
			ustart++
			istart = 0
		}
		amount += (iend - istart) * GENERATION_AMOUNT[ustart]
	}

	args, err := getApproveArgs(native, contract, genesis.OngContractAddress, address, balance, uint64(amount))
	if err != nil {
		return err
	}

	if _, err := native.ContextRef.AppCall(genesis.OngContractAddress, "approve", []byte{}, args); err != nil {
		return err
	}

	native.CloneCache.Add(scommon.ST_STORAGE, getAddressHeightKey(contract, address), utils.GetUInt32StorageItem(native.Height))
	return nil
}

func getApproveArgs(native *native.NativeService, contract, ongContract, address common.Address, balance, amount uint64) ([]byte, error) {
	bf := new(bytes.Buffer)
	approve := &State{
		From:  contract,
		To:    address,
		Value: balance * amount,
	}

	stateValue, err := utils.GetStorageUInt64(native, GetApproveKey(ongContract, approve.From, approve.To))
	if err != nil {
		return nil, err
	}

	approve.Value += stateValue

	if err := approve.Serialize(bf); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}
