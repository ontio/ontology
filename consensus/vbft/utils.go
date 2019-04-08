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

package vbft

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/vrf"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/overlaydb"
	gov "github.com/ontio/ontology/smartcontract/service/native/governance"
	nutils "github.com/ontio/ontology/smartcontract/service/native/utils"

	shardmgt "github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
)

func SignMsg(account *account.Account, msg ConsensusMsg) ([]byte, error) {

	data, err := msg.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal msg when signing: %s", err)
	}

	return signature.Sign(account, data)
}

func hashData(data []byte) common.Uint256 {
	t := sha256.Sum256(data)
	f := sha256.Sum256(t[:])
	return common.Uint256(f)
}

func HashMsg(msg ConsensusMsg) (common.Uint256, error) {

	// FIXME: has to do marshal on each call

	data, err := SerializeVbftMsg(msg)
	if err != nil {
		return common.Uint256{}, fmt.Errorf("failed to marshal block: %s", err)
	}

	return hashData(data), nil
}

type seedData struct {
	BlockNum          uint32         `json:"block_num"`
	PrevBlockProposer uint32         `json:"prev_block_proposer"` // TODO: change to NodeID
	BlockRoot         common.Uint256 `json:"block_root"`
	VrfValue          []byte         `json:"vrf_value"`
}

func getParticipantSelectionSeed(block *Block) vconfig.VRFValue {

	data, err := json.Marshal(&seedData{
		BlockNum:          block.getBlockNum() + 1,
		PrevBlockProposer: block.getProposer(),
		BlockRoot:         block.Block.Header.BlockRoot,
		VrfValue:          block.getVrfValue(),
	})
	if err != nil {
		return vconfig.VRFValue{}
	}

	t := sha512.Sum512(data)
	f := sha512.Sum512(t[:])
	return vconfig.VRFValue(f)
}

type vrfData struct {
	BlockNum uint32 `json:"block_num"`
	PrevVrf  []byte `json:"prev_vrf"`
}

func computeVrf(sk keypair.PrivateKey, blkNum uint32, prevVrf []byte) ([]byte, []byte, error) {
	data, err := json.Marshal(&vrfData{
		BlockNum: blkNum,
		PrevVrf:  prevVrf,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("computeVrf failed to marshal vrfData: %s", err)
	}

	return vrf.Vrf(sk, data)
}

func verifyVrf(pk keypair.PublicKey, blkNum uint32, prevVrf, newVrf, proof []byte) error {
	data, err := json.Marshal(&vrfData{
		BlockNum: blkNum,
		PrevVrf:  prevVrf,
	})
	if err != nil {
		return fmt.Errorf("verifyVrf failed to marshal vrfData: %s", err)
	}

	result, err := vrf.Verify(pk, data, newVrf, proof)
	if err != nil {
		return fmt.Errorf("verifyVrf failed: %s", err)
	}
	if !result {
		return fmt.Errorf("verifyVrf failed")
	}
	return nil
}
func GetVbftConfigInfo(memdb *overlaydb.MemDB, isRootShard bool) (*config.VBFTConfig, error) {
	//get governance view,change view
	changeview, err := GetChangeView(memdb, isRootShard)
	if err != nil {
		return nil, err
	}
	//get preConfig
	preCfg := new(nutils.PreConfig)

	contractAddress := nutils.GovernanceContractAddress
	key := gov.PRE_CONFIG
	vbft_key := gov.VBFT_CONFIG
	if !isRootShard {
		contractAddress = nutils.ShardMgmtContractAddress
		key = shardmgt.PRE_CONFIG
		vbft_key = shardmgt.VBFT_CONFIG
	}
	data, err := GetStorageValue(memdb, ledger.DefLedger, contractAddress, []byte(key))
	if err != nil && err != scommon.ErrNotFound {
		return nil, err
	}
	if data != nil {
		err = preCfg.Deserialize(bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
	}

	chainconfig := new(config.VBFTConfig)
	if preCfg.SetView == changeview.View {
		chainconfig = &config.VBFTConfig{
			N:                    uint32(preCfg.Configuration.N),
			C:                    uint32(preCfg.Configuration.C),
			K:                    uint32(preCfg.Configuration.K),
			L:                    uint32(preCfg.Configuration.L),
			BlockMsgDelay:        uint32(preCfg.Configuration.BlockMsgDelay),
			HashMsgDelay:         uint32(preCfg.Configuration.HashMsgDelay),
			PeerHandshakeTimeout: uint32(preCfg.Configuration.PeerHandshakeTimeout),
			MaxBlockChangeView:   uint32(preCfg.Configuration.MaxBlockChangeView),
		}
	} else {
		data, err := GetStorageValue(memdb, ledger.DefLedger, contractAddress, []byte(vbft_key))
		if err != nil {
			return nil, err
		}
		cfg := new(nutils.Configuration)
		err = cfg.Deserialize(bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
		chainconfig = &config.VBFTConfig{
			N:                    uint32(cfg.N),
			C:                    uint32(cfg.C),
			K:                    uint32(cfg.K),
			L:                    uint32(cfg.L),
			BlockMsgDelay:        uint32(cfg.BlockMsgDelay),
			HashMsgDelay:         uint32(cfg.HashMsgDelay),
			PeerHandshakeTimeout: uint32(cfg.PeerHandshakeTimeout),
			MaxBlockChangeView:   uint32(cfg.MaxBlockChangeView),
		}
	}
	return chainconfig, nil
}

func GetPeersConfig(memdb *overlaydb.MemDB, isRootShard bool) ([]*config.VBFTPeerStakeInfo, error) {
	changeview, err := GetChangeView(memdb, isRootShard)
	if err != nil {
		return nil, err
	}
	viewBytes := nutils.GetUint32Bytes(changeview.View)
	key := append([]byte(gov.PEER_POOL), viewBytes...)
	contractAddress := nutils.GovernanceContractAddress
	if !isRootShard {
		contractAddress = nutils.ShardMgmtContractAddress
		key = append([]byte(shardmgt.PEER_POOL), viewBytes...)
	}
	data, err := GetStorageValue(memdb, ledger.DefLedger, contractAddress, key)
	if err != nil {
		return nil, err
	}
	peerMap := &nutils.PeerPoolMap{
		PeerPoolMap: make(map[string]*nutils.PeerPoolItem),
	}
	err = peerMap.Deserialize(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	var peerstakes []*config.VBFTPeerStakeInfo
	for _, id := range peerMap.PeerPoolMap {
		if id.Status == gov.CandidateStatus || id.Status == gov.ConsensusStatus {
			peerStakeInfo := &config.VBFTPeerStakeInfo{
				Index:      uint32(id.Index),
				PeerPubkey: id.PeerPubkey,
				InitPos:    id.InitPos + id.TotalPos,
			}
			peerstakes = append(peerstakes, peerStakeInfo)
		}
	}
	return peerstakes, nil
}

func isUpdate(memdb *overlaydb.MemDB, view uint32, isRootShard bool) (bool, error) {
	changeview, err := GetChangeView(memdb, isRootShard)
	if err != nil {
		return false, err
	}
	if changeview.View > view {
		return true, nil
	}
	return false, nil
}

func getRawStorageItemFromMemDb(memdb *overlaydb.MemDB, addr common.Address, key []byte) (value []byte, unkown bool) {
	rawKey := make([]byte, 0, 1+common.ADDR_LEN+len(key))
	rawKey = append(rawKey, byte(scommon.ST_STORAGE))
	rawKey = append(rawKey, addr[:]...)
	rawKey = append(rawKey, key...)
	return memdb.Get(rawKey)
}

func GetStorageValue(memdb *overlaydb.MemDB, backend *ledger.Ledger, addr common.Address, key []byte) (value []byte, err error) {
	if memdb == nil {
		return backend.GetStorageItem(addr, key)
	}
	rawValue, unknown := getRawStorageItemFromMemDb(memdb, addr, key)
	if unknown {
		return backend.GetStorageItem(addr, key)
	}
	if len(rawValue) == 0 {
		return nil, scommon.ErrNotFound
	}

	value, err = states.GetValueFromRawStorageItem(rawValue)
	return
}

func GetChangeView(memdb *overlaydb.MemDB, isRootShard bool) (*nutils.ChangeView, error) {
	contractAddress := nutils.GovernanceContractAddress
	key := gov.GOVERNANCE_VIEW
	if !isRootShard {
		contractAddress = nutils.ShardMgmtContractAddress
		key = shardmgt.SHARD_VIEW
	}
	value, err := GetStorageValue(memdb, ledger.DefLedger, contractAddress, []byte(key))
	if err != nil {
		return nil, err
	}
	changeView := new(nutils.ChangeView)
	err = changeView.Deserialize(bytes.NewBuffer(value))
	if err != nil {
		return nil, err
	}
	return changeView, nil
}

func getChainConfig(memdb *overlaydb.MemDB, blkNum uint32, isRootShard bool) (*vconfig.ChainConfig, error) {
	config, err := GetVbftConfigInfo(memdb, isRootShard)
	if err != nil {
		return nil, fmt.Errorf("failed to get chainconfig from leveldb: %s", err)
	}

	peersinfo, err := GetPeersConfig(memdb, isRootShard)
	if err != nil {
		return nil, fmt.Errorf("failed to get peersinfo from leveldb: %s", err)
	}
	changeview, err := GetChangeView(memdb, isRootShard)
	if err != nil {
		return nil, fmt.Errorf("failed to get governanceview failed:%s", err)
	}

	cfg, err := vconfig.GenesisChainConfig(config, peersinfo, changeview.TxHash, blkNum)
	if err != nil {
		return nil, fmt.Errorf("GenesisChainConfig failed: %s", err)
	}
	cfg.View = changeview.View
	return cfg, err
}

func getShardGasBalance(memdb *overlaydb.MemDB) (uint64, error) {
	value, err := GetStorageValue(memdb, ledger.DefLedger, nutils.OngContractAddress, nutils.ShardGasMgmtContractAddress[:])
	if err != nil {
		return 0, err
	}
	balance, err := serialization.ReadUint64(bytes.NewBuffer(value))
	if err != nil {
		return 0, err
	}
	return balance, nil
}
