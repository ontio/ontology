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

package ethl2

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/scylladb/go-set/strset"
)

func InitETHL2() {
	native.Contracts[utils.ETHLayer2ContractAddress] = RegisterETHL2Contract
}

func RegisterETHL2Contract(native *native.NativeService) {
	native.Register(MethodPutName, Put)
	native.Register(MethodAppendAddress, AppendAuthedAddress)
	native.Register(MethodGetAddress, GetEthLayer2AuthAddress)
}

func Put(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	raw, err := utils.DecodeVarBytes(common.NewZeroCopySource(native.Input))
	if err != nil || len(raw) < 1 {
		return utils.BYTE_FALSE, err
	}

	ethtxType := raw[0]
	raweth := raw[1:]

	var s *State
	if ethtxType == EthEIP155Type {
		var tx types.Transaction
		txbin, err := hex.DecodeString(string(raweth))
		if err != nil {
			return utils.BYTE_FALSE, err
		}

		err = tx.UnmarshalBinary(txbin)
		if err != nil {
			return utils.BYTE_FALSE, err
		}
		chainID, err := GetEthLayer2ChainID(native)
		if err != nil {
			return utils.BYTE_FALSE, err
		}
		signer := types.NewEIP155Signer(big.NewInt(int64(chainID)))
		_, err = signer.Sender(&tx)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("eth eip 155 sign verify fail: %v", err)
		}

		s = &State{
			fName: MethodPutName,
			ethtx: raw,
		}

	} else {
		return utils.BYTE_FALSE, errors.New("work in progress")
	} // TODO EthSignedMessageType

	AddNotifications(native, contract, s)

	return utils.BYTE_TRUE, nil
}

func GetEthLayer2ChainID(native *native.NativeService) (uint64, error) {
	key := global_params.GenerateEthLayer2ChainIDKey(utils.ParamContractAddress)

	bin, err := native.CacheDB.Get(key)
	if err != nil {
		return 0, fmt.Errorf("eth layer2 chain id not found %v", err)
	}
	// in global param, we put value in little endian
	chainID := binary.LittleEndian.Uint64(bin)

	return chainID, nil
}

func AppendAuthedAddress(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	operator, err := global_params.GetStorageRole(native, global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil || operator == common.ADDRESS_EMPTY {
		return utils.BYTE_FALSE, fmt.Errorf("create snapshot, operator doesn't exist, caused by %v", err)
	}
	if !native.ContextRef.CheckWitness(operator) {
		return utils.BYTE_FALSE, errors.New("need global params admin to add address to this set, you have no permission to do this")
	}

	raw, err := utils.DecodeVarBytes(common.NewZeroCopySource(native.Input))
	if err != nil || len(raw) < 1 {
		return utils.BYTE_FALSE, err
	}

	ap := global_params.AddressParam{}
	authSet := strset.New(operator.ToHexString())

	// first read existed auth set
	b, err := native.CacheDB.Get(GetAppendAutAddressKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if len(b) > 0 {
		ap.Deserialization(common.NewZeroCopySource(b))
	}
	for _, addr := range ap.Contracts {
		authSet.Add(addr.ToHexString())
	}
	// append args

	if err := ap.Deserialization(common.NewZeroCopySource(raw)); err != nil {
		return utils.BYTE_FALSE, err
	}
	// contract is addr as well,
	for _, addr := range ap.Contracts {
		authSet.Add(addr.ToHexString())
	}

	ap.Contracts = make([]common.Address, 0, authSet.Size())
	for _, addrstr := range authSet.List() {
		addr, err := common.AddressFromHexString(addrstr)
		if err != nil || addr == common.ADDRESS_EMPTY {
			continue
		}
		ap.Contracts = append(ap.Contracts, addr)
	}

	sink := common.NewZeroCopySink(nil)
	ap.Serialization(sink)
	native.CacheDB.Put(GetAppendAutAddressKey(contract), sink.Bytes())

	AddAppendAddressNotification(native, contract, ap.Contracts)

	return utils.BYTE_TRUE, nil
}

func GetEthLayer2AuthAddress(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	b, err := native.CacheDB.Get(GetAppendAutAddressKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	return b, nil
}
