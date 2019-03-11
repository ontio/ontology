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

package shardmgmt

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/smartcontract/service/native/ont"

	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

/////////
//
// Shard management contract
//
//	. create shard
//	. config shard
//	. join shard
//	. activate shard
//
/////////

const (
	VERSION_CONTRACT_SHARD_MGMT = uint32(1)

	// function names
	INIT_NAME               = "init"
	CREATE_SHARD_NAME       = "createShard"
	CONFIG_SHARD_NAME       = "configShard"
	APPLY_JOIN_SHARD_NAME   = "applyJoinShard"
	APPROVE_JOIN_SHARD_NAME = "approveJoinShard"
	JOIN_SHARD_NAME         = "joinShard"
	ACTIVATE_SHARD_NAME     = "activateShard"
)

func InitShardManagement() {
	native.Contracts[utils.ShardMgmtContractAddress] = RegisterShardMgmtContract
}

func RegisterShardMgmtContract(native *native.NativeService) {
	native.Register(INIT_NAME, ShardMgmtInit)
	native.Register(CREATE_SHARD_NAME, CreateShard)
	native.Register(CONFIG_SHARD_NAME, ConfigShard)
	native.Register(APPLY_JOIN_SHARD_NAME, ApplyJoinShard)
	native.Register(APPROVE_JOIN_SHARD_NAME, ApproveJoinShard)
	native.Register(JOIN_SHARD_NAME, JoinShard)
	native.Register(ACTIVATE_SHARD_NAME, ActivateShard)

	registerShardGov(native)
}

func ShardMgmtInit(native *native.NativeService) ([]byte, error) {
	// check if admin
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	if err := utils.ValidateOwner(native, adminAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("init shard mgmt, checkWitness error: %v", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress

	// check if shard-mgmt initialized
	ver, err := getVersion(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("init shard mgmt, get version: %s", err)
	}
	if ver == 0 {
		// initialize shardmgmt version
		if err := setVersion(native, contract); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("init shard mgmt version: %s", err)
		}

		// initialize shard mgmt
		globalState := &shardstates.ShardMgmtGlobalState{NextSubShardIndex: 1}
		if err := setGlobalState(native, contract, globalState); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("init shard mgmt global state: %s", err)
		}

		// initialize shard states
		shardState := &shardstates.ShardState{
			ShardID:             native.ShardID,
			GenesisParentHeight: native.Height,
			State:               shardstates.SHARD_STATE_ACTIVE,
		}
		if err := setShardState(native, contract, shardState); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("init shard mgmt main shard state: %s", err)
		}
		return utils.BYTE_TRUE, nil
	}

	if ver < VERSION_CONTRACT_SHARD_MGMT {
		// make upgrade
		return utils.BYTE_FALSE, fmt.Errorf("upgrade TBD")
	} else if ver > VERSION_CONTRACT_SHARD_MGMT {
		return utils.BYTE_FALSE, fmt.Errorf("version downgrade from %d to %d", ver, VERSION_CONTRACT_SHARD_MGMT)
	}

	return utils.BYTE_TRUE, nil
}

func CreateShard(native *native.NativeService) ([]byte, error) {
	cp := new(CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("config shard, invalid cmd param: %s", err)
	}

	params := new(CreateShardParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("create shard, invalid param: %s", err)
	}
	if params.ParentShardID.ToUint64() != 0 {
		return utils.BYTE_FALSE, fmt.Errorf("create shard, invalid parent shard: %d", params.ParentShardID)
	}
	if params.ParentShardID != native.ShardID {
		return utils.BYTE_FALSE, fmt.Errorf("CreateShard: parent ShardID is not current shard")
	}

	if err := utils.ValidateOwner(native, params.Creator); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateShard: invalid creator: %s", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	if ok, err := checkVersion(native, contract); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateShard: check version: %s", err)
	}

	globalState, err := getGlobalState(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateShard: get global state: %s", err)
	}

	subShardID, err := native.ShardID.GenSubShardID(globalState.NextSubShardIndex)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	shard := &shardstates.ShardState{
		ShardID: subShardID,
		Creator: params.Creator,
		State:   shardstates.SHARD_STATE_CREATED,
		Peers:   make(map[keypair.PublicKey]*shardstates.PeerShardStakeInfo),
	}
	globalState.NextSubShardIndex += 1

	// update global state
	if err := setGlobalState(native, contract, globalState); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateShard: update global state: %s", err)
	}
	// save shard
	if err := setShardState(native, contract, shard); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateShard: set shard state: %s", err)
	}

	// transfer create shard fee to root chain governance contract
	err = ont.AppCallTransfer(native, utils.OntContractAddress, params.Creator, utils.GovernanceContractAddress,
		shardstates.SHARD_CREATE_FEE)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateShard: recharge create shard fee failed, err: %s", err)
	}

	evt := &shardstates.CreateShardEvent{
		SourceShardID: native.ShardID,
		Height:        native.Height,
		NewShardID:    shard.ShardID,
	}
	if err := AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateShard: add notification: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func ConfigShard(native *native.NativeService) ([]byte, error) {
	cp := new(CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("config shard, invalid cmd param: %s", err)
	}
	params := new(ConfigShardParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("config shard, invalid param: %s", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	if ok, err := checkVersion(native, contract); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("config shard, check version: %s", err)
	}

	shard, err := GetShardState(native, contract, params.ShardID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("config shard, get shard: %s", err)
	}

	if err := utils.ValidateOwner(native, shard.Creator); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("config shard, invalid configurator: %s", err)
	}
	if shard.ShardID.ParentID() != native.ShardID {
		return utils.BYTE_FALSE, fmt.Errorf("config shard, not on parent shard")
	}

	if params.NetworkMin < 1 {
		return utils.BYTE_FALSE, fmt.Errorf("config shard, invalid shard network size")
	}

	// TODO: support other stake
	if params.StakeAssetAddress.ToHexString() != utils.OntContractAddress.ToHexString() {
		return utils.BYTE_FALSE, fmt.Errorf("config shard, only support ONT staking")
	}
	if params.GasAssetAddress.ToHexString() != utils.OngContractAddress.ToHexString() {
		return utils.BYTE_FALSE, fmt.Errorf("config shard, only support ONG gas")
	}

	// TODO: validate input config
	shard.Config = &shardstates.ShardConfig{
		NetworkSize:       params.NetworkMin,
		StakeAssetAddress: params.StakeAssetAddress,
		GasAssetAddress:   params.GasAssetAddress,
		VbftConfigData:    params.VbftConfigData,
	}
	shard.State = shardstates.SHARD_STATE_CONFIGURED

	if err := setShardState(native, contract, shard); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("config shard, update shard state: %s", err)
	}

	evt := &shardstates.ConfigShardEvent{
		Height: native.Height,
		Config: shard.Config,
	}
	evt.SourceShardID = native.ShardID
	evt.ShardID = native.ShardID
	if err := AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ConfigShard: add notification: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func ApplyJoinShard(native *native.NativeService) ([]byte, error) {
	params := new(ApplyJoinShardParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApplyJoinShard: invalid param: %s", err)
	}
	// verify peer is exist in root chain consensus
	if _, err := getRootCurrentViewPeerItem(native, params.PeerPubKey); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApplyJoinShard: failed, err: %s", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	err := setShardPeerState(native, contract, params.ShardId, state_applied, params.PeerPubKey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApplyJoinShard: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func ApproveJoinShard(native *native.NativeService) ([]byte, error) {
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveJoinShard: get admin failed, err: %s", err)
	}
	if err := utils.ValidateOwner(native, adminAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveJoinShard: invalid configurator: %s", err)
	}

	params := new(ApproveJoinShardParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveJoinShard: invalid param: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	for _, pubKey := range params.PeerPubKey {
		state, err := getShardPeerState(native, contract, params.ShardId, pubKey)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ApproveJoinShard: faile, err: %s", err)
		}
		if state != state_applied {
			return utils.BYTE_FALSE, fmt.Errorf("ApproveJoinShard: peer %s hasn't applied", pubKey)
		}
		err = setShardPeerState(native, contract, params.ShardId, state_approved, pubKey)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ApproveJoinShard: update peer %s state faield, err: %s", pubKey, err)
		}
	}
	return utils.BYTE_TRUE, nil
}

func JoinShard(native *native.NativeService) ([]byte, error) {
	cp := new(CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("join shard, invalid cmd param: %s", err)
	}
	params := new(JoinShardParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("join shard, invalid param: %s", err)
	}

	if err := utils.ValidateOwner(native, params.PeerOwner); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: invalid peer owner: %s", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	if ok, err := checkVersion(native, contract); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: check version: %s", err)
	}

	shard, err := GetShardState(native, contract, params.ShardID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: get shard: %s", err)
	}
	if shard.ShardID.ParentID() != native.ShardID {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: not on parent shard")
	}

	state, err := getShardPeerState(native, contract, params.ShardID, params.PeerPubKey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: failed, err: %s", err)
	}
	if state != state_approved {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: peer state %s unmatch", state)
	}
	err = setShardPeerState(native, contract, params.ShardID, state_joined, params.PeerPubKey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: failed, err: %s", err)
	}
	rootChainPeerItem, err := getRootCurrentViewPeerItem(native, params.PeerPubKey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: failed, err: %s", err)
	}
	if rootChainPeerItem.TotalPos < params.StakeAmount {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: shard stake amount should less than root chain")
	}

	pubKeyData, err := hex.DecodeString(params.PeerPubKey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: decode param pub key failed, err: %s", err)
	}
	paramPubkey, err := keypair.DeserializePublicKey(pubKeyData)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: deserialize param pub key failed, err: %s", err)
	}
	if _, present := shard.Peers[paramPubkey]; present {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: peer already in shard")
	} else {
		peerStakeInfo := &shardstates.PeerShardStakeInfo{
			PeerOwner:        params.PeerOwner,
			PeerPubKey:       params.PeerPubKey,
			StakeAmount:      params.StakeAmount,
			MaxAuthorization: 0,
		}
		if shard.Peers == nil {
			shard.Peers = make(map[keypair.PublicKey]*shardstates.PeerShardStakeInfo)
		}
		peerStakeInfo.Index = uint32(len(shard.Peers) + 1)
		shard.Peers[paramPubkey] = peerStakeInfo
	}

	if err := setShardState(native, contract, shard); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: update shard state: %s", err)
	}

	// transfer stake asset
	err = ont.AppCallTransfer(native, shard.Config.StakeAssetAddress, params.PeerOwner, contract, params.StakeAmount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: transfer stake asset failed, err: %s", err)
	}

	evt := &shardstates.PeerJoinShardEvent{
		Height:     native.Height,
		PeerPubKey: params.PeerPubKey,
	}
	evt.SourceShardID = native.ShardID
	evt.ShardID = native.ShardID
	if err := AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: add notification: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func ActivateShard(native *native.NativeService) ([]byte, error) {
	cp := new(CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("activate shard, invalid cmd param: %s", err)
	}
	params := new(ActivateShardParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("activate shard, invalid param: %s", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	if ok, err := checkVersion(native, contract); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("activate shard, check version: %s", err)
	}

	shard, err := GetShardState(native, contract, params.ShardID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("activate shard, get shard: %s", err)
	}

	if err := utils.ValidateOwner(native, shard.Creator); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("activate shard, invalid configurator: %s", err)
	}
	if shard.State != shardstates.SHARD_STATE_CONFIGURED {
		return utils.BYTE_FALSE, fmt.Errorf("activate shard, invalid shard state: %d", shard.State)
	}
	if shard.ShardID.ParentID() != native.ShardID {
		return utils.BYTE_FALSE, fmt.Errorf("activate shard, not on parent shard")
	}

	// TODO: validate input config
	if uint32(len(shard.Peers)) < shard.Config.NetworkSize {
		return utils.BYTE_FALSE, fmt.Errorf("activae shard, not enough peer: %d vs %d",
			len(shard.Peers), shard.Config.NetworkSize)
	}

	shard.GenesisParentHeight = native.Height
	shard.State = shardstates.SHARD_STATE_ACTIVE
	if err := setShardState(native, contract, shard); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("activae shard, update shard state: %s", err)
	}

	evt := &shardstates.ShardActiveEvent{Height: native.Height}
	evt.SourceShardID = native.ShardID
	evt.ShardID = shard.ShardID
	if err := AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("activae shard, add notification: %s", err)
	}

	return utils.BYTE_TRUE, nil
}
