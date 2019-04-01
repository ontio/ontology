/*
 * Copyright (C) 2019 The ontology Authors
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

package shardccmc

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/shardccmc/states"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	// function names
	INIT_NAME        = "init"
	CC_REGISTER_NAME = "register"
	CC_FREEZE_NAME   = "freeze"
	CC_MIGRATE_NAME  = "migrate"

	// Key prefix
	KEY_VERSION     = "version"
	KEY_CCMC_STATE  = "ccmcState"
	KEY_CC_INFO     = "ccInfo"     // index CC with CCID
	KEY_CC_CONTRACT = "ccContract" // index CC with contract-addr

	INIT_CCID = 100
)

var ShardCCMCVersion = shardmgmt.VERSION_CONTRACT_SHARD_MGMT

func InitShardCCMC() {
	native.Contracts[utils.ShardCCMCAddress] = RegisterShardCCMC
}

func RegisterShardCCMC(native *native.NativeService) {
	native.Register(INIT_NAME, ShardCCMCInit)
	native.Register(CC_REGISTER_NAME, ShardCCMCRegister)
}

func ShardCCMCInit(native *native.NativeService) ([]byte, error) {
	// check if admin
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	if err := utils.ValidateOwner(native, adminAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("init shard gas, checkWitness error: %v", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress

	// check if shard-mgmt initialized
	ver, err := getVersion(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("init shard gas, get version: %s", err)
	}
	if ver == 0 {
		// initialize shardmgmt version
		if err := setVersion(native, contract); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("init shard gas version: %s", err)
		}

		ccmcState := &ccmc_states.ShardCCMCState{NextCCID: INIT_CCID}
		setCCMCState(native, contract, ccmcState)
		return utils.BYTE_TRUE, nil
	}

	if ver < ShardCCMCVersion {
		// make upgrade
		return utils.BYTE_FALSE, fmt.Errorf("upgrade TBD")
	} else if ver > ShardCCMCVersion {
		return utils.BYTE_FALSE, fmt.Errorf("version downgrade from %d to %d", ver, ShardCCMCVersion)
	}

	return utils.BYTE_TRUE, nil
}

func ShardCCMCRegister(native *native.NativeService) ([]byte, error) {
	cp := new(shardmgmt.CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("register cc, invalid cmd param: %s", err)
	}

	params := RegisterCCParam{}
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("register cc, invalid input param: %s", err)
	}
	if err := utils.ValidateOwner(native, params.Owner); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("register cc, invalid creator: %s", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	ccid, err := getCCID(native, contract, params.ContractAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("register cc, getCCID: %s", err)
	}
	if ccid != 0 {
		return utils.BYTE_FALSE, fmt.Errorf("register cc, contract registered with ID: %s", ccid)
	}

	ccmc, err := getCCMCState(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("register cc, failed to load ccmc: %s", err)
	}

	// TODO: validate shardID
	// TODO: verify dependencies
	// TODO: transfer registration fee

	ccInfo := &ccmc_states.ShardCCInfo{
		CCID:         ccmc.NextCCID,
		ShardID:      params.ShardID,
		Owner:        params.Owner,
		ContractAddr: params.ContractAddr,
		Dependencies: params.Dependencies,
	}
	ccmc.NextCCID += 1

	setCCMCState(native, contract, ccmc)
	if err := setCCInfo(native, contract, ccInfo); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("register cc, failed to save ccinfo: %s", err)
	}

	return utils.BYTE_FALSE, nil
}
