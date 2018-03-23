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
	"github.com/Ontology/smartcontract/event"
	"github.com/Ontology/smartcontract/service/native/states"
	cstates "github.com/Ontology/core/states"
	"github.com/Ontology/common"
	"bytes"
	"github.com/Ontology/account"
)

var (
	totalSupplyName = []byte("totalSupply")
	decimals = big.NewInt(8)
	totalSupply = new(big.Int).Mul(big.NewInt(1000000000), (new(big.Int).Exp(big.NewInt(10), decimals, nil)))
)

func OntInit(native *NativeService) (bool, error) {
	booKeepers := account.GetBookkeepers()

	amount, err := getBalance(native, getOntTotalSupplyKey())
	if err != nil {
		return false, err
	}

	if amount != nil && amount.Sign() != 0 {
		return false, errors.NewErr("Init ont has been completed!")
	}

	ts := new(big.Int).Div(totalSupply, big.NewInt(int64(len(booKeepers))))
	for _, v := range booKeepers {
		address := ctypes.AddressFromPubKey(v)
		native.CloneCache.Add(scommon.ST_Storage, append(getOntContext(), address.ToArray()...), &cstates.StorageItem{Value: ts.Bytes()})
		native.CloneCache.Add(scommon.ST_Storage, getOntTotalSupplyKey(), &cstates.StorageItem{Value: ts.Bytes()})
		native.Notifications = append(native.Notifications, &event.NotifyEventInfo{
			Container: native.Tx.Hash(),
			CodeHash: genesis.OntContractAddress,
			States: []interface{}{nil, address, ts},
		})
	}

	return true, nil
}

func Transfer(native *NativeService) (bool, error) {
	transfers := new(states.Transfers)
	if err := transfers.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[Transfer] Transfers deserialize error!")
	}
	for _, p := range transfers.Params {
		for _, s := range p.States {
			if s.Value.Cmp(big.NewInt(0)) < 0 {
				return false, errors.NewErr("[Transfer] transfer amount invalid!")
			}

			if s.From == s.To {
				return true, nil
			}

			if !checkWitness(native, s.From) {
				return false, errors.NewErr("[Transfer] Authentication failed!")
			}

			fromKey := append(p.Contract.ToArray(), s.From.ToArray()...)
			fromBalance, err := getBalance(native, fromKey); if err != nil {
				return false, err
			}
			balance := fromBalance.Cmp(s.Value)
			if balance < 0 {
				return false, errors.NewErr("[Transfer] balance insufficient!")
			}
			if balance == 0 {
				native.CloneCache.Delete(scommon.ST_Storage, fromKey)
			} else {
				native.CloneCache.Add(scommon.ST_Storage, fromKey, getFromAmountStorageItem(fromBalance, s.Value))
			}

			toKey := append(p.Contract.ToArray(), s.To.ToArray()...)
			toBalance, err := getBalance(native, toKey); if err != nil {
				return false, err
			}

			native.CloneCache.Add(scommon.ST_Storage, toKey, getToAmountStorageItem(toBalance, s.Value))
			native.Notifications = append(native.Notifications, &event.NotifyEventInfo{
				Container: native.Tx.Hash(),
				CodeHash: p.Contract,
				States: []interface{}{s.From, s.To, s.Value},
			})
		}
	}
	return true, nil
}

func getBalance(native *NativeService, key []byte) (*big.Int, error) {
	balance, err := native.CloneCache.Get(scommon.ST_Storage, key)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[getBalance] storage error!")
	}
	if balance == nil {
		return big.NewInt(0), nil
	}
	item, ok := balance.(*cstates.StorageItem); if !ok {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[getBalance] get amount error!")
	}
	return new(big.Int).SetBytes(item.Value), nil
}

func checkWitness(native *NativeService, u160 common.Address) bool {
	addresses := native.Tx.GetSignatureAddresses()
	for _, v := range addresses {
		if v == u160 {
			return true
		}
	}
	return false
}

func getOntContext() []byte {
	return genesis.OntContractAddress.ToArray()
}

func getOngContext() []byte {
	return genesis.OngContractAddress.ToArray()
}

func getOntTotalSupplyKey() []byte {
	return append(getOntContext(), totalSupplyName...)
}

func getOngTotalSupplyKey() []byte {
	return append(getOngContext(), totalSupplyName...)
}

func getFromAmountStorageItem(fromBalance, value *big.Int) *cstates.StorageItem {
	return &cstates.StorageItem{Value: new(big.Int).Sub(fromBalance, value).Bytes()}
}

func getToAmountStorageItem(toBalance, value *big.Int) *cstates.StorageItem {
	return &cstates.StorageItem{Value: new(big.Int).Add(toBalance, value).Bytes()}
}

