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
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/errors"
	"github.com/Ontology/core/genesis"
	ctypes "github.com/Ontology/core/types"
	"math/big"
	"github.com/Ontology/smartcontract/service/native/states"
	cstates "github.com/Ontology/core/states"
	"bytes"
	"github.com/Ontology/account"
	"github.com/Ontology/common"
)

var (
	decrementInterval = uint32(2000000)
	generationAmount = [17]uint32{80, 70, 60, 50, 40, 30, 20, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
	gl = uint32(len(generationAmount))
	ontTotalSupply = big.NewInt(1000000000)
)

func OntInit(native *NativeService) error {
	booKeepers := account.GetBookkeepers()

	contract := native.ContextRef.CurrentContext().ContractAddress
	amount, err := getStorageBigInt(native, getTotalSupplyKey(contract))
	if err != nil {
		return err
	}

	if amount != nil && amount.Sign() != 0 {
		return errors.NewErr("Init ont has been completed!")
	}

	ts := new(big.Int).Div(ontTotalSupply, big.NewInt(int64(len(booKeepers))))
	for _, v := range booKeepers {
		address := ctypes.AddressFromPubKey(v)
		native.CloneCache.Add(scommon.ST_STORAGE, append(contract[:], address[:]...), &cstates.StorageItem{Value: ts.Bytes()})
		native.CloneCache.Add(scommon.ST_STORAGE, getTotalSupplyKey(contract), &cstates.StorageItem{Value: ts.Bytes()})
		addNotifications(native, contract, &states.State{To: address, Value: ts})
	}

	return nil
}

func OntTransfer(native *NativeService) error {
	transfers := new(states.Transfers)
	if err := transfers.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Transfer] Transfers deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	for _, v := range transfers.States {
		fromBalance, toBalance, err := transfer(native, contract, v); if err != nil {
			return err
		}

		fromStartHeight, err := getStartHeight(native, contract, v.From); if err != nil {
			return err
		}

		toStartHeight, err := getStartHeight(native, contract, v.From); if err != nil {
			return err
		}

		if err := grantOng(native, contract, v.From, fromBalance, fromStartHeight); err != nil {
			return err
		}

		if err := grantOng(native, contract, v.To, toBalance, toStartHeight); err != nil {
			return err
		}

		addNotifications(native, contract, v)
	}
	return nil
}

func OntTransferFrom(native *NativeService) error {
	state := new(states.TransferFrom)
	if err := state.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[OntTransferFrom] State deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	if err := transferFrom(native, contract, state); err != nil {
		return err
	}
	addNotifications(native, contract, &states.State{From: state.From, To: state.To, Value: state.Value})
	return nil
}

func OntApprove(native *NativeService) error {
	state := new(states.State)
	if err := state.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[OngApprove] state deserialize error!")
	}
	if err := isApproveValid(native, state); err != nil {
		return err
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	native.CloneCache.Add(scommon.ST_STORAGE, getApproveKey(contract, state), &cstates.StorageItem{Value: state.Value.Bytes()})
	return nil
}

func grantOng(native *NativeService, contract, address common.Address, balance *big.Int, startHeight uint32) error {
	var amount uint32 = 0
	ustart := startHeight / decrementInterval
	if ustart < gl {
		istart := startHeight % decrementInterval
		uend := native.Height / decrementInterval
		iend := native.Height % decrementInterval
		if uend >= gl {
			uend = gl
			iend = 0
		}
		if iend == 0 {
			uend--
			iend = decrementInterval
		}
		for {
			if ustart >= uend {
				break
			}
			amount += (decrementInterval - istart) * generationAmount[ustart]
			ustart++
			istart = 0
		}
		amount += (iend - istart) * generationAmount[ustart]
	}

	args, err := getApproveArgs(native, contract, genesis.OngContractAddress, address, balance, amount); if err != nil {

		return err
	}

	if err := native.AppCall(genesis.OngContractAddress, "approve", args); err != nil {
		return err
	}

	native.CloneCache.Add(scommon.ST_STORAGE, getAddressHeightKey(contract, address), getHeightStorageItem(native.Height))
	return nil
}


func getApproveArgs(native *NativeService, contract, ongContract, address common.Address, balance *big.Int, amount uint32) ([]byte, error) {
	bf := new(bytes.Buffer)
	approve := &states.State {
		From: contract,
		To: address,
		Value: new(big.Int).Mul(balance, big.NewInt(int64(amount))),
	}


	stateValue, err := getStorageBigInt(native, getApproveKey(ongContract, approve)); if err != nil {
		return nil, err
	}

	approve.Value = new(big.Int).Add(approve.Value, stateValue)

	if err := approve.Serialize(bf); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

