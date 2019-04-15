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
	"fmt"
	"sort"
	"strings"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/shard_stake"
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
	EXIT_SHARD_NAME         = "exitShard"
	ACTIVATE_SHARD_NAME     = "activateShard"
	COMMIT_DPOS_NAME        = "commitDpos"

	// TODO: child shard commit dpos
	SHARD_COMMIT_DPOS = "shardCommitDpos"
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
	native.Register(COMMIT_DPOS_NAME, CommitDpos)
	native.Register(EXIT_SHARD_NAME, ExitShard)
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
		setGlobalState(native, contract, globalState)

		// initialize shard states
		shardState := &shardstates.ShardState{
			ShardID:             native.ShardID,
			GenesisParentHeight: native.Height,
			State:               shardstates.SHARD_STATE_ACTIVE,
			Config:              &shardstates.ShardConfig{VbftCfg: &config.VBFTConfig{}},
			Peers:               make(map[string]*shardstates.PeerShardStakeInfo),
		}
		setShardState(native, contract, shardState)
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
	params := new(CreateShardParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
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
		Config:  &shardstates.ShardConfig{VbftCfg: &config.VBFTConfig{}},
		Peers:   make(map[string]*shardstates.PeerShardStakeInfo),
	}
	globalState.NextSubShardIndex += 1

	// update global state
	setGlobalState(native, contract, globalState)
	// save shard
	setShardState(native, contract, shard)

	// transfer create shard fee to root chain governance contract
	err = ont.AppCallTransfer(native, utils.OngContractAddress, params.Creator, utils.GovernanceContractAddress,
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
	params := new(ConfigShardParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("config shard, invalid param: %s", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	if ok, err := checkVersion(native, contract); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ConfigShard: check version: %s", err)
	}

	shard, err := GetShardState(native, contract, params.ShardID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ConfigShard: get shard: %s", err)
	}

	if err := utils.ValidateOwner(native, shard.Creator); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ConfigShard: invalid configurator: %s", err)
	}
	if shard.ShardID.ParentID() != native.ShardID {
		return utils.BYTE_FALSE, fmt.Errorf("ConfigShard: not on parent shard")
	}

	if params.NetworkMin < 1 {
		return utils.BYTE_FALSE, fmt.Errorf("ConfigShard: invalid shard network size")
	}

	// TODO: reset default values
	if params.GasPrice == 0 && params.GasLimit == 0 {
		params.GasPrice = 500
		params.GasLimit = 200000
	}

	// TODO: support other stake
	if params.StakeAssetAddress.ToHexString() != utils.OntContractAddress.ToHexString() {
		return utils.BYTE_FALSE, fmt.Errorf("ConfigShard: only support ONT staking")
	}
	if params.GasAssetAddress.ToHexString() != utils.OngContractAddress.ToHexString() {
		return utils.BYTE_FALSE, fmt.Errorf("ConfigShard: only support ONG gas")
	}

	// TODO: validate input config
	shard.Config = &shardstates.ShardConfig{
		NetworkSize:       params.NetworkMin,
		StakeAssetAddress: params.StakeAssetAddress,
		GasAssetAddress:   params.GasAssetAddress,
		GasPrice:          params.GasPrice,
		GasLimit:          params.GasLimit,
	}
	cfg, err := params.GetConfig()
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ConfigShard: decode config failed, err: %s", err)
	}
	shard.Config.VbftCfg = cfg
	shard.State = shardstates.SHARD_STATE_CONFIGURED

	if err := initStakeContractShard(native, params.ShardID, uint64(cfg.MinInitStake), params.StakeAssetAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CreateShard: failed, err: %s", err)
	}

	setShardState(native, contract, shard)

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
	if err := utils.ValidateOwner(native, params.PeerOwner); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApplyJoinShard: check witness faield, err: %s", err)
	}
	// verify peer is exist in root chain consensus
	if config.DefConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_VBFT {
		if _, err := getRootCurrentViewPeerItem(native, params.PeerPubKey); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ApplyJoinShard: failed, err: %s", err)
		}
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	state, err := getShardPeerState(native, contract, params.ShardId, params.PeerPubKey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApplyJoinShard: faile, err: %s", err)
	}
	if state != state_default {
		return utils.BYTE_FALSE, fmt.Errorf("ApplyJoinShard: peer %s hasn't applied", params.PeerPubKey)
	}
	setShardPeerState(native, contract, params.ShardId, state_applied, params.PeerPubKey)
	return utils.BYTE_TRUE, nil
}

func ApproveJoinShard(native *native.NativeService) ([]byte, error) {
	params := new(ApproveJoinShardParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveJoinShard: invalid param: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	shard, err := GetShardState(native, contract, params.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveJoinShard: cannot get shard %d, err: %s", params.ShardId, err)
	}
	if err := utils.ValidateOwner(native, shard.Creator); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveJoinShard: check witness failed, err: %s", err)
	}
	for _, pubKey := range params.PeerPubKey {
		state, err := getShardPeerState(native, contract, params.ShardId, pubKey)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ApproveJoinShard: faile, err: %s", err)
		}
		if state != state_applied {
			return utils.BYTE_FALSE, fmt.Errorf("ApproveJoinShard: peer %s hasn't applied", pubKey)
		}
		setShardPeerState(native, contract, params.ShardId, state_approved, pubKey)
	}
	return utils.BYTE_TRUE, nil
}

func JoinShard(native *native.NativeService) ([]byte, error) {
	params := new(JoinShardParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
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
	setShardPeerState(native, contract, params.ShardID, state_joined, params.PeerPubKey)
	if config.DefConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_VBFT {
		rootChainPeerItem, err := getRootCurrentViewPeerItem(native, params.PeerPubKey)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("JoinShard: failed, err: %s", err)
		}
		if rootChainPeerItem.InitPos < params.StakeAmount {
			return utils.BYTE_FALSE, fmt.Errorf("JoinShard: shard stake amount should less than root chain")
		}
	}

	if _, present := shard.Peers[strings.ToLower(params.PeerPubKey)]; present {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: peer already in shard")
	} else {
		if shard.Peers == nil {
			shard.Peers = make(map[string]*shardstates.PeerShardStakeInfo)
		}
		peerStakeInfo := &shardstates.PeerShardStakeInfo{
			IpAddress:  params.IpAddress,
			PeerOwner:  params.PeerOwner,
			PeerPubKey: params.PeerPubKey,
		}
		shard.Peers[strings.ToLower(params.PeerPubKey)] = peerStakeInfo
		if shard.Config.VbftCfg.Peers == nil {
			shard.Config.VbftCfg.Peers = make([]*config.VBFTPeerStakeInfo, 0)
		}
		vbftPeerInfo := &config.VBFTPeerStakeInfo{
			PeerPubkey: strings.ToLower(params.PeerPubKey),
			Address:    params.PeerOwner.ToBase58(),
			InitPos:    params.StakeAmount,
		}
		shard.Config.VbftCfg.Peers = append(shard.Config.VbftCfg.Peers, vbftPeerInfo)
	}

	setShardState(native, contract, shard)

	// call shard stake contract
	if err := peerInitStake(native, params); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("JoinShard: failed, err: %s", err)
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

func ExitShard(native *native.NativeService) ([]byte, error) {
	param := new(ExitShardParam)
	if err := param.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ExitShard: invalid param: %s", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	if ok, err := checkVersion(native, contract); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ExitShard: check version: %s", err)
	}
	if err := utils.ValidateOwner(native, param.PeerOwner); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ExitShard: check witness failed, err: %s", err)
	}
	shard, err := GetShardState(native, contract, param.ShardId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ExitShard: get shard state failed, err: %s", err)
	}
	shardPeerInfo, ok := shard.Peers[strings.ToLower(param.PeerPubKey)]
	if !ok {
		return utils.BYTE_FALSE, fmt.Errorf("ExitShard: peer not exist in shard, err: %s", err)
	}
	if shardPeerInfo.PeerOwner != param.PeerOwner {
		return utils.BYTE_FALSE, fmt.Errorf("ExitShard: peer owner unmatch")
	}
	if err := peerExit(native, param.ShardId, param.PeerPubKey); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ExitShard: failed, err: %s", err)
	}
	if shardPeerInfo.NodeType == shardstates.CONSENSUS_NODE {
		shardPeerInfo.NodeType = shardstates.QUIT_CONSENSUS_NODE
	} else if shardPeerInfo.NodeType == shardstates.CONDIDATE_NODE {
		shardPeerInfo.NodeType = shardstates.QUITING_CONSENSUS_NODE
	} else {
		return utils.BYTE_FALSE, fmt.Errorf("ExitShard: peer has already exit")
	}
	shard.Peers[strings.ToLower(param.PeerPubKey)] = shardPeerInfo
	setShardState(native, contract, shard)

	return utils.BYTE_TRUE, nil
}

func ActivateShard(native *native.NativeService) ([]byte, error) {
	params := new(ActivateShardParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
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
	peers := shard.Config.VbftCfg.Peers
	sort.SliceStable(peers, func(i, j int) bool {
		return peers[i].InitPos > peers[j].InitPos
	})
	for index, peer := range peers {
		peer.Index = uint32(index) + 1
		shardPeer, ok := shard.Peers[peer.PeerPubkey]
		if !ok {
			return utils.BYTE_FALSE, fmt.Errorf("activate shard, unmatch peer pub key %s", peer.PeerPubkey)
		}
		shardPeer.Index = uint32(index) + 1
		if uint32(index) < shard.Config.VbftCfg.K {
			shardPeer.NodeType = shardstates.CONSENSUS_NODE
		} else {
			shardPeer.NodeType = shardstates.CONDIDATE_NODE
		}
		shard.Peers[peer.PeerPubkey] = shardPeer
	}
	shard.Config.VbftCfg.Peers = peers
	shard.GenesisParentHeight = native.Height
	shard.State = shardstates.SHARD_STATE_ACTIVE
	setShardState(native, contract, shard)

	evt := &shardstates.ShardActiveEvent{Height: native.Height}
	evt.SourceShardID = native.ShardID
	evt.ShardID = shard.ShardID
	if err := AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("activae shard, add notification: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func CommitDpos(native *native.NativeService) ([]byte, error) {
	if !native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: only can be invoked at root shard")
	}
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardGasMgmtContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: only can be invoked by shard gas contract")
	}
	params := new(CommitDposParam)
	if err := params.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: invalid param: %s", err)
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	if ok, err := checkVersion(native, contract); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: check version: %s", err)
	}
	shard, err := GetShardState(native, contract, params.ShardID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: get shard: %s", err)
	}
	quitPeers := make([]string, 0)
	// check peer exit shard
	for peer, info := range shard.Peers {
		if info.NodeType == shardstates.QUIT_CONSENSUS_NODE {
			info.NodeType = shardstates.QUITING_CONSENSUS_NODE
			shard.Peers[peer] = info
		} else if info.NodeType == shardstates.QUITING_CONSENSUS_NODE {
			// delete peer at mgmt contract
			delete(shard.Peers, peer)
			quitPeers = append(quitPeers, peer)
		}
	}
	// delete peers at stake contract
	if len(quitPeers) > 0 {
		if err = deletePeer(native, params.ShardID, quitPeers); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
		}
	}
	// TODO: update shard mgmt peer state
	shardView, err := shard_stake.GetShardCurrentView(native, params.ShardID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
	}
	viewInfo, err := shard_stake.GetShardViewInfo(native, params.ShardID, shardView)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
	}
	wholeNodeStakeAmount := uint64(0)
	for _, info := range viewInfo.Peers {
		wholeNodeStakeAmount += info.UserStakeAmount + info.InitPos
	}
	// TODO: check viewInfo.Peers is existed in shard states
	// TODO: candidate node and consensus node different rate
	feeInfo := make([]*shard_stake.PeerAmount, 0)
	for peer, info := range viewInfo.Peers {
		peerFee := (info.UserStakeAmount + info.InitPos) * params.FeeAmount / wholeNodeStakeAmount
		feeInfo = append(feeInfo, &shard_stake.PeerAmount{PeerPubKey: peer, Amount: peerFee})
	}
	if err := commitDpos(native, params.ShardID, feeInfo); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("CommitDpos: failed, err: %s", err)
	}
	setShardState(native, contract, shard)
	return utils.BYTE_TRUE, nil
}
