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

package header_sync

import (
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	ccom "github.com/ontio/ontology/smartcontract/service/native/cross_chain/common"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	//function name
	SYNC_GENESIS_HEADER = "syncGenesisHeader"
	SYNC_BLOCK_HEADER   = "syncBlockHeader"

	//key prefix
	BLOCK_HEADER   = "blockHeader"
	CURRENT_HEIGHT = "currentHeight"
	HEADER_INDEX   = "headerIndex"
	CONSENSUS_PEER = "consensusPeer"
	KEY_HEIGHTS    = "keyHeights"
)

//Init governance contract address
func InitHeaderSync() {
	native.Contracts[utils.HeaderSyncContractAddress] = RegisterHeaderSyncContract
}

//Register methods of governance contract
func RegisterHeaderSyncContract(native *native.NativeService) {
	native.Register(SYNC_GENESIS_HEADER, SyncGenesisHeader)
	native.Register(SYNC_BLOCK_HEADER, SyncBlockHeader)
}

func SyncGenesisHeader(native *native.NativeService) ([]byte, error) {
	params := new(SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, contract params deserialize error: %v", err)
	}

	// get admin from database
	operatorAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, checkWitness error: %v", err)
	}

	header, err := ccom.HeaderFromRawBytes(params.GenesisHeader)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, deserialize header err: %v", err)
	}
	//block header storage
	err = PutBlockHeader(native, header, params.GenesisHeader)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, put blockHeader error: %v", err)
	}

	//consensus node pk storage
	err = UpdateConsensusPeer(native, header)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, update ConsensusPeer error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func SyncBlockHeader(native *native.NativeService) ([]byte, error) {
	params := new(SyncBlockHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}
	for _, v := range params.Headers {
		header, err := ccom.HeaderFromRawBytes(v)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, new_types.HeaderFromRawBytes error: %v", err)
		}
		h, err := GetHeaderByHeight(native, header.ChainID, header.Height)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, %d, %d", header.ChainID, header.Height)
		}
		if h == nil {
			if err := ProcessHeader(native, header, v); err != nil {
				return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, error:%s", err)
			}
		}
	}
	return utils.BYTE_TRUE, nil
}
